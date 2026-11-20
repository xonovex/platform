// Package hook checks hook configuration files: the event names, the matcher
// groups and the fields of each handler type.
package hook

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"slices"
	"sort"
	"strings"

	"github.com/xonovex/platform/packages/library/shared-contract-go/pkg/findings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-validate-go/internal/check"
)

// Name is the name of the check.
const Name = "hook"

// Options tune the check.
type Options struct {
	// File is the base name of a hook configuration, which sits in a directory
	// named Directory.
	File string `json:"file"`
	// Directory is the name of the directory that holds a hook configuration.
	Directory string `json:"directory"`
}

func defaults() Options { return Options{File: "hooks.json", Directory: "hooks"} }

// Check is the hook check.
type Check struct{}

// Name returns the name of the check.
func (Check) Name() string { return Name }

// Summary describes the check.
func (Check) Summary() string { return "hook event groups and handler configuration" }

// DefaultPaths returns the repository root.
func (Check) DefaultPaths() []string { return []string{"."} }

// Run checks the hook configurations below the paths.
func (Check) Run(ctx check.Context) ([]findings.Finding, error) {
	options := defaults()
	if err := check.DecodeOptions(ctx.Options, &options); err != nil {
		return nil, err
	}
	report := check.NewReporter(Name)
	files, err := ctx.Tree.WalkAll(ctx.Paths, func(p string) bool {
		return path.Base(p) == options.File && path.Base(path.Dir(p)) == options.Directory
	})
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		text, err := ctx.Tree.Read(file)
		if err != nil {
			return nil, err
		}
		var value any
		if err := json.Unmarshal([]byte(text), &value); err != nil {
			report.Fail("hook.json", file, 1, "invalid JSON: %v", err)
			continue
		}
		for _, i := range validate(value) {
			pointer := i.pointer
			if pointer == "" {
				pointer = "/"
			}
			report.Fail("hook.schema", file, lineOf(text, i.pointer), "hook schema %s %s", pointer, i.message)
		}
	}
	return report.Findings(), nil
}

var numeric = regexp.MustCompile(`^\d+$`)

// lineOf finds the line of the last key of a pointer. An index points at no
// single line, so the first line stands for it.
func lineOf(text, pointer string) int {
	segments := strings.FieldsFunc(pointer, func(r rune) bool { return r == '/' })
	if len(segments) == 0 || numeric.MatchString(segments[len(segments)-1]) {
		return 1
	}
	needle := `"` + segments[len(segments)-1] + `"`
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, needle) {
			return i + 1
		}
	}
	return 1
}

var eventNames = []string{
	"ConfigChange", "CwdChanged", "DirectoryAdded", "Elicitation", "ElicitationResult",
	"FileChanged", "InstructionsLoaded", "MessageDisplay", "Notification", "PermissionDenied",
	"PermissionRequest", "PostCompact", "PostToolBatch", "PostToolUse", "PostToolUseFailure",
	"PreCompact", "PreToolUse", "SessionEnd", "SessionStart", "Setup", "Stop", "StopFailure",
	"SubagentStart", "SubagentStop", "TaskCompleted", "TaskCreated", "TeammateIdle",
	"UserPromptExpansion", "UserPromptSubmit", "WorktreeCreate", "WorktreeRemove",
}

type kind string

const (
	nonEmptyString kind = "a non-empty string"
	boolean        kind = "true or false"
	positiveNumber kind = "a number greater than 0"
	stringList     kind = "a list of strings"
	uniqueStrings  kind = "a list of unique non-empty strings"
	stringMapping  kind = "a mapping of strings"
	anyMapping     kind = "a mapping"
	shellName      kind = `either "bash" or "powershell"`
)

type handlerType struct {
	name     string
	required []string
	fields   map[string]kind
}

var commonFields = map[string]kind{
	"if": nonEmptyString, "once": boolean, "statusMessage": nonEmptyString, "timeout": positiveNumber,
}

var handlerTypes = []handlerType{
	{"command", []string{"command"}, map[string]kind{"args": stringList, "async": boolean, "asyncRewake": boolean, "command": nonEmptyString, "shell": shellName}},
	{"http", []string{"url"}, map[string]kind{"allowedEnvVars": uniqueStrings, "headers": stringMapping, "url": nonEmptyString}},
	{"mcp_tool", []string{"server", "tool"}, map[string]kind{"input": anyMapping, "server": nonEmptyString, "tool": nonEmptyString}},
	{"prompt", []string{"prompt"}, map[string]kind{"model": nonEmptyString, "prompt": nonEmptyString}},
	{"agent", []string{"prompt"}, map[string]kind{"model": nonEmptyString, "prompt": nonEmptyString}},
}

type issue struct{ pointer, message string }

type issues []issue

