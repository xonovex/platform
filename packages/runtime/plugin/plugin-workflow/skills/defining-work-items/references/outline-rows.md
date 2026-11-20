# Outline rows in the cut

## Whose rows

The operator's team owns the rows to define. Ask once which team the operator speaks for when the invocation does not say. Rows of other teams stay as they are, including their criteria, so their owners define them. The operator may ask for a row of another team by name; then it is part of the cut and the proposal says so. A dependency on another team's row is allowed; the row is referenced and never edited.

## Rows that wait

A row that waits on an open, paused or partly settled decision stays `waiting`, keeps its `waits_on`, and is defined only as far as its outline: the objective and the criterion the outline carries. Do not add criteria or an item split that the open decision would decide. The proposal lists these rows with the decisions they wait on, so the operator sees what is left and why.

When a decision a row waits on has since settled, the row's readiness is out of date. Do not change it here: report it, because **revising-plan** owns readiness and the plan.

## Splitting and merging outline rows

- Split: keep the original file for the part that continues it, write the other parts as new files with new names, and give each part its own criteria. Update the `depends_on` of other rows that pointed at the original when they depend on the new part.
- Merge: keep one file as the survivor and move the criteria into it. The file the merge makes redundant is not deleted or renamed. Set `status: deprecated` on it with a line in its `# Notes` that names the survivor, and update the `depends_on` of other rows to the survivor.
