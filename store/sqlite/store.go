// Package sqlite provides a sqlite-backed implementation of store.Store.
//
// Uses modernc.org/sqlite (pure Go) so no CGO is needed.
package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/geekgonecrazy/training-log/store"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Store struct {
	db *sql.DB
}

// Open opens (and creates if needed) the SQLite DB at the given path,
// applies any pending migrations, and returns a ready Store.
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("ensure db dir: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)&_pragma=busy_timeout(5000)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	// SQLite + WAL allows many concurrent readers and one writer. We keep a
	// modest pool — write contention is resolved by busy_timeout (set in DSN).
	db.SetMaxOpenConns(8)
	db.SetMaxIdleConns(4)
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Users() store.UserStore                 { return &userStore{db: s.db} }
func (s *Store) RefreshTokens() store.RefreshTokenStore { return &refreshTokenStore{db: s.db} }
func (s *Store) Machines() store.MachineStore           { return &machineStore{db: s.db} }
func (s *Store) Exercises() store.ExerciseStore         { return &exerciseStore{db: s.db} }
func (s *Store) Routines() store.RoutineStore           { return &routineStore{db: s.db} }
func (s *Store) Sessions() store.SessionStore           { return &sessionStore{db: s.db} }
func (s *Store) Reports() store.ReportStore             { return &reportStore{db: s.db} }

func (s *Store) migrate() error {
	ctx := context.Background()
	if _, err := s.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		filename TEXT PRIMARY KEY,
		applied_at INTEGER NOT NULL
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		var exists int
		if err := s.db.QueryRowContext(ctx,
			`SELECT 1 FROM schema_migrations WHERE filename = ?`, name).Scan(&exists); err == nil {
			continue
		}

		body, err := migrationsFS.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin tx for %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx, string(body)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations(filename, applied_at) VALUES(?, strftime('%s','now'))`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record %s: %w", name, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit %s: %w", name, err)
		}
	}
	return nil
}
