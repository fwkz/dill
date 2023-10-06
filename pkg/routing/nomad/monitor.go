package nomad

import (
	"log/slog"
	"os"
	"time"

	"github.com/spf13/viper"

	"dill/pkg/proxy"
)

var waitTime time.Duration = 5 * time.Second

// MonitorServices fetches healthy services that was tagged as `dill`
func MonitorServices(c chan<- *proxy.RoutingTable) {
	cfg := nomadConfig{}
	viper.UnmarshalKey("routing.nomad", &cfg)

	nomadClient, err := newNomadClient(&cfg)
	if err != nil {
		slog.Error("Invalid configuration of 'nomad' routing provider", "error", err)
		os.Exit(1)
	}

	var index uint64
	var exposedServices []string
	for {
		rt := proxy.RoutingTable{Table: map[string][]proxy.Upstream{}, ConsulIndex: 0}
		exposedServices, index, err = nomadClient.fetchExposedServices(index)
		if err != nil {
			slog.Warn("Failed to query Nomad's service catalog", "error", err)
			time.Sleep(waitTime)
			continue
		}

		for _, s := range exposedServices {
			details, err := nomadClient.fetchMatchingAllocations(s)
			if err != nil {
				slog.Warn("Failed to fetch service details", "error", err, "name", s)
				continue
			}
			for _, d := range details {
				rt.Update(&service{details: d})
			}
		}
		c <- &rt
		time.Sleep(waitTime)
	}
}
