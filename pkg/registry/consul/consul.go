package consul

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type service struct {
	ID      string   `json:"ID"`
	Name_   string   `json:"Service"`
	Tags    []string `json:"Tags"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`
}

func (s *service) Routing() ([]string, string) {
	listeners := []string{}
	for _, t := range s.Tags {
		if strings.HasPrefix(t, "dill.listener=") {
			v := strings.Split(t, "=")
			listeners = append(listeners, v[1])
		}
	}
	return listeners, fmt.Sprintf("%s:%d", s.Address, s.Port)
}

func (s *service) Name() string {
	return s.Name_
}

func fetchHealthyServices(index int) ([]string, int, error) {
	// TODO allow for stale reads
	req, err := http.NewRequest(
		"GET",
		viper.GetString("consul.address")+"/v1/health/state/passing",
		nil,
	)
	if err != nil {
		return nil, -1, err
	}
	q := req.URL.Query()
	q.Add("filter", "ServiceTags contains `dill`")
	if index <= 0 {
		index = 1
	}
	q.Add("index", strconv.Itoa(index))
	req.URL.RawQuery = q.Encode()

	c := &http.Client{}
	res, err := c.Do(req)
	if err != nil {
		return nil, -1, err
	}
	defer res.Body.Close()
	data, _ := ioutil.ReadAll(res.Body)
	var healthyServices []struct {
		Name string `json:"ServiceName"`
	}
	err = json.Unmarshal([]byte(data), &healthyServices)
	if err != nil {
		return nil, -1, err
	}

	unique := []string{}
	keys := map[string]struct{}{}
	for _, s := range healthyServices {
		if _, ok := keys[s.Name]; !ok {
			keys[s.Name] = struct{}{}
			unique = append(unique, s.Name)
		}
	}

	newIndex, err := strconv.Atoi(res.Header.Get("X-Consul-Index"))
	if err != nil {
		newIndex = 1
	}
	if newIndex < index {
		newIndex = 1
	}
	return unique, newIndex, nil
}

func fetchServiceDetails(name string) ([]service, error) {
	res, err := http.Get(
		fmt.Sprintf(
			"%s/v1/health/service/%s?passing=true",
			viper.GetString("consul.address"),
			name,
		),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, _ := ioutil.ReadAll(res.Body)

	var parsed []struct {
		Node struct {
			Address string `json:"Address"`
		} `json:"Node"`
		Service service `json:"Service"`
	}
	err = json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		return nil, err
	}
	services := []service{}
	for _, r := range parsed {
		s := r.Service
		// IP address of the service host — if empty, node address should be used
		// Ref. https://www.consul.io/api-docs/catalog#serviceaddress
		if s.Address == "" {
			s.Address = r.Node.Address
		}
		services = append(services, s)
	}
	return services, nil
}
