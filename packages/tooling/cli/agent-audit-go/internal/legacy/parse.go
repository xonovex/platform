package legacy

import (
	"path"
	"regexp"
	"strings"
)

var (
	h1Line        = regexp.MustCompile(`^#\s+(.*\S)\s*$`)
	h2Line        = regexp.MustCompile(`^##\s+(.*\S)\s*$`)
	fieldLine     = regexp.MustCompile(`^[-*]\s+\*\*([^*:]+):\*\*\s*(.*?)\s*$`)
	itemLine      = regexp.MustCompile(`^\s+[-*]\s+(.*?)\s*$`)
	urlPattern    = regexp.MustCompile("https?://[^\\s<>\"`]+")
	bareURL       = regexp.MustCompile(`^https?://\S+$`)
	referenceIn   = regexp.MustCompile(`references/([a-z0-9][a-z0-9-]*\.md)`)
	arrowTo       = regexp.MustCompile("(?:->|→)\\s*`?([a-z0-9][a-z0-9-]*\\.md)`?")
	referenceName = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*\.md$`)
	hex64         = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)
	watchLine     = regexp.MustCompile(`^(.+?)\s*(?:->|→)\s*(.+\S)$`)
	allRefs       = regexp.MustCompile("(?i)\\ball `references/`|\\ball references\\b|\\bevery\\b.*\\breference")
	versionTick   = regexp.MustCompile("^`([^`]+)`")
)

// referencesDir is the directory the feeds of a legacy catalog live in.
const referencesDir = "references"

// block is the content of one `##` section of a legacy SOURCES.md.
type block struct {
	title string

	urls       []string
	urlNotes   []string
	citation   string
	provenance string
	reviewed   string
	version    string
	commit     string
	checkout   string
	watch      []watchEntry
	contentSHA string
	outputSHA  string
	usage      []string
	aspects    []string
	covers     bool
	feeds      []string
	remarks    []string

	// body holds every line of the section, for notes.
	body []string
}

type watchEntry struct {
	path  string
	feeds []string
}

func (b *block) isSource() bool {
	return len(b.urls) > 0 || len(b.urlNotes) > 0 || b.provenance != ""
}

// section is one `##` heading with the lines below it.
type section struct {
	title string
	lines []string
}

// split cuts a legacy file into its heading, the prose before the first
// section, and the sections.
func split(text string) (title string, preamble []string, sections []section) {
	var current *section
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		if m := h2Line.FindStringSubmatch(line); m != nil {
			sections = append(sections, section{title: m[1]})
			current = &sections[len(sections)-1]
			continue
		}
		if current != nil {
			current.lines = append(current.lines, line)
			continue
		}
		if m := h1Line.FindStringSubmatch(line); m != nil && title == "" {
			title = m[1]
			continue
		}
		preamble = append(preamble, line)
	}
	return title, preamble, sections
}

// parseBlock reads the fields of one section. A field is a top-level bullet
// written as - **Label:** value, and the indented bullets below it belong to
// it. Every other line is prose of the source.
func parseBlock(s section) *block {
	b := &block{title: s.title, body: s.lines}
	label := ""
	paragraphBreak := false
	for _, line := range s.lines {
		b.collectReferences(line)
		if m := fieldLine.FindStringSubmatch(line); m != nil {
			label = strings.ToLower(strings.TrimSpace(m[1]))
			b.field(label, m[2], line)
			paragraphBreak = false
			continue
		}
		if m := itemLine.FindStringSubmatch(line); m != nil && label != "" {
			b.field(label, m[1], line)
			continue
		}
		if strings.TrimSpace(line) == "" {
			paragraphBreak = len(b.remarks) > 0
			continue
		}
		label = ""
		if paragraphBreak {
			b.remarks = append(b.remarks, "")
			paragraphBreak = false
		}
		b.remarks = append(b.remarks, strings.TrimRight(line, " \t"))
	}
	return b
}

// collectReferences adds the reference files a line names, in order of first
// mention: references/x.md paths and the files after an arrow.
func (b *block) collectReferences(line string) {
	for _, m := range referenceIn.FindAllStringSubmatch(line, -1) {
		b.addFeed(m[1])
	}
	for _, m := range arrowTo.FindAllStringSubmatch(line, -1) {
		b.addFeed(m[1])
	}
}

func (b *block) addFeed(name string) {
	feed := path.Join(referencesDir, name)
	for _, existing := range b.feeds {
		if existing == feed {
			return
		}
	}
	b.feeds = append(b.feeds, feed)
}

// field stores one value of a field. The raw line is kept as prose when the
// field is not one the format knows, so no text is lost.
func (b *block) field(label, value, raw string) {
	switch label {
	case "url", "urls":
		b.addURLs(value)
	case "title":
		b.setOnce(&b.citation, value, raw)
	case "provenance":
		b.setOnce(&b.provenance, value, raw)
	case "last reviewed":
		b.setOnce(&b.reviewed, value, raw)
	case "version":
		if m := versionTick.FindStringSubmatch(value); m != nil {
			value = m[1]
		}
		b.setOnce(&b.version, value, raw)
	case "commit":
		b.setOnce(&b.commit, value, raw)
	case "checkout":
		b.setOnce(&b.checkout, value, raw)
	case "content sha256":
		b.setDigest(&b.contentSHA, value, raw)
	case "cli output sha256":
		b.setDigest(&b.outputSHA, value, raw)
	case "watch":
		b.addWatch(value, raw)
	case "references":
		b.addReferences(value)
	case "used for", "also used for":
		b.addText(&b.usage, value)
		b.covers = b.covers || allRefs.MatchString(value)
	case "aspects extracted":
		b.addText(&b.aspects, value)
	default:
		b.remarks = append(b.remarks, strings.TrimRight(raw, " \t"))
	}
}

// setOnce keeps the first value of a single-valued field and the raw line of a
// repeat, so a second value is never dropped.
func (b *block) setOnce(target *string, value, raw string) {
	if value == "" {
		return
	}
	if *target != "" {
		b.remarks = append(b.remarks, strings.TrimRight(raw, " \t"))
		return
	}
	*target = value
}

func (b *block) setDigest(target *string, value, raw string) {
	if value != "" && hex64.MatchString(value) {
		b.setOnce(target, "sha256:"+strings.ToLower(value), raw)
		return
	}
	if value != "" {
		b.remarks = append(b.remarks, strings.TrimRight(raw, " \t"))
	}
}

func (b *block) addText(target *[]string, value string) {
	if value != "" {
		*target = append(*target, value)
	}
}

// addURLs reads the URLs of one line of a URL field. A line that says more than
// a bare URL is kept whole next to the list.
func (b *block) addURLs(value string) {
	if value == "" {
		return
	}
	found := urlPattern.FindAllString(value, -1)
	exact := len(found) == 1 && bareURL.MatchString(value)
	for _, url := range found {
		trimmed := trimURL(url)
		exact = exact && trimmed == url
		if !containsString(b.urls, trimmed) {
			b.urls = append(b.urls, trimmed)
		}
	}
	if !exact {
		b.urlNotes = append(b.urlNotes, value)
	}
}

// trimURL drops the punctuation that ends a sentence, and a closing bracket
// that has no opening one in the URL.
func trimURL(url string) string {
	for {
		trimmed := strings.TrimRight(url, ".,;:·")
		if strings.HasSuffix(trimmed, ")") && strings.Count(trimmed, ")") > strings.Count(trimmed, "(") {
			trimmed = strings.TrimSuffix(trimmed, ")")
		}
		if trimmed == url {
			return url
		}
		url = trimmed
	}
}

func (b *block) addWatch(value, raw string) {
	if value == "" {
		return
	}
	m := watchLine.FindStringSubmatch(value)
	if m == nil {
		b.remarks = append(b.remarks, strings.TrimRight(raw, " \t"))
		return
	}
	entry := watchEntry{path: strings.TrimSpace(m[1])}
	for _, name := range strings.Split(m[2], ",") {
		name = path.Base(strings.TrimSpace(name))
		if strings.HasSuffix(name, ".md") {
			entry.feeds = append(entry.feeds, path.Join(referencesDir, name))
			b.addFeed(name)
		}
	}
	b.watch = append(b.watch, entry)
}

// addReferences reads the value of a **References:** field: the word all, or a
// comma-separated list of reference files.
func (b *block) addReferences(value string) {
	value = strings.TrimSpace(strings.TrimLeft(strings.TrimSpace(value), "-:"))
	if strings.EqualFold(value, "all") || strings.EqualFold(value, "all references") {
		b.covers = true
		return
	}
	for _, name := range strings.Split(value, ",") {
		name = path.Base(strings.ReplaceAll(strings.TrimSpace(name), "`", ""))
		if referenceName.MatchString(name) {
			b.addFeed(name)
		}
	}
}

func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
