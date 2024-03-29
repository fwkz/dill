package file

import (
	"log/slog"

	"github.com/spf13/viper"

	"dill/pkg/proxy"
)

type RoutingConfig []struct {
	Name     string
	Listener string
	Backends []string
	Proxy    string
}

func readRoutingConfig(v *viper.Viper) RoutingConfig {
	p := viper.GetString("routing.file.path")
	slog.Info("Reading routing config file", "path", p)

	v.SetConfigFile(p)
	err := v.ReadInConfig()
	if err != nil {
		slog.Error("Invalid routing config", "error", err)
		return nil
	}
	cfg := RoutingConfig{}
	v.UnmarshalKey("services", &cfg)

	return cfg
}

type service struct {
	name     string
	listener string
	backend  string
	proxy    string
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Routing() ([]string, string) {
	return []string{s.listener}, s.backend
}

func (s *service) Proxy() string {
	return s.proxy
}

// BuildRoutingTable builds routing table ouf of static routing configuration file
func BuildRoutingTable(cfg RoutingConfig) proxy.RoutingTable {
	rt := proxy.RoutingTable{ConsulIndex: 1, Table: map[string][]proxy.Upstream{}}
	for _, e := range cfg {
		for _, b := range e.Backends {
			srv := service{
				name:     e.Name,
				listener: e.Listener,
				backend:  b,
				proxy:    e.Proxy,
			}
			rt.Update(&srv)
		}
	}
	return rt
}
