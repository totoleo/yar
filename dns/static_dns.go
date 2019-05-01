package dns

import "net"

type StaticResolver struct {
	domains  map[string][]net.IP
	fallback bool
}

func NewStaticResolver(fallback bool) *StaticResolver {
	return &StaticResolver{domains: make(map[string][]net.IP, 8), fallback: fallback}
}
func (r *StaticResolver) Add(domain string, ips ...net.IP) {
	r.domains[domain] = ips
}

func (r *StaticResolver) Lookup(host string) ([]net.IP, error) {
	if ips, ok := r.domains[host]; ok {
		return ips, nil
	}
	if !r.fallback {
		return nil, &net.DNSError{Err: "no such host", Name: host}
	}
	ips, err := net.LookupIP(host)
	if err != nil {
		return nil, err
	}
	return ips, nil
}
