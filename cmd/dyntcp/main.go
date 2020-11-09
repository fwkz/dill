package main

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/controller"
	"dyntcp/pkg/operations"
	"dyntcp/pkg/registry/consul"
)

var sch <-chan os.Signal

func init() {
	operations.SetupLogging()
	sch = operations.ShutdownChannel()
}

func main() {
	procs := runtime.NumCPU()
	log.WithField("count", procs).Info("Setting GOMAXPROCS")
	runtime.GOMAXPROCS(procs)

	c := make(chan *controller.RoutingTable)
	go consul.MonitorServices(c)
	controller.ControlRoutes(c, sch)
}
