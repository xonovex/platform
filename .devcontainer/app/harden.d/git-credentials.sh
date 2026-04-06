# Git credential helpers — VS Code injects these to route git credential
# requests through its own helper, which proxies to the host's credential
# store. An agent could use this to exfiltrate tokens.

unset VSCODE_GIT_ASKPASS_NODE 2>/dev/null
unset VSCODE_GIT_ASKPASS_MAIN 2>/dev/null
unset VSCODE_GIT_ASKPASS_EXTRA_ARGS 2>/dev/null
unset GIT_ASKPASS 2>/dev/null

# VS Code injects a credential helper via git system config. The empty
# string set at build time (Dockerfile) overrides it, but VS Code may
# re-inject at the system level. This clears it again.
git config --global --unset credential.helper 2>/dev/null || true
