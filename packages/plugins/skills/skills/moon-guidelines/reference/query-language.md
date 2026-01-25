# query-language: Advanced Project and Task Filtering

**Guideline:** Use Moon Query Language (MQL) for advanced filtering of projects and tasks.

**Rationale:** MQL allows precise selection of projects using language, type, tags, and other criteria for targeted task execution and querying.

**Example:**

```bash
# Query projects by language and type
moon query projects "language=javascript && projectType=library"

# Run build for typescript projects
moon run :build --query "language=typescript"

# Complex filtering
moon run :test --query "(language=javascript || language=typescript) && projectType=application"
```

**Techniques:**
- Operators: `=` (equals), `!=` (not), `~` (regex), `!~` (not regex), `&&` (AND), `||` (OR)
- Project fields: language, projectType, tags, projectName, projectAlias, projectSource
- #tag:task syntax: Use for simple single-tag filtering
- --query flag: Use for complex multi-criterion conditions
- Performance: Prefer exact matches (`=`) over regex when possible
- Parentheses: Group conditions for complex boolean logic
- Field types: language=javascript|typescript, projectType=library|application
