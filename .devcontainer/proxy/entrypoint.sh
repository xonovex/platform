#!/bin/sh
set -e

# =============================================================================
# Proxy Entrypoint — Config Generation and Service Startup
#
# Generates squid and dnsmasq configs from a shared domain allowlist,
# then starts both services. This ensures a single source of truth:
# edit allowed-domains.conf, and both HTTP proxy filtering and DNS
# filtering update automatically.
# =============================================================================

ALLOWED_DOMAINS="/etc/proxy/allowed-domains.conf"
SQUID_CONF="/etc/squid/squid.conf"
DNSMASQ_CONF="/etc/dnsmasq.d/allowed-domains.conf"

# Docker's embedded DNS resolver. Resolves both container names and
# external domains (forwarded to host DNS). Dnsmasq selectively forwards
# allowed domains here; everything else returns NXDOMAIN.
UPSTREAM_DNS="127.0.0.11"

# ---------------------------------------------------------------------------
# Generate squid config from the shared domain allowlist
# ---------------------------------------------------------------------------
generate_squid_conf() {
  cat <<'SQUID_STATIC'
# === Core Settings ===
http_port 3128
pid_filename none
max_filedesc 4096
coredump_dir /var/cache/squid

# === Logging ===
# Enhanced format with security-relevant fields: timestamp, response time,
# client IP, status, size, method, URL, server IP, content type, user agent.
logfile_rotate 0
logformat security %ts.%03tu %6tr %>a:%>p %Ss/%03>Hs %<st %rm %ru %Sh/%<a %mt "%{User-Agent}>h"
access_log stdio:/proc/self/fd/1 security
cache_log stdio:/proc/self/fd/2

# === Cache (disabled — forwarding proxy only) ===
cache deny all
cache_mem 0 MB

# === Timeout Hardening ===
# Tightened from defaults to reclaim resources and fail fast.
# Default connect_timeout is 1 minute — too slow for a devcontainer.
connect_timeout 15 seconds
# Close idle server connections after 5 min (default: 15 min)
read_timeout 5 minutes
# Max total lifetime of any client connection (default: 1 day)
client_lifetime 2 hours
# Idle keep-alive timeout (default: 1 min)
persistent_request_timeout 30 seconds
# On shutdown, wait this long for active connections (default: 30s)
shutdown_lifetime 5 seconds
# Time to wait for a complete HTTP request (default: 5 min)
request_timeout 1 minute

# === Size Limits ===
# Prevents exfiltration of large data blobs through the proxy.
# 50 MB covers npm publish, git push, Docker layer uploads.
request_body_max_size 50 MB
request_header_max_size 64 KB
reply_header_max_size 64 KB
# reply_body_max_size not set — npm packages and Docker layers can be 100s of MB

# === Connection Limits ===
# Prevents runaway processes from exhausting file descriptors.
# 128 is generous for a single devcontainer client.
client_ip_max_connections 128

# === Disable Unused Protocols ===
# ICP/HTCP are inter-cache protocols for cache hierarchies.
# Not needed for a standalone proxy. Disabling eliminates attack surface
# including CVE-2026-32748 (Use-After-Free in ICP, affects squid <= 7.4).
icp_port 0
htcp_port 0

# === Resource Management ===
# Close half-closed connections immediately (default: keep open)
half_closed_clients off
# Disable memory pooling — smaller footprint, no stale-data risk in
# recycled buffers. Acceptable for a devcontainer.
memory_pools off

# === Security Hardening ===
# Hide squid version from error pages and responses.
httpd_suppress_version_string on
# Strip proxy headers that leak internal network topology.
via off
forwarded_for delete
# Strip additional proxy-revealing headers from responses.
reply_header_access X-Cache deny all
reply_header_access X-Cache-Lookup deny all
reply_header_access X-Squid-Error deny all
# Belt-and-suspenders with forwarded_for delete.
request_header_access X-Forwarded-For deny all

# === ACL Definitions ===
acl SSL_ports port 443
acl CONNECT method CONNECT

# Restrict to safe HTTP methods. CONNECT is included because all HTTPS
# requests use it for proxy tunneling. Blocks TRACE (header leakage),
# PURGE (cache manipulation), and non-standard methods.
acl SAFE_METHODS method GET POST PUT PATCH DELETE HEAD OPTIONS CONNECT

# Generated domain ACL from allowed-domains.conf
SQUID_STATIC

  # Collect all domains into a single ACL line
  printf "acl allowed_domains dstdomain"
  while IFS= read -r line || [ -n "$line" ]; do
    line=$(echo "$line" | sed 's/#.*//' | tr -d '[:space:]')
    [ -z "$line" ] && continue
    printf " %s" "$line"
  done < "$ALLOWED_DOMAINS"
  printf "\n"

  cat <<'SQUID_RULES'

# === Access Rules (order matters — first match wins) ===

# Deny cache manager access (exposes config, stats, shutdown capability)
http_access deny manager
# Block unsafe HTTP methods
http_access deny !SAFE_METHODS
# Only allow CONNECT (HTTPS tunneling) to port 443
http_access deny CONNECT !SSL_ports
# Allow requests to domains in the allowlist
http_access allow allowed_domains
# Default deny — block everything not explicitly allowed
http_access deny all
SQUID_RULES
}

