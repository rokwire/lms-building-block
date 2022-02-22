package cacheadapter

import (
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

// CacheAdapter structure
type CacheAdapter struct {
	cache *cache.Cache
}

// NewCacheAdapter creates new instance
func NewCacheAdapter(defaultCacheExpirationSeconds string) *CacheAdapter {

	val, err := strconv.ParseInt(defaultCacheExpirationSeconds, 0, 64)
	var duration time.Duration
	if val == 0 || err != nil {
		duration = 1800 * time.Second
	} else {
		duration = time.Duration(val) * time.Second
	}

	cache := cache.New(duration, duration)

	return &CacheAdapter{
		cache: cache,
	}
}
