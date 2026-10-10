package cliproxy

import (
	"context"
	"encoding/json"
)

func (c *Client) Login(ctx context.Context, profile string) (Login, error) {
	data, err := c.call(ctx, command{Action: "login", Profile: profile})
	var login Login
	if err == nil {
		err = json.Unmarshal(data, &login)
	}
	return login, err
}
func (c *Client) Poll(ctx context.Context, profile, state string) (Login, error) {
	data, err := c.call(ctx, command{Action: "poll", Profile: profile, State: state})
	var login Login
	if err == nil {
		err = json.Unmarshal(data, &login)
	}
	return login, err
}
func (c *Client) Cancel(ctx context.Context, state string) error {
	_, err := c.call(ctx, command{Action: "cancel", State: state})
	return err
}
func (c *Client) Logout(ctx context.Context, profile string) error {
	_, err := c.call(ctx, command{Action: "logout", Profile: profile})
	return err
}
func (c *Client) Sync(ctx context.Context, profile string, disabled bool) error {
	_, err := c.call(ctx, command{Action: "sync", Profile: profile, Disabled: disabled})
	return err
}
func (c *Client) Accounts(ctx context.Context) ([]Account, error) {
	data, err := c.call(ctx, command{Action: "accounts"})
	var accounts []Account
	if err == nil {
		err = json.Unmarshal(data, &accounts)
	}
	return accounts, err
}
