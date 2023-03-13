package nomad

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dill/pkg/proxy"
)

var waitTime time.Duration = 5 * time.Second

// MonitorServices fetches healthy services that was tagged as `dill`
func MonitorServices(c chan<- *proxy.RoutingTable) {
	log.Info("Starting service monitor")

	cfg := nomadConfig{}
	viper.UnmarshalKey("routing.nomad", &cfg)

	nomadClient, err := newNomadClient(&cfg)
	if err != nil {
		log.Fatal("Invalid 'routing.nomad' configuration")
	}

	var index uint64
	var exposedServices []string
	for {
		rt := proxy.RoutingTable{Table: map[string][]proxy.Upstream{}, ConsulIndex: 0}
		exposedServices, index, err = nomadClient.fetchExposedServices(index)
		if err != nil {
			log.WithError(err).Warning("Failed to query Nomad's service catalog")
			time.Sleep(waitTime)
			continue
		}

		for _, s := range exposedServices {
			details, err := nomadClient.fetchMatchingAllocations(s)
			if err != nil {
				log.WithField("name", s).WithError(err).Warning("Failed to fetch service details")
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
