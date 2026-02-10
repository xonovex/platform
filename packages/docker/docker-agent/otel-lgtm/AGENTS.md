# AI Agent OTEL LGTM

OpenTelemetry + Grafana LGTM stack for Claude Code monitoring.

## Structure

- `dashboards/` — Grafana dashboard JSON
- `provisioning/` — Grafana provisioning config

## Workflow

- Config mounts → `/otel-lgtm/` (not `/etc/`)
- Dashboard import → init container via Grafana API on startup
- Metrics → Claude Code → OTLP (4317) → OTEL Collector → Prometheus (8889) → Grafana

## Dashboard Style

- Default Grafana palette only (`palette-classic`), no custom overrides
- Stat panels: `colorMode: "none"`, plain text, no colored backgrounds
- No thresholds for cost/tokens (informational only)
- Charts: smooth interpolation, gradient fills, donut, `thresholdsStyle: "off"`
- Clean, minimal, information-focused — no cost judgment indicators

## Integration

- Upstream → Claude Code (OTLP metrics and logs)
- Downstream → Grafana (3000), Prometheus (9090)

## Guidelines

- See [docker-guidelines](../../../guide/guide-docker/index.md)
- See [shell-scripting-guidelines](../../../guide/guide-shell-scripting/index.md)
