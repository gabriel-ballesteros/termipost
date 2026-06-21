package model

import (
	"github.com/gabriel-ballesteros/termipost/internal/domain"
	"github.com/gabriel-ballesteros/termipost/internal/store"
	"github.com/gabriel-ballesteros/termipost/internal/vars"
)

// App holds the shared, persisted application state plus the store used to
// persist it. Screens read and mutate this through the root Model.
type App struct {
	store        *store.Store
	cfg          domain.Config
	collections  []domain.Collection
	environments []domain.Environment
	secrets      domain.Secrets
}

// NewApp builds an App from loaded data.
func NewApp(s *store.Store, d *store.Data) *App {
	secrets := d.Secrets
	if secrets == nil {
		secrets = domain.Secrets{}
	}
	return &App{
		store:        s,
		cfg:          d.Config,
		collections:  d.Collections,
		environments: d.Environments,
		secrets:      secrets,
	}
}

// ActiveEnv returns the currently active environment, or nil if none.
func (a *App) ActiveEnv() *domain.Environment {
	for i := range a.environments {
		if a.environments[i].ID == a.cfg.ActiveEnvironmentID {
			return &a.environments[i]
		}
	}
	return nil
}

// Resolver builds a variable resolver from the active environment + secrets.
func (a *App) Resolver() *vars.Resolver {
	var envVars map[string]string
	if e := a.ActiveEnv(); e != nil {
		envVars = e.Vars
	}
	return vars.New(envVars, a.secrets)
}

// collectionNameTaken reports whether name is used by a collection other than
// the one with id exceptID.
func (a *App) collectionNameTaken(name, exceptID string) bool {
	for _, c := range a.collections {
		if c.ID != exceptID && c.Name == name {
			return true
		}
	}
	return false
}

func (a *App) envNameTaken(name, exceptID string) bool {
	for _, e := range a.environments {
		if e.ID != exceptID && e.Name == name {
			return true
		}
	}
	return false
}

// findCollection returns a pointer to the collection with id, or nil.
func (a *App) findCollection(id string) *domain.Collection {
	for i := range a.collections {
		if a.collections[i].ID == id {
			return &a.collections[i]
		}
	}
	return nil
}

func (a *App) saveCollection(c domain.Collection) error { return a.store.SaveCollection(c) }
func (a *App) saveConfig() error                        { return a.store.SaveConfig(a.cfg) }
func (a *App) saveSecrets() error                       { return a.store.SaveSecrets(a.secrets) }
func (a *App) saveEnvironment(e domain.Environment) error {
	return a.store.SaveEnvironment(e)
}

func (a *App) deleteCollection(id string) error {
	if err := a.store.DeleteCollection(id); err != nil {
		return err
	}
	a.collections = removeByID(a.collections, id, func(c domain.Collection) string { return c.ID })
	return nil
}

func (a *App) deleteEnvironment(id string) error {
	if err := a.store.DeleteEnvironment(id); err != nil {
		return err
	}
	a.environments = removeByID(a.environments, id, func(e domain.Environment) string { return e.ID })
	if a.cfg.ActiveEnvironmentID == id {
		a.cfg.ActiveEnvironmentID = ""
		_ = a.saveConfig()
	}
	return nil
}

func removeByID[T any](items []T, id string, idOf func(T) string) []T {
	out := items[:0]
	for _, it := range items {
		if idOf(it) != id {
			out = append(out, it)
		}
	}
	return out
}
