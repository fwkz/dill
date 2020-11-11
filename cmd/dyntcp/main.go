package main

import (
	"os"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dyntcp/pkg/controller"
	"dyntcp/pkg/operations"
	"dyntcp/pkg/registry/consul"
)

var sch <-chan os.Signal

func init() {
	operations.SetupLogging()
	operations.SetupConfig()
	sch = operations.ShutdownChannel()
}

func main() {
	procs := viper.GetInt("gomaxprocs")
	log.WithField("count", procs).Info("Setting GOMAXPROCS")
	runtime.GOMAXPROCS(procs)

	c := make(chan *controller.RoutingTable)
	go consul.MonitorServices(c)
	controller.ControlRoutes(c, sch)
}
