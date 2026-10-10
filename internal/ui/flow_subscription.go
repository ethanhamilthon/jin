package ui

import (
	"context"
	"errors"
	"jin/internal/cliproxy"
	"jin/internal/sources"
	"time"
)

func (a *app) addSubscription() {
	a.ensureProxy(func() {
		sel := a.openList("Subscription", onboardingProfiles, "codex", func(profile string) error {
			return a.addProfile(profile)
		})
		sel.twoLines = true
	})
}

func (a *app) addProfile(profile string) error {
	if err := sources.AddManaged(a.store, profile); err != nil {
		return err
	}
	if err := a.reloadProviders(); err != nil {
		return err
	}
	a.openSubscription("subscription-" + profile)
	return nil
}

func (a *app) openSubscription(id string) {
	entry, err := a.store.Provider(id)
	if err != nil {
		a.report(err)
		return
	}
	if entry.Source != "cliproxy" {
		a.report(errors.New("this is not a subscription provider"))
		return
	}
	client, root, err := sources.Managed()
	if err != nil {
		a.report(err)
		return
	}
	if cliproxy.Installed(root).Version == "" {
		a.ensureProxy(func() { a.openSubscription(id) })
		return
	}
	a.openLoading(entry.Name, id, func(ctx context.Context) ([]option, error) {
		accounts, err := client.Accounts(ctx)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			if account.Profile == entry.Profile {
				return []option{{label: "Account connected · " + account.Status, value: "model"}, {label: "Sign out", detail: "Remove this subscription account", value: "logout"}}, nil
			}
		}
		return []option{{label: "Sign in", detail: "Opens the browser on this computer", value: "login"}}, nil
	}, func(action string) error {
		if action == "model" {
			a.openCatalogPicker()
			return nil
		}
		if action == "logout" {
			a.confirmDelete("subscription account", func(string) error {
				a.settingsAction("Sign out", func(ctx context.Context) error { return client.Logout(ctx, entry.Profile) }, func() { a.openSubscription(id) })
				return nil
			}, func() { a.openSubscription(id) })
			return nil
		}
		a.settingsAction("Sign in · complete the browser flow", func(ctx context.Context) error {
			login, err := client.Login(ctx, entry.Profile)
			if err != nil {
				return err
			}
			defer func() {
				cancel, stop := context.WithTimeout(context.Background(), 5*time.Second)
				defer stop()
				_ = client.Cancel(cancel, login.State)
			}()
			if err = openSubscriptionBrowser(login.URL); err != nil {
				return err
			}
			for {
				status, err := client.Poll(ctx, entry.Profile, login.State)
				if err != nil {
					return err
				}
				if status.Status == "ok" {
					return nil
				}
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(time.Second):
				}
			}
		}, func() { _ = a.reloadProviders(); a.openCatalogPicker() })
		return nil
	})
}
