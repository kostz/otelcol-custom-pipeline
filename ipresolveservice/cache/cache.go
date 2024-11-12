package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Get(key string) (interface{}, bool)
	Set(key string, data interface{})
}

type syncMapCache struct {
	CleanupInterval time.Duration
	ExpiryTime      time.Duration
	Data            *sync.Map
}

type Entry struct {
	Data       interface{}
	ExpiryTime time.Time
}

func (c *syncMapCache) Get(key string) (interface{}, bool) {
	val, ok := c.Data.Load(key)
	if !ok {
		return nil, false
	}
	entry, ok := val.(*Entry)
	if !ok {
		return nil, false
	}
	if entry.ExpiryTime.Before(time.Now()) {
		return nil, false
	}
	return entry.Data, true
}

func (c *syncMapCache) Set(key string, data interface{}) {
	c.Data.Store(key, &Entry{Data: data, ExpiryTime: time.Now().Add(c.ExpiryTime)})
}

func (c *syncMapCache) Expire() {
	for {
		time.Sleep(c.CleanupInterval)
		c.Data.Range(c.cleanup)
	}
}

func (c *syncMapCache) cleanup(key, val any) bool {
	data, ok := val.(*Entry)
	if !ok {
		c.Data.Delete(key)
	}
	if data.ExpiryTime.Before(time.Now()) {
		c.Data.Delete(key)
	}
	return true
}

func NewSyncCache() Cache {
	cache := &syncMapCache{
		CleanupInterval: 30 * time.Second,
		ExpiryTime:      10 * time.Minute,
		Data:            &sync.Map{},
	}
	go cache.Expire()
	return cache
}
