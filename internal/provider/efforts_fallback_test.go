package provider

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestEffortsFallback(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		statusCode int
		body       string
		want       []string
	}{
		{
			name: "anthropic kind returns DefaultEfforts without probe",
			kind: KindAnthropic,
			want: DefaultEfforts,
		},
		{
			name:       "openai probe fails with 500 error",
			kind:       KindOpenAI,
			statusCode: http.StatusInternalServerError,
			body:       `{"error":"internal error"}`,
			want:       DefaultEfforts,
		},
		{
			name:       "openai probe unparseable rejection",
			kind:       KindOpenAI,
			statusCode: http.StatusBadRequest,
			body:       `{"error":{"message":"unparseable bad request"}}`,
			want:       DefaultEfforts,
		},
		{
			name:       "openai probe accepted invalid effort returns nil",
			kind:       KindOpenAI,
			statusCode: http.StatusOK,
			body:       `{"id":"chatcmpl-123"}`,
			want:       nil,
		},
		{
			name:       "openai probe parseable rejection returns parsed levels",
			kind:       KindOpenAI,
			statusCode: http.StatusBadRequest,
			body:       `{"error":{"message":"reasoning_effort: valid levels are 'low', 'medium', 'high'"}}`,
			want:       []string{"low", "medium", "high"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.kind == KindAnthropic {
					t.Fatal("anthropic kind should not probe server")
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(Config{Kind: tt.kind, BaseURL: server.URL, APIKey: "key"})
			got, err := client.Efforts(t.Context(), "model")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
