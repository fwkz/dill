package main

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
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
	log.WithField("version", version).Info("Starting dill")

	procs := viper.GetInt("runtime.gomaxprocs")
	log.WithField("count", procs).Info("Setting GOMAXPROCS")
	runtime.GOMAXPROCS(procs)

	l := viper.GetString("peek.listener")
	if l != "" {
		go operations.Peek(l)
	}

	c := make(chan *proxy.RoutingTable)

	name, monitor, err := routing.GetRoutingMonitor()
	if err != nil {
		log.WithError(err).Fatal("Failed to setup routing provider")
	}
	log.WithField("provider", name).Info("Starting routing provider")
	go monitor(c)

	proxy.ControlRoutes(c, sch)
}
