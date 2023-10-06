package proxy

import (
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Service interface {
	Routing() ([]string, string)
	Name() string
	Proxy() string
}

type RoutingTable struct {
	Table       map[string][]Upstream
	ConsulIndex int
}

// Update validates the service's routing settings
// and updates routing table if it's valid.
func (rt *RoutingTable) Update(service Service) {
	listeners, upstreamAddr := service.Routing()
	if len(listeners) == 0 {
		slog.Warn("No listeners found",
			"service_name", service.Name(),
			"upstream", upstreamAddr,
		)
		return
	}

	for _, addr := range listeners {
		a := strings.Split(addr, ":")
		if len(a) != 2 {
			slog.Warn("Invalid listener address, port missing.",
				"address", addr,
				"service_name", service.Name(),
				"upstream", upstreamAddr,
			)
			continue
		}
		label, port := a[0], a[1]

		if p, err := strconv.Atoi(port); p <= viper.GetInt("listeners.port_min") ||
			p >= viper.GetInt("listeners.port_max") ||
			err != nil {
			slog.Warn("Invalid listener port",
				"port", port,
				"service_name", service.Name(),
				"upstream", upstreamAddr,
			)
			continue
		}

		allowedListeners := viper.GetStringMapString("listeners.allowed")
		labelValid := false
		for l, ip := range allowedListeners {
			if label == l {
				labelValid = true
				addr = fmt.Sprintf("%s:%s", ip, a[1])
				break
			}
		}
		if !labelValid {
			slog.Warn("Invalid listener label",
				"label", label,
				"allowed_listeners", allowedListeners,
				"service_name", service.Name(),
				"upstream", upstreamAddr,
			)
			continue
		}

		rt.update(addr, Upstream{address: upstreamAddr, prx: service.Proxy()})
	}
}

func (rt *RoutingTable) update(addr string, upstream Upstream) {
	if t, ok := rt.Table[addr]; ok {
		rt.Table[addr] = append(t, upstream)
	} else {
		rt.Table[addr] = []Upstream{upstream}
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
