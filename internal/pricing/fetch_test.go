package pricing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchHTTPStatus(t *testing.T) {
	cases := []struct {
		status  int
		wantErr bool
	}{
		{http.StatusOK, false},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, true},
		{http.StatusTooManyRequests, true},
	}

	for _, c := range cases {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(c.status)
			if c.status == http.StatusOK {
				_, _ = w.Write([]byte(`{}`))
			}
		}))

		_, _, err := fetch(context.Background(), server.URL, func(b []byte) (Table, error) {
			return Table{}, nil
		})
		server.Close()

		if (err != nil) != c.wantErr {
			t.Errorf("status %d: err = %v, wantErr = %v", c.status, err, c.wantErr)
		}
	}
}
