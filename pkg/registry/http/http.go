package http

import (
	"bytes"
	"crypto/sha256"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"dill/pkg/controller"
	"dill/pkg/registry/file"
)

func computeHash(value []byte) []byte {
	h := sha256.New()
	h.Write(value)
	sum := h.Sum(nil)
	return sum
}

func MonitorServices(c chan<- *controller.RoutingTable) {
	prevSum := []byte{}
	consulIndex := 0
	e := viper.GetString("routing.http.endpoint")
	pollInterval := viper.GetDuration("routing.http.poll_interval")

	for {
		res, err := http.Get(e)
		if err != nil {
			log.WithError(err).Error("Failed to fetch routing configuration")
			time.Sleep(pollInterval)
			continue
		}

		data, _ := io.ReadAll(res.Body)
		res.Body.Close()

		sum := computeHash(data)
		if bytes.Equal(sum, prevSum) {
			time.Sleep(pollInterval)
			continue
		}
		consulIndex += 1
		prevSum = sum

		v := viper.New()
		v.SetConfigType("toml")
		err = v.ReadConfig(bytes.NewBuffer(data))
		if err != nil {
			log.WithError(err).Info("Invalid config")
		}

		cfg := file.RoutingConfig{}
		v.UnmarshalKey("services", &cfg)

		rt := file.BuildRoutingTable(cfg)
		c <- &rt
		time.Sleep(pollInterval)
	}
}
