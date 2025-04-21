package bench

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/dangermike/associative_entity_redux/connections"
	"github.com/dangermike/associative_entity_redux/logging"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

func GetSamplesPg(ctx context.Context, db string, schema string, count int) (people []string, companies []string, err error) {
	conn, err := connections.PgConn(ctx, db, schema)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close(ctx)

	if people, err = GetSamplesPgTable(ctx, conn, "people", count); err != nil {
		return nil, nil, err
	}
	if companies, err = GetSamplesPgTable(ctx, conn, "companies", count); err != nil {
		return nil, nil, err
	}

	return people, companies, err
}

func GetSamplesPgTable(ctx context.Context, conn *pgx.Conn, table string, count int) ([]string, error) {
	start := time.Now()
	log := logging.FromContext(ctx).With(zap.String("table", table))

	q := fmt.Sprintf("SELECT count('x') from %s", cleanObj(table))
	var rowcnt int
	if err := conn.QueryRow(ctx, q).Scan(&rowcnt); err != nil {
		return nil, fmt.Errorf("failed to get count: %w", err)
	}

	sampleRate := 100 * 1.5 * float64(count) / float64(rowcnt)
	q = fmt.Sprintf("SELECT * FROM (SELECT name from %s TABLESAMPLE BERNOULLI (%f)) t LIMIT %d", cleanObj(table), sampleRate, count)
	rows, err := conn.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	vals, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) {
		var s string
		err := row.Scan(&s)
		return s, err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get samples: %w", err)
	}
	log.Info("queried samples", zap.Int("samples", len(vals)), zap.Int("table_size", rowcnt), zap.Duration("duration", time.Since(start)))
	return vals, err
}

var rxObjNeg = regexp.MustCompile("[^a-zA-Z_]")

func cleanObj(name string) string {
	return rxObjNeg.ReplaceAllString(name, "_")
}
