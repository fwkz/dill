package http

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

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
			slog.Error("Failed to fetch routing configuration", "error", err)
			time.Sleep(pollInterval)
			continue
		}
		if !(res.StatusCode >= 200 && res.StatusCode <= 299) {
			slog.Error("Failed to fetch routing configuration", "status_code", res.StatusCode)
			time.Sleep(pollInterval)
			continue
		}

		data, _ := io.ReadAll(res.Body)
		res.Body.Close()

		ct := res.Header.Get("Content-Type")
		v := viper.New()
		err = setConfigType(ct, v)
		if err != nil {
			slog.Error("Failed to set routing config type", "content_type", ct, "error", err)
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
			slog.Info("Invalid routing config", "error", err)
		}

		cfg := file.RoutingConfig{}
		v.UnmarshalKey("services", &cfg)

		rt := file.BuildRoutingTable(cfg)
		c <- &rt
		time.Sleep(pollInterval)
	}
}
