// Package session
// in fact, session should not belong to web framework
// under the web package now because it is only used for practice
package session

import (
	"code-practise/web"
	"github.com/google/uuid"
)

type Manager struct {
	Propagator
	Store
}

func (m *Manager) GetSession(ctx *web.Context) (Session, error) {
	sid, err := m.Extract(ctx.Req)
	if err != nil {
		return nil, err
	}
	sess, err := m.Get(ctx.Req.Context(), sid)
	if err != nil {
		return nil, err
	}
	return sess, err
}

func (m *Manager) InitSession(ctx *web.Context) (Session, error) {
	sess, err := m.Generate(ctx.Req.Context(), uuid.New().String())
	if err != nil {
		return nil, err
	}
	// inject http response cookie
	err = m.Inject("id", ctx.Resp)
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
