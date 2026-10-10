package ui

import (
	"context"
	"errors"

	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/store"
	"jin/internal/web"

	"github.com/skip2/go-qrcode"
)

func (a *app) openRemoteFlow() {
	if a.backend == nil {
		a.report(errors.New("Remote access requires the daemon"))
		return
	}
	var state web.RemoteState
	if err := a.backend.Remote(a.ctx, daemon.RemoteCommand{Action: "state", Path: a.dir}, &state); err != nil {
		a.report(err)
		return
	}
	toggle := "Enable remote access"
	if state.Enabled {
		toggle = "Disable remote access"
	}
	options := []option{{label: toggle, detail: "Tailscale on this computer and phone; persists without local clients", value: "toggle"}}
	if state.Enabled {
		options = append(options, option{label: "Connect a phone", detail: state.URL, value: "pair"})
	}
	var devices []store.Device
	if err := a.backend.Remote(a.ctx, daemon.RemoteCommand{Action: "devices"}, &devices); err != nil {
		a.report(err)
		return
	}
	for _, device := range devices {
		options = append(options, option{label: device.Name, detail: "Rename or revoke this browser", value: device.ID})
	}
	sel := a.openList("Remote access", options, "", func(value string) error {
		switch value {
		case "toggle":
			a.settingsAction("Remote access", func(ctx context.Context) error {
				return a.backend.Remote(ctx, daemon.RemoteCommand{Action: "set", Enabled: !state.Enabled}, nil)
			}, a.openRemoteFlow)
		case "pair":
			return a.remotePair()
		default:
			for _, device := range devices {
				if device.ID == value {
					a.remoteDevice(device)
					break
				}
			}
		}
		return nil
	})
	sel.twoLines = true
	if state.Error != "" {
		sel.hint = state.Error
	}
}

func (a *app) remotePair() error {
	var pair struct{ URL string }
	if err := a.backend.Remote(a.ctx, daemon.RemoteCommand{Action: "pair"}, &pair); err != nil {
		return err
	}
	code, err := qrcode.New(pair.URL, qrcode.Medium)
	if err != nil {
		return err
	}
	a.active.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Scan with your phone. One use; expires in 5 minutes.\n\n```\n" + code.ToSmallString(false) + "```\n\n" + pair.URL})
	a.sel = nil
	return nil
}

func (a *app) remoteDevice(device store.Device) {
	a.openList(device.Name, []option{{label: "Rename", value: "rename"}, {label: "Revoke access", value: "revoke"}}, "", func(value string) error {
		if value == "rename" {
			a.openField("Device name", device.Name, false, func(name string) error {
				if err := a.backend.Remote(a.ctx, daemon.RemoteCommand{Action: "rename", ID: device.ID, Name: name}, nil); err != nil {
					return err
				}
				a.openRemoteFlow()
				return nil
			})
			return nil
		}
		if err := a.backend.Remote(a.ctx, daemon.RemoteCommand{Action: "revoke", ID: device.ID}, nil); err != nil {
			return err
		}
		a.openRemoteFlow()
		return nil
	})
}
