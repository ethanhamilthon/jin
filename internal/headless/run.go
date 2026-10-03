package headless

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/startup"
	"jin/internal/store"
	"jin/internal/tools"
)

const (
	// maxDepth stops jin from starting jin without end: a run started from
	// the bash tool of another headless run carries the depth in JIN_DEPTH.
	maxDepth = 3
	maxTitle = 60

	exitOK          = 0
	exitError       = 1
	exitInterrupted = 130

	pricingWait = 5 * time.Second
)

// loadPricing is replaced in tests so they never touch the network.
var loadPricing = pricing.Load

type ioSet struct {
	in     io.Reader
	piped  bool
	out    io.Writer
	err    io.Writer
	getenv func(string) string
}

// Main runs a headless command and returns the exit code. The caller has
// opened the database and recovered interrupted requests.
func Main(args []string, db *store.DB, dir string) int {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	piped := false
	if info, err := os.Stdin.Stat(); err == nil {
		piped = info.Mode()&os.ModeCharDevice == 0
	}
	return Run(context.Background(), args, db, dir, signals, ioSet{os.Stdin, piped, os.Stdout, os.Stderr, os.Getenv})
}

// Run executes `jin models`, `jin refresh-models` or `jin -p`.
func Run(ctx context.Context, args []string, db *store.DB, dir string, signals <-chan os.Signal, io_ ioSet) int {
	if len(args) > 0 && (args[0] == "models" || args[0] == "refresh-models") {
		return runModels(ctx, args[0], args[1:], db, io_)
	}
	opt, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return exitError
	}
	out := newWriter(opt.Format, io_.out, io_.err)
	fail := func(err error) int {
		out.Result(result{Err: err.Error()})
		return exitError
	}
	depth, _ := strconv.Atoi(io_.getenv("JIN_DEPTH"))
	if depth >= maxDepth {
		return fail(errors.New("subagent depth limit"))
	}
	os.Setenv("JIN_DEPTH", strconv.Itoa(depth+1))
	prompt, err := BuildPrompt(opt.Prompt, io_.in, io_.piped)
	if err != nil {
		return fail(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil {
		return fail(err)
	}
	env := applyEnv(&cfg, io_.getenv)
	if !cfg.Provider.Ready() {
		return fail(errors.New("provider is not configured: set JIN_BASE_URL and JIN_API_KEY, or configure it in the TUI"))
	}
	names, err := toolNames(cfg.ToolsDisabled, opt)
	if err != nil {
		return fail(err)
	}
	prices := make(chan pricing.Table, 1)
	go func() {
		priceCtx, cancel := context.WithTimeout(ctx, pricingWait)
		defer cancel()
		prices <- loadPricing(priceCtx)
	}()

	record, history, err := openSession(db, opt, dir)
	if err != nil {
		return fail(err)
	}
	model := resolveModel(opt.Model, env, record, cfg)
	if model == "" {
		return fail(errors.New("no model selected: pass --model or set JIN_MODEL (run `jin models` to list them)"))
	}
	effort := resolveEffort(opt.Effort, env, record, cfg, model)
	save := !opt.NoSession
	id := record.ID
	if id == "" && save {
		id = newID()
	}
	if save {
		title := record.Title
		if title == "" {
			title = truncate(firstLine(prompt), maxTitle)
		}
		if err := db.Touch(id, dir, model, effort, title); err != nil {
			return fail(err)
		}
		for _, msg := range core.InterruptedToolMessages(history) {
			if err := db.AppendMessage(id, msg); err != nil {
				return fail(err)
			}
			history = append(history, msg)
		}
		_ = db.SetRunning(id, true)
		defer db.SetRunning(id, false)
	} else {
		history = append(history, core.InterruptedToolMessages(history)...)
	}
	var todos tools.TodoStore = &tools.MemoryTodos{}
	if save {
		todos = store.SessionTodos{DB: db, ID: id}
	}
	registry := tools.Build(names, todos)
	// Commands in the system prompt file and hooks run here, before the agent
	// starts. #prompts are not used in headless mode, so they are not read.
	rendered := startup.Render(ctx, startup.Input{
		Dir: dir, SessionID: id, ToolNames: names, HooksDisabled: cfg.HooksDisabled,
	}, nil)
	for _, warning := range rendered.Warnings {
		out.Progress("jin: " + warning)
	}
	client := provider.NewClient(cfg.Provider)
	client.SetStallTimeout(cfg.StallTimeout)
	agent := core.NewAgent(client, rendered.System, registry)
	agent.SetSidePrompts(rendered.Compact, rendered.Handoff)

	table := <-prices
	entry, known := table.Lookup(model)
	request := core.Request{
		Prompt: prompt, Model: model, Effort: effort, Window: entry.MaxInputTokens,
		NoVision: known && entry.VisionKnown && !entry.Vision,
	}
	out.Session(id, dir, model, effort)

	runCtx, stop := context.WithCancel(ctx)
	defer stop()
	requests := make(chan core.Request, 1)
	updates := make(chan core.Update, 64)
	go agent.Run(runCtx, core.SinceLastSummary(history), requests, updates)
	requests <- request
	started := time.Now()

	var timeout <-chan time.Time
	if opt.Timeout > 0 {
		timer := time.NewTimer(opt.Timeout)
		defer timer.Stop()
		timeout = timer.C
	}
	usage := record.Usage
	var (
		answer      string
		runErr      string
		final       bool
		interrupted bool
		timedOut    bool
		finished    bool
	)
	retry := time.NewTicker(50 * time.Millisecond)
	defer retry.Stop()
	for !finished {
		select {
		case u, ok := <-updates:
			if !ok {
				finished = true
				break
			}
			switch u.Kind {
			case core.UpdateHistory:
				if save {
					if err := db.AppendMessage(id, u.Message); err != nil {
						out.Progress("jin: history was not saved: " + err.Error())
					}
				}
				out.Message(u.Message)
				if u.Message.Role == "assistant" && len(u.Message.ToolCalls) == 0 {
					answer = u.Message.Content
				}
			case core.UpdateUsage, core.UpdateCompacted:
				usage.Add(u.Usage, u.Model, table)
				if u.Kind == core.UpdateCompacted && u.Usage.Known {
					usage.Context = u.Usage.Output
				}
				if save {
					_ = db.SaveUsage(id, usage)
				}
			case core.UpdateToolCall:
				out.Progress(u.Tool + ": " + u.Text)
			case core.UpdateInfo:
				out.Progress(u.Text)
			case core.UpdateError:
				runErr = u.Text
				out.Progress("error: " + u.Text)
			case core.UpdateDone:
				final = u.Final
				finished = true
			}
		case <-timeout:
			timedOut, timeout = true, nil
			agent.Interrupt()
		case <-signals:
			interrupted = true
			agent.Interrupt()
		case <-retry.C:
			// An interrupt that lands before the turn starts is lost, so
			// keep asking until the run ends.
			if timedOut || interrupted {
				agent.Interrupt()
			}
		}
	}
	close(requests)

	res := result{Text: answer, SessionID: id, Duration: time.Since(started).Milliseconds(), Usage: usage}
	code := exitOK
	switch {
	case interrupted:
		res.Err, code = "interrupted", exitInterrupted
	case timedOut:
		res.Err, code = "timed out after "+opt.Timeout.String(), exitError
	case runErr != "":
		res.Err, code = runErr, exitError
	case !final:
		res.Err, code = "the request did not finish", exitError
	}
	out.Result(res)
	return code
}

// openSession finds the session to continue, if any, and its messages.
func openSession(db *store.DB, opt Options, dir string) (store.Session, []provider.Message, error) {
	var record store.Session
	switch {
	case opt.Session != "":
		found, ok, err := db.GetSession(opt.Session)
		if err != nil {
			return record, nil, err
		}
		if !ok {
			return record, nil, fmt.Errorf("no session with id %q", opt.Session)
		}
		record = found
	case opt.Continue:
		list, err := db.ListByPath(dir)
		if err != nil {
			return record, nil, err
		}
		if len(list) == 0 {
			return record, nil, errors.New("no session to continue in this directory")
		}
		record = list[0]
	default:
		return record, nil, nil
	}
	messages, err := db.LoadMessages(record.ID)
	return record, messages, err
}

func newID() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}
