# shellcheck shell=bash

# VS Code IPC sockets — primary container escape vector.
# These variables point to Unix sockets that allow direct communication
# with the VS Code process running on the HOST. An agent with access to
# these sockets can execute arbitrary commands outside the container.

unset VSCODE_IPC_HOOK_CLI 2>/dev/null
unset VSCODE_GIT_IPC_HANDLE 2>/dev/null

find /tmp -maxdepth 2 \( -name "vscode-ipc-*" -o -name "vscode-git-*" \) -type s -delete 2>/dev/null || true
