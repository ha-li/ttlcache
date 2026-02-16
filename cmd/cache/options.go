package cache

type Option func(*ttlCache)

func WithTTL(ttl int64) Option {
	return func(c *ttlCache) {
		c.ttl = ttl
	}
}
