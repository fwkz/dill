package proxy

import (
	"io"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"dyntcp/pkg/backend"
	"dyntcp/pkg/frontend"
)

var proxies = make(map[string]*Proxy)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}

func New(frontend *frontend.Frontend, backend *backend.Backend) *Proxy {
	p := Proxy{frontend: frontend, backend: backend}
	proxies[frontend.Address] = &p
	return &p
}

func Lookup(listenerAddr string) *Proxy {
	if p, ok := proxies[listenerAddr]; ok {
		return p
	}
	return nil
}

func CloseProxies() {
	for _, p := range proxies {
		p.Close()
	}
}

type Proxy struct {
	frontend *frontend.Frontend
	backend  *backend.Backend
}

func (p *Proxy) ListenAndServe() {
	l, err := p.frontend.Listen()
	if err != nil {
		log.Warning("Can't establish frontend listener: %s", err)
		return
	}
	p.serve(l)
}

func (p *Proxy) Close() {
	p.frontend.Shutdown()
	delete(proxies, p.frontend.Address)
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
