# =============================================================================
# Proxy Container — Squid HTTP Proxy + Dnsmasq DNS Forwarder
#
# This is the security boundary between the app and the internet.
# All HTTP/HTTPS traffic and DNS queries from the app pass through here.
#
# Runs as the squid user (non-root) with cap_drop ALL. dnsmasq binds
# port 53 via a setcap file capability (cap_net_bind_service), not via
# Docker's cap_add — keeping the bounding set completely empty.
#
# no-new-privileges is NOT used on this container because squid 7.x
# requires setgid/initgroups for user switching, and setcap file
# capabilities also require the ability to gain privileges at execve.
# This is acceptable because: cap_drop ALL empties the bounding set,
# read_only prevents binary modification, and Alpine ships no setuid
# binaries.
# =============================================================================

# SECURITY: Pinned to specific digest to prevent supply chain attacks.
# Floating tags silently change — a compromised tag could inject malware.
# Update digest periodically after verifying the new image.
FROM alpine@sha256:25109184c71bdad752c8312a8623239686a9a2071e8825f20acb8f2198c3f659

# dnsmasq-dnssec (not plain dnsmasq) includes DNSSEC validation support.
# Alpine's base dnsmasq package is compiled without it.
# libcap-utils provides setcap for granting file capabilities.
RUN apk update && \
    apk add --no-cache ca-certificates squid tzdata dnsmasq-dnssec libcap-utils && \
    rm -rf /var/cache/apk/*

# SECURITY: Grant dnsmasq the ability to bind port 53 as non-root.
# This is a file capability (stored in the binary's extended attributes),
# not a process capability — it only takes effect when this specific
# binary is executed. Combined with read_only: true, the binary cannot
# be replaced with a malicious one that inherits this capability.
RUN setcap cap_net_bind_service=+ep /usr/sbin/dnsmasq

# Pre-create directories. At runtime these are overlaid with tmpfs
# owned by uid=31 (squid) — see docker-compose.yml. The paths must
# exist in the image for squid -z (cache initialization) to succeed.
RUN mkdir -p /var/log/squid /var/cache/squid /run/squid /etc/dnsmasq.d /etc/proxy

# SECURITY: Preserve squid's default mime.conf in a location that
# survives the tmpfs overlay. The read_only + tmpfs over /etc/squid
# wipes all files including mime.conf, which squid requires to start.
# The entrypoint restores it at startup.
RUN cp /etc/squid/mime.conf /etc/proxy/mime.conf

# Domain allowlist — single source of truth for both squid ACLs and
# dnsmasq forwarding rules. The entrypoint generates both configs from
# this file. Mounted read-only at runtime via docker-compose.
COPY allowed-domains.conf /etc/proxy/allowed-domains.conf

COPY entrypoint.sh /entrypoint.sh
RUN chmod 755 /entrypoint.sh

EXPOSE 3128 53/udp 53/tcp

# Run as non-root squid user. Squid 7.x refuses to run as root.
USER squid

CMD ["/entrypoint.sh"]
