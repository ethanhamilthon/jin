package desktop

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"jin/internal/daemon"
	"jin/internal/datadir"
	"jin/internal/web"
)

type link struct {
	base   *url.URL
	cookie string
	client *http.Client
}

func connect(ctx context.Context, version, project string) (*link, error) {
	root, err := datadir.Current()
	if err != nil {
		return nil, err
	}
	daemonClient, err := daemon.Ensure(ctx, root, version)
	if err != nil {
		return nil, err
	}
	raw, err := daemonClient.Web(ctx, project, web.Options{NoOpen: true})
	if err != nil {
		return nil, err
	}
	base, err := parseUpstream(raw)
	if err != nil {
		return nil, err
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, base.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	for _, cookie := range jar.Cookies(base) {
		if cookie.Name == "jin_device" {
			return &link{base: &url.URL{Scheme: base.Scheme, Host: base.Host}, cookie: cookie.Value, client: &http.Client{CheckRedirect: client.CheckRedirect}}, nil
		}
	}
	return nil, errors.New("daemon did not issue a desktop device cookie")
}
