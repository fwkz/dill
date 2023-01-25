package proxy

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type strategy interface {
	Select(*[]Upstream) Upstream
	Name() string
}

type roundrobin struct {
	next uint32
}

func (r *roundrobin) Select(upstreams *[]Upstream) Upstream {
	n := atomic.AddUint32(&r.next, 1)
	return (*upstreams)[(int(n)-1)%len(*upstreams)]
}

func (r *roundrobin) Name() string {
	return "round_robin"
}

type random struct {
}

func (r *random) Select(upstreams *[]string) string {
	rand.Seed(time.Now().UnixNano())
	return (*upstreams)[rand.Intn(len(*upstreams))]
}

func (r *random) Name() string {
	return "random"
}
