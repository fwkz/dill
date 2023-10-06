package consul

import (
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/exp/slices"
)

var validConsistencyModes = []string{"stale", "consistent", "leader"}

type consulConfig struct {
	Address         string
	Token           string
	Datacenter      string
	Namespace       string
	Wait            string
	ConsistencyMode string `mapstructure:"consistency_mode"`
}

func (c *consulConfig) Validate() {
	if c.ConsistencyMode != "" && !slices.Contains(validConsistencyModes, c.ConsistencyMode) {
		slog.Error("Invalid Consul's consistency mode")
		os.Exit(1)
	}
}

type httpClient struct {
	client http.Client
	config *consulConfig
}

func (c *httpClient) Do(req *http.Request) (*http.Response, error) {
	if c.config.Token != "" {
		req.Header.Add("X-Consul-Token", c.config.Token)
	}

	q := req.URL.Query()
	if c.config.Datacenter != "" {
		q.Add("dc", c.config.Datacenter)
	}
	if c.config.Namespace != "" {
		q.Add("ns", c.config.Namespace)
	}
	if c.config.ConsistencyMode != "" {
		q.Add(c.config.ConsistencyMode, "")
	}
	req.URL.RawQuery = q.Encode()
	return c.client.Do(req)
}
