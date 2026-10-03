# Operators

Resolve operator toolchains through the plugin interface and keep concrete implementations out of controllers.

- Resolve operator toolchains through `internal/plugins.ResolveToolchain`; controllers and pod hardening must not name concrete toolchains.
