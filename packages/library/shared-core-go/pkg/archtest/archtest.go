// Package archtest reads the Go source of a module for its architecture
// fitness tests and reports the facts those tests judge: the imports of a file,
// whether it reads the clock, whether it prints, and which package state it
// keeps. It reads through an fs.FS, so the caller decides which tree it sees.
package archtest

import (
	"go/ast"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// File is one parsed Go source file of a module.
type File struct {
	// Path is the slash-separated path of the file inside the tree.
	Path string
	// Dir is the slash-separated directory of the file: its package.
	Dir string
	// Test marks a _test.go file.
	Test    bool
	Imports []string
	Syntax  *ast.File
}

// Load parses every .go file under each root of fsys. It skips the
// directories the go tool ignores: testdata and a name that begins with a dot
// or an underscore.
func Load(fsys fs.FS, roots ...string) ([]File, error) {
	fset := token.NewFileSet()
	var files []File
	for _, root := range roots {
		err := fs.WalkDir(fsys, root, func(name string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if name != root && ignored(entry.Name()) {
					return fs.SkipDir
				}
				return nil
			}
			if !strings.HasSuffix(name, ".go") {
				return nil
			}
			file, err := parse(fsys, fset, name)
			if err != nil {
				return err
			}
			files = append(files, file)
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func ignored(dir string) bool {
	return dir == "testdata" || strings.HasPrefix(dir, ".") || strings.HasPrefix(dir, "_")
}

func parse(fsys fs.FS, fset *token.FileSet, name string) (File, error) {
	source, err := fs.ReadFile(fsys, name)
	if err != nil {
		return File{}, err
	}
	syntax, err := parser.ParseFile(fset, name, source, parser.ParseComments|parser.SkipObjectResolution)
	if err != nil {
		return File{}, err
	}
	file := File{Path: name, Dir: path.Dir(name), Test: strings.HasSuffix(name, "_test.go"), Syntax: syntax}
	for _, spec := range syntax.Imports {
		// The parser accepts only a well-formed string literal as an import
		// path, so unquoting it cannot fail.
		value, _ := strconv.Unquote(spec.Path.Value)
		file.Imports = append(file.Imports, value)
	}
	return file, nil
}

// Production keeps the files that are not tests.
func Production(files []File) []File {
	var out []File
	for _, f := range files {
		if !f.Test {
			out = append(out, f)
		}
	}
	return out
}

// BuildsOnlyWith reports whether the //go:build line of the file holds when
// tag is set and fails when no tag is set: the file belongs to the tier that tag
// names.
func (f File) BuildsOnlyWith(tag string) bool {
	for _, group := range f.Syntax.Comments {
		if group.Pos() >= f.Syntax.Package {
			break
		}
		for _, comment := range group.List {
			if !constraint.IsGoBuild(comment.Text) {
				continue
			}
			expr, err := constraint.Parse(comment.Text)
			if err != nil {
				return false
			}
			return expr.Eval(func(name string) bool { return name == tag }) && !expr.Eval(func(string) bool { return false })
		}
	}
	return false
}

// IsStandard reports whether an import path names a standard library package:
// its first element holds no dot.
func IsStandard(importPath string) bool {
	first, _, _ := strings.Cut(importPath, "/")
	return !strings.Contains(first, ".")
}

// IsHost reports whether an import path reaches the real disk, the environment,
// the process or the network.
func IsHost(importPath string) bool {
	switch importPath {
	case "os", "os/exec", "os/user", "os/signal", "io/ioutil", "net", "net/http", "syscall":
		return true
	}
	return false
}

// ReadsClock reports whether the file reads the clock through time.Now,
// time.Since or time.Until.
func ReadsClock(file *ast.File) bool {
	found := false
	ast.Inspect(file, func(node ast.Node) bool {
		if pkg, name, ok := qualified(node); ok && pkg == "time" && (name == "Now" || name == "Since" || name == "Until") {
			found = true
		}
		return true
	})
	return found
}

// Prints lists each place where the file writes to the standard streams of the
// process: a fmt.Print, fmt.Printf or fmt.Println call, a print or println
// builtin call, or a reference to os.Stdout or os.Stderr.
func Prints(file *ast.File) []string {
	var found []string
	ast.Inspect(file, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok {
			if ident, ok := call.Fun.(*ast.Ident); ok && (ident.Name == "print" || ident.Name == "println") {
				found = append(found, ident.Name)
			}
		}
		pkg, name, ok := qualified(node)
		switch {
		case !ok:
		case pkg == "fmt" && (name == "Print" || name == "Printf" || name == "Println"):
			found = append(found, pkg+"."+name)
		case pkg == "os" && (name == "Stdout" || name == "Stderr"):
			found = append(found, pkg+"."+name)
		}
		return true
	})
	return found
}

// PackageState lists the package state a file keeps: every init function, and
// every package variable that a later assignment could change. A variable
// passes when each of its values is a literal, the address of a literal, a
// compiled pattern, a sentinel error or a value computed once, and when it is a
// blank assertion or an embedded file set.
func PackageState(file *ast.File) []string {
	var found []string
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil && d.Name.Name == "init" {
				found = append(found, "func init")
			}
		case *ast.GenDecl:
			if d.Tok != token.VAR {
				continue
			}
			for _, spec := range d.Specs {
				value := spec.(*ast.ValueSpec)
				if embedded(d, value) || fixed(value) {
					continue
				}
				for _, name := range value.Names {
					if name.Name != "_" {
						found = append(found, "var "+name.Name)
					}
				}
			}
		}
	}
	return found
}

// embedded reports a variable that a //go:embed directive fills at build time.
func embedded(decl *ast.GenDecl, spec *ast.ValueSpec) bool {
	for _, doc := range []*ast.CommentGroup{decl.Doc, spec.Doc} {
		if doc == nil {
			continue
		}
		for _, comment := range doc.List {
			if strings.HasPrefix(comment.Text, "//go:embed ") {
				return true
			}
		}
	}
	return false
}

// fixed reports a variable initialised once with values that nothing
// reassigns by design.
func fixed(spec *ast.ValueSpec) bool {
	if len(spec.Values) == 0 {
		return false
	}
	for _, value := range spec.Values {
		switch v := value.(type) {
		case *ast.CompositeLit, *ast.BasicLit, *ast.UnaryExpr:
		case *ast.SelectorExpr:
			// Another package's sentinel error, which callers match by identity.
			if _, name, ok := qualified(v); !ok || !strings.HasPrefix(name, "Err") {
				return false
			}
		case *ast.CallExpr:
			pkg, name, ok := qualified(v.Fun)
			if !ok || !onceCall(pkg, name) {
				return false
			}
		default:
			return false
		}
	}
	return true
}

// onceCall names the calls whose result stays the same for the life of the
// process: a compiled pattern, a sentinel error and a memoized value.
func onceCall(pkg, name string) bool {
	switch pkg + "." + name {
	case "regexp.MustCompile", "errors.New", "sync.OnceValue", "sync.OnceValues":
		return true
	}
	return false
}

// qualified returns the package and the name of a pkg.Name selector.
func qualified(node ast.Node) (string, string, bool) {
	selector, ok := node.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	pkg, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", "", false
	}
	return pkg.Name, selector.Sel.Name, true
}
