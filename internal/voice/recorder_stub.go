//go:build !cgo

package voice

import "errors"

var errNoCgo = errors.New("this jin build has no microphone support (needs cgo)")

// Recorder is a stub: the microphone needs a build with cgo.
type Recorder struct{}

func NewRecorder() (*Recorder, error) { return nil, errNoCgo }
func (r *Recorder) Start() error      { return errNoCgo }
func (r *Recorder) Stop() error       { return errNoCgo }
func (r *Recorder) Snapshot() []byte  { return nil }
func (r *Recorder) Take() []byte      { return nil }
func (r *Recorder) Close()            {}
