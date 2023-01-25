package consul

import (
	"time"

	log "github.com/sirupsen/logrus"

	"dill/pkg/proxy"
)

var waitTime time.Duration = 5 * time.Second

// MonitorServices fetches healthy services that was tagged as `dill`
func MonitorServices(c chan<- *proxy.RoutingTable) {
	log.Info("Starting service monitor")
	index := 1
	for {
		services, newIndex, err := fetchHealthyServices(index)
		if err != nil {
			log.WithField("error", err).Warning("Fetching healthy services failed")
			time.Sleep(waitTime)
			continue
		}
		index = newIndex
		rt := &proxy.RoutingTable{Table: map[string][]proxy.Upstream{}, ConsulIndex: newIndex}
		for _, s := range services {
			details, err := fetchServiceDetails(s)
			if err != nil {
				log.WithFields(
					log.Fields{"error": err, "service": s},
				).Warning("Fetching service details failed")
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
