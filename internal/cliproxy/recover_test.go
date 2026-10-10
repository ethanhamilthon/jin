package cliproxy

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestProxyCrashRecoversForTheNextRequest(t *testing.T) {
	root := t.TempDir()
	installFixture(t, root, DefaultVersion)
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	client := NewClient(root)
	defer client.Close()
	endpoint, key, release, err := client.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	request, _ := http.NewRequestWithContext(ctx, "POST", endpoint+"/test-exit", nil)
	request.Header.Set("Authorization", "Bearer "+key)
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	release()
	time.Sleep(200 * time.Millisecond)
	next, _, end, err := client.Acquire(ctx)
	if err != nil {
		t.Fatal(err)
	}
	end()
	if next == endpoint {
		t.Fatal("dead proxy endpoint reused")
	}
	client.Close()
	waitSocketClosed(t, root)
}
