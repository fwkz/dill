package file

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dill/pkg/controller"
)

type RoutingConfig []struct {
	Name     string
	Listener string
	Backends []string
}

func readRoutingConfig(v *viper.Viper) RoutingConfig {
	p := viper.GetString("routing.file.path")
	log.WithField("path", p).Info("Reading routing config file")

	v.SetConfigFile(p)
	err := v.ReadInConfig()
	if err != nil {
		log.WithError(err).Error("Invalid routing config")
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
}

func (s *service) Name() string {
	return s.name
}

func (s *service) Routing() ([]string, string) {
	return []string{s.listener}, s.backend
}

// BuildRoutingTable builds routing table ouf of static routing configuration file
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
