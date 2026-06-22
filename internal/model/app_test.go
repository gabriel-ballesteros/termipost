package model

import (
	"testing"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/store"
)

func newTestApp(t *testing.T) *App {
	t.Helper()
	s := store.New(t.TempDir())
	if err := s.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	data, err := s.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	return NewApp(s, data)
}

func TestNewAppNilSecrets(t *testing.T) {
	a := NewApp(&store.Store{}, &store.Data{}) // Secrets nil
	if a.secrets == nil {
		t.Fatal("NewApp must initialize a non-nil secrets map")
	}
}

func TestActiveEnvAndResolver(t *testing.T) {
	a := newTestApp(t)
	a.environments = []domain.Environment{
		{ID: "e1", Name: "Dev", Vars: map[string]string{"host": "dev.local"}},
		{ID: "e2", Name: "Prod", Vars: map[string]string{"host": "prod.local"}},
	}
	a.secrets = domain.Secrets{"key": "abc"}

	if a.ActiveEnv() != nil {
		t.Fatal("ActiveEnv should be nil when no active id set")
	}
	// Resolver with no active env still resolves secrets.
	if got, _ := a.Resolver().Resolve("{{key}}"); got != "abc" {
		t.Errorf("Resolve secret = %q, want abc", got)
	}

	a.cfg.ActiveEnvironmentID = "e2"
	if e := a.ActiveEnv(); e == nil || e.Name != "Prod" {
		t.Fatalf("ActiveEnv = %+v, want Prod", e)
	}
	if got, _ := a.Resolver().Resolve("{{host}}"); got != "prod.local" {
		t.Errorf("Resolve env var = %q, want prod.local", got)
	}
}

func TestNameTaken(t *testing.T) {
	a := newTestApp(t)
	a.collections = []domain.Collection{{ID: "c1", Name: "API"}}
	a.environments = []domain.Environment{{ID: "e1", Name: "Dev"}}

	if !a.collectionNameTaken("API", "") {
		t.Error("collectionNameTaken should report API as taken")
	}
	if a.collectionNameTaken("API", "c1") {
		t.Error("collectionNameTaken should ignore the excepted id")
	}
	if a.collectionNameTaken("Other", "") {
		t.Error("collectionNameTaken false positive")
	}
	if !a.envNameTaken("Dev", "") {
		t.Error("envNameTaken should report Dev as taken")
	}
	if a.envNameTaken("Dev", "e1") {
		t.Error("envNameTaken should ignore the excepted id")
	}
}

func TestFindCollection(t *testing.T) {
	a := newTestApp(t)
	a.collections = []domain.Collection{{ID: "c1", Name: "A"}, {ID: "c2", Name: "B"}}
	if c := a.findCollection("c2"); c == nil || c.Name != "B" {
		t.Fatalf("findCollection(c2) = %+v", c)
	}
	if c := a.findCollection("nope"); c != nil {
		t.Fatalf("findCollection(nope) = %+v, want nil", c)
	}
}

func TestDeleteCollection(t *testing.T) {
	a := newTestApp(t)
	col := domain.Collection{ID: "c1", Name: "A"}
	if err := a.saveCollection(col); err != nil {
		t.Fatalf("save: %v", err)
	}
	a.collections = []domain.Collection{col, {ID: "c2", Name: "B"}}

	if err := a.deleteCollection("c1"); err != nil {
		t.Fatalf("deleteCollection: %v", err)
	}
	if a.findCollection("c1") != nil {
		t.Error("c1 still present after delete")
	}
	if a.findCollection("c2") == nil {
		t.Error("c2 wrongly removed")
	}
}

func TestDeleteEnvironmentClearsActive(t *testing.T) {
	a := newTestApp(t)
	env := domain.Environment{ID: "e1", Name: "Dev", Vars: map[string]string{}}
	if err := a.saveEnvironment(env); err != nil {
		t.Fatalf("save env: %v", err)
	}
	a.environments = []domain.Environment{env}
	a.cfg.ActiveEnvironmentID = "e1"
	if err := a.saveConfig(); err != nil {
		t.Fatalf("save config: %v", err)
	}

	if err := a.deleteEnvironment("e1"); err != nil {
		t.Fatalf("deleteEnvironment: %v", err)
	}
	if len(a.environments) != 0 {
		t.Errorf("environments not emptied: %+v", a.environments)
	}
	if a.cfg.ActiveEnvironmentID != "" {
		t.Errorf("active env id not cleared: %q", a.cfg.ActiveEnvironmentID)
	}
}

func TestRemoveByID(t *testing.T) {
	type item struct{ id string }
	idOf := func(i item) string { return i.id }
	in := []item{{"a"}, {"b"}, {"c"}}
	out := removeByID(in, "b", idOf)
	if len(out) != 2 || out[0].id != "a" || out[1].id != "c" {
		t.Fatalf("removeByID = %+v", out)
	}
	// Removing a missing id leaves the slice intact.
	if got := removeByID([]item{{"x"}}, "y", idOf); len(got) != 1 {
		t.Fatalf("removeByID(missing) = %+v", got)
	}
}
