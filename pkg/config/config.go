package config

import (
	"strings"
)

type Config struct {
	FrontendBind []string "dyntcp.frontend.bind"
}

func NewFromTags(tags []string) *Config {
	c := Config{}
	for _, t := range tags {
		if strings.HasPrefix(t, "dyntcp.frontend.bind=") {
			v := strings.Split(t, "=")
			c.FrontendBind = append(c.FrontendBind, v[1])
		}
	}
	return &c
}
