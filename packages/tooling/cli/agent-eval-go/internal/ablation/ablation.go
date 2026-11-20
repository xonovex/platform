// Package ablation turns an instruction file into the text a condition runs
// with. Present keeps the file. Absent removes the rule that an anchor names,
// so the run shows what the agent does without it. Ablated removes only a named
// phrase inside the rule, so the run shows which words carry the behavior.
package ablation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xonovex/platform/packages/tooling/cli/agent-eval-go/internal/model"
)

// ErrAnchor reports an anchor that matches nothing or more than one place.
var ErrAnchor = errors.New("anchor")

// Apply returns the instruction text for a condition.
func Apply(instructions string, anchor model.Anchor, ablate string, condition model.Condition) (string, error) {
	switch condition {
	case model.Present:
		return instructions, nil
	case model.Absent, model.Ablated:
	default:
		return "", fmt.Errorf("unknown condition %q", condition)
	}
	start, end, err := locate(instructions, anchor)
	if err != nil {
		return "", err
	}
	if condition == model.Absent {
		return instructions[:start] + instructions[end:], nil
	}
	rule := instructions[start:end]
	if ablate == "" {
		return "", errors.New("an ablated case needs the text to ablate")
	}
	if strings.Count(rule, ablate) != 1 {
		return "", fmt.Errorf("%w: ablate text %q must occur once inside the anchored rule", ErrAnchor, ablate)
	}
	return instructions[:start] + strings.Replace(rule, ablate, "", 1) + instructions[end:], nil
}

// locate returns the byte range of the anchored rule.
func locate(instructions string, anchor model.Anchor) (int, int, error) {
	if anchor.Text != "" {
		if strings.Count(instructions, anchor.Text) != 1 {
			return 0, 0, fmt.Errorf("%w: text %q must occur once in the instructions", ErrAnchor, anchor.Text)
		}
		start := strings.Index(instructions, anchor.Text)
		return start, start + len(anchor.Text), nil
	}
	return section(instructions, anchor.Heading)
}

// section returns the range from a heading to the next heading of the same or a
// higher level. A line in a fenced code block is never a heading.
func section(instructions, heading string) (int, int, error) {
	want := headingText(heading)
	if want == "" {
		return 0, 0, fmt.Errorf("%w: needs a heading or a text", ErrAnchor)
	}
	offset, fenced := 0, false
	start, level := -1, 0
	for _, line := range strings.SplitAfter(instructions, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			fenced = !fenced
		} else if !fenced {
			if l := headingLevel(line); l > 0 {
				if start >= 0 && l <= level {
					return start, offset, nil
				}
				if start < 0 && headingText(line) == want {
					start, level = offset, l
				}
			}
		}
		offset += len(line)
	}
	if start < 0 {
		return 0, 0, fmt.Errorf("%w: heading %q not found", ErrAnchor, heading)
	}
	return start, len(instructions), nil
}

func headingLevel(line string) int {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return 0
	}
	return level
}

func headingText(line string) string {
	return strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(line), "#"))
}
