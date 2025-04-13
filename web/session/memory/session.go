package memory

import (
	"code-practise/web/session"
	"context"
	"errors"
	cache "github.com/patrickmn/go-cache"
	"sync"
	"time"
)

var (
	// errKeyNotFound sentinel error
	errorKeyNotFound     = errors.New("key not found")
	errorSessionNotFound = errors.New("session not found")
)

type Store struct {
	mutex      sync.RWMutex
	sessions   *cache.Cache
	expiration time.Duration
}

//func NewStore(ms int) *Store {
//	return &Store{}
//}
//
//func NewStore(ms int, unit TimeUnit) *Store {
//	return &Store{}
//}

func NewStore(expiration time.Duration) *Store {
	return &Store{
		sessions:   cache.New(expiration, time.Second),
		expiration: expiration,
	}
}

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	sess := &Session{
		id:     id,
		values: sync.Map{},
	}
	s.sessions.Set(id, sess, s.expiration)
	return sess, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	val, ok := s.sessions.Get(id)
	if !ok {
		return errorSessionNotFound
	}
	s.sessions.Set(id, val, s.expiration)
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.sessions.Delete(id)
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	val, ok := s.sessions.Get(id)
	if !ok {
		return nil, errorSessionNotFound
	}
	return val.(*Session), nil
}

type Session struct {
	id string

	// mutex sync.RWMutex
	// values map[string]any

	values sync.Map
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, ok := s.values.Load(key)
	if !ok {
		// return nil, fmt.Errorf("%w: %s", errSessionNotFound, key)
		return nil, errorKeyNotFound
	}
	return val, nil
}

func (s *Session) Set(ctx context.Context, key string, value any) error {
	s.values.Store(key, value)
	return nil
}

func (s *Session) ID() string {
	// read only, do not need to think about concurrency security
	return s.id
}
