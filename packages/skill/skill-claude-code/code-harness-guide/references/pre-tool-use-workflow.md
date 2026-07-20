# PreToolUse Workflow Adapter

Use `scripts/workflow-pre-tool-use.sh` only as a transport from Claude Code to a workflow adapter. Set `XONOVEX_WORKFLOW_HOOK_EXECUTABLE` to one executable that reads the native hook JSON from standard input and returns the native hook exit status. The transport does not choose tools, policy, enforcement mode, evidence, or failure behavior.

Choose the `PreToolUse` matcher in the owning Claude settings layer. The matcher defines which tool calls produce workflow trigger events; it is not part of the workflow itself. The target executable should authenticate or validate the event as needed, map it to a trusted workflow template, run the selected composition, and translate a denied result to Claude Code's documented blocking response.

Keep command arguments in a separate wrapper executable so settings and environment parsing cannot change their meaning. Verify the exact matcher, executable, exit behavior, timeout, sibling hooks, and active settings scope before treating the adapter as an enforcement point.
