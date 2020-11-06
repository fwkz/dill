package backend

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type strategy interface {
	Select(*[]string) string
}

type roundrobin struct {
	next uint32
}

func (r *roundrobin) Select(upstreams *[]string) string {
	n := atomic.AddUint32(&r.next, 1)
	return (*upstreams)[(int(n)-1)%len(*upstreams)]
}

type random struct {
}

func (r *random) Select(upstreams *[]string) string {
	rand.Seed(time.Now().UnixNano())
	return (*upstreams)[rand.Intn(len(*upstreams))]
}
