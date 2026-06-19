package cli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, dbPath string, args ...string) (int, string, string) {
	t.Helper()

	app := New("test", "abc123", "2026-06-20").WithEnv(nil)
	fullArgs := append([]string{"--db", dbPath}, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := app.Run(fullArgs, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestCLICommands(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "readmarker.db")

	code, stdout, stderr := runCLI(t, dbPath, "get", "slack:workspace:channel")
	if code != ExitOK {
		t.Fatalf("get exit = %d, stderr = %q", code, stderr)
	}
	if stdout != "0\n" {
		t.Fatalf("get stdout = %q, want %q", stdout, "0\n")
	}

	code, stdout, stderr = runCLI(t, dbPath, "advance", "slack:workspace:channel", "10")
	if code != ExitOK {
		t.Fatalf("advance exit = %d, stderr = %q", code, stderr)
	}
	if stdout != "10\n" {
		t.Fatalf("advance stdout = %q, want %q", stdout, "10\n")
	}

	code, stdout, stderr = runCLI(t, dbPath, "advance", "slack:workspace:channel", "4")
	if code != ExitOK {
		t.Fatalf("advance rewind exit = %d, stderr = %q", code, stderr)
	}
	if stdout != "10\n" {
		t.Fatalf("advance rewind stdout = %q, want %q", stdout, "10\n")
	}

	code, stdout, stderr = runCLI(t, dbPath, "set", "slack:workspace:channel", "4")
	if code != ExitOK {
		t.Fatalf("set exit = %d, stderr = %q", code, stderr)
	}
	if stdout != "4\n" {
		t.Fatalf("set stdout = %q, want %q", stdout, "4\n")
	}

	code, stdout, stderr = runCLI(t, dbPath, "list")
	if code != ExitOK {
		t.Fatalf("list exit = %d, stderr = %q", code, stderr)
	}
	if stdout != "slack:workspace:channel\t4\n" {
		t.Fatalf("list stdout = %q", stdout)
	}
}

func TestCLIListJSON(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "readmarker.db")

	code, _, stderr := runCLI(t, dbPath, "set", "github:owner/repo#1", "123")
	if code != ExitOK {
		t.Fatalf("set exit = %d, stderr = %q", code, stderr)
	}

	code, stdout, stderr := runCLI(t, dbPath, "list", "--json")
	if code != ExitOK {
		t.Fatalf("list --json exit = %d, stderr = %q", code, stderr)
	}
	if !strings.Contains(stdout, `"source_key": "github:owner/repo#1"`) {
		t.Fatalf("list --json stdout = %q", stdout)
	}
	if !strings.Contains(stdout, `"cursor": 123`) {
		t.Fatalf("list --json stdout = %q", stdout)
	}
}

func TestCLIUsageErrors(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "readmarker.db")

	code, _, stderr := runCLI(t, dbPath, "wat")
	if code != ExitUsage {
		t.Fatalf("unknown command exit = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(stderr, `unknown command "wat"`) {
		t.Fatalf("unknown command stderr = %q", stderr)
	}

	code, _, stderr = runCLI(t, dbPath, "advance", "source", "-1")
	if code != ExitUsage {
		t.Fatalf("invalid pos exit = %d, want %d", code, ExitUsage)
	}
	if !strings.Contains(stderr, "non-negative base-10 integer") {
		t.Fatalf("invalid pos stderr = %q", stderr)
	}
}
