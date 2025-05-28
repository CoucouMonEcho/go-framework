package net

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

type Pool struct {
	idlesConns chan *idlesConn
	// reqQueue use slices instead of chan for dynamic capacity
	reqQueue []connReq

	maxCnt int
	//initCnt int
	cnt int

	maxIdleTime time.Duration

	factory func() (net.Conn, error)

	mutex sync.RWMutex
}

func NewPool(initCnt, maxIdleCnt, maxCnt int,
	maxIdleTime time.Duration,
	factory func() (net.Conn, error)) (*Pool, error) {

	if initCnt > maxIdleCnt {
		return nil, fmt.Errorf("micro: initCnt greater than maxIdleCnt")
	}
	idlesConns := make(chan *idlesConn, maxIdleCnt)
	for _ = range initCnt {
		conn, err := factory()
		if err != nil {
			return nil, err
		}
		idlesConns <- &idlesConn{
			c:                conn,
			lastActivityTime: time.Now(),
		}
	}

	res := &Pool{
		idlesConns:  idlesConns,
		maxCnt:      maxCnt,
		maxIdleTime: maxIdleTime,
		factory:     factory,
	}
	return res, nil
}

func (p *Pool) Get(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	for {
		select {
		case ic := <-p.idlesConns:
			// idles conn
			if ic.lastActivityTime.Add(p.maxIdleTime).Before(time.Now()) {
				_ = ic.c.Close()
				continue
			}
			return ic.c, nil
		default:
			p.mutex.Lock()
			if p.cnt >= p.maxCnt {
				req := connReq{make(chan net.Conn, 1)}
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

func (p *Pool) Put(ctx context.Context, c net.Conn) error {
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
	case p.idlesConns <- ic:
	default:
		_ = c.Close()
		//p.mutex.Lock()
		p.cnt--
		//p.mutex.Unlock()
	}
	return nil
}

type idlesConn struct {
	c                net.Conn
	lastActivityTime time.Time
}

type connReq struct {
	connChan chan net.Conn
}
