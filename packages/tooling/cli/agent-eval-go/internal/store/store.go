// Package store keeps eval results on disk as one checkpointed file per run, so
// an interrupted run resumes at the first case it lacks. The run id names the
// protocol, the subject and a digest of the inputs, so a resumer finds only a
// run whose inputs have not changed.
package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"path"
	"regexp"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// ResultsFile is the name of the checkpoint file inside a run directory.
const ResultsFile = "results.json"

var unsafeName = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// DigestPrefix starts every digest, as the results schema requires.
const DigestPrefix = "sha256:"

// Digest hashes the inputs of a run. The parts are length-prefixed so two
// different splits of the same bytes never collide.
func Digest(parts ...string) string {
	h := sha256.New()
	for _, part := range parts {
		h.Write([]byte{byte(len(part) >> 24), byte(len(part) >> 16), byte(len(part) >> 8), byte(len(part))})
		h.Write([]byte(part))
	}
	return DigestPrefix + hex.EncodeToString(h.Sum(nil))
}

// RunID names a run by protocol, subject and the first eight digits of the
// inputs digest.
func RunID(protocol model.Protocol, subject, digest string) string {
	name := strings.Trim(unsafeName.ReplaceAllString(subject, "-"), "-")
	if name == "" {
		name = "run"
	}
	short := strings.TrimPrefix(digest, DigestPrefix)
	if len(short) > 8 {
		short = short[:8]
	}
	return string(protocol) + "-" + name + "-" + short
}

// Store reads and writes run results below a directory.
type Store struct {
	FS     fs.FileSystem
	Reader document.Reader
	Dir    string
}

// File returns the checkpoint path of a run.
func (s Store) File(runID string) string { return path.Join(s.Dir, runID, ResultsFile) }

// Load returns the saved results of a run. The bool is false when the run has no
// checkpoint yet.
func (s Store) Load(runID string) (model.Results, bool, error) {
	file := s.File(runID)
	if ok, err := fs.Exists(s.FS, file); err != nil || !ok {
		return model.Results{}, false, err
	}
	results, err := s.Reader.Results(file)
	if err != nil {
		return model.Results{}, false, err
	}
	return results, true, nil
}

// Save writes the results of a run, replacing the earlier checkpoint.
func (s Store) Save(results model.Results) error {
	if results.RunID == "" {
		return errors.New("results have no run id")
	}
	data, err := model.Encode(results)
	if err != nil {
		return err
	}
	return s.FS.WriteFile(s.File(results.RunID), data)
}
