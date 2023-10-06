package proxy

import (
	"log/slog"
	"os"
)

// ControlRoutes waits for a new version of the routing table,
// compares it with previous version and apply approriate changes.
func ControlRoutes(c <-chan *RoutingTable, shutdown <-chan os.Signal) {
	slog.Info("Starting routes controller")
	prt := &RoutingTable{map[string][]Upstream{}, 1}
	for {
		select {
		case rt := <-c:
			updateRouting(rt, prt)
			prt = rt
		case <-shutdown:
			slog.Info("Closing routes controller")
			Shutdown()
			return
		}
	}
}

func updateRouting(routingTable *RoutingTable, previousRoutingTable *RoutingTable) {
	slog.Info(
		"Change occurred, updating the routing.",
		"consul_index", routingTable.ConsulIndex,
	)
	for _, f := range difference(
		previousRoutingTable.FrontendAddresses(),
		routingTable.FrontendAddresses(),
	) {
		if p := Lookup(f); p != nil {
			slog.Info("No backends for frontend", "address", f)
			p.Close()
		}
	}
	for port, upstreams := range routingTable.Table {
		if p := Lookup(port); p == nil {
			p = NewProxy(
				NewFrontend(port),
				NewBackend(upstreams),
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
