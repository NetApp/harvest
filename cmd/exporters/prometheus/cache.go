package main

import (
	"sync"
	"time"
)

type cache struct {
	*sync.Mutex
	data   map[string][][]byte
	timers map[string]time.Time
	expire time.Duration
}

func newCache(d time.Duration) *cache {
	c := cache{Mutex: &sync.Mutex{}, expire: d}
	c.data = make(map[string][][]byte)
	c.timers = make(map[string]time.Time)
	return &c
}

func (c *cache) Get() map[string][][]byte {
	c.Clean()
	return c.data
}

func (c *cache) Put(key string, data [][]byte) {
	c.data[key] = data
	c.timers[key] = time.Now()
}

func (c *cache) Clean() {
	for k, t := range c.timers {
		if time.Since(t) > c.expire {
			delete(c.timers, k)
			delete(c.data, k)
		}
	}
}
