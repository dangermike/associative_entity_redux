package bench

import (
	"context"
	"fmt"

	"github.com/dangermike/associative_entity_redux/connections"
	"github.com/jackc/pgx/v5"
)

var _ TestTarget = new(PgTestTarget)

type PgTestTarget struct {
	name string
	conn *pgx.Conn
}

func NewPgTestTarget(ctx context.Context, name string, db string, schema string) (*PgTestTarget, error) {
	trg := PgTestTarget{name: name}
	var err error
	trg.conn, err = connections.PgConn(ctx, db, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres '%s': %w", name, err)
	}
	if err := trg.conn.Ping(ctx); err != nil {
		_ = trg.Close(ctx)
		return nil, fmt.Errorf("failed to ping postgres '%s': %w", name, err)
	}
	return &trg, nil
}

func (trg *PgTestTarget) Kind() string {
	return "postgres"
}

func (trg *PgTestTarget) Name() string {
	return trg.name
}

func (trg *PgTestTarget) Baseline(ctx context.Context, personName string) (int, error) {
	const q = `SELECT count(distinct name) FROM people AS p WHERE p.name = $1`
	var res int
	err := trg.conn.QueryRow(ctx, q, personName).Scan(&res)
	return res, err
}

func (trg *PgTestTarget) P2C(ctx context.Context, personName string) (int, error) {
	const p2cQ = `SELECT count(distinct c.name)
	FROM
		people AS p INNER JOIN
		people_companies AS pc ON
			p.id = pc.person_id INNER JOIN
		companies AS c ON
			pc.company_id = c.id
	WHERE
		p.name = $1`
	var res int
	err := trg.conn.QueryRow(ctx, p2cQ, personName).Scan(&res)
	return res, err
}

func (trg *PgTestTarget) C2P(ctx context.Context, companyName string) (int, error) {
	const c2pQ = `SELECT count(distinct p.name)
	FROM
		people AS p INNER JOIN
		people_companies AS pc ON
			p.id = pc.person_id INNER JOIN
		companies AS c ON
			pc.company_id = c.id
	WHERE
		c.name = $1`
	var res int
	err := trg.conn.QueryRow(ctx, c2pQ, companyName).Scan(&res)
	return res, err
}

func (trg *PgTestTarget) Close(ctx context.Context) error {
	return trg.conn.Close(ctx)
}

func (trg *PgTestTarget) Conn() *pgx.Conn {
	return trg.conn
}
