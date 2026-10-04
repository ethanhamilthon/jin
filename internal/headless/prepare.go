package headless

import (
	"context"
	"errors"
	"fmt"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/store"
)

// runState is one `jin -p` request, ready to send.
type runState struct {
	db         *store.DB
	opt        Options
	out        writer
	id, dir    string
	save       bool
	record     store.Session
	history    []provider.Message
	provider   string
	endpoint   string
	budget     *budget
	saveErr    error
	changeTurn int
	request    core.Request
	agent      *core.Agent
	prices     <-chan pricing.Table
	table      pricing.Table
	close      func()
}

// prepare reads the settings, opens or creates the session and builds the
// agent. Nothing is sent to the model yet.
func prepare(ctx context.Context, db *store.DB, dir string, opt Options, prompt string, getenv func(string) string, out writer) (*runState, error) {
	cfg, err := db.LoadConfig()
	if err != nil {
		return nil, err
	}
	record, history, err := openSession(db, opt, dir)
	if err != nil {
		return nil, err
	}
	providerID, err := pinProvider(&cfg, opt.Provider, record, getenv)
	if err != nil {
		return nil, err
	}
	env := applyEnv(&cfg, getenv)
	if !cfg.Provider.Ready() {
		return nil, errors.New("provider is not configured: set JIN_BASE_URL and JIN_API_KEY, or configure it in the TUI")
	}
	names, err := toolNames(cfg.ToolsDisabled, opt)
	if err != nil {
		return nil, err
	}
	prices := make(chan pricing.Table, 1)
	load := loadPricing
	go func() {
		priceCtx, cancel := context.WithTimeout(ctx, pricingWait)
		defer cancel()
		prices <- load(priceCtx)
	}()
	model := resolveModel(opt.Model, env, record, cfg)
	if model == "" {
		return nil, errors.New("no model selected: pass --model or set JIN_MODEL (run `jin models` to list them)")
	}
	r := &runState{budget: newBudget(opt, record.Usage), endpoint: endpointOf(cfg.Provider.BaseURL), db: db, dir: dir, opt: opt, out: out, save: !opt.NoSession, record: record, provider: providerID, history: history, prices: prices, close: func() {}}
	r.request = core.Request{Prompt: prompt, Model: model, Effort: resolveEffort(opt.Effort, env, record, cfg, model)}
	if err := r.persist(dir, prompt); err != nil {
		return nil, err
	}
	var bodies map[string]string
	r.agent, bodies = buildAgent(ctx, db, dir, r.id, prompt, names, cfg, r.save, out)
	r.request.Prompt = prompts.Expand(prompt, bodies)
	r.agent.SetContextSize(record.Usage.Context)
	return r, nil
}

// claim makes this process the owner of the session before anything is
// written to it; a session another live jin process uses is refused.
func (r *runState) claim() error {
	err := r.db.SetRunning(r.id, true)
	var busy store.ErrSessionBusy
	if errors.As(err, &busy) {
		return fmt.Errorf("session %s is in use by process %d", r.id, busy.PID)
	}
	return nil
}

// persist creates or touches the session and closes tool calls an earlier
// run left without an answer.
func (r *runState) persist(dir, prompt string) error {
	r.id = r.record.ID
	if !r.save {
		r.history = append(r.history, core.InterruptedToolMessages(r.history)...)
		return nil
	}
	if r.id == "" {
		r.id = newID()
	}
	if err := r.claim(); err != nil {
		return err
	}
	title := r.record.Title
	if title == "" {
		title = truncate(firstLine(prompt), maxTitle)
	}
	if err := r.db.TouchProvider(r.id, dir, r.request.Model, r.request.Effort, title, r.provider); err != nil {
		return err
	}
	for _, msg := range core.InterruptedToolMessages(r.history) {
		if err := r.db.AppendMessage(r.id, msg); err != nil {
			return err
		}
		r.history = append(r.history, msg)
	}
	r.close = func() { _ = r.db.SetRunning(r.id, false) }
	return nil
}
