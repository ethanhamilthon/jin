package cliproxy

import (
	"context"
	"net/url"
	"strings"
)

func (b *broker) logout(ctx context.Context, profile string) error {
	if _, err := profileName(profile); err != nil {
		return err
	}
	resume, err := b.pause(ctx)
	if err != nil {
		return err
	}
	defer resume()
	for state, source := range b.logins {
		if source == profile {
			_, _ = b.cancelLogin(ctx, state)
		}
	}
	accounts, err := b.accounts(ctx)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.Profile == profile {
			_, err = b.proxy.management(ctx, "DELETE", "/auth-files?name="+url.QueryEscape(account.Name), nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ProfilePrefix(profile string) string { return strings.TrimSpace(profile) + "/" }
