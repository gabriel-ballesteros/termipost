package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

func TestDefaultDir(t *testing.T) {
	dir, err := DefaultDir()
	if err != nil {
		t.Fatalf("DefaultDir: %v", err)
	}
	if !strings.HasSuffix(filepath.ToSlash(dir), "termipost") {
		t.Fatalf("DefaultDir = %q, want a termipost path", dir)
	}
}

func TestDeleteCollection(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	col := domain.Collection{ID: "c1", Name: "A"}
	if err := s.SaveCollection(col); err != nil {
		t.Fatalf("save: %v", err)
	}
	path := filepath.Join(s.Dir, collectionsDir, "c1.json")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("collection file missing after save: %v", err)
	}
	if err := s.DeleteCollection("c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("collection file should be gone, stat err = %v", err)
	}
	// Deleting a missing collection is not an error.
	if err := s.DeleteCollection("c1"); err != nil {
		t.Fatalf("deleting missing collection should be a no-op, got %v", err)
	}
}

func TestDeleteEnvironment(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := s.SaveEnvironment(domain.Environment{ID: "e1", Name: "Dev"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	if err := s.DeleteEnvironment("e1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := s.DeleteEnvironment("e1"); err != nil {
		t.Fatalf("deleting missing environment should be a no-op, got %v", err)
	}
}

func TestSaveSecretsNilBecomesEmpty(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := s.SaveSecrets(nil); err != nil {
		t.Fatalf("save nil secrets: %v", err)
	}
	d, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if d.Secrets == nil || len(d.Secrets) != 0 {
		t.Fatalf("nil secrets should persist as empty map, got %+v", d.Secrets)
	}
}

func TestLoadConfigPresent(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	if err := s.SaveConfig(domain.Config{ActiveEnvironmentID: "e9"}); err != nil {
		t.Fatalf("save config: %v", err)
	}
	d, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if d.Config.ActiveEnvironmentID != "e9" {
		t.Fatalf("config not loaded, got %+v", d.Config)
	}
}

func TestEnsureGitignoreAppends(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, gitignoreFile)
	// Pre-existing .gitignore without a trailing newline and without our entry.
	if err := os.WriteFile(gi, []byte("*.log"), 0o644); err != nil {
		t.Fatalf("seed gitignore: %v", err)
	}
	s := New(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	data, err := os.ReadFile(gi)
	if err != nil {
		t.Fatalf("read gitignore: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "*.log") {
		t.Fatalf("existing entry clobbered: %q", content)
	}
	if !strings.Contains(content, secretsFile) {
		t.Fatalf("secrets entry not appended: %q", content)
	}

	// Re-running Init must not duplicate the entry.
	if err := s.Init(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	data, _ = os.ReadFile(gi)
	if strings.Count(string(data), secretsFile) != 1 {
		t.Fatalf("secrets entry duplicated: %q", string(data))
	}
}

func TestLoadReportsDirReadError(t *testing.T) {
	dir := t.TempDir()
	// Put a regular file where the collections directory is expected, so
	// os.ReadDir fails with a non-"not exist" error and Load records it.
	if err := os.WriteFile(filepath.Join(dir, collectionsDir), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	d, err := New(dir).Load()
	if err != nil {
		t.Fatalf("Load should not hard-fail: %v", err)
	}
	if len(d.LoadErrors) == 0 {
		t.Fatal("expected a LoadError for the unreadable collections dir")
	}
}

func TestReadJSONEmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.json")
	if err := os.WriteFile(path, []byte("   \n"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	var cfg domain.Config
	if err := readJSON(path, &cfg); err != nil {
		t.Fatalf("empty file should read as empty value, got %v", err)
	}
}

func TestWriteJSONMarshalError(t *testing.T) {
	// channels cannot be marshalled to JSON.
	err := writeJSON(filepath.Join(t.TempDir(), "x.json"), make(chan int))
	if err == nil {
		t.Fatal("expected marshal error for an unmarshallable value")
	}
}

func TestRemoveIfExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := removeIfExists(path); err != nil {
		t.Fatalf("remove existing: %v", err)
	}
	if err := removeIfExists(path); err != nil {
		t.Fatalf("remove missing should be no-op, got %v", err)
	}
}
