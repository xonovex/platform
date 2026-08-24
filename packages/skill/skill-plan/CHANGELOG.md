# @xonovex/skill-plan

## 5.2.0 - 2026-08-23

### Added

- Followup operation: close out a completed, paused, or handed-over plan with an
  inline record of evidenced status, gates and blockers, an open-decision register,
  residue, follow-up plan seeds, a review brief, and a retrospective. Results are
  inline; plan seeds persist through Create.
- Distill operation: turn a completed plan's branch, record, and execution logs into
  a self-contained replayable skill suite, with provenance quarantine and mechanical
  suite verification.

## 5.1.0 - 2026-08-06

### Breaking changes

- Narrowed the skill to planning and read-only code research. Early-lifecycle,
  design, decision, approval, rejection, and role-governance procedures are no longer
  owned by this package.
- Plan status is descriptive metadata. Subplan expansion no longer requires an
  approval transition.
