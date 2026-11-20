package workspace

import (
	"encoding/json"
	"path"
	"sort"
	"strings"
)

// npm creates a node_modules/.bin link only when the bin target already exists
// as the package is linked, and building the target afterwards adds no link, so
// npx falls through to the registry and fails. A private package is linked
// from the working tree, so its target has to be committed; a published
// package is linked from its unpacked tarball, so its target has to be inside
// the files allowlist.

type binManifest struct {
	Name    string          `json:"name"`
	Private bool            `json:"private"`
	Files   []string        `json:"files"`
	Bin     json.RawMessage `json:"bin"`
}

type binEntry struct{ name, target string }

// binEntries reads the bin field. A string takes the package name, which npm
// strips to the last segment for a scoped package.
func binEntries(raw json.RawMessage, name string) []binEntry {
	if len(raw) == 0 {
		return nil
	}
	var single string
	if json.Unmarshal(raw, &single) == nil {
		return []binEntry{{name: name[strings.LastIndex(name, "/")+1:], target: single}}
	}
	var many map[string]string
	if json.Unmarshal(raw, &many) != nil {
		return nil
	}
	names := make([]string, 0, len(many))
	for n := range many {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]binEntry, 0, len(names))
	for _, n := range names {
		out = append(out, binEntry{name: n, target: many[n]})
	}
	return out
}

func normalize(p string) string {
	return strings.TrimSuffix(strings.TrimPrefix(p, "./"), "/")
}

// A files entry covers the target when it names it or is a directory above it.
func isPacked(target string, files []string) bool {
	for _, entry := range files {
		packed := normalize(entry)
		if target == packed || strings.HasPrefix(target, packed+"/") {
			return true
		}
	}
	return false
}

func (r *run) binTargets() error {
	var manifests []struct {
		dir  string
		file string
		m    binManifest
	}
	for _, p := range r.projects {
		file := path.Join(p.dir, "package.json")
		text, found, err := r.ctx.Tree.Optional(file)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		var m binManifest
		if err := json.Unmarshal([]byte(text), &m); err != nil {
			r.report.Fail("workspace.bin-manifest", file, 1, "%s is not valid JSON: %v", file, err)
			continue
		}
		if len(m.Bin) == 0 {
			continue
		}
		manifests = append(manifests, struct {
			dir  string
			file string
			m    binManifest
		}{p.dir, file, m})
	}
	if len(manifests) == 0 {
		return nil
	}
	if r.ctx.Repository == nil {
		r.report.Fail("workspace.bin-git", "", 0, "git could not list the committed files")
		return nil
	}
	tracked, err := r.ctx.Repository.TrackedFiles()
	if err != nil {
		r.report.Fail("workspace.bin-git", "", 0, "git could not list the committed files: %v", err)
		return nil
	}
	committed := make(map[string]bool, len(tracked))
	for _, t := range tracked {
		committed[t] = true
	}
	for _, mf := range manifests {
		for _, bin := range binEntries(mf.m.Bin, mf.m.Name) {
			target := normalize(bin.target)
			if committed[mf.dir+"/"+target] {
				continue
			}
			if !mf.m.Private && isPacked(target, mf.m.Files) {
				continue
			}
			if mf.m.Private {
				r.report.Fail("workspace.bin-target", mf.file, 1, "%s bin %s points at %s, which is not committed, so a cold install links no %s", mf.file, bin.name, target, bin.name)
			} else {
				r.report.Fail("workspace.bin-target", mf.file, 1, "%s bin %s points at %s, which is neither committed nor inside the files allowlist", mf.file, bin.name, target)
			}
		}
	}
	return nil
}
