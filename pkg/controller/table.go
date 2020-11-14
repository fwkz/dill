package controller

import (
	"fmt"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Service interface {
	Routing() ([]string, string)
	Name() string
}

type RoutingTable struct {
	Table       map[string][]string
	ConsulIndex int
}

// Update validates the service's routing settings
// and updates routing table if it's valid.
func (rt *RoutingTable) Update(service Service) {
	listeners, upstream := service.Routing()
	if len(listeners) == 0 {
		log.WithFields(log.Fields{
			"service_name": service.Name(),
			"upstream":     upstream,
		}).Warn("No listeners found")
		return
	}

	for _, addr := range listeners {
		a := strings.Split(addr, ":")
		if len(a) != 2 {
			log.WithFields(log.Fields{
				"address":      addr,
				"service_name": service.Name(),
				"upstream":     upstream,
			}).Warn("Invalid listener address, port missing.")
			continue
		}
		label, port := a[0], a[1]

		if p, err := strconv.Atoi(port); p <= viper.GetInt("ports.min") ||
			p >= viper.GetInt("ports.max") ||
			err != nil {
			log.WithFields(log.Fields{
				"port":         port,
				"service_name": service.Name(),
				"upstream":     upstream,
			}).Warn("Invalid listener port")
			continue
		}

		allowedListeners := viper.GetStringMapString("allowed_listeners")
		labelValid := false
		for l, ip := range allowedListeners {
			if label == l {
				labelValid = true
				addr = fmt.Sprintf("%s:%s", ip, a[1])
				break
			}
		}
		if !labelValid {
			log.WithFields(log.Fields{
				"label":             label,
				"allowed_listeners": allowedListeners,
				"service_name":      service.Name(),
				"upstream":          upstream,
			}).Warn("Invalid listener label")
			continue
		}

		rt.update(addr, upstream)
	}
}

func (rt *RoutingTable) update(addr string, upstream string) {
	if t, ok := rt.Table[addr]; ok {
		rt.Table[addr] = append(t, upstream)
	} else {
		rt.Table[addr] = []string{upstream}
	}
}

func (rt *RoutingTable) FrontendAddresses() []string {
	addrs := make([]string, len(rt.Table))
	i := 0
	for k := range rt.Table {
		addrs[i] = k
		i++
	}
	return addrs
}

func (rt *RoutingTable) Dump() {
	for key, value := range rt.Table {
		fmt.Println(key)
		for _, ip := range value {
			fmt.Printf("  --> %s\n", ip)
		}
	}
}
