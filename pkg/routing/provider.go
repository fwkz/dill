package routing

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dill/pkg/proxy"
	"dill/pkg/routing/consul"
	"dill/pkg/routing/file"
	"dill/pkg/routing/http"
)

type routingeMonitor func(chan<- *proxy.RoutingTable)

var routingMonitors = map[string]routingeMonitor{
	"http":   http.MonitorServices,
	"consul": consul.MonitorServices,
	"file":   file.MonitorServices,
}

// GetRoutingMonitor selects routing provider based on configuration file
func GetRoutingMonitor() routingeMonitor {
	cfg := viper.GetStringMap("routing")
	var monitor routingeMonitor
	for k := range cfg {
		monitor = routingMonitors[k]
		log.WithField("provider", k).Info("Monitoring upstream services")
	}
	return monitor
}
