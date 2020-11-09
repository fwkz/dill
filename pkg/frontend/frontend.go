package frontend

import (
	"net"

	log "github.com/sirupsen/logrus"
)

func New(address string) *Frontend {
	f := &Frontend{Address: address}
	return f
}

type Frontend struct {
	Address  string
	listener net.Listener
	conns    []net.Conn
}

func (f *Frontend) Listen() (net.Listener, error) {
	log.WithField("address", f.Address).Info("Frontend is starting to listen")
	l, err := net.Listen("tcp", f.Address)
	if err != nil {
		return nil, err
	}
	f.listener = l
	return l, nil
}

func (f *Frontend) Close() {
	log.WithField("address", f.Address).Info("Closing frontend")
	for _, c := range f.conns {
		c.Close()
	}
	if f.listener != nil {
		f.listener.Close()
	}
}
