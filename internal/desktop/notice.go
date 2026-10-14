//go:build desktop

package desktop

import (
	"encoding/json"
	"net/http"
	"time"
)

const noticeWait = 4 * time.Second

// noticeBody is a problem the page shows in a banner. Kind is "path" when
// ~/.local/bin is not in the PATH of the login shell and "link" when the jin
// command could not be linked.
type noticeBody struct {
	Kind    string `json:"kind"`
	Dir     string `json:"dir,omitempty"`
	Line    string `json:"line,omitempty"`
	Message string `json:"message,omitempty"`
}

// notice is answered at /desktop/notice by the app itself, not by the daemon.
type notice struct {
	done chan struct{}
	body *noticeBody
}

func newNotice() *notice { return &notice{done: make(chan struct{})} }

// finish publishes the result of the background check; nil means no problem.
func (n *notice) finish(body *noticeBody) {
	n.body = body
	close(n.done)
}

func (n *notice) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !shellOriginAllowed(r.Header.Get("Origin")) {
		http.Error(w, "desktop request refused", http.StatusForbidden)
		return
	}
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	select {
	case <-n.done:
	case <-time.After(noticeWait):
	case <-r.Context().Done():
		return
	}
	select {
	case <-n.done:
	default:
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if n.body == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n.body)
}

// checkCLI links the jin command and looks at the PATH of the login shell.
func checkCLI(exe, home string) *noticeBody {
	linked, err := linkCLI(exe, home)
	if err != nil {
		return &noticeBody{Kind: "link", Message: "Could not link the jin command in ~/.local/bin: " + err.Error()}
	}
	if !linked {
		return nil
	}
	if in, known := inLoginPath(linkDir(home)); known && !in {
		return &noticeBody{Kind: "path", Dir: "~/.local/bin", Line: `export PATH="$HOME/.local/bin:$PATH"`}
	}
	return nil
}
