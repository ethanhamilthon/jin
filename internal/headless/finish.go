package headless

import (
	"jin/internal/core"
	"jin/internal/store"
)

// outcome is what the update loop saw by the time the request ended.
type outcome struct {
	answer, runErr               string
	final, interrupted, timedOut bool
}

// apply handles one update; it reports whether the request has ended.
func (r *runState) apply(u core.Update, o *outcome, usage *store.Usage) bool {
	switch u.Kind {
	case core.UpdateHistory:
		if r.save {
			if err := r.db.AppendMessage(r.id, u.Message); err != nil {
				r.out.Progress("jin: history was not saved: " + err.Error())
			}
		}
		r.out.Message(u.Message)
		if u.Message.Role == "assistant" && len(u.Message.ToolCalls) == 0 {
			o.answer = u.Message.Content
		}
	case core.UpdateUsage, core.UpdateCompacted:
		usage.Add(u.Usage, u.Model, r.table)
		if u.Kind == core.UpdateCompacted && u.Usage.Known {
			usage.Context = u.Usage.Output
		}
		if u.Kind == core.UpdateCompacted && u.Text != "" {
			r.out.Progress(u.Text)
		}
		if r.save {
			_ = r.db.SaveUsage(r.id, *usage)
		}
	case core.UpdateToolCall:
		r.out.Progress(u.Tool + ": " + u.Text)
	case core.UpdateInfo:
		r.out.Progress(u.Text)
	case core.UpdateError:
		o.runErr = u.Text
		r.out.Progress("error: " + u.Text)
	case core.UpdateDone:
		o.final = u.Final
		return true
	}
	return false
}

func (r *runState) finish(o outcome, res result) int {
	code := exitOK
	switch {
	case o.interrupted:
		res.Err, code = "interrupted", exitInterrupted
	case o.timedOut:
		res.Err, code = "timed out after "+r.opt.Timeout.String(), exitError
	case o.final:
	case o.runErr != "":
		res.Err, code = o.runErr, exitError
	default:
		res.Err, code = "the request did not finish", exitError
	}
	r.out.Result(res)
	return code
}
