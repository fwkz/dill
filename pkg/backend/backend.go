package backend

import (
	"strings"
	"sync"
)

func New(upstreams []string) *Backend {
	b := Backend{upstreams: upstreams, strategy: &roundrobin{}}
	return &b
}

type Backend struct {
	upstreams []string
	strategy  strategy
	rwm       sync.RWMutex
}

func (b *Backend) Select() string {
	b.rwm.RLock()
	u := b.strategy.Select(&b.upstreams)
	b.rwm.RUnlock()
	return u
}

func (b *Backend) SetUpstreams(upstreams []string) {
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
		bd.WriteString(u)
		bd.WriteString("\n")
	}
	b.rwm.RUnlock()
	return bd.String()
}
