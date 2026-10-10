package cliproxy

import (
	"context"
	"encoding/json"
	"errors"
)

type Account struct {
	Name     string `json:"name"`
	Profile  string `json:"profile"`
	Disabled bool   `json:"disabled"`
	Status   string `json:"status"`
}

func profileName(value string) (string, error) {
	switch value {
	case "claude", "codex", "antigravity":
		return value, nil
	}
	return "", errors.New("unsupported subscription provider")
}

func (b *broker) accounts(ctx context.Context) ([]Account, error) {
	raw, err := b.proxy.management(ctx, "GET", "/auth-files", nil)
	if err != nil {
		return nil, err
	}
	var body struct {
		Files []struct {
			Name     string
			Provider string
			Type     string
			Disabled bool
			Status   string
		}
	}
	if err = json.Unmarshal(raw, &body); err != nil {
		return nil, errors.New("invalid proxy accounts response")
	}
	result := []Account{}
	for _, file := range body.Files {
		profile := file.Provider
		if profile == "" {
			profile = file.Type
		}
		if profile == "anthropic" {
			profile = "claude"
		}
		if _, err := profileName(profile); err != nil {
			continue
		}
		result = append(result, Account{Name: file.Name, Profile: profile, Disabled: file.Disabled, Status: file.Status})
	}
	return result, nil
}

func (b *broker) syncProfile(ctx context.Context, profile string, disabled bool) error {
	if _, err := profileName(profile); err != nil {
		return err
	}
	resume, err := b.pause(ctx)
	if err != nil {
		return err
	}
	defer resume()
	accounts, err := b.accounts(ctx)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.Profile != profile {
			continue
		}
		fields := map[string]any{"name": account.Name, "prefix": profile, "request_retry": 0}
		if profile == "claude" {
			fields["disable_claude_cloak_mode"] = true
		}
		if _, err = b.proxy.management(ctx, "PATCH", "/auth-files/fields", fields); err != nil {
			return err
		}
		if _, err = b.proxy.management(ctx, "PATCH", "/auth-files/status", map[string]any{"name": account.Name, "disabled": disabled}); err != nil {
			return err
		}
	}
	return nil
}
