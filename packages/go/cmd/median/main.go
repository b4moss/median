// Package main is the median CLI (migrate dump).
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/b4moss/median/packages/go/db"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run implements CLI for tests. Returns process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, usage())
		return 2
	}
	switch args[0] {
	case "migrate":
		return runMigrate(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprintln(stderr, usage())
		return 0
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n%s\n", args[0], usage())
		return 2
	}
}

func runMigrate(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "migrate: missing subcommand (want dump)\n\n"+usage())
		return 2
	}
	switch args[0] {
	case "dump":
		return runMigrateDump(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "migrate: unknown subcommand %q (want dump)\n\n%s\n", args[0], usage())
		return 2
	}
}

func runMigrateDump(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("migrate dump", flag.ContinueOnError)
	fs.SetOutput(stderr)
	dialect := fs.String("dialect", "", "goose dialect: mysql|postgres|sqlite3 (required)")
	idStrategy := fs.String("id-strategy", db.IDAutoIncrement, "id strategy: auto_increment|uuid_v4|uuid_v7|ulid")
	table := fs.String("table", "", "table name (default media)")
	fs.Usage = func() {
		fmt.Fprintln(stderr, usage())
	}
	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if strings.TrimSpace(*dialect) == "" {
		fmt.Fprintln(stderr, "migrate dump: --dialect is required\n\n"+usage())
		return 2
	}

	sql, err := db.CreateMediaSQL(*dialect, *idStrategy, *table)
	if err != nil {
		fmt.Fprintf(stderr, "migrate dump: %v\n", err)
		return 1
	}
	sql = strings.TrimSpace(sql)
	if _, err := fmt.Fprintln(stdout, sql); err != nil {
		fmt.Fprintf(stderr, "migrate dump: write stdout: %v\n", err)
		return 1
	}
	return 0
}

func usage() string {
	return strings.TrimSpace(`
median — media library CLI

Usage:
  median migrate dump --dialect <mysql|postgres|sqlite3> [--id-strategy <strategy>] [--table <name>]

Options:
  --dialect       Required. mysql | postgres | sqlite3
  --id-strategy   Default auto_increment. auto_increment | uuid_v4 | uuid_v7 | ulid
  --table         Default media

SQL is written to stdout only. Errors and usage go to stderr.
`)
}
