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

const modeAPI = "api"

var onboardingKinds = []option{
	{label: "OpenAI-compatible", detail: "Automatic protocol selection · OpenAI-compatible APIs", value: "auto"},
	{label: "Anthropic", detail: "/v1/messages · Claude and Anthropic-compatible APIs", value: provider.KindAnthropic},
}

func (a *app) onboardingChoices() []option { return onboardingKinds }

func (a *app) chooseOnboarding(i int) {
	a.kindIndex = i
	a.addProviderOfKind(a.onboardingChoices()[i].value)
}

// onboarding reports whether jin still lacks a connected provider or a model.
func (a *app) onboarding() bool {
	if a.cfg.Provider.Managed || a.removedProvider(a.cfg.ActiveProvider) {
		return true
	}
	return a.cfg.Model == "" || (len(a.cfg.Providers) == 0 && !a.cfg.Provider.Ready())
}

func (a *app) removedProvider(id string) bool {
	entry, ok := a.configuredProvider(id)
	return ok && entry.Source == "cliproxy"
}

func (a *app) continueOnboarding() { a.openModelFlow() }

func (a *app) startOnboarding() {
	if a.cfg.Provider.Ready() && a.cfg.Model == "" && a.sel == nil {
		a.continueOnboarding()
	}
}

// onboardingKey handles first-run setup keys.
func (a *app) onboardingKey(ev *tcell.EventKey) {
	n := len(a.onboardingChoices())
	switch {
	case isCtrl(ev, 'c', false):
		a.quit = true
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
