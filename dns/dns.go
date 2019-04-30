package dns

import "net"

type Resolver interface {
	Lookup(domain string) ([]net.IP, error)
}
