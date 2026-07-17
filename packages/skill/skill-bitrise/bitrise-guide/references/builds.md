# Bitrise builds — find, read, triage, re-trigger

## Find the build for a commit or PR

Bitrise posts each build's status back to the git host, so the git host is the index from commit -> Bitrise build. On a Bitbucket Server host this is the build-status endpoint (HTTP basic auth via `~/.netrc`); it needs **no** Bitrise token:

```bash
SHA=<full-commit-sha>
curl -s -n "https://<git-host>/rest/build-status/1.0/commits/$SHA" | python3 -m json.tool
```

(The `/rest/build-status/1.0/` path is Bitbucket Server's; other hosts expose an equivalent commit-status API. Authenticate to the git host with whatever that host expects — for a Bitbucket Server host that is HTTP basic auth via `~/.netrc`, which Bitrise itself never uses.)

Each entry has:

- `state` — `SUCCESSFUL` / `FAILED` / `INPROGRESS`.
- `key` — a display label like `Bitrise - Build run-tests`. This is **not** a triggerable workflow id (see re-trigger below).
- `name` — the build's display name with workflow and number, e.g. `Bitrise <App> (<workflow>) #<number>`.
- `url` — `https://app.bitrise.io/app/<APP_SLUG>/build/<BUILD_SLUG>`; split it for the app and build slugs.

To check whether a failure is specific to one PR or hitting many (a flakiness signal), compare the same workflow's state across several open PRs' head commits — a mix of FAILED and SUCCESSFUL on an unrelated change points to flaky infra, not the change.

## Inspect a build

```bash
BUILD=<build-slug>
curl -s -H "Authorization: $TOKEN" \
  "https://api.bitrise.io/v0.1/apps/$APP/builds/$BUILD" \
  | python3 -c "import sys,json;d=json.load(sys.stdin)['data'];print(json.dumps({k:d.get(k) for k in ['status_text','triggered_workflow','branch','pull_request_id','pull_request_target_branch','original_build_params']},indent=2))"
```

The key field for re-triggering is `original_build_params.workflow_id` — the **real** workflow name, which can differ from the build-status `key` label (e.g. label `run-tests` vs real `run-tests-from-source-control`).

## Read or download a build log

```bash
curl -s -H "Authorization: $TOKEN" \
  "https://api.bitrise.io/v0.1/apps/$APP/builds/$BUILD/log" \
  | python3 -c "import sys,json;d=json.load(sys.stdin);print('is_archived',d['is_archived'],'chunks',len(d.get('log_chunks',[])))"
```

- **Finished build** (`is_archived: true`) — the whole raw log sits at `expiring_raw_log_url`, a short-lived signed URL. Fetch it with **no** Authorization header (the header can make the storage backend reject the request):

```bash
URL=$(curl -s -H "Authorization: $TOKEN" \
  "https://api.bitrise.io/v0.1/apps/$APP/builds/$BUILD/log" \
  | python3 -c "import sys,json;print(json.load(sys.stdin)['expiring_raw_log_url'])")
curl -s "$URL" -o /tmp/build_log.txt
```

- **Running build** (`is_archived: false`) — the log streams in as paginated `log_chunks`; page with `next_before_timestamp` / `next_after_timestamp`. Usually it is easier to wait for archival and read the whole thing.

## Triage: infra flakiness vs a real failure

Read the log before concluding anything. Bitrise marks the whole build FAILED whether a test asserted wrong or an emulator dropped offline — only the log distinguishes them.

```bash
grep -nE "device offline|AdbCommandRejectedException|Expected [0-9]+ tests, received|[1-9][0-9]* failed|AssertionError|Task .* FAILED|What went wrong" /tmp/build_log.txt
```

**Infra / emulator flakiness** (not your code) looks like:

- `Exception thrown during onAfterAll invocation of plugin AndroidTestApkInstallerPlugin: device offline`
- `com.android.ddmlib.AdbCommandRejectedException: device offline`
- `Test run failed to complete. Expected N tests, received 1` — a shard's managed device disconnected before running its tests.
- A shard that already `Finished N tests on <device>` with `(0 failed)` just before the offline error — the tests passed, the device died in teardown.

The fix for flakiness is to **re-trigger**, not to change code.

**A real failure** looks like:

- A running tally ending in `(N failed)` with `N > 0`.
- An `AssertionError` / exception stack trace from a test class, or a Paparazzi snapshot mismatch.
- A compile / lint task `FAILED` with the offending file and message.

## Re-trigger a build or PR check

```bash
curl -s -X POST -H "Authorization: $TOKEN" -H "Content-Type: application/json" \
  "https://api.bitrise.io/v0.1/apps/$APP/builds" \
  -d '{
    "hook_info": {"type": "bitrise"},
    "build_params": {
      "branch": "<source-branch>",
      "branch_dest": "main",
      "commit_hash": "<full-sha>",
      "commit_message": "<message>",
      "pull_request_id": <pr-id>,
      "workflow_id": "<real-workflow-id-from-original_build_params>"
    }
  }' | python3 -m json.tool
```

- For a plain branch build (not a PR) drop `branch_dest` / `pull_request_id` and keep `branch`, `commit_hash`, `workflow_id`.
- Success is `{"status":"ok","message":"webhook processed", "build_slug":..., "build_number":..., "build_url":...}`. "webhook processed" is the success message.
- If you pass a `workflow_id` that the app config does not define, you get HTTP **200** with `{"status":"error","message":"workflow (...) did not match any workflows defined in app config"}` and **no build starts**. This is the most common mistake — it happens when you copy the workflow name from the build-status `key` label instead of reading `original_build_params.workflow_id` from a prior build. Always confirm the name against a real build first.
