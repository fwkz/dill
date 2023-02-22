package http

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"golang.org/x/exp/slices"

	"dill/pkg/proxy"
	"dill/pkg/routing/file"
)

var supportedFormats = []string{
	"json",
	"toml",
	"yaml",
}

func computeHash(value []byte) []byte {
	h := sha256.New()
	h.Write(value)
	sum := h.Sum(nil)
	return sum
}

// setConfigType sets config type based on Content-Type HTTP header.
func setConfigType(content_type string, v *viper.Viper) error {
	err := errors.New("unsupported config type")
	ct := strings.Split(content_type, "/")
	if len(ct) != 2 {
		return err
	}

	t := ct[len(ct)-1]
	if !slices.Contains(supportedFormats, t) {
		return err
	}
	v.SetConfigType(t)
	return nil
}

func MonitorServices(c chan<- *proxy.RoutingTable) {
	prevSum := []byte{}
	consulIndex := 0
	e := viper.GetString("routing.http.endpoint")
	pollInterval := viper.GetDuration("routing.http.poll_interval")
	pollTimeout := viper.GetDuration("routing.http.poll_timeout")
	client := http.Client{Timeout: pollTimeout}

	for {
		res, err := client.Get(e)
		if err != nil {
			log.WithError(err).Error("Failed to fetch routing configuration")
			time.Sleep(pollInterval)
			continue
		}
		if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
			log.WithField("status_code", res.StatusCode).Error("Failed to fetch routing configuration")
			time.Sleep(pollInterval)
			continue
		}

		data, _ := io.ReadAll(res.Body)
		res.Body.Close()

		ct := res.Header.Get("Content-Type")
		v := viper.New()
		err = setConfigType(ct, v)
		if err != nil {
			log.WithField("content_type", ct).WithError(err).Error("Failed to set routing config type")
			time.Sleep(pollInterval)
			continue
		}

		sum := computeHash(data)
		if bytes.Equal(sum, prevSum) {
			time.Sleep(pollInterval)
			continue
		}
		consulIndex += 1
		prevSum = sum

		err = v.ReadConfig(bytes.NewBuffer(data))
		if err != nil {
			log.WithError(err).Info("Invalid routing config")
		}

		cfg := file.RoutingConfig{}
		v.UnmarshalKey("services", &cfg)

		rt := file.BuildRoutingTable(cfg)
		c <- &rt
		time.Sleep(pollInterval)
	}
}
