package bench

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/dangermike/associative_entity_redux/connections"
)

var _ TestTarget = new(MySQLTestTarget)

type MySQLTestTarget struct {
	name string
	conn *sql.DB
}

func NewMySQLTestTarget(ctx context.Context, name string, db string) (*MySQLTestTarget, error) {
	trg := MySQLTestTarget{name: name}
	var err error
	trg.conn, err = connections.MySqlConn(ctx, db)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql '%s': %w", name, err)
	}
	if err := trg.conn.PingContext(ctx); err != nil {
		_ = trg.Close(ctx)
		return nil, fmt.Errorf("failed to ping mysql '%s': %w", name, err)
	}
	return &trg, nil
}

func (trg *MySQLTestTarget) Kind() string {
	return "mysql"
}

func (trg *MySQLTestTarget) Name() string {
	return trg.name
}

func (trg *MySQLTestTarget) Baseline(ctx context.Context, personName string) (int, error) {
	const q = `SELECT count(distinct name) FROM people AS p WHERE p.name = ?`
	var res int
	err := trg.conn.QueryRowContext(ctx, q, personName).Scan(&res)
	return res, err
}

func (trg *MySQLTestTarget) P2C(ctx context.Context, personName string) (int, error) {
	const p2cQ = `SELECT count(distinct c.name)
	FROM
		people AS p INNER JOIN
		people_companies AS pc ON
			p.id = pc.person_id INNER JOIN
		companies AS c ON
			pc.company_id = c.id
	WHERE
		p.name = ?`
	var res int
	err := trg.conn.QueryRowContext(ctx, p2cQ, personName).Scan(&res)
	return res, err
}

func (trg *MySQLTestTarget) C2P(ctx context.Context, companyName string) (int, error) {
	const c2pQ = `SELECT count(distinct p.name)
	FROM
		people AS p INNER JOIN
		people_companies AS pc ON
			p.id = pc.person_id INNER JOIN
		companies AS c ON
			pc.company_id = c.id
	WHERE
		c.name = ?`
	var res int
	err := trg.conn.QueryRowContext(ctx, c2pQ, companyName).Scan(&res)
	return res, err
}

func (trg *MySQLTestTarget) Close(ctx context.Context) error {
	return trg.conn.Close()
}
