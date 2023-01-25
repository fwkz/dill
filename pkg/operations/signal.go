package operations

import (
	"os"
	"os/signal"
	"syscall"
)

func ShutdownChannel() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
