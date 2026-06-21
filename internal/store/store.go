// Package store handles persistence of termipost data as human-readable JSON
// files. Layout under the data directory:
//
//	config.json              app settings + active environment
//	collections/<id>.json    one collection with its requests and assertions
//	environments/<id>.json   one environment's variables
//	secrets.json             global secret values (gitignored)
//	.gitignore               excludes secrets.json
package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gbrlballesteros/termipost/internal/domain"
)

const (
	configFile      = "config.json"
	secretsFile     = "secrets.json"
	gitignoreFile   = ".gitignore"
	collectionsDir  = "collections"
	environmentsDir = "environments"
)

// Store reads and writes termipost data under a single directory.
type Store struct {
	Dir string
}

// Data is the full set of persisted state loaded from disk.
type Data struct {
	Config       domain.Config
	Collections  []domain.Collection
	Environments []domain.Environment
	Secrets      domain.Secrets
	// LoadErrors collects per-file problems (e.g. malformed JSON) encountered
	// during Load. They are surfaced to the user without aborting startup.
	LoadErrors []error
}

// DefaultDir resolves the data directory, preferring the OS config home
// (e.g. ~/.config/termipost) and falling back to ~/.termipost.
func DefaultDir() (string, error) {
	if cfg, err := os.UserConfigDir(); err == nil && cfg != "" {
		return filepath.Join(cfg, "termipost"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".termipost"), nil
}

// New returns a Store rooted at dir.
func New(dir string) *Store { return &Store{Dir: dir} }

// Init creates the directory tree and writes the .gitignore that keeps secrets
// out of version control. It is safe to call repeatedly.
func (s *Store) Init() error {
	for _, d := range []string{s.Dir, filepath.Join(s.Dir, collectionsDir), filepath.Join(s.Dir, environmentsDir)} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}
	return s.ensureGitignore()
}

func (s *Store) ensureGitignore() error {
	path := filepath.Join(s.Dir, gitignoreFile)
	want := secretsFile
	if data, err := os.ReadFile(path); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(line) == want {
				return nil // already excluded
			}
		}
		// Append without clobbering existing entries.
		content := string(data)
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += want + "\n"
		return os.WriteFile(path, []byte(content), 0o644)
	}
	return os.WriteFile(path, []byte(want+"\n"), 0o644)
}

// Load reads all data. Missing files are treated as empty. A malformed file is
// recorded in Data.LoadErrors and skipped; it is never overwritten.
func (s *Store) Load() (*Data, error) {
	d := &Data{Secrets: domain.Secrets{}}

	if err := readJSON(filepath.Join(s.Dir, configFile), &d.Config); err != nil && !errors.Is(err, os.ErrNotExist) {
		d.LoadErrors = append(d.LoadErrors, fmt.Errorf("%s: %w", configFile, err))
	}

	if err := readJSON(filepath.Join(s.Dir, secretsFile), &d.Secrets); err != nil && !errors.Is(err, os.ErrNotExist) {
		d.LoadErrors = append(d.LoadErrors, fmt.Errorf("%s: %w", secretsFile, err))
	}
	if d.Secrets == nil {
		d.Secrets = domain.Secrets{}
	}

	cols, errs := loadDir[domain.Collection](filepath.Join(s.Dir, collectionsDir))
	d.Collections = cols
	d.LoadErrors = append(d.LoadErrors, errs...)

	envs, errs := loadDir[domain.Environment](filepath.Join(s.Dir, environmentsDir))
	d.Environments = envs
	d.LoadErrors = append(d.LoadErrors, errs...)

	return d, nil
}

// loadDir reads every *.json file in dir into a slice of T, sorted by name.
// Malformed files become errors rather than failing the whole load.
func loadDir[T any](dir string) ([]T, []error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, []error{fmt.Errorf("read dir %s: %w", dir, err)}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	var out []T
	var loadErrs []error
	for _, name := range names {
		var v T
		if err := readJSON(filepath.Join(dir, name), &v); err != nil {
			loadErrs = append(loadErrs, fmt.Errorf("%s: %w", name, err))
			continue
		}
		out = append(out, v)
	}
	return out, loadErrs
}

// SaveConfig persists application settings.
func (s *Store) SaveConfig(c domain.Config) error {
	return writeJSON(filepath.Join(s.Dir, configFile), c)
}

// SaveSecrets persists the global secrets store.
func (s *Store) SaveSecrets(sec domain.Secrets) error {
	if sec == nil {
		sec = domain.Secrets{}
	}
	return writeJSON(filepath.Join(s.Dir, secretsFile), sec)
}

// SaveCollection persists a single collection (with its requests/assertions).
func (s *Store) SaveCollection(c domain.Collection) error {
	return writeJSON(filepath.Join(s.Dir, collectionsDir, c.ID+".json"), c)
}

// DeleteCollection removes a collection's file. Missing files are not an error.
func (s *Store) DeleteCollection(id string) error {
	return removeIfExists(filepath.Join(s.Dir, collectionsDir, id+".json"))
}

// SaveEnvironment persists a single environment.
func (s *Store) SaveEnvironment(e domain.Environment) error {
	return writeJSON(filepath.Join(s.Dir, environmentsDir, e.ID+".json"), e)
}

// DeleteEnvironment removes an environment's file. Missing files are not an error.
func (s *Store) DeleteEnvironment(id string) error {
	return removeIfExists(filepath.Join(s.Dir, environmentsDir, id+".json"))
}

func removeIfExists(path string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func readJSON(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil // empty file: treat as empty value
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

// writeJSON writes v as indented JSON atomically: it writes a temp file in the
// same directory and renames it into place so an interrupted write cannot
// corrupt the existing file.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	data = append(data, '\n')

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName) // no-op if the rename succeeded

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}
