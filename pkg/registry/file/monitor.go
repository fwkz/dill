package file

import (
	"dill/pkg/controller"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

type service struct {
	name     string
	listener string
	backend  string
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Routing() ([]string, string) {
	return []string{s.listener}, s.backend
}

func BuildRoutingTable(cfg RoutingConfig) controller.RoutingTable {
	rt := controller.RoutingTable{ConsulIndex: 1, Table: map[string][]string{}}
	for _, e := range cfg {
		for _, b := range e.Backends {
			srv := service{
				name:     e.Name,
				listener: e.Listener,
				backend:  b,
			}
			rt.Update(&srv)
		}
	}
	return rt
}

func MonitorServices(c chan<- *controller.RoutingTable) {
	v := viper.New()
	cfg := readRoutingConfig(v)
	rt := BuildRoutingTable(cfg)
	c <- &rt

	if viper.GetBool("routing.file.watch") {
		v.OnConfigChange(func(e fsnotify.Event) {
			rc := readRoutingConfig(viper.New())
			t := BuildRoutingTable(rc)
			c <- &t
		})
		v.WatchConfig()
	}
}
