package task

import (
	"context"
	"fmt"
)

func toString(a interface{}) string {
	switch v := a.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return fmt.Sprint(v)
	}
}

func firstContext(ctx []context.Context, def context.Context) context.Context {
	if len(ctx) > 0 && ctx[0] != nil {
		return ctx[0]
	}
	return def
}

func isCanceledContext(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
