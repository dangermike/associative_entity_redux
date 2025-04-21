package connections

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
)

func PgConn(ctx context.Context, db string, schema string) (*pgx.Conn, error) {
	pgconnstr := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=disable", "postgres", "wont_tell", "localhost", 5432, db)
	pgconn, err := pgx.Connect(ctx, pgconnstr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres (db=%s): %w", db, err)
	}
	if err := pgconn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres (db=%s): %w", db, err)
	}

	if len(schema) > 0 && schema != "public" {
		if _, err := pgconn.Exec(ctx, fmt.Sprintf("SET search_path TO %s,public", schema)); err != nil {
			return nil, fmt.Errorf("failed to set postgres schema (db=%s): %w", db, err)
		}
	}

	return pgconn, err
}

func MySqlConn(ctx context.Context, db string) (*sql.DB, error) {
	mycfg := mysql.Config{
		User:   "root",
		Passwd: "wont_tell",
		Net:    "tcp",
		DBName: db,
	}

	myconn, err := sql.Open("mysql", mycfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql (db=%s): %w", db, err)
	}

	if err := myconn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping mysql (db=%s): %w", db, err)
	}
	return myconn, nil
}
