package headless

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"jin/internal/pricing"
	"jin/internal/store"
)

const (
	// maxDepth stops jin from starting jin without end: a run started from
	// the bash tool of another headless run carries the depth in JIN_DEPTH.
	maxDepth = 3
	maxTitle = 60

	exitOK          = 0
	exitError       = 1
	exitBudget      = 3
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

// Run executes `jin models`, `jin refresh-models`, `jin projects` or `jin -p`.
func Run(ctx context.Context, args []string, db *store.DB, dir string, signals <-chan os.Signal, io_ ioSet) int {
	if len(args) > 0 && (args[0] == "models" || args[0] == "refresh-models") {
		return runModels(ctx, args[0], args[1:], db, io_)
	}
	if len(args) > 0 && args[0] == "projects" {
		return runProjects(args[1:], db, dir, io_)
	}
	if len(args) > 1 && args[0] == "sessions" {
		ctx, stop := withLimits(ctx, 0, signals)
		defer stop()
		return runSessionAction(ctx, args[1:], db, io_)
	}
	opt, err := ParseArgs(args)
	if err != nil {
		fmt.Fprintln(io_.err, "jin:", err)
		return exitError
	}
	out := newWriter(opt.Format, io_.out, io_.err)
	ctx, stop := withLimits(ctx, opt.Timeout, signals)
	defer stop()
	fail := func(err error) int {
		if ctx.Err() != nil {
			err = endedBy(ctx, opt.Timeout)
		}
		out.Result(result{Err: err.Error()})
		return exitCode(ctx)
	}
	if opt.Cwd != "" {
		if dir, err = enterDir(opt.Cwd); err != nil {
			return fail(err)
		}
	}
	depth, _ := strconv.Atoi(io_.getenv("JIN_DEPTH"))
	if depth >= maxDepth {
		return fail(errors.New("depth limit"))
	}
	os.Setenv("JIN_DEPTH", strconv.Itoa(depth+1))
	prompt, err := BuildPrompt(ctx, opt.Prompt, io_.in, io_.piped)
	if err != nil {
		return fail(err)
	}
	r, err := prepare(ctx, db, dir, opt, prompt, io_.getenv, out)
	if err != nil {
		return fail(err)
	}
	defer r.close()
	return r.run(ctx)
}
