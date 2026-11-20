# Role briefs

A migration research has one read-only explorer for each role. Every brief states the question, why the answer matters to a decision, the repositories to read and the required output: `Observed` and `Inference` headings, one anchor per line as repository root plus path and line, and an `Unknowns` section. An explorer never writes a file. A vague brief makes explorers duplicate work and leave gaps, so each brief also names what the role does not cover.

The examples use an invented publisher that sends order notices through a legacy relay, a delivery service that replaces the relay, a mobile client and a web client.

## Contents

- The current producer
- The target
- Each consumer app
- After the reports

## The current producer

Question: what does the producer send, through which path, and what does it do with the answer? It matters because every change the migration forces on the producer starts here.

Read for:

- The send path and the one seam every send goes through, and any place that bypasses the seam. A bypass is a second path to migrate.
- The payload and every field the consumers read. A field a consumer reads and the target does not carry is a gap.
- The stored identifiers, and the tables or indexes that hold them. A migration that changes the identifier changes the storage.
- How the response is handled, including what triggers a sign-off, a deletion or a state change.
- Retry behavior, and what a failure does to the caller of the producer.
- Configuration for each environment, and where its credentials are kept.

Example: the order publisher calls the legacy relay from a single `NoticeSender`, and a batch job also calls the relay directly. The brief asks for both paths.

## The target

Question: what does the target accept, and what does it do when the input is wrong? It matters because the gap table compares this contract with the producer's behavior.

Read for:

- The accepted recipient types and what each one needs.
- Authentication, and whether anything restricts a caller to a channel.
- What the target does with a rejected recipient, an unknown recipient and a switched-off channel, and whether any of it reaches the caller.
- Whether documented behavior exists in the code, for example persistence or deduplication on an idempotency key. A design document is a claim, not evidence.
- Capabilities removed in history and capabilities on unmerged branches. Say which branch each is on.
- Neighbor repositories that already call the target. They are the reference callers.

When the target does not exist yet, point the explorer at its contract or design record instead, and require the answer to say that no running target was read.

Example: the delivery service documents deduplication on an idempotency key, and the brief asks the explorer to find the code that does it.

## Each consumer app

Use one brief and one explorer for each client app. A mobile client and a web client release on different cadences and handle messages differently.

Question: what does the app do with the old and the new message shape, and which versions are in use? It matters because the consumers decide when the producer can switch.

Read for:

- Which identifiers the app knows and which it sends on every call.
- How it handles old and new message shapes, for each version.
- What the user's action on a message does, and where that route is declared.
- What happens on a failed update, and when the app registers again.
- The versions that matter and the release cadence, so the supported versions can be computed.

Example: the web client opens a message route by path and the mobile client by a link scheme. The brief asks where each is declared.

## After the reports

Read the explorers' pointers to anything outside the default branch yourself. Re-check the two or three claims that would change a decision, and state which side you checked: the code, a design record, or a running system.
