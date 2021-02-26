package controller

import (
	"os"

	log "github.com/sirupsen/logrus"

	"dill/pkg/backend"
	"dill/pkg/frontend"
	"dill/pkg/proxy"
)

// ControlRoutes waits for a new version of the routing table,
// compares it with previous version and apply approriate changes.
func ControlRoutes(c <-chan *RoutingTable, shutdown <-chan os.Signal) {
	log.Info("Starting routes controller")
	prt := &RoutingTable{map[string][]string{}, 1}
	for {
		select {
		case rt := <-c:
			updateRouting(rt, prt)
			prt = rt
		case <-shutdown:
			log.Info("Closing routes controller")
			proxy.Shutdown()
			return
		}
	}
}

func updateRouting(routingTable *RoutingTable, previousRoutingTable *RoutingTable) {
	log.WithField(
		"consul_index", routingTable.ConsulIndex,
	).Info("Change occurred, updating the routing.")
	for _, f := range difference(
		previousRoutingTable.FrontendAddresses(),
		routingTable.FrontendAddresses(),
	) {
		if p := proxy.Lookup(f); p != nil {
			log.WithField("address", f).Info("No backends for frontend")
			p.Close()
		}
	}
	for port, upstreams := range routingTable.Table {
		if p := proxy.Lookup(port); p == nil {
			p = proxy.New(
				frontend.New(port),
				backend.New(upstreams),
			)
			p.ListenAndServe()
		} else {
			p.UpdateBackend(upstreams)
		}
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
