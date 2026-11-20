package artifact

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// FromTree reads the artifacts of one source out of its files and names the
// source in each of them. It fails closed on a name that is not kebab case, a
// skill whose front matter name differs from its directory or has no
// description, a skill directory without SKILL.md and a file that is not valid
// front matter.
func FromTree(source string, tree Tree) (Set, error) {
	var set Set
	withFile := map[string]bool{}
	for _, file := range tree.Under(SkillsDirectory) {
		name, rest, nested := strings.Cut(file.Path, "/")
		if !nested {
			continue
		}
		withFile[name] = withFile[name] || rest == SkillFile
	}
	for _, name := range slices.Sorted(maps.Keys(withFile)) {
		if !ValidName(name) {
			return Set{}, fmt.Errorf("%s: skill directory %q is not a kebab-case name", source, name)
		}
		if !withFile[name] {
			return Set{}, fmt.Errorf("%s: skill %q has no %s", source, name, SkillFile)
		}
		skill, err := readSkill(source, name, tree)
		if err != nil {
			return Set{}, err
		}
		set.Skills = append(set.Skills, skill)
	}
	if err := readFlat(source, tree, AgentsDirectory, func(name, text string) {
		set.Agents = append(set.Agents, Agent{Name: name, Source: source, Document: text})
	}); err != nil {
		return Set{}, err
	}
	if err := readFlat(source, tree, CommandsDirectory, func(name, text string) {
		set.Commands = append(set.Commands, Command{Name: name, Source: source, Document: text})
	}); err != nil {
		return Set{}, err
	}
	if file, ok := tree.Get(InstructionFile); ok {
		set.Instructions = append(set.Instructions, Instruction{Source: source, Content: string(file.Content)})
	}
	sortSet(&set)
	return set, nil
}

func readSkill(source, name string, tree Tree) (Skill, error) {
	skillFile, _ := tree.Get(SkillsDirectory + "/" + name + "/" + SkillFile)
	document, err := ParseDocument(string(skillFile.Content))
	if err != nil {
		return Skill{}, fmt.Errorf("%s: skill %q: %w", source, name, err)
	}
	if declared := document.Fields["name"]; declared != name {
		return Skill{}, fmt.Errorf("%s: skill %q declares the name %q", source, name, declared)
	}
	if document.Fields["description"] == "" {
		return Skill{}, fmt.Errorf("%s: skill %q has no description", source, name)
	}
	skill := Skill{Name: name, Source: source, Document: string(skillFile.Content)}
	for _, file := range tree.Under(SkillsDirectory + "/" + name) {
		if file.Path != SkillFile {
			skill.Files = append(skill.Files, file)
		}
	}
	return skill, nil
}

// readFlat reads the Markdown files that sit directly in directory.
func readFlat(source string, tree Tree, directory string, add func(name, text string)) error {
	for _, file := range tree.Under(directory) {
		name, isMarkdown := strings.CutSuffix(file.Path, ".md")
		if !isMarkdown || strings.Contains(name, "/") {
			continue
		}
		if !ValidName(name) {
			return fmt.Errorf("%s: %s file %q is not a kebab-case name", source, directory, file.Path)
		}
		if _, err := ParseDocument(string(file.Content)); err != nil {
			return fmt.Errorf("%s: %s/%s: %w", source, directory, file.Path, err)
		}
		add(name, string(file.Content))
	}
	return nil
}

// Merge combines sets into one. A name that two sets define fails closed,
// because the harness would pick one of them without saying so.
func Merge(sets ...Set) (Set, error) {
	var merged Set
	seen := map[string]string{}
	claim := func(kind Kind, name, source string) error {
		key := string(kind) + " " + name
		if other, found := seen[key]; found {
			return fmt.Errorf("%s %q is defined by %s and by %s", kind, name, other, source)
		}
		seen[key] = source
		return nil
	}
	for _, set := range sets {
		for _, skill := range set.Skills {
			if err := claim(KindSkill, skill.Name, skill.Source); err != nil {
				return Set{}, err
			}
		}
		for _, agent := range set.Agents {
			if err := claim(KindAgent, agent.Name, agent.Source); err != nil {
				return Set{}, err
			}
		}
		for _, command := range set.Commands {
			if err := claim(KindCommand, command.Name, command.Source); err != nil {
				return Set{}, err
			}
		}
		merged.Skills = append(merged.Skills, set.Skills...)
		merged.Agents = append(merged.Agents, set.Agents...)
		merged.Commands = append(merged.Commands, set.Commands...)
		merged.Instructions = append(merged.Instructions, set.Instructions...)
	}
	sortSet(&merged)
	return merged, nil
}

// sortSet orders the named lists by name; instructions keep their source order.
func sortSet(set *Set) {
	slices.SortStableFunc(set.Skills, func(a, b Skill) int { return strings.Compare(a.Name, b.Name) })
	slices.SortStableFunc(set.Agents, func(a, b Agent) int { return strings.Compare(a.Name, b.Name) })
	slices.SortStableFunc(set.Commands, func(a, b Command) int { return strings.Compare(a.Name, b.Name) })
}
