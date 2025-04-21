package logging

import (
	"context"

	"go.uber.org/zap"
)

type keyType struct{}

var key keyType

func FromContext(ctx context.Context) *zap.Logger {
	switch log := ctx.Value(key).(type) {
	case *zap.Logger:
		return log
	default:
		return zap.L()
	}
}

func NewContext(parent context.Context, log *zap.Logger) context.Context {
	if log == nil {
		return parent
	}
	return context.WithValue(parent, key, log)
}
