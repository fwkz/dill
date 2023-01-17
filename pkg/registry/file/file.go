package file

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
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
