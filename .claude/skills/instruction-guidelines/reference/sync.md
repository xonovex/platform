# sync: Sync AGENTS.md with Filesystem

**Guideline:** Update AGENTS.md to reflect current directory structure, files, and configuration state while preserving technology names.

**Rationale:** Directory structures evolve. Syncing keeps instruction files accurate and up-to-date with the actual filesystem state so developers have reliable reference documentation.

**Example:**

```markdown
# Before sync (missing new directories)

## Subdirectories
- **services/**: TypeScript applications
- **clusters/**: Kubernetes GitOps
- **docs/**: Documentation

# After scan, detect new directories
# New: games/ (with C++, TypeScript), assets/, insights/

# After sync (directories now reflect filesystem)

## Subdirectories
- **services/**: TypeScript/JavaScript apps (package.json, tsconfig.json)
- **clusters/**: Kubernetes GitOps (kustomization.yaml)
- **docs/**: Technical documentation (*.md)
- **games/**: C++ game dev with TypeScript scripting (CMakeLists.txt, tsconfig.json)
- **assets/**: Brand assets and fonts (*.ttf, *.png)
- **insights/**: Development lessons (*.md)

# Updated commands from detected config files
# From moon.yml: "npx moon run <project>:<task>"
# From package.json: "npm install", "npm run build"
```

**Techniques:**
- Scan subdirectories 1 level deep, excluding node_modules, .git, build, dist
- Identify config files (package.json, moon.yml, CMakeLists.txt, Dockerfile)
- Format files inline: `(main.tf, vars, backend.sh)`
- Detect `<name>/` patterns for similar directories
- Add new directories, update existing entries, remove deleted ones
- Extract updated commands from config files while preserving technology names
