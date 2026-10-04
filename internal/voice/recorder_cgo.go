//go:build cgo

package voice

import (
	"sync"

	"github.com/gen2brain/malgo"
)

// Recorder captures the default microphone into memory while it runs.
type Recorder struct {
	mu  sync.Mutex
	ctx *malgo.AllocatedContext
	dev *malgo.Device
	pcm []byte
}

func NewRecorder() (*Recorder, error) {
	ctx, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		return nil, err
	}
	r := &Recorder{ctx: ctx}
	cfg := malgo.DefaultDeviceConfig(malgo.Capture)
	cfg.Capture.Format = malgo.FormatS16
	cfg.Capture.Channels = 1
	cfg.SampleRate = SampleRate
	r.dev, err = malgo.InitDevice(ctx.Context, cfg, malgo.DeviceCallbacks{Data: r.onData})
	if err != nil {
		r.Close()
		return nil, err
	}
	return r, nil
}

func (r *Recorder) onData(_, in []byte, _ uint32) {
	r.mu.Lock()
	r.pcm = append(r.pcm, in...)
	r.mu.Unlock()
}

func (r *Recorder) Start() error { return r.dev.Start() }

func (r *Recorder) Stop() error { return r.dev.Stop() }

// Snapshot copies the audio captured since the last Take.
func (r *Recorder) Snapshot() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]byte(nil), r.pcm...)
}

// Take returns the captured audio and starts the next piece empty.
func (r *Recorder) Take() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	pcm := r.pcm
	r.pcm = nil
	return pcm
}

func (r *Recorder) Close() {
	if r == nil {
		return
	}
	if r.dev != nil {
		r.dev.Uninit()
	}
	if r.ctx != nil {
		_ = r.ctx.Uninit()
		r.ctx.Free()
	}
}
