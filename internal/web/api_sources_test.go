package web

import (
	"net/http"
	"testing"
)

func TestProviderSwitchAndCombinedModels(t *testing.T) {
	server, _ := newTestServer(t)
	var result any
	for _, profile := range []string{"codex", "claude"} {
		if status := call(t, server, "POST", "/api/cliproxy/providers", `{"profile":"`+profile+`"}`, &result); status != 200 {
			t.Fatalf("add=%d", status)
		}
	}
	for _, id := range []string{"subscription-codex", "subscription-claude"} {
		if status := call(t, server, "POST", "/api/providers/"+id+"/enabled", `{"enabled":false}`, &result); status != 200 {
			t.Fatalf("switch=%d", status)
		}
	}
	var catalog struct {
		Models []any
		Errors []any
	}
	if status := call(t, server, "GET", "/api/models", "", &catalog); status != 200 || len(catalog.Models) != 0 || len(catalog.Errors) != 0 {
		t.Fatalf("catalog status=%d result=%+v", status, catalog)
	}
}

func TestManagedSignInIsLocalOnly(t *testing.T) {
	server, _ := newTestServer(t)
	request, err := http.NewRequest("POST", server.URL+"/api/cliproxy/login", nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Host = "phone.tailnet.ts.net"
	response, err := server.Client().Do(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		t.Fatal("remote sign-in accepted")
	}
}
