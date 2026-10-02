package provider

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestModelsEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		kind     string
		wantPath string
		checkHdr func(t *testing.T, h http.Header)
		body     string
		want     []string
	}{
		{
			name:     "anthropic models parsed and deduped",
			kind:     KindAnthropic,
			wantPath: "/v1/models",
			checkHdr: func(t *testing.T, h http.Header) {
				if h.Get("x-api-key") != "key" || h.Get("anthropic-version") != "2023-06-01" {
					t.Fatalf("unexpected headers: %v", h)
				}
			},
			body: `{"data":[{"id":"claude-3-haiku"},{"id":"claude-3-5-sonnet"},{"id":"claude-3-haiku"},{"id":""}]}`,
			want: []string{"claude-3-5-sonnet", "claude-3-haiku"},
		},
		{
			name:     "openai models path and auth header",
			kind:     KindOpenAI,
			wantPath: "/models",
			checkHdr: func(t *testing.T, h http.Header) {
				if h.Get("Authorization") != "Bearer key" {
					t.Fatalf("unexpected auth: %v", h.Get("Authorization"))
				}
			},
			body: `{"data":[{"id":"gpt-4o"},{"id":"gpt-3.5-turbo"}]}`,
			want: []string{"gpt-3.5-turbo", "gpt-4o"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Errorf("got path %s, want %s", r.URL.Path, tt.wantPath)
				}
				tt.checkHdr(t, r.Header)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(Config{Kind: tt.kind, BaseURL: server.URL, APIKey: "key"})
			got, err := client.Models(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got models %v, want %v", got, tt.want)
			}
		})
	}
}
