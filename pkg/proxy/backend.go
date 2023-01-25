package proxy

import (
	"strings"
	"sync"
)

type Upstream struct {
	address string
	prx     string
}

func NewBackend(upstreams []Upstream) *Backend {
	b := Backend{upstreams: upstreams, strategy: &roundrobin{}}
	return &b
}

type Backend struct {
	upstreams []Upstream
	strategy  strategy
	rwm       sync.RWMutex
}

func (b *Backend) Select() Upstream {
	b.rwm.RLock()
	u := b.strategy.Select(&b.upstreams)
	b.rwm.RUnlock()
	return u
}

func (b *Backend) SetUpstreams(upstreams []Upstream) {
	b.rwm.Lock()
	b.upstreams = upstreams
	b.rwm.Unlock()
}

func (b *Backend) Dump() string {
	var bd strings.Builder
	bd.WriteString("  ├ ")
	bd.WriteString(b.strategy.Name())
	bd.WriteString("\n")
	b.rwm.RLock()
	for _, u := range b.upstreams {
		bd.WriteString("  ├──➤ ")
		if u.prx != "" {
			bd.WriteString(u.prx)
			bd.WriteString(" ─➤ ")
		}
		bd.WriteString(u.address)
		bd.WriteString("\n")
	}
	b.rwm.RUnlock()
	return bd.String()
}
