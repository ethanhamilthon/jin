//go:build cgo

package voice

import (
	"encoding/binary"
	"sync"

	"github.com/gen2brain/malgo"
)

// Recorder captures the default microphone into memory while it runs.
type Recorder struct {
	mu  sync.Mutex
	ctx *malgo.AllocatedContext
	dev *malgo.Device
	pcm []byte

	sumSquares float64
	count      int
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
	for i := 0; i+1 < len(in); i += 2 {
		v := float64(int16(binary.LittleEndian.Uint16(in[i:])))
		r.sumSquares += v * v
		r.count++
	}
	r.mu.Unlock()
}

// Level is the loudness, 0 to 1, of the audio since the last call.
func (r *Recorder) Level() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	level := loudness(r.sumSquares, r.count)
	r.sumSquares, r.count = 0, 0
	return level
}

func (r *Recorder) Start() error { return r.dev.Start() }

func (r *Recorder) Stop() error { return r.dev.Stop() }

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
