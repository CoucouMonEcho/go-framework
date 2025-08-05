// Package session
// in fact, session should not belong to web framework
// under the web package now because it is only used for practice
package session

import (
	"github.com/google/uuid"
	"go-framework/web"
)

type Manager struct {
	Propagator
	Store
	CtxSessKey string
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	// get from cache
	if ctx.UserValues == nil {
		ctx.UserValues = make(map[string]any)
	}
	val, ok := ctx.UserValues[m.CtxSessKey]
	// val := ctx.Req.Context().Value(m.CtxSessKey)
	if ok {
		return val.(Session), nil
	}
	// extract session
	sid, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	sess, err := m.Get(ctx.Req.Context(), sid)
	if err != nil {
		return nil, err
	}
	// add cache
	ctx.UserValues[m.CtxSessKey] = sess
	// ctx.Req = ctx.Req.WithContext(context.WithValue(ctx.Req.Context(), m.CtxSessKey, sess))
	return sess, err
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {
	id := uuid.New().String()
	sess, err := m.Generate(ctx.Req.Context(), id)
	if err != nil {
		return nil, err
	}
	// inject http response cookie
	err = m.Inject(id, ctx.Resp)
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func (m *Manager) RefreshSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Refresh(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) RemoveSession(ctx *web.Context) error {
	sess, err := m.GetSession(ctx)
	if err != nil {
		return err
	}
	err = m.Store.Remove(ctx.Req.Context(), sess.ID())
	if err != nil {
		return err
	}
	err = m.Propagator.Remove(ctx.Resp)
	if err != nil {
		return err
	}
	return nil
}
