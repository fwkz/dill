package operations

import (
	"os"
	"os/signal"
	"syscall"
)

func ShutdownChannel() <-chan os.Signal {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	return c
}
