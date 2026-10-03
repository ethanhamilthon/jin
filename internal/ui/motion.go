package ui

import "jin/internal/store"

// motionSteps is how far the decorative animations (input glow, logo
// shimmer) move per tick, in tenths of a frame. Spinners are not decoration:
// they show that work goes on, so they always turn.
var motionSteps = map[string]int{
	store.MotionOff:    0,
	store.MotionSlow:   5,
	store.MotionNormal: 10,
	store.MotionFast:   20,
}

// motionOrder is the order /motion lists the speeds in.
var motionOrder = []string{store.MotionOff, store.MotionSlow, store.MotionNormal, store.MotionFast}

// tick advances the spinners and, unless motion is off or the terminal is
// not focused, the glow.
func (a *app) tick() {
	a.frame++
	if a.blurred {
		return
	}
	step, ok := motionSteps[a.cfg.Motion]
	if !ok {
		step = motionSteps[store.MotionNormal]
	}
	a.glowTenths += step
}

// glowFrame is the frame of the decorative animations.
func (a *app) glowFrame() int { return a.glowTenths / 10 }

// moving reports whether the glow animates at all.
func (a *app) moving() bool { return a.cfg.Motion != store.MotionOff }

// openMotionFlow sets the speed of the glow and the shimmer, or turns them
// off. The change shows at once.
func (a *app) openMotionFlow() {
	options := make([]option, len(motionOrder))
	labels := map[string]string{
		store.MotionOff: "Off · no glow, no shimmer", store.MotionSlow: "Slow",
		store.MotionNormal: "Normal", store.MotionFast: "Fast",
	}
	for i, m := range motionOrder {
		options[i] = option{label: labels[m], value: m}
	}
	current := a.cfg.Motion
	if current == "" {
		current = store.MotionNormal
	}
	a.openList("Motion · input glow and logo shimmer", options, current, func(m string) error {
		if err := a.store.SaveMotion(m); err != nil {
			return err
		}
		a.cfg.Motion = m
		return nil
	})
}
