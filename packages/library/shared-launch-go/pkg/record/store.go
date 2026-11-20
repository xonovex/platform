package record

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Files is the port the store reaches the disk through.
type Files interface {
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	Stat(path string) (fs.FileInfo, error)
	MkdirAll(path string, perm fs.FileMode) error
	WriteFile(path string, data []byte, perm fs.FileMode) error
}

// DefaultDir returns the directory of the run records of the tool app:
// app/runs below the XDG state directory when it is absolute, else below
// ~/.local/state.
func DefaultDir(getenv func(string) string, home, app string) string {
	state := getenv("XDG_STATE_HOME")
	if !path.IsAbs(state) {
		state = path.Join(home, ".local", "state")
	}
	return path.Join(state, app, "runs")
}

var idPattern = regexp.MustCompile(`^[0-9]{8}T[0-9]{6}Z(-[0-9]+)?$`)

// ValidID reports whether id has the form Write assigns, so that an id read from
// the command line cannot name a path.
func ValidID(id string) bool { return idPattern.MatchString(id) }

const (
	extension = ".json"
	// maxSuffix bounds the search for a free id among runs that started in one second.
	maxSuffix = 1000
)

// Store is the directory of run records.
type Store struct {
	files Files
	dir   string
}

// NewStore returns the store at dir, which must be a clean absolute path.
func NewStore(files Files, dir string) (Store, error) {
	if !path.IsAbs(dir) || path.Clean(dir) != dir {
		return Store{}, fmt.Errorf("the run record directory %q is not a clean absolute path", dir)
	}
	return Store{files: files, dir: dir}, nil
}

// Dir returns the directory of the store.
func (s Store) Dir() string { return s.dir }

// Write assigns the record an id from its start time, writes it as a private file
// and returns the id. Two runs that start in one second get distinct ids.
func (s Store) Write(r Record) (string, error) {
	if err := s.files.MkdirAll(s.dir, 0o700); err != nil {
		return "", fmt.Errorf("create the run record directory: %w", err)
	}
	id, err := s.freeID(r.StartedAt)
	if err != nil {
		return "", err
	}
	r.ID = id
	data, err := Marshal(r)
	if err != nil {
		return "", err
	}
	if err := s.files.WriteFile(s.file(id), data, 0o600); err != nil {
		return "", fmt.Errorf("write the run record: %w", err)
	}
	return id, nil
}

// freeID returns the first id for the instant that no record uses.
func (s Store) freeID(at time.Time) (string, error) {
	base := at.UTC().Format("20060102T150405Z")
	for suffix := 1; suffix <= maxSuffix; suffix++ {
		id := base
		if suffix > 1 {
			id = fmt.Sprintf("%s-%d", base, suffix)
		}
		if _, err := s.files.Stat(s.file(id)); errors.Is(err, fs.ErrNotExist) {
			return id, nil
		} else if err != nil {
			return "", fmt.Errorf("check the run record %s: %w", id, err)
		}
	}
	return "", fmt.Errorf("no free run record id for %s", base)
}

func (s Store) file(id string) string { return path.Join(s.dir, id+extension) }

// Read returns the record with the id.
func (s Store) Read(id string) (Record, error) {
	if !ValidID(id) {
		return Record{}, fmt.Errorf("%q is not a run record id", id)
	}
	data, err := s.files.ReadFile(s.file(id))
	if err != nil {
		return Record{}, fmt.Errorf("read the run record %s: %w", id, err)
	}
	r, err := Unmarshal(data)
	if err != nil {
		return Record{}, fmt.Errorf("run record %s: %w", id, err)
	}
	return r, nil
}

// List returns the records, newest first. A store that does not exist yet has none.
func (s Store) List() ([]Record, error) {
	entries, err := s.files.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list the run records: %w", err)
	}
	var ids []string
	for _, entry := range entries {
		id, ok := strings.CutSuffix(entry.Name(), extension)
		if ok && !entry.IsDir() && ValidID(id) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return newer(ids[i], ids[j]) })
	records := make([]Record, 0, len(ids))
	for _, id := range ids {
		r, err := s.Read(id)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

// newer orders ids by their start second, then by their suffix, newest first.
func newer(a, b string) bool {
	baseA, suffixA, _ := strings.Cut(a, "-")
	baseB, suffixB, _ := strings.Cut(b, "-")
	if baseA != baseB {
		return baseA > baseB
	}
	return len(suffixA) > len(suffixB) || len(suffixA) == len(suffixB) && suffixA > suffixB
}
