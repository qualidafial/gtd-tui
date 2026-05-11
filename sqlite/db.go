package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	_ "modernc.org/sqlite"
)

var (
	ErrAlreadyInTx = errors.New("already in transaction")
)

//go:embed migrations/*.sql
var migrations embed.FS

type QueryContext interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// DB wraps a SQLite database connection.
type DB struct {
	db    QueryContext
	close func() error
}

// Open opens the SQLite database at the given path, creating it if necessary,
// and runs any pending migrations.
func Open(path string) (*DB, error) {
	conn, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	ctx := context.Background()

	if _, err := conn.ExecContext(ctx, `PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		conn.Close()
		return nil, fmt.Errorf("pragma: %w", err)
	}

	db := &DB{db: conn, close: conn.Close}
	if err := db.migrate(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (d *DB) Close() error {
	if d.close != nil {
		return d.close()
	}
	return nil
}

func (d *DB) RunTx(ctx context.Context, f func(context.Context, *DB) error) (err error) {
	db, ok := d.db.(interface{ Begin() (*sql.Tx, error) })
	if !ok {
		return ErrAlreadyInTx
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		rollbackErr := tx.Rollback()
		if rollbackErr == nil {
			return
		}
		if errors.Is(rollbackErr, sql.ErrTxDone) {
			return
		}
		err = errors.Join(err, rollbackErr)
	}()

	txDb := &DB{db: tx}
	if err := f(ctx, txDb); err != nil {
		return err
	}

	return tx.Commit()
}

func (d *DB) migrate(ctx context.Context) error {
	if _, err := d.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY)`); err != nil {
		return err
	}

	entries, err := fs.ReadDir(migrations, "migrations")
	if err != nil {
		return err
	}
	slices.SortFunc(entries, func(a, b fs.DirEntry) int {
		return strings.Compare(a.Name(), b.Name())
	})

	for _, entry := range entries {
		name := entry.Name()

		var exists bool
		err := d.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM migrations WHERE name = ?)`, name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if exists {
			continue
		}

		sql, err := fs.ReadFile(migrations, "migrations/"+name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		err = d.RunTx(ctx, func(ctx context.Context, db *DB) error {
			if _, err := db.db.ExecContext(ctx, string(sql)); err != nil {
				return fmt.Errorf("run migration %s: %w", name, err)
			}
			if _, err := db.db.ExecContext(ctx, `INSERT INTO migrations (name) VALUES (?)`, name); err != nil {
				return fmt.Errorf("record migration %s: %w", name, err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}
