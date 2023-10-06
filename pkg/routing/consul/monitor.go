package consul

import (
	"log/slog"
	"time"

	"github.com/spf13/viper"

	"dill/pkg/proxy"
)

var waitTime time.Duration = 5 * time.Second

// MonitorServices fetches healthy services that was tagged as `dill`
func MonitorServices(c chan<- *proxy.RoutingTable) {
	cfg := consulConfig{}
	viper.UnmarshalKey("routing.consul", &cfg)
	cfg.Validate()

	consulClient := &httpClient{config: &cfg}

	index := 1
	for {
		services, newIndex, err := fetchHealthyServices(index, consulClient)
		if err != nil {
			slog.Warn("Fetching healthy services failed", "error", err)
			time.Sleep(waitTime)
			continue
		}
		index = newIndex
		rt := &proxy.RoutingTable{Table: map[string][]proxy.Upstream{}, ConsulIndex: newIndex}
		for _, s := range services {
			details, err := fetchServiceDetails(s, consulClient)
			if err != nil {
				slog.Warn("Fetching service details failed",
					"error", err, "service", s,
				)
				continue
			}
			for _, i := range details {
				rt.Update(&i)
			}
		}
		c <- rt
		// TODO: naive rate limiting use something
		// more resilient like token bucket algorithm
		time.Sleep(waitTime)
	}
}
