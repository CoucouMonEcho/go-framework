package redis

import (
	"code-practise/web/session"
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	// errKeyNotFound sentinel error
	errorKeyNotFound     = errors.New("key not found")
	errorSessionNotFound = errors.New("session not found")
)

type Store struct {
	prefix     string
	client     redis.Cmdable
	expiration time.Duration
}

type StoreOption func(*Store) *Store

func (s *Store) Generate(ctx context.Context, id string) (session.Session, error) {
	key := redisKey(s.prefix, id)
	_, err := s.client.HSet(ctx, key, id, id).Result()
	if err != nil {
		return nil, err
	}
	_, err = s.client.Expire(ctx, key, s.expiration).Result()
	if err != nil {
		return nil, err
	}
	return &Session{
		id:     id,
		key:    key,
		client: s.client,
	}, nil
}

func (s *Store) Refresh(ctx context.Context, id string) error {
	key := redisKey(s.prefix, id)
	ok, err := s.client.Expire(ctx, key, s.expiration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errorSessionNotFound
	}
	return nil
}

func (s *Store) Remove(ctx context.Context, id string) error {
	key := redisKey(s.prefix, id)
	_, err := s.client.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) Get(ctx context.Context, id string) (session.Session, error) {
	key := redisKey(s.prefix, id)
	cnt, err := s.client.Exists(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if cnt == 0 {
		return nil, errorSessionNotFound
	}
	return &Session{
		id:     id,
		key:    key,
		client: s.client,
	}, nil
}

func NewStore(client redis.Cmdable, opts ...StoreOption) *Store {
	res := &Store{
		prefix:     "session",
		client:     client,
		expiration: time.Minute * 15,
	}
	for _, opt := range opts {
		res = opt(res)
	}
	return res
}

func StoreWithPrefix(prefix string) StoreOption {
	return func(store *Store) *Store {
		store.prefix = prefix
		return store
	}
}

type Session struct {
	id     string
	key    string
	client redis.Cmdable
}

func (s *Session) Get(ctx context.Context, key string) (any, error) {
	val, err := s.client.HGet(ctx, s.key, key).Result()
	return val, err
}

func (s *Session) Set(ctx context.Context, key string, value any) error {
	const lua = `
if redis.call("EXISTS", KEYS[1])
then
	return redis.call("HSET", KEYS[1], ARGV[1], ARGV[2])
else
	return -1
end
`
	res, err := s.client.Eval(ctx, lua, []string{s.key}, key, value).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return errorSessionNotFound
	}
	return nil
}

func (s *Session) ID() string {
	return s.id
}

func redisKey(prefix, id string) string {
	return fmt.Sprintf("%s:%s", prefix, id)
}
