package main

import (
	"runtime"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/controller"
	"dyntcp/pkg/operations"
	"dyntcp/pkg/registry/consul"
)

func init() {
	operations.SetupLogging()
	operations.SetupCloseHandler()
}

func main() {
	procs := runtime.NumCPU()
	log.WithField("count", procs).Info("Setting GOMAXPROCS")
	runtime.GOMAXPROCS(procs)

	c := make(chan *consul.RoutingTable)
	go consul.MonitorServices(c)
	controller.ControlRoutes(c)
}
