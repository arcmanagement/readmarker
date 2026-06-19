package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/arcmanagement/readmarker/internal/store"
)

const (
	ExitOK      = 0
	ExitRuntime = 1
	ExitUsage   = 2
)

type App struct {
	version string
	commit  string
	date    string
	env     []string
}

func New(version, commit, date string) *App {
	return &App{version: version, commit: commit, date: date, env: os.Environ()}
}

func (a *App) WithEnv(env []string) *App {
	copy := *a
	copy.env = env
	return &copy
}

func (a *App) Run(args []string, stdout, stderr io.Writer) int {
	var dbPath string
	var showVersion bool

	flags := flag.NewFlagSet("readmarker", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&dbPath, "db", "", "path to the readmarker database")
	flags.BoolVar(&showVersion, "version", false, "print version information")
	flags.Usage = func() {
		_, _ = fmt.Fprint(stderr, usage())
	}

	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}

	if showVersion {
		_, _ = fmt.Fprintf(stdout, "readmarker %s (%s, %s)\n", a.version, a.commit, a.date)
		return ExitOK
	}

	if flags.NArg() == 0 {
		_, _ = fmt.Fprint(stderr, usage())
		return ExitUsage
	}

	command := flags.Arg(0)
	commandArgs := flags.Args()[1:]
	if command == "help" {
		_, _ = fmt.Fprint(stdout, usage())
		return ExitOK
	}

	resolvedDB, err := a.resolveDBPath(dbPath)
	if err != nil {
		return runtimeError(stderr, err)
	}

	ledger, err := store.Open(resolvedDB)
	if err != nil {
		return runtimeError(stderr, err)
	}
	defer func() {
		_ = ledger.Close()
	}()

	switch command {
	case "get":
		return runGet(ledger, commandArgs, stdout, stderr)
	case "advance":
		return runAdvance(ledger, commandArgs, stdout, stderr)
	case "set":
		return runSet(ledger, commandArgs, stdout, stderr)
	case "list":
		return runList(ledger, commandArgs, stdout, stderr)
	default:
		return usageError(stderr, "unknown command %q", command)
	}
}

func runGet(store *store.Store, args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		return usageError(stderr, "usage: readmarker get <source_key>")
	}

	cursor, _, err := store.Get(args[0])
	if err != nil {
		return runtimeError(stderr, err)
	}
	_, _ = fmt.Fprintf(stdout, "%d\n", cursor)
	return ExitOK
}

func runAdvance(store *store.Store, args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		return usageError(stderr, "usage: readmarker advance <source_key> <pos>")
	}

	next, err := parseCursor(args[1])
	if err != nil {
		return usageError(stderr, "%v", err)
	}

	cursor, err := store.Advance(args[0], next)
	if err != nil {
		return runtimeError(stderr, err)
	}
	_, _ = fmt.Fprintf(stdout, "%d\n", cursor)
	return ExitOK
}

func runSet(store *store.Store, args []string, stdout, stderr io.Writer) int {
	if len(args) != 2 {
		return usageError(stderr, "usage: readmarker set <source_key> <pos>")
	}

	cursor, err := parseCursor(args[1])
	if err != nil {
		return usageError(stderr, "%v", err)
	}

	if err := store.Set(args[0], cursor); err != nil {
		return runtimeError(stderr, err)
	}
	_, _ = fmt.Fprintf(stdout, "%d\n", cursor)
	return ExitOK
}

func runList(store *store.Store, args []string, stdout, stderr io.Writer) int {
	var outputJSON bool
	flags := flag.NewFlagSet("readmarker list", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.BoolVar(&outputJSON, "json", false, "emit JSON")
	flags.Usage = func() {
		_, _ = fmt.Fprint(stderr, "usage: readmarker list [--json]\n")
	}
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage
	}
	if flags.NArg() != 0 {
		return usageError(stderr, "usage: readmarker list [--json]")
	}

	cursors, err := store.List()
	if err != nil {
		return runtimeError(stderr, err)
	}

	if outputJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(cursors); err != nil {
			return runtimeError(stderr, err)
		}
		return ExitOK
	}

	for _, cursor := range cursors {
		_, _ = fmt.Fprintf(stdout, "%s\t%d\n", cursor.SourceKey, cursor.Cursor)
	}
	return ExitOK
}

func parseCursor(raw string) (uint64, error) {
	cursor, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("pos must be a non-negative base-10 integer: %q", raw)
	}
	return cursor, nil
}

func (a *App) resolveDBPath(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}
	if envValue := a.getenv("READMARKER_DB"); envValue != "" {
		return envValue, nil
	}

	if runtime.GOOS == "darwin" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("find home directory: %w", err)
		}
		return filepath.Join(home, "Library", "Application Support", "readmarker", "readmarker.db"), nil
	}

	if xdgDataHome := a.getenv("XDG_DATA_HOME"); xdgDataHome != "" {
		return filepath.Join(xdgDataHome, "readmarker", "readmarker.db"), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("find home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share", "readmarker", "readmarker.db"), nil
}

func (a *App) getenv(key string) string {
	prefix := key + "="
	for _, item := range a.env {
		if strings.HasPrefix(item, prefix) {
			return strings.TrimPrefix(item, prefix)
		}
	}
	return ""
}

func runtimeError(stderr io.Writer, err error) int {
	_, _ = fmt.Fprintf(stderr, "readmarker: %v\n", err)
	return ExitRuntime
}

func usageError(stderr io.Writer, format string, args ...any) int {
	_, _ = fmt.Fprintf(stderr, "readmarker: "+format+"\n", args...)
	return ExitUsage
}

func usage() string {
	return `readmarker tracks read cursors for source-agnostic agent workflows.

Usage:
  readmarker [--db <path>] get <source_key>
  readmarker [--db <path>] advance <source_key> <pos>
  readmarker [--db <path>] set <source_key> <pos>
  readmarker [--db <path>] list [--json]

Environment:
  READMARKER_DB  database path used when --db is omitted

Exit codes:
  0  success
  1  runtime or storage error
  2  usage error
`
}
