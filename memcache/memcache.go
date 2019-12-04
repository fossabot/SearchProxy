package memcache

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type ValueType struct {
	Value   string
	Expires int64
}

type CacheType struct {
	cache  map[string]*ValueType
	m      sync.RWMutex
	ticker *time.Ticker
	done   chan struct{}
}

func (mc *CacheType) Get(key string) (value string, ok bool) {
	mc.m.RLock()
	defer mc.m.RUnlock()
	valueType, ok := mc.cache[key]

	if ok {
		value = valueType.Value
	}

	return
}

func (mc *CacheType) Set(key, value string) {
	mc.SetEx(key, value, 0)
}

func (mc *CacheType) SetEx(key, value string, expires int64) {
	mc.m.Lock()
	defer mc.m.Unlock()

	if expires > 0 {
		expires += time.Now().Unix()
	}

	mc.cache[key] = &ValueType{
		Value:   value,
		Expires: expires,
	}
}

func (mc *CacheType) Len() (cacheSize int) {
	cacheSize = len(mc.cache)
	return
}

func (mc *CacheType) Cache() (cache map[string]*ValueType) {
	cache = mc.cache
	return
}

func (mc *CacheType) UnsafeDelete(key string) {
	delete(mc.cache, key)
}

func (mc *CacheType) Delete(key string) {
	mc.m.Lock()
	defer mc.m.Unlock()

	mc.UnsafeDelete(key)
}

func (mc *CacheType) Evictor() {
	for {
		select {
		case <-mc.done:
			return
		case <-mc.ticker.C:
			mc.m.Lock()
			for key, value := range mc.cache {
				if value.Expires == 0 {
					continue
				}

				if value.Expires-time.Now().Unix() <= 0 {
					log.Printf("Evicting %s\n", key)
					mc.UnsafeDelete(key)
				}
			}
			mc.m.Unlock()
		}
	}
}

func (mc *CacheType) Stop() {
	mc.ticker.Stop()
	mc.done <- struct{}{}

	log.Debug("Memcache is saying goodbye!")
}

func New() (memCache *CacheType) {
	memCache = &CacheType{cache: make(map[string]*ValueType),
		done:   make(chan struct{}),
		ticker: time.NewTicker(1 * time.Second),
	}
	go memCache.Evictor()

	return
}
