#!/bin/bash
# Wait for Grafana to be ready, then import dashboards

GRAFANA_URL="${GRAFANA_URL:-http://localhost:3000}"
GRAFANA_USER="${GRAFANA_USER:-admin}"
GRAFANA_PASSWORD="${GRAFANA_PASSWORD:-admin}"
DASHBOARDS_DIR="${DASHBOARDS_DIR:-/dashboards}"

# Wait for Grafana to be ready
echo "Waiting for Grafana to be ready..."
until curl -s -u "${GRAFANA_USER}:${GRAFANA_PASSWORD}" "${GRAFANA_URL}/api/health" | grep -q "ok"; do
  sleep 1
done
echo "Grafana is ready"

# Import all dashboards
for dashboard in "${DASHBOARDS_DIR}"/*.json; do
  if [ -f "$dashboard" ]; then
    name=$(basename "$dashboard" .json)
    echo "Importing dashboard: ${name}"
    curl -s -u "${GRAFANA_USER}:${GRAFANA_PASSWORD}" \
      "${GRAFANA_URL}/api/dashboards/db" \
      -X POST \
      -H "Content-Type: application/json" \
      -d "{\"dashboard\": $(cat "$dashboard"), \"overwrite\": true}" \
      | grep -q '"status":"success"' && echo "  Success" || echo "  Failed"
  fi
done

echo "Dashboard import complete"
