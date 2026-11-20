package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/fs"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/catalog"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/document"
	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// resolve makes a path absolute against the working directory.
func resolve(dir, p string) string {
	if strings.HasPrefix(p, "/") {
		return path.Clean(p)
	}
	return path.Join(dir, p)
}

func (d Dependencies) reader() document.Reader {
	return document.Reader{FS: d.FS, Validator: d.Validator}
}

// loadProject reads the project file. An unset flag falls back to the default
// file in the working directory, and a project that is not there is empty.
func (d Dependencies) loadProject(flag string) (model.Project, string, error) {
	file := flag
	if file == "" {
		file = model.ProjectFile
	}
	file = resolve(d.Dir, file)
	exists, err := fs.Exists(d.FS, file)
	if err != nil {
		return model.Project{}, "", err
	}
	if !exists {
		if flag != "" {
			return model.Project{}, "", fmt.Errorf("project file %s: %w", file, fs.ErrNotExist)
		}
		return model.Project{Version: model.Version, Profiles: map[string]model.Profile{}}, d.Dir, nil
	}
	project, err := d.reader().Project(file)
	return project, path.Dir(file), err
}

// loadCatalog discovers the plugins in the project's plugin directories and the
// directories the flags add. With neither, it looks in the working directory.
func (d Dependencies) loadCatalog(project model.Project, projectDir string, extra []string) (catalog.Catalog, error) {
	var roots []string
	for _, dir := range project.PluginDirs {
		roots = append(roots, resolve(projectDir, dir))
	}
	for _, dir := range extra {
		roots = append(roots, resolve(d.Dir, dir))
	}
	if len(roots) == 0 {
		roots = []string{d.Dir}
	}
	return catalog.Load(d.FS, roots)
}
