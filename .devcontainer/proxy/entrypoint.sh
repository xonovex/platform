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
# Generate squid ACLs from the shared domain allowlist
# ---------------------------------------------------------------------------
generate_squid_conf() {
  cat <<'SQUID_STATIC'
# Core Settings
http_port 3128
pid_filename none
max_filedesc 4096
coredump_dir /var/cache/squid

# Logging Configuration
logfile_rotate 0
access_log stdio:/proc/self/fd/1
cache_log stdio:/proc/self/fd/2

# Cache Settings (disabled — this is a forwarding proxy, not a cache)
cache deny all
cache_mem 0 MB

# SECURITY: Hide squid version string from error pages and responses.
# Prevents version-specific exploit targeting.
httpd_suppress_version_string on

# SECURITY: Strip proxy headers that leak internal network topology.
# - via off: removes Via header (reveals proxy hostname/version)
# - forwarded_for delete: removes X-Forwarded-For (reveals client IP)
via off
forwarded_for delete

# ACL Definitions
acl SSL_ports port 443
acl CONNECT method CONNECT

# SECURITY: Restrict to safe HTTP methods. Blocks TRACE (can leak
# headers), PURGE (cache manipulation), and other non-standard methods.
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

# Access Rules (order matters — first match wins)

# SECURITY: Block unsafe HTTP methods before any allow rules
http_access deny !SAFE_METHODS

# SECURITY: Only allow CONNECT (HTTPS tunneling) to port 443.
# Prevents CONNECT-based tunneling to arbitrary ports.
http_access deny CONNECT !SSL_ports

# Allow requests to domains in the allowlist
http_access allow allowed_domains

# SECURITY: Default deny — block everything not explicitly allowed
http_access deny all
SQUID_RULES
}

# ---------------------------------------------------------------------------
# Generate dnsmasq domain forwarding rules from the shared allowlist
# ---------------------------------------------------------------------------
generate_dnsmasq_conf() {
  cat <<EOF
# SECURITY: Block all domains by default. Any domain not in the
# allowlist returns NXDOMAIN immediately. This is defense-in-depth:
# even if a tool bypasses HTTP_PROXY/HTTPS_PROXY env vars, it cannot
# resolve unauthorized domains on the internal network. Also prevents
# DNS-based data exfiltration (encoding data in DNS queries to
# attacker-controlled domains).
address=/#/

# SECURITY: DNSSEC validation. Cryptographically verifies DNS responses
# using the root zone trust anchor (KSK-2017, key tag 20326). Prevents
# DNS cache poisoning attacks where an attacker injects forged responses
# to redirect traffic to malicious servers.
dnssec
trust-anchor=.,20326,8,2,E06D44B80B8F1D39A95C0B0D7C65D08458E880409BBC683457104237C7F8EC8D

# SECURITY: DNS rebinding protection. Rejects DNS responses that map
# public domain names to private IP addresses (10.x, 172.16.x, 192.168.x).
# Prevents attacks where an attacker's domain resolves to an internal IP,
# tricking the app into making requests to internal services.
# rebind-localhost-ok permits localhost (needed for some dev tools).
stop-dns-rebind
rebind-localhost-ok

# Selectively forward only allowed domains to Docker's upstream DNS.
# Dnsmasq matches subdomains automatically (server=/example.com/ also
# matches sub.example.com), so leading dots from the allowlist are
# stripped.
EOF

  while IFS= read -r line || [ -n "$line" ]; do
    line=$(echo "$line" | sed 's/#.*//' | tr -d '[:space:]')
    [ -z "$line" ] && continue
    # Strip leading dot — dnsmasq server directive matches subdomains
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
# Note: tmpfs is owned by root but squid user needs write access.
# The entrypoint runs as squid user, so the tmpfs mount options in
# docker-compose must not restrict this (no uid/gid set = world-writable).
cp /etc/proxy/mime.conf /etc/squid/mime.conf

generate_squid_conf > "$SQUID_CONF"
DOMAIN_COUNT=$(grep -v '^\s*#' "$ALLOWED_DOMAINS" | grep -v '^\s*$' | wc -l)
echo "Squid: generated ACL with ${DOMAIN_COUNT} domains"

generate_dnsmasq_conf > "$DNSMASQ_CONF"
echo "Dnsmasq: generated ${DOMAIN_COUNT} forwarding rules (all others -> NXDOMAIN)"

# ---------------------------------------------------------------------------
# Start dnsmasq (needs NET_BIND_SERVICE capability for port 53)
# ---------------------------------------------------------------------------
echo "Starting dnsmasq..."
dnsmasq --no-daemon --log-facility=/dev/stderr --conf-dir=/etc/dnsmasq.d &

# ---------------------------------------------------------------------------
# Start squid
# ---------------------------------------------------------------------------
if [ ! -d /var/cache/squid/00 ]; then
  echo "Initializing squid cache..."
  squid -N -z
fi

echo "Starting squid..."
exec squid -NYC
