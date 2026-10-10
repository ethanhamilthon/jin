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

// onboardingKinds are the provider kinds the first-run screen offers.
var onboardingKinds = []option{
	{label: "OpenAI Responses", detail: "/responses · OpenAI, keeps reasoning between turns", value: provider.KindResponses},
	{label: "OpenAI Chat Completions", detail: "/chat/completions · most OpenAI-compatible APIs", value: provider.KindOpenAI},
	{label: "Anthropic", detail: "/v1/messages · Claude and Anthropic-compatible APIs", value: provider.KindAnthropic},
}

// onboarding reports whether jin still lacks a provider or a model. Until
// both are set, the first-run screen replaces the chat.
func (a *app) onboarding() bool {
	return len(a.cfg.Providers) == 0 && (!a.cfg.Provider.Ready() || a.cfg.Model == "")
}

// startOnboarding opens the model list right away when a provider exists
// but no model was chosen yet.
func (a *app) startOnboarding() {
	if a.cfg.Provider.Ready() && !a.cfg.Provider.Managed && a.cfg.Model == "" && a.sel == nil {
		a.openModelFlow()
	}
}

// onboardingKey moves through the kinds and starts the setup of one; s
// switches to another data folder, such as one /reset put aside. Ctrl+C
// quits: there is nothing to interrupt yet.
func (a *app) onboardingKey(ev *tcell.EventKey) {
	n := len(onboardingKinds)
	switch {
	case isCtrl(ev, 'c', false):
		a.quit = true
	case ev.Key() == tcell.KeyRune && ev.Str() == "p":
		a.addSubscription()
	case ev.Key() == tcell.KeyRune && ev.Str() == "i":
		a.openProxyFlow()
	case ev.Key() == tcell.KeyRune && ev.Str() == "s":
		a.openSwapFlow()
	case a.cfg.Provider.Ready() && ev.Key() == tcell.KeyEnter:
		a.openModelFlow()
	case ev.Key() == tcell.KeyUp:
		a.kindIndex = (a.kindIndex + n - 1) % n
	case ev.Key() == tcell.KeyDown || ev.Key() == tcell.KeyTab:
		a.kindIndex = (a.kindIndex + 1) % n
	case ev.Key() == tcell.KeyRune && ev.Str() >= "1" && ev.Str() <= "3":
		a.kindIndex = int(ev.Str()[0] - '1')
		a.addProviderOfKind(onboardingKinds[a.kindIndex].value)
	case ev.Key() == tcell.KeyEnter:
		a.addProviderOfKind(onboardingKinds[a.kindIndex].value)
	}
}
