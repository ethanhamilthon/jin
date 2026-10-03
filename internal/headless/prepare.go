package headless

import (
	"context"
	"errors"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
)

// runState is one `jin -p` request, ready to send.
type runState struct {
	db      *store.DB
	opt     Options
	out     writer
	id, dir string
	save    bool
	record  store.Session
	history []provider.Message
	request core.Request
	agent   *core.Agent
	prices  <-chan pricing.Table
	table   pricing.Table
	close   func()
}

// prepare reads the settings, opens or creates the session and builds the
// agent. Nothing is sent to the model yet.
func prepare(ctx context.Context, db *store.DB, dir string, opt Options, prompt string, getenv func(string) string, out writer) (*runState, error) {
	cfg, err := db.LoadConfig()
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
	go func() {
		priceCtx, cancel := context.WithTimeout(ctx, pricingWait)
		defer cancel()
		prices <- loadPricing(priceCtx)
	}()
	record, history, err := openSession(db, opt, dir)
	if err != nil {
		return nil, err
	}
	model := resolveModel(opt.Model, env, record, cfg)
	if model == "" {
		return nil, errors.New("no model selected: pass --model or set JIN_MODEL (run `jin models` to list them)")
	}
	r := &runState{db: db, dir: dir, opt: opt, out: out, save: !opt.NoSession, record: record, history: history, prices: prices, close: func() {}}
	r.request = core.Request{Prompt: prompt, Model: model, Effort: resolveEffort(opt.Effort, env, record, cfg, model)}
	if err := r.persist(dir, prompt); err != nil {
		return nil, err
	}
	r.agent = buildAgent(ctx, db, dir, r.id, names, cfg, r.save, out)
	return r, nil
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
	title := r.record.Title
	if title == "" {
		title = truncate(firstLine(prompt), maxTitle)
	}
	if err := r.db.Touch(r.id, dir, r.request.Model, r.request.Effort, title); err != nil {
		return err
	}
	for _, msg := range core.InterruptedToolMessages(r.history) {
		if err := r.db.AppendMessage(r.id, msg); err != nil {
			return err
		}
		r.history = append(r.history, msg)
	}
	_ = r.db.SetRunning(r.id, true)
	r.close = func() { _ = r.db.SetRunning(r.id, false) }
	return nil
}
