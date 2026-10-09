package headless

import (
	"jin/internal/core"
	"jin/internal/store"
)

// outcome is what the update loop saw by the time the request ended.
type outcome struct {
	answer, runErr               string
	partial, budget              string
	final, interrupted, timedOut bool
	// afterNote joins the next final answer to the one before it: the
	// background tasks note came between them.
	afterNote bool
}

// apply handles one update; it reports whether the request has ended.
func (r *runState) apply(u core.Update, o *outcome, usage *store.Usage) bool {
	switch u.Kind {
	case core.UpdateHistory:
		if r.save {
			r.saved(r.db.AppendMessage(r.id, u.Message))
		}
		r.out.Message(u.Message)
		if u.Message.Role == "assistant" && u.Message.Content != "" {
			o.partial = u.Message.Content
		}
		if u.Message.Role == "user" && core.IsTasksNote(u.Message.Content) {
			o.afterNote = true
		}
		if u.Message.Role == "assistant" && len(u.Message.ToolCalls) == 0 {
			o.answer = joinAnswer(o, u.Message.Content)
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
			r.saved(r.db.SaveUsage(r.id, *usage))
		}
	case core.UpdateToolResult:
		r.recordChanges(u.Changes)
	case core.UpdateToolCall:
		r.out.Progress(u.Tool + ": " + u.Text)
	case core.UpdateInfo:
		r.out.Progress(u.Text)
	case core.UpdateError:
		if r.budget.halted() {
			break
		}
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
	o.budget = r.budget.result(res.Usage)
	switch {
	case o.budget != "" && !o.timedOut && !o.interrupted:
		res.Text, res.Err, code = o.partial, "budget reached: "+o.budget, exitBudget
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
	if r.saveErr != nil {
		res.SaveErr = r.saveErr.Error()
		r.out.Progress("jin: could not save the session: " + res.SaveErr)
		code = max(code, exitError)
	}
	r.out.Result(res)
	return code
}
