package cache

import (
	"testing"
	"time"
)

func TestNewSingleflightCache(t *testing.T) {
	c := &BuildInMapCache{}
	NewSingleflightCache(c, c.Get, time.Second)
}
