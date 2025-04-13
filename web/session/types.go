// Package session
// in fact, session should not belong to web framework
// under the web package now because it is only used for practice
package session

import (
	"context"
	"net/http"
)

//var (
//	// errKeyNotFound sentinel error
//	errKeyNotFound = errors.New("key not found")
//)

type Store interface {
	// Generate id and timeout need or not
	Generate(ctx context.Context, id string) (Session, error)
	// Refresh and Remove is not need to get session first
	// Func(ctx context.Context, sess Session) error
	Refresh(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (Session, error)
}

type Session interface {
	Get(ctx context.Context, key string) (any, error)
	Set(ctx context.Context, key string, value any) error
	ID() string
}

type Propagator interface {
	Inject(id string, writer http.ResponseWriter) error
	Extract(req *http.Request) (string, error)
	Remove(writer http.ResponseWriter) error
}
