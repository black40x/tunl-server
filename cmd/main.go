package main

import (
	"context"
	"errors"
	"github.com/black40x/tunl-server/cmd/server"
	"github.com/black40x/tunl-server/cmd/tui"
	"os"
	"os/signal"
	"time"
)

var Version = "0.1.45"

func startTunlServer() {
	ctx := context.Background()
	conf, err := server.LoadConfig()
	if err != nil {
		tui.PrintError(errors.New("config load error"))
		os.Exit(1)
	}

	tunlHttp := server.NewTunlHttp(conf, ctx)
	tunlHttp.Start()

	tui.PrintServerStarted(conf.Base.HTTPAddr, conf.Base.HTTPPort, Version)

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
