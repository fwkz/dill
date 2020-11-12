package operations

import (
	"net"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/proxy"
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
		c.Write([]byte(proxy.Dump()))
		c.Close()
	}
}
