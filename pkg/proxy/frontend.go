package proxy

import (
	"log/slog"
	"net"
	"sync"
)

func NewFrontend(address string) *Frontend {
	f := &Frontend{Address: address, conns: make(map[string]net.Conn)}
	return f
}

type Frontend struct {
	Address  string
	listener net.Listener
	conns    map[string]net.Conn
	rwm      sync.RWMutex
}

func (f *Frontend) Listen() error {
	slog.Info("Frontend is starting to listen", "address", f.Address)
	l, err := net.Listen("tcp", f.Address)
	if err != nil {
		return err
	}
	f.listener = l
	return nil
}

func (f *Frontend) Close() {
	slog.Info("Closing frontend", "address", f.Address)
	f.rwm.Lock()
	for _, c := range f.conns {
		c.Close()
	}
	f.conns = make(map[string]net.Conn)
	f.rwm.Unlock()

	if f.listener != nil {
		f.listener.Close()
	}
}

func (f *Frontend) Accept() (net.Conn, error) {
	c, err := f.listener.Accept()
	if err != nil {
		return nil, err
	}
	f.rwm.Lock()
	f.conns[c.RemoteAddr().String()] = c
	f.rwm.Unlock()

	return c, nil
}

func (f *Frontend) RemoveConn(c net.Conn) {
	f.rwm.Lock()
	delete(f.conns, c.RemoteAddr().String())
	f.rwm.Unlock()
}
