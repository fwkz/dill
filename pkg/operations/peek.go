package operations

import (
	"net"

	log "github.com/sirupsen/logrus"

	"dill/pkg/proxy"
)

func Peek(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.WithField("error", err).Error("Failed to start Peek")
		return
	}
	log.WithField("address", addr).Info("Starting Peek")
	for {
		c, err := l.Accept()
		if err != nil {
			log.WithField("error", err).Warning("Failed to accept a connection")
			continue
		}

		d := proxy.Dump()
		if d == "" {
			d = "No registered backends. Please verify services' configuration.\n"
		}

		c.Write([]byte(d))
		c.Close()
	}
}
