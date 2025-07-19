package etcd

import (
	"code-practise/micro/registry"
	"context"
	"encoding/json"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Registry struct {
	c    *clientv3.Client
	sess *concurrency.Session
}

func NewRegistry(c *clientv3.Client) (*Registry, error) {
	sess, err := concurrency.NewSession(c)
	if err != nil {
		return nil, err
	}
	return &Registry{
		c:    c,
		sess: sess,
	}, nil
}

func (r *Registry) Register(ctx context.Context, si registry.ServiceInstance) error {
	val, err := json.Marshal(si)
	if err != nil {
		return err
	}
	_, err = r.c.Put(ctx, r.instanceKey(si), string(val), clientv3.WithLease(r.sess.Lease()))
	return err
}

func (r *Registry) UnRegister(ctx context.Context, si registry.ServiceInstance) error {
	_, err := r.c.Delete(ctx, r.instanceKey(si))
	return err
}

func (r *Registry) LisServices(ctx context.Context, name string) ([]registry.ServiceInstance, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Registry) Subscribe(name string) (<-chan registry.Event, error) {
	//TODO implement me
	panic("implement me")
}

func (r *Registry) Close() error {
	return r.sess.Close()
}

func (r *Registry) instanceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s/%s", si.Name, si.Address)
}

func (r *Registry) serviceKey(si registry.ServiceInstance) string {
	return fmt.Sprintf("/micro/%s", si.Name)
}
