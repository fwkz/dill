package consul

type RoutingTable struct {
	Table       map[string][]Service
	ConsulIndex int
}

func (rt *RoutingTable) Update(service Service) {
	for _, addr := range service.Cfg.FrontendBind {
		if t, ok := rt.Table[addr]; ok {
			t = append(t, service)
			rt.Table[addr] = t
		} else {
			rt.Table[addr] = []Service{service}
		}
	}
}

func (rt *RoutingTable) GetFrontendAddresses() []string {
	addrs := []string{}
	for a := range rt.Table {
		addrs = append(addrs, a)
	}
	return addrs
}
