package main

import (
	"os"
)

// 检测标准输入是否来自管道
func checkStdInFromPiped() bool {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}

type cacheData struct {
	cache     map[string]interface{}
	cacheList []string
	maxSize   int
}

func newCacheData(size int) *cacheData {
	return &cacheData{
		cache:     make(map[string]interface{}, size),
		cacheList: make([]string, 0, size),
		maxSize:   size,
	}
}

func (c *cacheData) get(key string) interface{} {
	return c.cache[key]
}

func (c *cacheData) getOrStore(key string, load func() interface{}) (interface{}, bool) {
	if c.maxSize == 0 {
		return load(), true
	}
	v, isOk := c.cache[key]
	if isOk {
		return v, false
	}
	value := load()
	c.set(key, value)
	return value, true
}

func (c *cacheData) set(key string, value interface{}) {
	if c.maxSize > 1 && len(c.cacheList) == c.maxSize {
		delete(c.cache, c.cacheList[0])
		c.cacheList = c.cacheList[1:]
	}
	c.cacheList = append(c.cacheList, key)
	c.cache[key] = value
}
