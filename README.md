a simple ttl cache in go.

a background job is spawned to refresh the cache every ttl/2 seconds.

when the cache is refreshed, the old cache is replaced with the new cache.