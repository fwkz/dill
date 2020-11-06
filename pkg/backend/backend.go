package backend

func New(upstreams []string) *Backend {
	b := Backend{upstreams: upstreams, strategy: &roundrobin{}}
	return &b
}

type Backend struct {
	upstreams []string
	strategy  strategy
}

func (b *Backend) Select() string {
	return b.strategy.Select(&b.upstreams)
}

func (b *Backend) AddUpstream(addr string) {
	b.upstreams = append(b.upstreams, addr)
}

func (b *Backend) SetUpstreams(upstreams []string) {
	b.upstreams = upstreams
}
