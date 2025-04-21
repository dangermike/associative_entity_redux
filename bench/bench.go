package bench

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"testing"

	"github.com/dangermike/associative_entity_redux/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/klauspost/cpuid/v2"
)

type ConnType byte

const (
	ConnTypePostgres = iota
	ConnTypeMySQL
)

type TestTarget interface {
	Kind() string
	Name() string
	Baseline(ctx context.Context, personName string) (int, error)
	P2C(ctx context.Context, personName string) (int, error)
	C2P(ctx context.Context, companyName string) (int, error)
	Close(ctx context.Context) error
}

func Command() *cobra.Command {
	return &cobra.Command{
		Use:  "bench",
		RunE: runBench,
	}
}

func runBench(cmd *cobra.Command, args []string) error {
	testing.Init()
	ctx := cmd.Context()

	log := logging.FromContext(ctx)
	log.Info("environment", zap.String("goos", runtime.GOOS), zap.String("goarch", runtime.GOARCH), zap.String("cpu_brand", cpuid.CPU.BrandName))

	people, companies, err := GetSamplesPg(ctx, "postgres", "cover", 10000)
	if err != nil {
		return fmt.Errorf("failed to get samples: %w", err)
	}

	testTargets, err := GetTestTargets(ctx)
	if err != nil {
		return fmt.Errorf("failed to get test targets: %w", err)
	}

	fmt.Println("engine,test,baseline,p2c,c2p,delta,delta_b")

	for _, tt := range testTargets {
		log := logging.FromContext(ctx).With(zap.String("kind", tt.Kind()), zap.String("name", tt.Name()))
		ctx := logging.NewContext(ctx, log)
		fmt.Printf("%s,%s,", tt.Kind(), tt.Name())
		baseline := runBenchmark(BenchmarkBaseline(ctx, tt, people))
		fmt.Printf("%d,", baseline)
		p2c := runBenchmark(BenchmarkP2C(ctx, tt, people))
		fmt.Printf("%d,", p2c)
		c2p := runBenchmark(BenchmarkC2P(ctx, tt, companies))
		fmt.Printf("%d,", c2p)
		fmt.Printf("%0.2f%%,", ((float64(p2c)/float64(c2p))-1)*100)
		fmt.Printf("%0.2f%%\n", ((float64(p2c-baseline)/float64(c2p-baseline))-1)*100)
	}

	return nil
}

func runBenchmark(f func(*testing.B)) int64 {
	runtime.GC()
	const rcnt = 10
	results := make([]testing.BenchmarkResult, rcnt)
	for i := range rcnt {
		results[i] = testing.Benchmark(f)
	}
	slices.SortFunc(results, func(a, b testing.BenchmarkResult) int {
		return int(a.NsPerOp() - b.NsPerOp())
	})

	// average ex. high and low
	var total int64
	for i := 1; i < rcnt-1; i++ {
		total += results[i].NsPerOp()
	}
	return total / (rcnt - 2)
}

func GetTestTargets(ctx context.Context) ([]TestTarget, error) {
	var testTargets []TestTarget

	schemata := []string{"cover", "nocover"}

	var errs []error
	for _, n := range schemata {
		tt, err := NewPgTestTarget(ctx, n, "postgres", n)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create postgres target '%s': %w", n, err))
			continue
		}
		testTargets = append(testTargets, tt)
	}
	for _, n := range schemata {
		tt, err := NewMySQLTestTarget(ctx, n, n)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to create mysql target '%s': %w", n, err))
			continue
		}
		testTargets = append(testTargets, tt)
	}
	if len(errs) > 0 {
		for _, tt := range testTargets {
			_ = tt.Close(ctx)
		}
		return nil, errors.Join(errs...)
	}
	return testTargets, nil
}