# ---------------------------------------------------------------------------
# Generate dnsmasq config from the shared domain allowlist
# ---------------------------------------------------------------------------
generate_dnsmasq_conf() {
  cat <<EOF
# === Network Exposure ===
# Only accept queries from directly-connected subnets (rejects remote).
local-service
# Bind to specific interfaces, not wildcard.
bind-interfaces
# Do not read /etc/resolv.conf — use only explicit server= directives.
# Prevents container orchestration from injecting unexpected resolvers.
no-resolv
no-poll

# === Query Leakage Prevention ===
# Never forward unqualified names (no domain part) upstream.
domain-needed
# Answer private reverse lookups locally, never forward to upstream.
bogus-priv
# Filter useless Windows periodic DNS requests.
filterwin2k
# Do not read /etc/hosts — prevents a compromised hosts file from
# influencing resolution.
no-hosts

# === DNSSEC ===
# Cryptographically verify DNS responses using the root zone trust
# anchor (KSK-2017, key tag 20326). Prevents cache poisoning.
dnssec
trust-anchor=.,20326,8,2,E06D44B80B8F1D39A95C0B0D7C65D08458E880409BBC683457104237C7F8EC8D
# Verify that unsigned responses are legitimately unsigned, not stripped.
dnssec-check-unsigned

# === Cache Poisoning Prevention ===
# EDNS UDP packet size set to 1232 (DNS Flag Day 2020 recommendation).
# Prevents fragmentation-based spoofing attacks (CVE-2023-28450).
edns-packet-max=1232
# Source port randomization is on by default — never pin query-port.

# === DNS Rebinding Protection ===
# Reject responses mapping public domains to private IPs (10.x, 172.16.x,
# 192.168.x). Prevents tricking apps into hitting internal services.
stop-dns-rebind
rebind-localhost-ok

# === Resource Limits ===
# Limit concurrent forwarded queries (default: 150). Lower value reduces
# the window for birthday-attack-style cache poisoning.
dns-forward-max=50
# Cache size — balance between performance and poisoning surface.
cache-size=1000
# Clamp TTLs to prevent rapid churn attacks and excessively stale data.
min-cache-ttl=60
max-cache-ttl=3600

# === Logging ===
# Extra format includes requestor IP and serial number for correlation.
log-queries=extra
# Log to stderr (container runtime handles collection).
log-facility=-
# Async logging to prevent blocking under load.
log-async=25

# === Default Block ===
# Return NXDOMAIN for all domains not explicitly forwarded below.
address=/#/

# === Allowed Domain Forwarding ===
# Selectively forward only allowed domains to Docker's upstream DNS.
# Dnsmasq matches subdomains automatically.
EOF

  while IFS= read -r line || [ -n "$line" ]; do
    line=$(echo "$line" | sed 's/#.*//' | tr -d '[:space:]')
    [ -z "$line" ] && continue
    domain=$(echo "$line" | sed 's/^\.//')
    echo "server=/${domain}/${UPSTREAM_DNS}"
  done < "$ALLOWED_DOMAINS"
}

# ---------------------------------------------------------------------------
# Generate configs from the shared allowlist
# ---------------------------------------------------------------------------
echo "=== Generating configs from allowed-domains.conf ==="

mkdir -p /etc/dnsmasq.d

# Restore mime.conf into the tmpfs. The read_only + tmpfs overlay on
# /etc/squid wipes all files from the image layer. mime.conf is required
# by squid for MIME type handling — without it, squid aborts on startup.
cp /etc/proxy/mime.conf /etc/squid/mime.conf

generate_squid_conf > "$SQUID_CONF"
DOMAIN_COUNT=$(grep -v '^\s*#' "$ALLOWED_DOMAINS" | grep -v '^\s*$' | wc -l)
echo "Squid: generated ACL with ${DOMAIN_COUNT} domains"

generate_dnsmasq_conf > "$DNSMASQ_CONF"
echo "Dnsmasq: generated ${DOMAIN_COUNT} forwarding rules (all others -> NXDOMAIN)"

# ---------------------------------------------------------------------------
# Start dnsmasq (binds port 53 via setcap cap_net_bind_service)
# ---------------------------------------------------------------------------
echo "Starting dnsmasq..."
dnsmasq --no-daemon --conf-dir=/etc/dnsmasq.d &

# ---------------------------------------------------------------------------
# Start squid
# ---------------------------------------------------------------------------
if [ ! -d /var/cache/squid/00 ]; then
  echo "Initializing squid cache..."
  squid -N -z
fi

echo "Starting squid..."
exec squid -NYC
