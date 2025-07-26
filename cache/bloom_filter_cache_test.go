package cache

import (
	"testing"
	"time"
)

func TestNewBloomFilterCache(t *testing.T) {
	var bf BloomFilter
	c := NewBuildInMapCache(10 * time.Second)
	NewBloomFilterCache(c, bf, c.Get)
}
