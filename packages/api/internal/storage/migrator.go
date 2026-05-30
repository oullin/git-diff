package storage

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	migratesqlite3 "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migrator struct {
	db *sql.DB
}

func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

func (m *Migrator) Run(ctx context.Context) error {
	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if err := mg.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}

	return nil
}

func (m *Migrator) newMigrate() (*migrate.Migrate, error) {
	src, err := iofs.New(migrationsFS, "migrations")

	if err != nil {
		return nil, fmt.Errorf("load migration source: %w", err)
	}

	drv, err := migratesqlite3.WithInstance(m.db, &migratesqlite3.Config{})

	if err != nil {
		return nil, fmt.Errorf("init sqlite3 migration driver: %w", err)
	}

	mg, err := migrate.NewWithInstance("iofs", src, "sqlite3", drv)

	if err != nil {
		return nil, fmt.Errorf("init migrator: %w", err)
	}

	// Callers must not invoke mg.Close(): the sqlite3 migrate driver wraps the
	// caller-owned *sql.DB and Close() would shut it down underneath us.
	return mg, nil
}

// Version reports the current schema version and whether the DB is in a dirty
// state. Returns version == 0 and ok == false when no migrations have ever run.
func (m *Migrator) Version() (version uint, dirty bool, ok bool, err error) {
	mg, err := m.newMigrate()

	if err != nil {
		return 0, false, false, err
	}

	v, d, err := mg.Version()

	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, false, nil
	}

	if err != nil {
		return 0, false, false, fmt.Errorf("read migration version: %w", err)
	}

	return v, d, true, nil
}

// Down rolls back N migration steps. Pass 0 to roll back everything.
func (m *Migrator) Down(steps int) error {
	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if steps <= 0 {
		if err := mg.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate down: %w", err)
		}

		return nil
	}

	if err := mg.Steps(-steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down %d: %w", steps, err)
	}

	return nil
}

// Force marks the DB at version v with dirty=false. Used to recover from a
// failed migration after the operator has manually inspected the state.
func (m *Migrator) Force(version int) error {
	mg, err := m.newMigrate()

	if err != nil {
		return err
	}

	if err := mg.Force(version); err != nil {
		return fmt.Errorf("force version %d: %w", version, err)
	}

	return nil
}

func (m *Migrator) tableExists(ctx context.Context, name string) (bool, error) {
	var got string

	err := m.db.QueryRowContext(ctx,
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?",
		name,
	).Scan(&got)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		return false, fmt.Errorf("query sqlite_master for %s: %w", name, err)
	}

	return true, nil
}
