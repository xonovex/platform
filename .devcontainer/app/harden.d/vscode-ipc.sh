# VS Code IPC sockets — primary container escape vector.
# These variables point to Unix sockets that allow direct communication
# with the VS Code process running on the HOST. An agent with access to
# these sockets can execute arbitrary commands outside the container.
#
# Sources:
# - The Red Guild: Leveraging VS Code Internals to Escape Containers
# - Daniel Demmel: Coding Agents in Secured VS Code Dev Containers

unset VSCODE_IPC_HOOK_CLI 2>/dev/null
unset VSCODE_GIT_IPC_HANDLE 2>/dev/null

# One-time cleanup of existing IPC sockets left in /tmp.
find /tmp -maxdepth 2 \( -name "vscode-ipc-*" -o -name "vscode-git-*" \) -type s -delete 2>/dev/null || true
