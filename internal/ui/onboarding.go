package ui

import (
	"github.com/gdamore/tcell/v3"

	"jin/internal/provider"
)

// bigLogo is the JIN of the first-run screen, drawn with the logo shimmer.
var bigLogo = []string{
	"  ██████  ██  ███    ██",
	"      ██  ██  ████   ██",
	"      ██  ██  ██ ██  ██",
	"██    ██  ██  ██  ██ ██",
	" ██████   ██  ██   ████",
}

const (
	modeAPI          = "api"
	modeSubscription = "subscription"
)

var onboardingModes = []option{
	{label: "API", detail: "your own key and base URL", value: modeAPI},
	{label: "Subscription", detail: "Claude, Codex or Antigravity through CLIProxyAPI", value: modeSubscription},
}

var onboardingKinds = []option{
	{label: "OpenAI Responses", detail: "/responses · OpenAI, keeps reasoning between turns", value: provider.KindResponses},
	{label: "OpenAI Chat Completions", detail: "/chat/completions · most OpenAI-compatible APIs", value: provider.KindOpenAI},
	{label: "Anthropic", detail: "/v1/messages · Claude and Anthropic-compatible APIs", value: provider.KindAnthropic},
}

var onboardingProfiles = []option{
	{label: "Claude", detail: "Anthropic subscription", value: "claude"},
	{label: "Codex", detail: "OpenAI subscription", value: "codex"},
	{label: "Antigravity", detail: "Google subscription", value: "antigravity"},
}

// onboardingChoices are the rows of the current first-run step.
func (a *app) onboardingChoices() []option {
	switch a.onboardMode {
	case modeAPI:
		return onboardingKinds
	case modeSubscription:
		return onboardingProfiles
	}
	return onboardingModes
}

// chooseOnboarding takes the choice of the current step.
func (a *app) chooseOnboarding(i int) {
	a.kindIndex = i
	value := a.onboardingChoices()[i].value
	switch a.onboardMode {
	case modeAPI:
		a.addProviderOfKind(value)
	case modeSubscription:
		if err := a.addProfile(value); err != nil {
			a.report(err)
		}
	default:
		if value == modeSubscription {
			a.ensureProxy(func() { a.onboardMode, a.kindIndex = modeSubscription, 0 })
			return
		}
		a.onboardMode, a.kindIndex = value, 0
	}
}

// onboarding reports whether jin still lacks a connected provider or a
// model. Until a model is chosen the first-run screen replaces the chat, so
// a subscription that is not signed in yet cannot open it either.
func (a *app) onboarding() bool {
	return a.cfg.Model == "" || (len(a.cfg.Providers) == 0 && !a.cfg.Provider.Ready())
}

// continueOnboarding opens the next step of a provider that exists but has
// no model yet: the sign-in of a subscription, or the model list.
func (a *app) continueOnboarding() {
	if a.cfg.Provider.Managed {
		a.openSubscription(a.cfg.ActiveProvider)
		return
	}
	a.openModelFlow()
}

// startOnboarding resumes the setup when a provider exists but no model was
// chosen yet.
func (a *app) startOnboarding() {
	if a.cfg.Provider.Ready() && a.cfg.Model == "" && a.sel == nil {
		a.continueOnboarding()
	}
}

// onboardingKey moves through the choices of a step; Esc goes back to the
// first one. s switches to another data folder, such as one /reset put
// aside. Ctrl+C quits: there is nothing to interrupt yet.
func (a *app) onboardingKey(ev *tcell.EventKey) {
	n := len(a.onboardingChoices())
	switch {
	case isCtrl(ev, 'c', false):
		a.quit = true
	case ev.Key() == tcell.KeyEscape && a.onboardMode != "":
		a.onboardMode, a.kindIndex = "", 0
	case ev.Key() == tcell.KeyRune && ev.Str() == "i":
		a.openProxyFlow()
	case ev.Key() == tcell.KeyRune && ev.Str() == "s":
		a.openSwapFlow()
	case a.cfg.Provider.Ready() && ev.Key() == tcell.KeyEnter:
		a.continueOnboarding()
	case ev.Key() == tcell.KeyUp:
		a.kindIndex = (a.kindIndex + n - 1) % n
	case ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyTab:
		a.kindIndex = (a.kindIndex + 1) % n
	case ev.Key() == tcell.KeyRune && ev.Str() >= "1" && int(ev.Str()[0]-'1') < n:
		a.chooseOnboarding(int(ev.Str()[0] - '1'))
	case ev.Key() == tcell.KeyEnter:
		a.chooseOnboarding(a.kindIndex)
	}
}
