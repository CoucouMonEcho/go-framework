package memery

import (
	"context"
	"go-framework/micro/registry"
	"sync"
)

type Registry struct {
	data     map[string][]registry.ServiceInstance
	watchers map[string][]chan registry.Event
	mutex    sync.Mutex
}

func NewRegistry() *Registry {
	return &Registry{
		data:     make(map[string][]registry.ServiceInstance),
		watchers: make(map[string][]chan registry.Event),
	}
}

func (r *Registry) Register(_ context.Context, si registry.ServiceInstance) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.data[si.Name] = append(r.data[si.Name], si)

	// notify watcher
	for _, ch := range r.watchers[si.Name] {
		select {
		case ch <- registry.Event{Type: "register"}:
		default:
		}
	}
	return nil
}

func (r *Registry) UnRegister(_ context.Context, si registry.ServiceInstance) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for i, instance := range r.data[si.Name] {
		if instance.Address == si.Address {
			r.data[si.Name] = append(r.data[si.Name][:i], r.data[si.Name][i+1:]...)
			break
		}
	}

	// notify watcher
	for _, ch := range r.watchers[si.Name] {
		select {
		case ch <- registry.Event{Type: "unregister"}:
		default:
		}
	}
	return nil
}

func (r *Registry) ListServices(_ context.Context, name string) ([]registry.ServiceInstance, error) {
	return r.data[name], nil
}

func (r *Registry) Subscribe(name string) (<-chan registry.Event, error) {
	ch := make(chan registry.Event, 4)
	r.watchers[name] = append(r.watchers[name], ch)
	return ch, nil
}

func (r *Registry) Close() error {
	for _, chs := range r.watchers {
		for _, ch := range chs {
			close(ch)
		}
	}
	return nil
}