func (r *issues) at(pointer, format string, args ...any) {
	*r = append(*r, issue{pointer, fmt.Sprintf(format, args...)})
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func validate(value any) issues {
	var r issues
	document, ok := value.(map[string]any)
	if !ok {
		r.at("/", "must be an object")
		return r
	}
	for _, key := range sortedKeys(document) {
		if key != "hooks" && key != "description" {
			r.at("/"+key, "is not a known top-level field")
		}
	}
	if description, present := document["description"]; present {
		field(&r, "/description", "description", nonEmptyString, description)
	}
	events, present := document["hooks"]
	if !present {
		r.at("/", "must have required property 'hooks'")
		return r
	}
	grouped, isMapping := events.(map[string]any)
	if !isMapping {
		r.at("/hooks", "must be an object keyed by event name")
		return r
	}
	if len(grouped) == 0 {
		r.at("/hooks", "must name at least one event")
	}
	for _, name := range sortedKeys(grouped) {
		if !slices.Contains(eventNames, name) {
			r.at("/hooks/"+name, "is not a known hook event")
			continue
		}
		groups(&r, "/hooks/"+name, grouped[name])
	}
	return r
}

func groups(r *issues, pointer string, value any) {
	list, ok := value.([]any)
	if !ok {
		r.at(pointer, "must be a list of matcher groups")
		return
	}
	if len(list) == 0 {
		r.at(pointer, "must have at least one matcher group")
		return
	}
	for i, entry := range list {
		at := fmt.Sprintf("%s/%d", pointer, i)
		group, isMapping := entry.(map[string]any)
		if !isMapping {
			r.at(at, "must be an object")
			continue
		}
		for _, key := range sortedKeys(group) {
			if key != "hooks" && key != "matcher" {
				r.at(at+"/"+key, "is not a matcher-group field")
			}
		}
		if matcher, present := group["matcher"]; present {
			if _, isText := matcher.(string); !isText {
				r.at(at+"/matcher", "must be a string")
			}
		}
		handlers, present := group["hooks"]
		if !present {
			r.at(at, "must have required property 'hooks'")
			continue
		}
		handlerList(r, at+"/hooks", handlers)
	}
}

func handlerList(r *issues, pointer string, value any) {
	list, ok := value.([]any)
	if !ok {
		r.at(pointer, "must be a list of handlers")
		return
	}
	if len(list) == 0 {
		r.at(pointer, "must have at least one handler")
		return
	}
	for i, entry := range list {
		handler(r, fmt.Sprintf("%s/%d", pointer, i), entry)
	}
}

func handler(r *issues, pointer string, value any) {
	h, ok := value.(map[string]any)
	if !ok {
		r.at(pointer, "must be an object")
		return
	}
	declared, present := h["type"]
	if !present {
		r.at(pointer, "must have required property 'type'")
		return
	}
	name, isText := declared.(string)
	if !isText {
		r.at(pointer+"/type", "must be a string")
		return
	}
	index := slices.IndexFunc(handlerTypes, func(t handlerType) bool { return t.name == name })
	if index == -1 {
		names := make([]string, len(handlerTypes))
		for i, t := range handlerTypes {
			names[i] = t.name
		}
		r.at(pointer+"/type", "must be one of %s", strings.Join(names, ", "))
		return
	}
	branch := handlerTypes[index]
	for _, key := range branch.required {
		if _, given := h[key]; !given {
			r.at(pointer, "must have required property '%s'", key)
		}
	}
	for _, key := range sortedKeys(h) {
		if key == "type" {
			continue
		}
		rule, own := branch.fields[key]
		if !own {
			if rule, own = commonFields[key]; !own {
				r.at(pointer+"/"+key, "is not a field of a %s handler", branch.name)
				continue
			}
		}
		field(r, pointer+"/"+key, key, rule, h[key])
	}
}

func field(r *issues, pointer, key string, rule kind, value any) {
	if valid(rule, value) {
		return
	}
	r.at(pointer, "%s must be %s", key, rule)
}

func valid(rule kind, value any) bool {
	switch rule {
	case nonEmptyString:
		s, ok := value.(string)
		return ok && strings.TrimSpace(s) != ""
	case boolean:
		_, ok := value.(bool)
		return ok
	case positiveNumber:
		n, ok := value.(float64)
		return ok && n > 0
	case stringList:
		return textList(value, false)
	case uniqueStrings:
		return textList(value, true)
	case stringMapping:
		m, ok := value.(map[string]any)
		if !ok {
			return false
		}
		for _, v := range m {
			if _, isText := v.(string); !isText {
				return false
			}
		}
		return true
	case anyMapping:
		_, ok := value.(map[string]any)
		return ok
	case shellName:
		return value == "bash" || value == "powershell"
	}
	return false
}

// textList reports whether a value is a list of strings. A unique list also
// needs each entry to be non-empty and to appear once.
func textList(value any, unique bool) bool {
	list, ok := value.([]any)
	if !ok {
		return false
	}
	seen := map[string]bool{}
	for _, item := range list {
		s, isText := item.(string)
		if !isText {
			return false
		}
		if unique {
			if s == "" || seen[s] {
				return false
			}
			seen[s] = true
		}
	}
	return true
}
