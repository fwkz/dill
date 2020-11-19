package proxy

import (
	"io"
	"net"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/backend"
	"dyntcp/pkg/frontend"
)

var (
	proxies = make(map[string]*Proxy)
	rwm     sync.RWMutex
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

func New(frontend *frontend.Frontend, backend *backend.Backend) *Proxy {
	p := &Proxy{frontend: frontend, backend: backend}
	rwm.Lock()
	proxies[frontend.Address] = p
	rwm.Unlock()
	return p
}

func Lookup(listenerAddr string) *Proxy {
	rwm.RLock()
	p, ok := proxies[listenerAddr]
	rwm.RUnlock()
	if ok {
		return p
	}
	return nil
}

// Shutdown gracefully closes proxies, its listeners and all the connections.
func Shutdown() {
	rwm.Lock()
	for _, p := range proxies {
		p.frontend.Close()
	}
	proxies = make(map[string]*Proxy)
	rwm.Unlock()
}

type Proxy struct {
	frontend *frontend.Frontend
	backend  *backend.Backend
}

func (p *Proxy) ListenAndServe() {
	l, err := p.frontend.Listen()
	if err != nil {
		log.WithField("error", err).Warning("Can't establish frontend listener")
		return
	}
	p.serve(l)
}

func (p *Proxy) Close() {
	p.frontend.Close()
	rwm.Lock()
	delete(proxies, p.frontend.Address)
	rwm.Unlock()
}

func (p *Proxy) UpdateBackend(upstreams []string) {
	p.backend.SetUpstreams(upstreams)
}

func (p *Proxy) serve(l net.Listener) error {
	defer l.Close()
	var delay time.Duration
	for {
		c, err := l.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if delay == 0 {
					delay = 5 * time.Millisecond
				} else {
					delay *= 2
					if delay > time.Second {
						delay = time.Second
					}
				}
				time.Sleep(delay)
				continue
			}
			return err
		}
		delay = 0
		go p.handle(c)
	}

}

func (p *Proxy) handle(in net.Conn) {
	out, err := net.Dial("tcp", p.backend.Select())
	if err != nil {
		in.Close()
		return
	}
	once := sync.Once{}
	cp := func(dst net.Conn, src net.Conn) {
		buf := bufferPool.Get().([]byte)
		defer bufferPool.Put(buf)
		_, err := io.CopyBuffer(dst, src, buf)
		if err != nil {
		}
		once.Do(func() {
			in.Close()
			out.Close()
		})
	}
	go cp(in, out)
	cp(out, in)
}

func Dump() string {
	var bd strings.Builder
	rwm.RLock()
	for addr, p := range proxies {
		bd.WriteString(addr)
		bd.WriteString("\n")
		bd.WriteString(p.backend.Dump())
	}
	rwm.RUnlock()
	return bd.String()
}
