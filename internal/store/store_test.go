package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

func TestSaveLoadRoundTrip(t *testing.T) {
	s := New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}

	col := domain.Collection{ID: "c-1", Name: "Demo", Requests: []domain.Request{{
		ID: "r-1", Name: "Get", Method: domain.GET, URL: "https://example.com",
		Assertions: []domain.Assertion{{Kind: domain.AssertStatusCode, Op: domain.OpEquals, Expected: "200"}},
	}}}
	env := domain.Environment{ID: "e-1", Name: "local", Vars: map[string]string{"base": "x"}}
	cfg := domain.Config{ActiveEnvironmentID: "e-1"}
	sec := domain.Secrets{"token": "abc"}

	for _, op := range []struct {
		name string
		fn   func() error
	}{
		{"collection", func() error { return s.SaveCollection(col) }},
		{"environment", func() error { return s.SaveEnvironment(env) }},
		{"config", func() error { return s.SaveConfig(cfg) }},
		{"secrets", func() error { return s.SaveSecrets(sec) }},
	} {
		if err := op.fn(); err != nil {
			t.Fatalf("save %s: %v", op.name, err)
		}
	}

	d, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if len(d.LoadErrors) != 0 {
		t.Fatalf("unexpected load errors: %v", d.LoadErrors)
	}
	if len(d.Collections) != 1 || d.Collections[0].Name != "Demo" || len(d.Collections[0].Requests) != 1 {
		t.Fatalf("collection round-trip mismatch: %+v", d.Collections)
	}
	if len(d.Collections[0].Requests[0].Assertions) != 1 {
		t.Fatalf("nested assertion lost: %+v", d.Collections[0].Requests[0])
	}
	if len(d.Environments) != 1 || d.Environments[0].Vars["base"] != "x" {
		t.Fatalf("environment round-trip mismatch: %+v", d.Environments)
	}
	if d.Config.ActiveEnvironmentID != "e-1" {
		t.Fatalf("config round-trip mismatch: %+v", d.Config)
	}
	if d.Secrets["token"] != "abc" {
		t.Fatalf("secrets round-trip mismatch: %+v", d.Secrets)
	}
}

func TestLoadMissingFilesIsEmpty(t *testing.T) {
	s := New(t.TempDir())
	d, err := s.Load()
	if err != nil {
		t.Fatalf("load on empty dir: %v", err)
	}
	if len(d.Collections) != 0 || len(d.Environments) != 0 || len(d.LoadErrors) != 0 {
		t.Fatalf("expected empty data, got %+v", d)
	}
	if d.Secrets == nil {
		t.Fatal("secrets should be non-nil empty map")
	}
}

func TestLoadMalformedFileIsReportedNotOverwritten(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	bad := filepath.Join(dir, collectionsDir, "broken.json")
	const garbage = "{not valid json"
	if err := os.WriteFile(bad, []byte(garbage), 0o644); err != nil {
		t.Fatal(err)
	}

	d, err := s.Load()
	if err != nil {
		t.Fatalf("load should not fail on malformed file: %v", err)
	}
	if len(d.LoadErrors) == 0 {
		t.Fatal("expected a load error for malformed file")
	}
	// The bad file must be left untouched.
	got, _ := os.ReadFile(bad)
	if string(got) != garbage {
		t.Fatalf("malformed file was modified: %q", got)
	}
}

func TestInitGeneratesGitignore(t *testing.T) {
	dir := t.TempDir()
	s := New(dir)
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, gitignoreFile))
	if err != nil {
		t.Fatalf("read gitignore: %v", err)
	}
	if !strings.Contains(string(data), secretsFile) {
		t.Fatalf("gitignore missing %s: %q", secretsFile, data)
	}

	// Idempotent: a second Init must not duplicate the entry.
	if err := s.Init(); err != nil {
		t.Fatalf("re-init: %v", err)
	}
	data2, _ := os.ReadFile(filepath.Join(dir, gitignoreFile))
	if strings.Count(string(data2), secretsFile) != 1 {
		t.Fatalf("gitignore entry duplicated: %q", data2)
	}
}
