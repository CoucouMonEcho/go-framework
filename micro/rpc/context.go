package rpc

import "context"

type onewayKey struct{}

func CtxWithOneway(ctx context.Context) context.Context {
	return context.WithValue(ctx, onewayKey{}, true)
}

func isOneway(ctx context.Context) bool {
	val := ctx.Value(onewayKey{})
	oneway, ok := val.(bool)
	return ok && oneway
}
