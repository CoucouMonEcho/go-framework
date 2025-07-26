package cookie

import "net/http"

type Propagator struct {
	cookieName   string
	cookieOption func(c *http.Cookie)
}

type Option func(*Propagator)

func NewPropagator(opts ...Option) *Propagator {
	res := &Propagator{
		cookieName: "session",
		cookieOption: func(c *http.Cookie) {

		},
	}
	for _, opt := range opts {
		opt(res)
	}
	return res
}

func WithCookieName(name string) Option {
	return func(p *Propagator) {
		p.cookieName = name
	}
}

func (p *Propagator) Inject(id string, writer http.ResponseWriter) error {
	c := &http.Cookie{
		Name:  p.cookieName,
		Value: id,
		// HttpOnly: true,
	}
	p.cookieOption(c)
	http.SetCookie(writer, c)
	return nil
}

func (p *Propagator) Extract(req *http.Request) (string, error) {
	c, err := req.Cookie(p.cookieName)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

func (p *Propagator) Remove(writer http.ResponseWriter) error {
	c := &http.Cookie{
		Name:   p.cookieName,
		MaxAge: -1,
		// HttpOnly: true,
	}
	http.SetCookie(writer, c)
	return nil
}
