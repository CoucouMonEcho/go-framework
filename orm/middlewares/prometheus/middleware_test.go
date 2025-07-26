package prometheus

import (
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	NewMiddlewareBuilder("namespace", "subsystem", "name", "help")
}
