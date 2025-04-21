package bench

import (
	"context"
	"testing"

	"github.com/dangermike/associative_entity_redux/logging"
	"go.uber.org/zap"
)

func BenchmarkBaseline(ctx context.Context, trg TestTarget, names []string) func(b *testing.B) {
	return func(b *testing.B) {
		log := logging.FromContext(ctx).With(zap.String("test", "Baseline"), zap.String("name", trg.Name()))
		var i int
		nlen := len(names)
		for b.Loop() {
			if _, err := trg.Baseline(ctx, names[i%nlen]); err != nil {
				log.Fatal("query failed", zap.Error(err))
			}
			i++
		}
	}
}

func BenchmarkP2C(ctx context.Context, trg TestTarget, names []string) func(b *testing.B) {
	return func(b *testing.B) {
		log := logging.FromContext(ctx).With(zap.String("test", "P2C"), zap.String("name", trg.Name()))
		var i int
		nlen := len(names)
		for b.Loop() {
			if _, err := trg.P2C(ctx, names[i%nlen]); err != nil {
				log.Fatal("query failed", zap.Error(err))
			}
			i++
		}
	}
}

func BenchmarkC2P(ctx context.Context, trg TestTarget, names []string) func(b *testing.B) {
	return func(b *testing.B) {
		log := logging.FromContext(ctx).With(zap.String("test", "C2P"), zap.String("name", trg.Name()))
		var i int
		nlen := len(names)
		for b.Loop() {
			if _, err := trg.C2P(ctx, names[i%nlen]); err != nil {
				log.Fatal("query failed", zap.Error(err))
			}
			i++
		}
	}
}
