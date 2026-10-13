//go:build desktop

package desktop

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"

	"jin/internal/web"
)

func Run(version string) error {
	dist, ok := web.Dist()
	if !ok {
		return errors.New("the web UI is not built: run npm run build in web")
	}
	shell := &proxy{}
	api := newRouter(shell)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			api.ServeHTTP(w, r)
			return
		}
		web.SPA(dist).ServeHTTP(w, r)
	})
	var window context.Context
	file := windowFile()
	size := loadWindow(file)
	startup := func(ctx context.Context) {
		window = ctx
		go func() {
			home, _ := os.UserHomeDir()
			to, err := connect(ctx, version, home)
			if err != nil {
				message := "Jin could not start: " + err.Error()
				shell.failed.Store(&message)
				return
			}
			shell.to.Store(to)
		}()
	}
	return wails.Run(&options.App{
		Title: "Jin", Width: size.Width, Height: size.Height, MinWidth: minWidth, MinHeight: minHeight,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "dev.jin.desktop",
			OnSecondInstanceLaunch: func(options.SecondInstanceData) {
				if window != nil {
					runtime.WindowUnminimise(window)
					runtime.Show(window)
				}
			},
		},
		OnBeforeClose: func(ctx context.Context) bool { saveWindow(ctx, file); return false },
		AssetServer:   &assetserver.Options{Handler: handler},
		Mac:           &mac.Options{},
		OnStartup:     startup,
	})
}
