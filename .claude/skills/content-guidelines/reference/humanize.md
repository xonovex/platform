# humanize: Make Writing Sound Human

**Guideline:** Remove AI patterns from documentation to make it sound natural and professional.

**Rationale:** AI-generated text contains recognizable patterns (forced casual tone, filler phrases, excessive formatting) that diminish credibility and readability.

**Example:**

Before:
```markdown
## Phase 1: Setup (Est. 2-3 days)
Let's dive in! TD-01: Configure Database ✅
This involves database connection. In other words:
- Connect to PostgreSQL - primary data store
- Set up migrations - easy fix though
```

After:
```markdown
## Setup
Configure database by connecting to PostgreSQL as primary data store and setting up migrations.
```

**Techniques:**
- Remove task IDs: Strip TD-01, TASK-001, phase structures, time estimations
- Replace filler: Remove "Let's dive in", "here's the thing", "easy fix though"
- Simplify punctuation: Use commas instead of dashes (", " not " - ")
- Prose over bullets: Use paragraphs where logical flow works better
- Remove enthusiasm: Eliminate excessive exclamation marks, emojis, meta-commentary
- Keep accuracy: Maintain technical precision while removing fluff
