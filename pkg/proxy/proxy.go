package proxy

import (
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"
	"time"

	proxy_ "golang.org/x/net/proxy"
)

var (
	proxies = make(map[string]*Proxy)
	rwm     sync.RWMutex
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		b := make([]byte, 32*1024)
		return &b
	},
}

func NewProxy(frontend *Frontend, backend *Backend) *Proxy {
	p := &Proxy{frontend: frontend, backend: backend, quit: make(chan struct{})}
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
		close(p.quit)
		p.frontend.Close()
		p.wg.Wait()
	}
	proxies = make(map[string]*Proxy)
	rwm.Unlock()
}

type Proxy struct {
	frontend *Frontend
	backend  *Backend
	quit     chan struct{}
	wg       sync.WaitGroup
}

func (p *Proxy) ListenAndServe() {
	err := p.frontend.Listen()
	if err != nil {
		slog.Warn("Can't establish frontend listener", "error", err)
		return
	}
	p.wg.Add(1)
	go func() {
		err := p.serve()
		if err != nil {
			slog.Error(
				"Can't serve requests",
				"error", err,
				"listener", p.frontend.Address,
			)
		}
		p.wg.Done()
	}()
}

func (p *Proxy) Close() {
	close(p.quit)
	p.frontend.Close()
	p.wg.Wait()

	rwm.Lock()
	delete(proxies, p.frontend.Address)
	rwm.Unlock()
}

func (p *Proxy) UpdateBackend(upstreams []Upstream) {
	p.backend.SetUpstreams(upstreams)
}

func (p *Proxy) serve() error {
	for {
		c, err := p.frontend.Accept()
		if err != nil {
			select {
			case <-p.quit:
				return nil
			default:
				if ne, ok := err.(net.Error); ok && ne.Temporary() {
					time.Sleep(50 * time.Millisecond)
					continue
				}
				return err
			}
		}
		p.wg.Add(1)
		go func() {
			p.handle(c, p.backend.Select())
			p.frontend.RemoveConn(c)
			p.wg.Done()
		}()
	}
}

type dialer interface {
	Dial(string, string) (net.Conn, error)
}

func (p *Proxy) dial(u Upstream) (net.Conn, error) {
	var d dialer
	if u.prx != "" {
		url, err := url.Parse(u.prx)
		if err != nil {
			slog.Info("Failed to parse proxy URL", "error", err)
			return nil, err
		}
		d, err = proxy_.FromURL(url, proxy_.Direct)
		if err != nil {
			slog.Info("SOCKS proxy connection failed", "error", err)
			return nil, err
		}
	} else {
		d = &net.Dialer{}
	}

	out, err := d.Dial("tcp", u.address)
	return out, err
}

func (p *Proxy) handle(in net.Conn, u Upstream) {
	out, err := p.dial(u)
	if err != nil {
		in.Close()
		slog.Info("Connection to upstream failed", "error", err)
		return
	}
	once := sync.Once{}
	cp := func(dst net.Conn, src net.Conn) {
		buf := bufferPool.Get().(*[]byte)
		defer bufferPool.Put(buf)
		_, _ = io.CopyBuffer(dst, src, *buf)
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
