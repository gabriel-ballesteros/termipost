package model

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/gabriel-ballesteros/termipost/internal/domain"
)

// openEnv navigates Collections -> Environments.
func openEnv(t *testing.T, m *Model) *envListScreen {
	t.Helper()
	send(t, m, keyMsg("e"))
	es, ok := m.top().(*envListScreen)
	if !ok {
		t.Fatalf("`e` should open environments, got %T", m.top())
	}
	return es
}

func TestEnvCreateRenameDelete(t *testing.T) {
	m := seededModel(t)
	openEnv(t, m)

	// Empty rejected.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("enter"))
	if !m.statusErr {
		t.Fatal("empty env name should error")
	}

	// Create "Dev".
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("Dev"))
	send(t, m, keyMsg("enter"))
	if len(m.app.environments) != 1 || m.app.environments[0].Name != "Dev" {
		t.Fatalf("env not created: %+v", m.app.environments)
	}

	// Duplicate rejected.
	send(t, m, keyMsg("n"))
	send(t, m, keyMsg("Dev"))
	send(t, m, keyMsg("enter"))
	if len(m.app.environments) != 1 {
		t.Fatalf("duplicate env added: %+v", m.app.environments)
	}

	// Rename Dev -> Prod.
	send(t, m, keyMsg("r"))
	send(t, m, tea.KeyMsg{Type: tea.KeyCtrlU})
	send(t, m, keyMsg("Prod"))
	send(t, m, keyMsg("enter"))
	if m.app.environments[0].Name != "Prod" {
		t.Fatalf("rename failed: %+v", m.app.environments)
	}

	// Delete it (confirm).
	send(t, m, keyMsg("d"))
	send(t, m, keyMsg("y"))
	if len(m.app.environments) != 0 {
		t.Fatalf("env not deleted: %+v", m.app.environments)
	}
}

func TestEnvSetActive(t *testing.T) {
	m := seededModel(t)
	es := openEnv(t, m)
	es.app.environments = []domain.Environment{{ID: "e1", Name: "Dev", Vars: map[string]string{}}}
	es.refresh()

	send(t, m, keyMsg("a")) // set active
	if m.app.cfg.ActiveEnvironmentID != "e1" {
		t.Fatalf("active env id = %q, want e1", m.app.cfg.ActiveEnvironmentID)
	}
	// Active marker appears in the list view.
	if !strings.Contains(m.View(), "active") {
		t.Fatalf("active marker missing:\n%s", m.View())
	}
}

func TestEnvEditVars(t *testing.T) {
	m := seededModel(t)
	es := openEnv(t, m)
	es.app.environments = []domain.Environment{{ID: "e1", Name: "Dev", Vars: map[string]string{}}}
	es.refresh()

	send(t, m, keyMsg("enter")) // open KV editor for vars
	if _, ok := m.top().(*kvEditorScreen); !ok {
		t.Fatalf("enter should open the variables KV editor, got %T", m.top())
	}
	send(t, m, keyMsg("a"))
	send(t, m, keyMsg("host: api.test"))
	send(t, m, keyMsg("enter"))
	send(t, m, keyMsg("esc")) // commit via onDone -> saveVars
	if got := m.app.environments[0].Vars["host"]; got != "api.test" {
		t.Fatalf("var not saved, host = %q (vars=%+v)", got, m.app.environments[0].Vars)
	}
}

func TestEnvOpensSecrets(t *testing.T) {
	m := seededModel(t)
	openEnv(t, m)
	send(t, m, keyMsg("s"))
	if _, ok := m.top().(*secretsScreen); !ok {
		t.Fatalf("`s` should open secrets, got %T", m.top())
	}
}

func TestMapKVRoundTrip(t *testing.T) {
	in := map[string]string{"b": "2", "a": "1"}
	kvs := mapToKVs(in)
	// Sorted by key.
	if len(kvs) != 2 || kvs[0].Key != "a" || kvs[1].Key != "b" {
		t.Fatalf("mapToKVs not sorted: %+v", kvs)
	}
	// Round-trip, dropping empty-key entries.
	kvs = append(kvs, domain.KV{Key: "", Value: "ignored"})
	out := kvsToMap(kvs)
	if len(out) != 2 || out["a"] != "1" || out["b"] != "2" {
		t.Fatalf("kvsToMap = %+v", out)
	}
}
