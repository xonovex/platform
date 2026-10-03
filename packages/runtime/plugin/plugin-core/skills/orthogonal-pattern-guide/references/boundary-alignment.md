# Boundary Alignment

A correctly placed axis contains one change concept and has one owner. A tidy directory on the wrong seam still leaks work across modules.

Check a proposed boundary:

1. Name one likely change in domain language.
2. List every file and team that change would touch.
3. Keep the boundary when the change stays within one axis and one owner.
4. Merge axes that always change together; split an axis whose parts change independently.

Use **ddd-guide** to discover bounded contexts and the vocabulary that defines domain seams. Use **connascence-guide** to measure coupling across the proposed seam. This guide only maps independently varying decisions onto the resulting module structure.

A boundary is not real merely because a folder exists. Its owner should be able to change one variant without coordinating edits in sibling axes; the enforcement mechanism belongs to the port or architecture-test owner.
