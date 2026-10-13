package provider

import (
	"sync/atomic"
	"testing"
)

func TestProtocolStoreAndFallback(t *testing.T) {
	client := NewClient(Config{Kind: KindAuto, Protocol: KindOpenAI})
	if got := client.ActualProtocol("m"); got != KindOpenAI {
		t.Fatalf("fallback=%q", got)
	}
	var loads, saves atomic.Int32
	client.SetProtocolStore(func(model string) (string, error) {
		loads.Add(1)
		if model != "m" {
			t.Fatalf("model=%q", model)
		}
		if saves.Load() > 0 {
			return KindOpenAI, nil
		}
		return KindResponses, nil
	}, func(model, protocol string) error {
		saves.Add(1)
		return nil
	})
	if got := client.ActualProtocol("m"); got != KindResponses || loads.Load() != 1 {
		t.Fatalf("protocol=%q loads=%d", got, loads.Load())
	}
	client.rememberProtocol(client.Config(), "m", KindOpenAI)
	if got := client.ActualProtocol("m"); got != KindOpenAI || saves.Load() != 1 {
		t.Fatalf("protocol=%q saves=%d", got, saves.Load())
	}
}

func TestProtocolSaveFailureDoesNotBlockRequestState(t *testing.T) {
	client := NewClient(Config{Kind: KindAuto})
	client.SetProtocolStore(nil, func(string, string) error { return errStore })
	client.rememberProtocol(client.Config(), "m", KindOpenAI)
	if got := client.ActualProtocol("m"); got != KindOpenAI {
		t.Fatalf("protocol=%q", got)
	}
}

var errStore = errorString("store failure")

type errorString string

func (e errorString) Error() string { return string(e) }
