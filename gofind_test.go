package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
)

func TestJoinSearchRoot(t *testing.T) {
	sep := string(os.PathSeparator)
	tests := []struct {
		root string
		child string
		want string
	}{
		{root: ".", child: "child", want: "." + sep + "child"},
		{root: "." + sep + "base", child: "child", want: "." + sep + "base" + sep + "child"},
		{root: "base", child: "child", want: filepath.Join("base", "child")},
		{root: filepath.Join(sep, "tmp", "base"), child: "child", want: filepath.Join(sep, "tmp", "base", "child")},
	}

	for _, tt := range tests {
		got := joinSearchRoot(tt.root, tt.child)
		if got != tt.want {
			t.Fatalf("joinSearchRoot(%q, %q) = %q, want %q", tt.root, tt.child, got, tt.want)
		}
	}
}

func TestGoFindMatchesFindForRelativeRoot(t *testing.T) {
	sourceDir := sourceDir(t)
	bin := buildGoFind(t, sourceDir)

	fixtureDir := t.TempDir()
	mustMkdirAll(t, filepath.Join(fixtureDir, "alpha", "nested"))
	mustMkdirAll(t, filepath.Join(fixtureDir, "beta"))
	mustWriteFile(t, filepath.Join(fixtureDir, "root.go"))
	mustWriteFile(t, filepath.Join(fixtureDir, "alpha", "a.go"))
	mustWriteFile(t, filepath.Join(fixtureDir, "alpha", "nested", "b.go"))
	mustWriteFile(t, filepath.Join(fixtureDir, "beta", "c.go"))
	mustWriteFile(t, filepath.Join(fixtureDir, "beta", "ignore.txt"))

	got := sortedLines(t, runCmd(t, fixtureDir, bin, ".", "-name", "*.go"))
	want := sortedLines(t, runCmd(t, fixtureDir, "find", ".", "-name", "*.go"))

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("gofind output mismatch\n got: %v\nwant: %v", got, want)
	}
}

func TestGoFindMultipleRootsDoNotLeakExclusions(t *testing.T) {
	sourceDir := sourceDir(t)
	bin := buildGoFind(t, sourceDir)

	fixtureDir := t.TempDir()
	rootA := filepath.Join(fixtureDir, "rootA")
	rootB := filepath.Join(fixtureDir, "rootB")

	mustMkdirAll(t, filepath.Join(rootA, "skip"))
	mustMkdirAll(t, filepath.Join(rootB, "skip"))
	mustWriteFile(t, filepath.Join(rootA, "skip", "a.txt"))
	mustWriteFile(t, filepath.Join(rootB, "skip", "b.txt"))

	got := sortedLines(t, runCmd(t, fixtureDir, bin, rootA, rootB, "-name", "*.txt"))
	want := sortedLines(t, runCmd(t, fixtureDir, "find", rootA, rootB, "-name", "*.txt"))

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("multi-root output mismatch\n got: %v\nwant: %v", got, want)
	}
}

func TestGoFindPreservesStdoutAndExitCodeOnFindError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission-based traversal test is not portable to Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("permission-based traversal test is not meaningful as root")
	}

	sourceDir := sourceDir(t)
	bin := buildGoFind(t, sourceDir)

	fixtureDir := t.TempDir()
	mustMkdirAll(t, filepath.Join(fixtureDir, "ok"))
	mustMkdirAll(t, filepath.Join(fixtureDir, "deny", "nested"))
	mustWriteFile(t, filepath.Join(fixtureDir, "ok", "a.txt"))
	mustWriteFile(t, filepath.Join(fixtureDir, "deny", "nested", "b.txt"))

	denyDir := filepath.Join(fixtureDir, "deny")
	if err := os.Chmod(denyDir, 0); err != nil {
		t.Fatalf("chmod deny dir: %v", err)
	}
	defer func() {
		_ = os.Chmod(denyDir, 0o755)
	}()

	stdout, stderr, code := runCmdWithStatus(t, fixtureDir, bin, fixtureDir, "-name", "*.txt")
	if code == 0 {
		t.Fatalf("gofind exit code = 0, want non-zero")
	}

	lines := sortedLines(t, stdout)
	wantPath := filepath.Join(fixtureDir, "ok", "a.txt")
	if !reflect.DeepEqual(lines, []string{wantPath}) {
		t.Fatalf("gofind stdout mismatch\n got: %v\nwant: %v", lines, []string{wantPath})
	}
	if !strings.Contains(stderr, "Permission denied") {
		t.Fatalf("stderr missing permission error: %q", stderr)
	}
}

func sourceDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(file)
}

func buildGoFind(t *testing.T, sourceDir string) string {
	t.Helper()

	bin := filepath.Join(t.TempDir(), "gofind-test-bin")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", bin, "gofind.go")
	cmd.Dir = sourceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build gofind.go failed: %v\n%s", err, output)
	}
	return bin
}

func runCmd(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v failed: %v\n%s", name, args, err, output)
	}
	return string(output)
}

func runCmdWithStatus(t *testing.T, dir string, name string, args ...string) (string, string, int) {
	t.Helper()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.Output()

	var stderr string
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = string(exitErr.Stderr)
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("%s %v failed unexpectedly: %v", name, args, err)
		}
	}

	return string(out), stderr, exitCode
}

func sortedLines(t *testing.T, output string) []string {
	t.Helper()

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return nil
	}

	lines := strings.Split(trimmed, "\n")
	sort.Strings(lines)
	return lines
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
