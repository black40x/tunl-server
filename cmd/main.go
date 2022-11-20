package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/black40x/golog"
	"github.com/black40x/tunl-server/cmd/server"
	"github.com/black40x/tunl-server/cmd/tui"
	"os"
	"os/signal"
	"time"
)

func startTunlServer() {
	if ver := server.CheckUpdates(); ver != nil {
		tui.PrintWarning(fmt.Sprintf("update available: %s\n", ver.String()))
	}

	ctx := context.Background()
	conf, err := server.LoadConfig()
	if err != nil {
		tui.PrintError(errors.New("config load error"))
		os.Exit(1)
	}

	var logger *golog.Logger
	if conf.Log.Enabled {
		logger = golog.NewLogger(&golog.Options{
			LogDir:  conf.Log.LogDir,
			Daily:   conf.Log.LogDaily,
			LogName: "tunl-server",
		}, golog.Ltime|golog.Ldate)
	}

	tunlHttp := server.NewTunlHttp(conf, logger, ctx)
	tunlHttp.Start()

	tui.PrintServerStarted(
		conf.Server.HTTPAddr,
		conf.Server.HTTPPort,
		server.Version,
	)

	var wait time.Duration
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	tunlHttp.Shutdown()
	tui.PrintInfo("Shutting down tunl server")
	os.Exit(0)
}

func main() {
	startTunlServer()
}
