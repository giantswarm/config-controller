package cache

import (
	"fmt"
	"time"

	gocache "github.com/patrickmn/go-cache"

	"github.com/giantswarm/config-controller/pkg/github/internal/gitrepo"
)

const (
	defaultExpiration = 0
)

type StoreCache struct {
	underlying *gocache.Cache
}

func NewStoreCache(expiration time.Duration) *StoreCache {
	c := &StoreCache{
		underlying: gocache.New(expiration, expiration/2),
	}
	return c
}

func (c *StoreCache) Set(owner, name, ref string, value *gitrepo.Store) {
	key := storeKey(owner, name, ref)
	c.underlying.Set(key, value, defaultExpiration)
}

func (c *StoreCache) Get(owner, name, ref string) (*gitrepo.Store, bool) {
	key := storeKey(owner, name, ref)
	v, exists := c.underlying.Get(key)
	if v == nil {
		return nil, exists
	}

	store, ok := v.(*gitrepo.Store)
	if !ok {
		return nil, exists
	}

	return store, exists
}

type TagCache struct {
	underlying *gocache.Cache
}

func NewTagCache(expiration time.Duration) *TagCache {
	c := &TagCache{
		underlying: gocache.New(expiration, expiration/2),
	}
	return c
}

func (c *TagCache) Set(owner, name, major string, value string) {
	key := tagKey(owner, name, major)
	c.underlying.Set(key, value, defaultExpiration)
}

func (c *TagCache) Get(owner, name, major string) (string, bool) {
	key := tagKey(owner, name, major)
	v, exists := c.underlying.Get(key)
	if v == nil {
		return "", false
	}

	value, ok := v.(string)
	if !ok {
		return "", false
	}

	return value, exists
}

func storeKey(owner, name, ref string) string {
	return fmt.Sprintf("%s/%s@%s", owner, name, ref)
}

func tagKey(owner, name, major string) string {
	return fmt.Sprintf("%s/%s@%s", owner, name, major)
}
