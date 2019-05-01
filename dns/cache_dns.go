package dns

import (
	"net"
	"sync"
	"time"
)

type ResolverResult struct {
	list    []net.IP
	Expired time.Time
}

type CacheResolver struct {
	cache  map[string]ResolverResult
	lock   sync.RWMutex
	Expire time.Duration
	Max    int
}

func NewResolver(max int, expires time.Duration) *CacheResolver {
	r := new(CacheResolver)
	r.Max = max
	r.Expire = expires
	r.cache = make(map[string]ResolverResult, max/2+1)
	return r
}

func (r *CacheResolver) Lookup(domain string) ([]net.IP, error) {
	now := time.Now()
	r.lock.RLock()
	ret, ok := r.cache[domain]
	r.lock.RUnlock()
	if ok {
		if ret.Expired.IsZero() {
			return ret.list, nil
		}
		if ret.Expired.Before(now) {
			return ret.list, nil
		}
	}
	ips, err := net.LookupIP(domain)
	if err != nil {
		return nil, err
	}
	ret = ResolverResult{
		list: ips,
	}
	if r.Expire > 0 {
		ret.Expired = now
	}

	r.lock.Lock()
	r.cache[domain] = ret
	if r.Max > 0 && r.Max < len(r.cache) {
		i := 0
		newCache := make(map[string]ResolverResult, r.Max)
		for k, item := range r.cache {
			newCache[k] = item
			i++
			if i > r.Max/2 {
				break
			}
		}
		r.cache = newCache
	}
	r.lock.Unlock()
	return ips, err
}
