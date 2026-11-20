# Operators

Resolve operator toolchains through the plugin interface and keep concrete implementations out of controllers.

- Resolve operator toolchains through the injected `ToolchainResolver` port that `internal/wiring` binds to `plugins.Registry`; controllers and pod hardening must not name concrete toolchains.
