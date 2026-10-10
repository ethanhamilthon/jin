package ui

import (
	"context"
	"jin/internal/cliproxy"
	"jin/internal/sources"
)

func (a *app) openProxyFlow() {
	_, root, err := sources.Managed()
	if err != nil {
		a.report(err)
		return
	}
	installed := cliproxy.Installed(root)
	options := []option{{label: "Install " + cliproxy.DefaultVersion, detail: "Verified official binary; independent of Jin", value: "install"}}
	if installed.Version != "" {
		options = []option{{label: "Current: " + installed.Version, detail: "Choose a compatible v8.0.x version", value: "version"}, {label: "Check versions", detail: "Official GitHub releases", value: "versions"}}
		if installed.Previous != "" {
			options = append(options, option{label: "Roll back to " + installed.Previous, value: installed.Previous})
		}
	}
	a.openList("CLIProxyAPI", options, "", func(action string) error {
		if action == "versions" {
			a.openLoading("Compatible CLIProxyAPI releases", installed.Version, func(ctx context.Context) ([]option, error) {
				versions, err := cliproxy.Versions(ctx)
				return plainOptions(versions), err
			}, func(version string) error { a.proxyUpdate(version); return nil })
		} else if action == "version" {
			a.openField("CLIProxyAPI version", installed.Version, false, func(version string) error { a.proxyUpdate(version); return nil })
		} else if action == "install" {
			a.settingsAction("Install CLIProxyAPI", func(ctx context.Context) error { return cliproxy.InstallInitial(ctx, root, cliproxy.DefaultVersion) }, a.openProxyFlow)
		} else {
			a.proxyUpdate(action)
		}
		return nil
	})
}

func (a *app) proxyUpdate(version string) {
	client, _, err := sources.Managed()
	if err != nil {
		a.report(err)
		return
	}
	a.settingsAction("Update CLIProxyAPI · waits for active requests", func(ctx context.Context) error { return client.Update(ctx, version) }, a.openProxyFlow)
}
