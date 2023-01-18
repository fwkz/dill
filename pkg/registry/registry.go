package registry

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dill/pkg/controller"
	"dill/pkg/registry/consul"
	"dill/pkg/registry/file"
	"dill/pkg/registry/http"
)

type serviceMonitor func(chan<- *controller.RoutingTable)

var serviceMonitors = map[string]serviceMonitor{
	"http":   http.MonitorServices,
	"consul": consul.MonitorServices,
	"file":   file.MonitorServices,
}

func GetServicesMonitor() serviceMonitor {
	cfg := viper.GetStringMap("routing")
	var monitor serviceMonitor
	for k := range cfg {
		monitor = serviceMonitors[k]
		log.WithField("monitor", k).Info("Monitoring upstream services")
	}
	return monitor
}
