package gcommon

import "context"

func AddToContext(ctx context.Context) context.Context {
	newCtx := context.WithValue(ctx, "X-Username", "")
	newCtx = context.WithValue(ctx, "X-Unitcode", "")
	newCtx = context.WithValue(ctx, "X-Unitname", "")

	return newCtx
}
