#!/bin/bash
# =============================================================================
# Shell Hardening — Module Loader
#
# Sources all hardening modules from /usr/local/lib/harden.d/.
# Each module mitigates a specific container escape vector.
#
# To add a new hardening module: create a .sh file in harden.d/.
# It will be automatically sourced on every shell init.
#
# Sourced at line 1 of .bashrc BEFORE the interactive guard, because
# coding agents invoke bash as non-interactive login shells where
# anything after the "[ -z "$PS1" ] && return" guard never executes.
#
# Must complete quickly — VS Code's env probe times out after 10 seconds.
# =============================================================================

for _harden_module in /usr/local/lib/harden.d/*.sh; do
  [ -f "$_harden_module" ] && . "$_harden_module"
done
unset _harden_module
