# OTEL LGTM

- Config mounts → `/otel-lgtm/`; dashboard import via Grafana API init container
- Claude Code → OTLP (4317) → OTEL Collector → Prometheus (8889) → Grafana (3000)
- Default palette only (`palette-classic`), no custom overrides
- Stat panels: `colorMode: "none"`, plain text, no colored backgrounds
- No thresholds for cost/tokens (informational); charts: smooth, gradient, donut, `thresholdsStyle: "off"`
