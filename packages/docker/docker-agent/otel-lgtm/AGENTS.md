# AI Agent OTEL LGTM

OpenTelemetry + Grafana LGTM stack for Claude Code monitoring.

## Structure

- `dashboards/` — Grafana dashboard JSON definitions
- `provisioning/` — Grafana provisioning configuration

## Workflow

- **Config mounts**: Custom configs mount to `/otel-lgtm/` (not `/etc/`)
- **Dashboard import**: Init container imports dashboards via Grafana API on startup
- **Metrics flow**: Claude Code → OTLP (4317) → OTEL Collector → Prometheus exporter (8889) → Prometheus scrape → Grafana query

## Dashboard Style

- **Colors**: Default Grafana palette only (`palette-classic`), no custom overrides
- **Stat panels**: `colorMode: "none"`, plain text values, no colored backgrounds
- **Thresholds**: None for cost/tokens (unlimited plan, informational only)
- **Charts**: Smooth interpolation, gradient fills, donut charts, `thresholdsStyle: "off"`
- **Philosophy**: Clean, minimal, information-focused—no visual indicators implying cost judgments

## Integration

- **Upstream**: Claude Code (OTLP metrics and logs)
- **Downstream**: Grafana (port 3000), Prometheus (port 9090)
