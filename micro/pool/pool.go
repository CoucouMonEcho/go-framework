package pool

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Conn any

type Pool struct {
	idlesConnes chan *idlesConn
	// reqQueue use slices instead of chan for dynamic capacity
	reqQueue []connReq

	maxCnt int
	//initCnt int
	cnt int

	maxIdleTime time.Duration

	factory func() (Conn, error)
	close   func(conn any) error

	mutex sync.RWMutex
}

func NewPool(config *Config) (*Pool, error) {

	if config.InitialCap > config.MaxIdle {
		return nil, fmt.Errorf("micro: initCnt greater than maxIdleCnt")
	}
	idlesConnes := make(chan *idlesConn, config.MaxIdle)
	for range config.InitialCap {
		conn, err := config.Factory()
		if err != nil {
			return nil, err
		}
		idlesConnes <- &idlesConn{
			c:                conn,
			lastActivityTime: time.Now(),
		}
	}

	res := &Pool{
		idlesConnes: idlesConnes,
		maxCnt:      config.MaxCap,
		maxIdleTime: config.IdleTimeout,
		factory: func() (Conn, error) {
			return config.Factory()
		},
		close: config.Close,
	}
	return res, nil
}

func (p *Pool) Get(ctx context.Context) (Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for {
		select {
		case ic := <-p.idlesConnes:
			// idles conn
			if ic.lastActivityTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = p.close(ic.c)
				continue
			}
			return ic.c, nil
		default:
			p.mutex.Lock()
			if p.cnt >= p.maxCnt {
				req := connReq{make(chan Conn, 1)}
				p.reqQueue = append(p.reqQueue, req)
				p.mutex.Unlock()
				select {
				// 1 delete from queue
				// 2 forward
				case <-ctx.Done():
					go func() {
						c := <-req.connChan
						_ = p.Put(context.Background(), c)
					}()
				case c := <-req.connChan:
					// ping to checkout
					return c, nil
				}
			}
			c, err := p.factory()
			if err != nil {
				return nil, err
			}
			p.cnt++
			p.mutex.Unlock()
			return c, nil
		}
	}
}

func (p *Pool) Put(_ context.Context, c Conn) error {
	p.mutex.Lock()
	if ql := len(p.reqQueue); ql > 0 {
		req := p.reqQueue[ql-1]
		p.reqQueue = p.reqQueue[:ql-1]
		p.mutex.Unlock()
		req.connChan <- c
	}
	defer p.mutex.Unlock()
	ic := &idlesConn{
		c:                c,
		lastActivityTime: time.Now(),
	}
	select {
	case p.idlesConnes <- ic:
	default:
		_ = p.close(c)
		//p.mutex.Lock()
		p.cnt--
		//p.mutex.Unlock()
	}
	return nil
}

type Config struct {
	InitialCap  int
	MaxCap      int
	MaxIdle     int
	IdleTimeout time.Duration
	Factory     func() (any, error)
	Close       func(conn any) error
}

type idlesConn struct {
	c                Conn
	lastActivityTime time.Time
}

type connReq struct {
	connChan chan Conn
}
