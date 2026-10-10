package cliproxy

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
)

type Login struct {
	State  string `json:"state"`
	URL    string `json:"url"`
	Status string `json:"status"`
}

func (b *broker) login(ctx context.Context, profile string) (Login, error) {
	name, err := profileName(profile)
	if err != nil {
		return Login{}, err
	}
	if name == "claude" {
		name = "anthropic"
	}
	raw, err := b.proxy.management(ctx, "GET", "/"+name+"-auth-url?is_webui=true", nil)
	if err != nil {
		return Login{}, err
	}
	var login Login
	if err = json.Unmarshal(raw, &login); err != nil {
		return login, errors.New("invalid sign-in response")
	}
	u, err := url.Parse(login.URL)
	hosts := map[string]string{"claude": "claude.ai", "codex": "auth.openai.com", "antigravity": "accounts.google.com"}
	if err != nil || u.Scheme != "https" || u.Hostname() != hosts[profile] || u.User != nil || login.State == "" {
		return Login{}, errors.New("invalid sign-in URL")
	}
	if b.logins == nil {
		b.logins = map[string]string{}
	}
	b.logins[login.State] = profile
	return login, nil
}

func (b *broker) poll(ctx context.Context, profile, state string) (Login, error) {
	if _, err := profileName(profile); err != nil {
		return Login{}, err
	}
	if state == "" || b.logins[state] != profile {
		return Login{}, errors.New("sign-in state does not match this provider")
	}
	raw, err := b.proxy.management(ctx, "GET", "/get-auth-status?state="+url.QueryEscape(state), nil)
	if err != nil {
		return Login{}, err
	}
	var result struct {
		Status string
		Error  string
	}
	if err = json.Unmarshal(raw, &result); err != nil {
		return Login{}, errors.New("invalid sign-in status")
	}
	if result.Status == "error" {
		return Login{Status: "error"}, errors.New("subscription sign-in failed or expired")
	}
	if result.Status == "ok" {
		if err = b.syncProfile(ctx, profile, false); err != nil {
			return Login{}, err
		}
	}
	return Login{State: state, Status: result.Status}, nil
}

func (b *broker) cancelLogin(ctx context.Context, state string) (any, error) {
	if state == "" {
		return nil, errors.New("sign-in state is required")
	}
	result, err := b.proxy.management(ctx, "DELETE", "/oauth-session?state="+url.QueryEscape(state), nil)
	delete(b.logins, state)
	return result, err
}
