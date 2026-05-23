package migratex

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/oullin/git-diff/internal/storage"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Home   string
	Stderr io.Writer
	Stdout io.Writer
}

// Run dispatches `git-diff migrate <subcommand> [args...]`.
func Run(args []string, cfg Config) int {
	if cfg.Stdout == nil {
		cfg.Stdout = os.Stdout
	}

	if cfg.Stderr == nil {
		cfg.Stderr = os.Stderr
	}

	if len(args) == 0 {
		usage(cfg.Stderr)

		return 2
	}

	sub := args[0]
	rest := args[1:]

	switch sub {
	case "up":
		return runUp(rest, cfg)
	case "down":
		return runDown(rest, cfg)
	case "version":
		return runVersion(rest, cfg)
	case "force":
		return runForce(rest, cfg)
	case "help", "-h", "--help":
		usage(cfg.Stdout)

		return 0
	default:
		fmt.Fprintf(cfg.Stderr, "unknown migrate subcommand %q\n\n", sub)
		usage(cfg.Stderr)

		return 2
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: git-diff migrate <subcommand> [flags]")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "subcommands:")
	fmt.Fprintln(w, "  up                 apply all pending migrations")
	fmt.Fprintln(w, "  down [N]           roll back N steps (default: all)")
	fmt.Fprintln(w, "  version            print current version and dirty flag")
	fmt.Fprintln(w, "  force <version>    mark DB at <version> with dirty=false")
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "flags (apply to up/down/version/force):")
	fmt.Fprintln(w, "  --db <path>        override database path")
}

func runUp(args []string, cfg Config) int {
	fs := flag.NewFlagSet("migrate up", flag.ContinueOnError)
	dbPath := fs.String("db", "", "database path (default: standard app path)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	db, exit := openDB(cfg, *dbPath)

	if exit != 0 {
		return exit
	}

	defer db.Close()

	if err := storage.NewMigrator(db).Run(context.Background()); err != nil {
		fmt.Fprintf(cfg.Stderr, "migrate up: %v\n", err)

		return 1
	}

	fmt.Fprintln(cfg.Stdout, "migrations up to date")

	return 0
}

func runDown(args []string, cfg Config) int {
	fs := flag.NewFlagSet("migrate down", flag.ContinueOnError)
	dbPath := fs.String("db", "", "database path (default: standard app path)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	steps := 0
	rest := fs.Args()

	if len(rest) > 0 {
		n, err := strconv.Atoi(rest[0])

		if err != nil || n < 1 {
			fmt.Fprintf(cfg.Stderr, "migrate down: invalid step count %q\n", rest[0])

			return 2
		}

		steps = n
	}

	db, exit := openDB(cfg, *dbPath)

	if exit != 0 {
		return exit
	}

	defer db.Close()

	if err := storage.NewMigrator(db).Down(steps); err != nil {
		fmt.Fprintf(cfg.Stderr, "migrate down: %v\n", err)

		return 1
	}

	if steps == 0 {
		fmt.Fprintln(cfg.Stdout, "rolled back all migrations")
	} else {
		fmt.Fprintf(cfg.Stdout, "rolled back %d migration(s)\n", steps)
	}

	return 0
}

func runVersion(args []string, cfg Config) int {
	fs := flag.NewFlagSet("migrate version", flag.ContinueOnError)
	dbPath := fs.String("db", "", "database path (default: standard app path)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	db, exit := openDB(cfg, *dbPath)

	if exit != 0 {
		return exit
	}

	defer db.Close()

	v, dirty, ok, err := storage.NewMigrator(db).Version()

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "migrate version: %v\n", err)

		return 1
	}

	if !ok {
		fmt.Fprintln(cfg.Stdout, "no migrations applied")

		return 0
	}

	fmt.Fprintf(cfg.Stdout, "version=%d dirty=%t\n", v, dirty)

	return 0
}

func runForce(args []string, cfg Config) int {
	fs := flag.NewFlagSet("migrate force", flag.ContinueOnError)
	dbPath := fs.String("db", "", "database path (default: standard app path)")

	if err := fs.Parse(args); err != nil {
		return 2
	}

	rest := fs.Args()

	if len(rest) != 1 {
		fmt.Fprintln(cfg.Stderr, "migrate force: expected exactly one version argument")

		return 2
	}

	version, err := strconv.Atoi(rest[0])

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "migrate force: invalid version %q\n", rest[0])

		return 2
	}

	db, exit := openDB(cfg, *dbPath)

	if exit != 0 {
		return exit
	}

	defer db.Close()

	if err := storage.NewMigrator(db).Force(version); err != nil {
		fmt.Fprintf(cfg.Stderr, "migrate force: %v\n", err)

		return 1
	}

	fmt.Fprintf(cfg.Stdout, "forced version %d (dirty=false)\n", version)

	return 0
}

func openDB(cfg Config, override string) (*sql.DB, int) {
	path := override

	if path == "" {
		path = storage.DefaultPath(cfg.Home)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		fmt.Fprintf(cfg.Stderr, "create database directory: %v\n", err)

		return nil, 1
	}

	db, err := sql.Open("sqlite3", dsn(path))

	if err != nil {
		fmt.Fprintf(cfg.Stderr, "open sqlite database: %v\n", err)

		return nil, 1
	}

	return db, 0
}

func dsn(path string) string {
	u := url.URL{Scheme: "file", Path: path}
	q := u.Query()
	q.Add("_foreign_keys", "on")
	u.RawQuery = q.Encode()

	return u.String()
}
