package querylog

import (
	"testing"
	"time"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	NewMiddlewareBuilder(time.Second)
}
