package controller

type Service interface {
	Routing() ([]string, string)
}

type RoutingTable struct {
	Table       map[string][]string
	ConsulIndex int
}

func (rt *RoutingTable) Update(service Service) {
	frontends, upstream := service.Routing()
	for _, addr := range frontends {
		if t, ok := rt.Table[addr]; ok {
			t = append(t, addr)
			rt.Table[addr] = t
		} else {
			rt.Table[addr] = []string{upstream}
		}
	}
}

func (rt *RoutingTable) GetFrontendAddresses() []string {
	addrs := make([]string, len(rt.Table))
	i := 0
	for k := range rt.Table {
		addrs[i] = k
		i++
	}
	return addrs
}
