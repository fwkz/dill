package consul

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
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

func (s *service) Proxy() string {
	for _, t := range s.Tags {
		if strings.HasPrefix(t, "dill.proxy=") {
			v := strings.Split(t, "=")
			return v[1]
		}
	}
	return ""
}

func fetchHealthyServices(index int, consulClient *httpClient) ([]string, int, error) {
	req, err := http.NewRequest(
		"GET",
		consulClient.config.Address+"/v1/health/state/passing",
		nil,
	)
	if err != nil {
		return nil, 0, err
	}
	q := url.Values{}
	q.Add("filter", "ServiceTags contains `dill`")
	if index <= 0 {
		index = 1
	}
	q.Add("index", strconv.Itoa(index))
	if consulClient.config.Wait != "" {
		q.Add("wait", consulClient.config.Wait)
	}
	req.URL.RawQuery = q.Encode()
	res, err := consulClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return nil, 0, fmt.Errorf("%d: %s", res.StatusCode, data)
	}

	var healthyServices []struct {
		Name string `json:"ServiceName"`
	}
	err = json.Unmarshal([]byte(data), &healthyServices)
	if err != nil {
		return nil, 0, err
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
	if err != nil || newIndex < index || newIndex < 0 {
		newIndex = 1
	}
	return unique, newIndex, nil
}

func fetchServiceDetails(name string, consulClient *httpClient) ([]service, error) {
	req, err := http.NewRequest(
		"GET",
		consulClient.config.Address+"/v1/health/service/"+name,
		nil,
	)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("passing", "true")
	req.URL.RawQuery = q.Encode()
	res, err := consulClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("%d: %s", res.StatusCode, data)
	}

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
