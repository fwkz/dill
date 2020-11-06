package controller

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/backend"
	"dyntcp/pkg/frontend"
	"dyntcp/pkg/proxy"
	"dyntcp/pkg/registry/consul"
)

func ControlRoutes(c <-chan *consul.RoutingTable) {
	log.Info("Starting routes controller")
	lastFrontends := []string{}
	for {
		rt := <-c
		currentFrontends := rt.GetFrontendAddresses()
		for _, f := range difference(lastFrontends, currentFrontends) {
			if p := proxy.Lookup(f); p != nil {
				log.WithField("address", f).Info("No backends for frontend")
				p.Close()
			}
		}
		for port, services := range rt.Table {
			upstreams := []string{}
			for _, s := range services {
				upstreams = append(
					upstreams,
					fmt.Sprintf("%s:%d", s.Address, s.Port),
				)
			}
			if p := proxy.Lookup(port); p == nil {
				p = proxy.New(
					frontend.New(port),
					backend.New(upstreams),
				)
				go p.ListenAndServe()
			} else {
				p.UpdateBackend(upstreams)
			}
		}
		lastFrontends = currentFrontends
	}
}

// difference returns elements in `a` that aren't in `b`
func difference(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
