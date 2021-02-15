package consul

import (
	"time"

	log "github.com/sirupsen/logrus"

	"dill/pkg/controller"
)

func MonitorServices(c chan<- *controller.RoutingTable) {
	log.Info("Starting service monitor")
	index := 1
	for {
		services, newIndex, err := fetchHealthyServices(index)
		if err != nil {
			log.WithField("error", err).Warning("Fetching healthy services failed")
		}
		index = newIndex
		rt := &controller.RoutingTable{Table: map[string][]string{}, ConsulIndex: newIndex}
		for _, s := range services {
			details, err := fetchServiceDetails(s)
			if err != nil {
				log.WithFields(
					log.Fields{"error": err, "service": s},
				).Warning("Fetching service details for failed")
				continue
			}
			for _, i := range details {
				rt.Update(&i)
			}
		}
		c <- rt
		// TODO: naive rate limiting use something
		// more resilient like token bucket algorithm
		time.Sleep(5 * time.Second)
	}
}
