package load

import (
	"bufio"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"time"

	"github.com/dangermike/associative_entity_redux/connections"
	"github.com/dangermike/associative_entity_redux/logging"
	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

//go:embed data/words.zst
var wordsfs embed.FS

//go:embed schemata/*
var schemata embed.FS

var words = readWords()

func Command() *cobra.Command {
	return &cobra.Command{
		Use:  "load",
		RunE: runLoad,
	}
}

func runLoad(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	log := logging.FromContext(ctx)

	start := time.Now()
	pgConns, err := CreatePostgres(ctx)
	if err != nil {
		return err
	}
	for _, c := range pgConns {
		defer c.Close(ctx)
	}

	myConns, err := CreateMySQL(ctx)
	if err != nil {
		return err
	}
	for _, c := range myConns {
		defer c.Close()
	}
	log.Info("created tables", zap.Duration("duration", time.Since(start)))

	start = time.Now()

	if err := Fill(ctx, pgConns, myConns); err != nil {
		return err
	}

	log.Info("filled tables", zap.Duration("duration", time.Since(start)))

	return nil
}

func CreatePostgres(ctx context.Context) ([]*pgx.Conn, error) {
	log := logging.FromContext(ctx).With(zap.String("kind", "postgres"))
	var pgConns []*pgx.Conn

	pgroot, err := connections.PgConn(ctx, "postgres", "public")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres %s: %w", "public", err)
	}
	if err = fs.WalkDir(schemata, "schemata/postgresql", func(path string, d fs.DirEntry, err error) error {
		name := strings.SplitN(d.Name(), ".", 2)[0]
		log := log.With(zap.String("name", name))
		if !d.Type().IsRegular() {
			return nil
		}
		if err != nil {
			return err
		}
		start := time.Now()
		f, err := schemata.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		scn := bufio.NewScanner(f)
		scn.Split(ScanQueries)
		var queryno int
		for scn.Scan() {
			q := strings.Trim(scn.Text(), " \t\n")
			if len(q) == 0 {
				continue
			}
			if _, err := pgroot.Exec(ctx, q); err != nil {
				return fmt.Errorf("query error from '%s #%d': %w\n%s", path, queryno, err, scn.Text())
			}
			queryno++
		}

		log.Debug("created", zap.Int("queries", queryno), zap.Duration("duration", time.Since(start)))

		newconn, err := connections.PgConn(ctx, "postgres", name)
		if err != nil {
			return fmt.Errorf("failed to connect to postgres %s: %w", name, err)
		}
		pgConns = append(pgConns, newconn)

		return nil
	}); err != nil {
		return nil, err
	}

	return pgConns, nil
}

func CreateMySQL(ctx context.Context) ([]*sql.DB, error) {
	log := logging.FromContext(ctx).With(zap.String("kind", "mysql"))
	var sqlConns []*sql.DB

	sqlroot, err := connections.MySqlConn(ctx, "mysql")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql %s: %w", "mysql", err)
	}
	if err = fs.WalkDir(schemata, "schemata/mysql", func(path string, d fs.DirEntry, err error) error {
		name := strings.SplitN(d.Name(), ".", 2)[0]
		log := log.With(zap.String("name", name))
		if !d.Type().IsRegular() {
			return nil
		}
		if err != nil {
			return err
		}
		start := time.Now()
		f, err := schemata.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		scn := bufio.NewScanner(f)
		scn.Split(ScanQueries)
		var queryno int
		for scn.Scan() {
			q := strings.Trim(scn.Text(), " \t\n")
			if len(q) == 0 {
				continue
			}
			if _, err := sqlroot.ExecContext(ctx, q); err != nil {
				return fmt.Errorf("query error from '%s #%d': %w\n%s", path, queryno, err, scn.Text())
			}
			queryno++
		}

		log.Debug("created", zap.Int("queries", queryno), zap.Duration("duration", time.Since(start)))

		newconn, err := connections.MySqlConn(ctx, name)
		if err != nil {
			return fmt.Errorf("failed to connect to mysql %s: %w", name, err)
		}
		sqlConns = append(sqlConns, newconn)

		return nil
	}); err != nil {
		return nil, err
	}

	return sqlConns, nil
}

func Fill(ctx context.Context, pgConns []*pgx.Conn, myConns []*sql.DB) error {
	const numNames, numAssoc = 1_000_000, 5_000_000
	const batchSize = 1000
	log := logging.FromContext(ctx)

	var sb strings.Builder
	for _, tbl := range []string{"people", "companies"} {
		log := log.With(zap.String("table", tbl))
		start := time.Now()
		var cnt int

		iinfo := InsertNameSpec{Table: tbl, Names: make([]string, batchSize)}
		log.Info("populating")
		for iinfo.StartIx = 0; iinfo.StartIx < numNames; iinfo.StartIx += batchSize {
			sb.Reset()
			fillWords(iinfo.Names)
			cnt += len(iinfo.Names)
			if err := templates.ExecuteTemplate(&sb, "insert_names.sql.gtmpl", iinfo); err != nil {
				return fmt.Errorf("failed to format query: %w", err)
			}
			query := sb.String()
			var errs errgroup.Group
			for _, pgc := range pgConns {
				errs.Go(func() error {
					if _, err := pgc.Exec(ctx, query); err != nil {
						return fmt.Errorf("failed to insert into postgres: %w\n%s", err, query)
					}
					return nil
				})
			}
			for _, sqlc := range myConns {
				errs.Go(func() error {
					if _, err := sqlc.ExecContext(ctx, query); err != nil {
						return fmt.Errorf("failed to insert into mysql: %w\n%s", err, query)
					}
					return nil
				})
			}
			if err := errs.Wait(); err != nil {
				return err
			}
		}
		log.Info("populated", zap.Int("count", cnt), zap.Duration("duration", time.Since(start)))
	}

	{
		log := log.With(zap.String("table", "people_companies"))
		start := time.Now()
		var cnt int
		j := joiner(numNames, numNames)
		ainfo := InsertAssocSpec{
			"people_companies",
			make([][2]int, batchSize),
		}

		log.Info("populating")
		for i := 0; i < numAssoc; i += batchSize {
			cnt += batchSize
			j(ainfo.Assocs)
			sb.Reset()
			if err := templates.ExecuteTemplate(&sb, "insert_assocs.sql.gtmpl", ainfo); err != nil {
				return fmt.Errorf("failed to format query: %w", err)
			}

			query := sb.String()
			var errs errgroup.Group
			for _, pgc := range pgConns {
				errs.Go(func() error {
					if _, err := pgc.Exec(ctx, query); err != nil {
						return fmt.Errorf("failed to insert into postgres: %w\n%s", err, query)
					}
					return nil
				})
			}
			for _, sqlc := range myConns {
				errs.Go(func() error {
					if _, err := sqlc.ExecContext(ctx, query); err != nil {
						return fmt.Errorf("failed to insert into mysql: %w\n%s", err, query)
					}
					return nil
				})
			}
			if err := errs.Wait(); err != nil {
				return err
			}
		}
		log.Info("populated", zap.Int("count", cnt), zap.Duration("duration", time.Since(start)))
	}

	return nil
}
