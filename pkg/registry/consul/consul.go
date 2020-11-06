package consul

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/config"
)

func MonitorServices(c chan<- *RoutingTable) {
	log.Info("Starting service monitor")
	index := 1
	for {
		services, newIndex, err := fetchHealthyServices(index)
		if err != nil {
			log.WithField("error", err).Warning("Fetching healthy services failed")
		}
		index = newIndex
		rt := &RoutingTable{Table: map[string][]Service{}, ConsulIndex: newIndex}
		for _, s := range services {
			details, err := fetchServiceDetails(s)
			if err != nil {
				log.WithFields(
					log.Fields{"error": err, "service": s},
				).Warning("Fetching service details for failed")
			}
			for _, i := range details {
				rt.Update(i)
			}
		}
		c <- rt
		// TODO: naive rate limiting use something
		// more resilient like token bucket algorithm
		time.Sleep(5 * time.Second)
	}
}

type Service struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Service"`
	Tags    []string `json:"Tags"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`

	Cfg *config.Config
}

func fetchHealthyServices(index int) ([]string, int, error) {
	// TODO allow for stale reads
	req, err := http.NewRequest(
		"GET",
		"http://127.0.0.1:8500/v1/health/state/passing",
		nil,
	)
	if err != nil {
		return nil, -1, err
	}
	q := req.URL.Query()
	q.Add("filter", "ServiceTags contains `dyntcp`")
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

func fetchServiceDetails(name string) ([]Service, error) {
	res, err := http.Get(
		"http://127.0.0.1:8500/v1/health/service/" + name + "?passing=true",
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
		Service Service `json:"Service"`
	}
	err = json.Unmarshal([]byte(data), &parsed)
	if err != nil {
		return nil, err
	}
	services := []Service{}
	for _, r := range parsed {
		s := r.Service
		// IP address of the service host — if empty, node address should be used
		// Ref. https://www.consul.io/api-docs/catalog#serviceaddress
		if s.Address == "" {
			s.Address = r.Node.Address
		}
		s.Cfg = config.NewFromTags(s.Tags)
		services = append(services, s)
	}
	return services, nil
}
