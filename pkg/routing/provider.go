package routing

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"

	"dill/pkg/proxy"
	"dill/pkg/routing/consul"
	"dill/pkg/routing/file"
	"dill/pkg/routing/http"
	"dill/pkg/routing/nomad"
)

type routingeMonitor func(chan<- *proxy.RoutingTable)

var routingMonitors = map[string]routingeMonitor{
	"http":   http.MonitorServices,
	"consul": consul.MonitorServices,
	"nomad":  nomad.MonitorServices,
	"file":   file.MonitorServices,
}

// GetRoutingMonitor selects routing provider based on configuration file
func GetRoutingMonitor() (string, routingeMonitor, error) {
	cfg := viper.GetStringMap("routing")
	if len(cfg) > 1 {
		return "", nil, errors.New("multiple routing providers declared")
	}

	for k := range cfg {
		monitor, ok := routingMonitors[k]
		if !ok {
			return k, nil, fmt.Errorf("unknown routing provider: %s", k)
		}
		return k, monitor, nil
	}
	return "", nil, errors.New("no routing provider declared")
}
