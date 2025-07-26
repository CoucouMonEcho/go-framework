package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Option func(*App)

type ShutdownCallback func(ctx context.Context)

func WithShutdownCallbacks(cbs ...ShutdownCallback) Option {
	return func(app *App) {
		app.cbs = cbs
	}
}

type App struct {
	servers []*Server

	// exit the overall timeout
	shutdownTimeout time.Duration
	// wait for the request processed timeout
	waitTime time.Duration
	// callback timeout
	cbTimeout time.Duration

	cbs []ShutdownCallback
}

func NewApp(servers []*Server, opts ...Option) *App {
	res := &App{
		shutdownTimeout: 30 * time.Second,
		waitTime:        10 * time.Second,
		cbTimeout:       3 * time.Second,
		servers:         servers,
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func (app *App) Run() {
	for _, server := range app.servers {
		srv := server
		go func() {
			if err := srv.Start(); err != nil {
				if errors.Is(err, http.ErrServerClosed) {
					log.Printf("server %s has stopped\n", srv.name)
				} else {
					log.Printf("server %s exit unexpectedly\n", srv.name)
				}
			}
		}()
	}
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, signals...)
	<-ch
	println("shutdown is being prepared...")
	go func() {
		select {
		case <-ch:
			log.Printf("forced shutdown\n")
			os.Exit(1)
		case <-time.After(app.shutdownTimeout):
			log.Printf("timed out forced shutdown\n")
			os.Exit(1)
		}
	}()
	app.shutdown()
}

func (app *App) shutdown() {
	// use http.server can omit this step
	log.Println("reject subsequent requests")
	for _, server := range app.servers {
		// no need to think about concurrency
		server.rejectReq()
	}

	// use http.server can omit this step
	log.Println("wait for the existing request to be executed")
	// modified to count requests being processed
	time.Sleep(app.waitTime)

	log.Println("start shutdown")
	var wg sync.WaitGroup
	wg.Add(len(app.servers))
	for _, server := range app.servers {
		srv := server
		go func() {
			if err := srv.stop(); err != nil {
				log.Printf("shutdown of service %v failed\n", srv)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	log.Println("execute shutdown callbacks")
	wg.Add(len(app.cbs))
	for _, cb := range app.cbs {
		c := cb
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), app.cbTimeout)
			c(ctx)
			cancel()
			wg.Done()
		}()
	}
	wg.Wait()

	log.Println("free up resources")
	app.close()
}

func (app *App) close() {
	log.Println("the app closes")
}

type Server struct {
	srv  *http.Server
	name string
	mux  *serverMux
}

func (s *Server) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

func (s *Server) Start() error {
	return s.srv.ListenAndServe()
}

func (s *Server) rejectReq() {
	s.mux.reject = true
}

func (s *Server) stop() error {
	log.Printf("%s shutting down\n", s.name)
	return s.srv.Shutdown(context.Background())
}

//func (s *Server) stop(ctx context.Context) error {
//	log.Printf("%s shutting down\n", s.name)
//	return s.srv.Shutdown(ctx)
//}

var _ http.Handler = serverMux{}

// serverMux decorator pattern or delegation pattern
type serverMux struct {
	reject bool
	*http.ServeMux
}

func (s *serverMux) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	if s.reject {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("the service is down"))
		return
	}
	s.ServeMux.ServeHTTP(w, r)
}

func NewServer(name string, addr string) *Server {
	mux := &serverMux{ServeMux: http.NewServeMux()}
	return &Server{
		name: name,
		mux:  mux,
		srv: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
	}
}
