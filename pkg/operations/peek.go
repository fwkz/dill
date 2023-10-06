package operations

import (
	"log/slog"
	"net"

	"dill/pkg/proxy"
)

func Peek(addr string) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		slog.Error("Failed to start Peek", "error", err)
		return
	}
	slog.Info("Starting Peek", "address", addr)
	for {
		c, err := l.Accept()
		if err != nil {
			slog.Warn("Failed to accept a connection", "error", err)
			continue
		}

		go func() {
			d := proxy.Dump()
			if d == "" {
				d = "No registered backends. Please verify routing configuration.\n"
			}

			c.Write([]byte(d))
			c.Close()
		}()
	}
}
