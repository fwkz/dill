package main

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/spf13/viper"

	"dill/pkg/operations"
	"dill/pkg/proxy"
	"dill/pkg/routing"
)

var version = "0.0.0" // will inserted dynamically at build time
var sch <-chan os.Signal

func init() {
	operations.SetupLogging()
	operations.SetupConfig()
	sch = operations.ShutdownChannel()
}

func main() {
	slog.Info("Starting dill", "version", version)

	procs := viper.GetInt("runtime.gomaxprocs")
	slog.Info("Setting GOMAXPROCS", "count", procs)
	runtime.GOMAXPROCS(procs)

	l := viper.GetString("peek.listener")
	if l != "" {
		go operations.Peek(l)
	}

	c := make(chan *proxy.RoutingTable)

	name, monitor, err := routing.GetRoutingMonitor()
	if err != nil {
		slog.Error("Failed to setup routing provider", "error", err)
		os.Exit(1)
	}
	slog.Info("Starting routing provider", "provider", name)
	go monitor(c)

	proxy.ControlRoutes(c, sch)
}
