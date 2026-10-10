package daemon

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"jin/internal/session"
)

func TestDuplicateCommandsReturnOriginalResult(t *testing.T) {
	manager := session.NewManager(context.Background(), nil, "v1", nil)
	handler := commandRoute("v1", manager)
	run := func(command Command) Result {
		body, _ := json.Marshal(command)
		response := httptest.NewRecorder()
		handler(response, httptest.NewRequest("POST", "/command", bytes.NewReader(body)))
		if response.Code != 200 {
			t.Fatalf("status: %d", response.Code)
		}
		var result Result
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	first := run(Command{ID: "once", Version: "v1", Action: "unknown"})
	repeated := run(Command{ID: "once", Version: "v1", Action: "stop", Session: "missing"})
	if first.Error != repeated.Error || repeated.Error != "unknown daemon command" {
		t.Fatalf("replayed result changed: %+v", repeated)
	}
}

func TestCommandVersionRequired(t *testing.T) {
	manager := session.NewManager(context.Background(), nil, "v1", nil)
	response := httptest.NewRecorder()
	commandRoute("v1", manager)(response, httptest.NewRequest("POST", "/command", bytes.NewBufferString(`{"id":"one","version":"v2","action":"live"}`)))
	if response.Code != http.StatusConflict {
		t.Fatalf("status: %d", response.Code)
	}
}
