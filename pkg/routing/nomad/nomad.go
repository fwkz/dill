package nomad

import (
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/nomad/api"
	"golang.org/x/exp/slices"
)

type service struct {
	details *api.ServiceRegistration
}

func (s *service) Routing() ([]string, string) {
	listeners := []string{}
	for _, t := range s.details.Tags {
		if strings.HasPrefix(t, "dill.listener=") {
			v := strings.Split(t, "=")
			listeners = append(listeners, v[1])
		}
	}
	return listeners, fmt.Sprintf("%s:%d", s.details.Address, s.details.Port)
}

func (s *service) Name() string {
	return s.details.ServiceName
}

func (s *service) Proxy() string {
	for _, t := range s.details.Tags {
		if strings.HasPrefix(t, "dill.proxy=") {
			v := strings.Split(t, "=")
			return v[1]
		}
	}
	return ""
}

type nomadConfig struct {
	Address   string
	Token     string
	Namespace string
	Wait      time.Duration
	Stale     bool
	TLS       struct {
		CA       string
		Cert     string
		Key      string
		Insecure bool
	}
}

type nomadClient struct {
	client *api.Client
	config *nomadConfig
}

// fetchExposedServices fetches services that was registered
// in Nomad's service catalog with tag `dill`
func (c *nomadClient) fetchExposedServices(index uint64) ([]string, uint64, error) {
	namespacedServices, meta, err := c.client.Services().List(
		&api.QueryOptions{AllowStale: c.config.Stale, WaitIndex: index, WaitTime: c.config.Wait},
	)
	if err != nil {
		return nil, 0, err
	}

	services := []string{}
	for _, ns := range namespacedServices {
		for _, s := range ns.Services {
			if slices.Contains(s.Tags, "dill") {
				services = append(services, s.ServiceName)
			}
		}
	}
	return services, meta.LastIndex, nil
}

// fetchMatchingAllocations fetches all alocations of the service tagged as `dill`.
// As `api.Client.Services().List()` of `fetchExposedServices()` will return services
// with aggregated list of tags from old and new allocations, matching tags once again
// is required to cover case when you deploy already running job with new `dill` tags.
func (c *nomadClient) fetchMatchingAllocations(name string) ([]*api.ServiceRegistration, error) {
	services, _, err := c.client.Services().Get(
		name,
		&api.QueryOptions{AllowStale: c.config.Stale},
	)
	if err != nil {
		return nil, err
	}

	var matchingServices []*api.ServiceRegistration
	for _, s := range services {
		if slices.Contains(s.Tags, "dill") {
			matchingServices = append(matchingServices, s)
		}
	}
	return matchingServices, nil
}

func newNomadClient(config *nomadConfig) (*nomadClient, error) {
	client, err := api.NewClient(&api.Config{
		Address:   config.Address,
		SecretID:  config.Token,
		Namespace: config.Namespace,
		WaitTime:  config.Wait,
		TLSConfig: &api.TLSConfig{
			CACert:     config.TLS.CA,
			ClientCert: config.TLS.Cert,
			ClientKey:  config.TLS.Key,
			Insecure:   config.TLS.Insecure,
		},
	})
	if err != nil {
		return nil, err
	}

	return &nomadClient{client: client, config: config}, nil
}
