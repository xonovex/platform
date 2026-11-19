# Attaching items to an existing work package

When the invoker names a work package, the items attach to it and no new package is proposed.

## Check the package

1. The named file exists in `work/` and has `kind: work-package`. If the name is a work item, or the file is missing, report it and stop. Do not guess another package and do not create one in its place.
2. Read its objective. The items must contribute to it. If the cut shows an item that does not, say so in the proposal and ask whether it belongs.

## Attach

- Set `parent` on each item to the package's concept ID, for example `work/invoice-feed-rollout-k3ne`.
- Leave the package file unchanged. If its objective, team or readiness is wrong, say so in the report and let its owner or **revising-plan** change it. The one exception is the operator asking for the change by name in the cut.
- The package may belong to another team. The items still carry the operator's team in `team`.

## Items already attached

Rows that already have a `parent` keep it. A different parent in the proposal is a change the operator must see in the proposal's table before approving it.
