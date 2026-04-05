#!/bin/sh
set -e

sudo /bin/chown -R squid:squid /var/log/squid
sudo /bin/chown -R squid:squid /var/cache/squid

# Start dnsmasq as a DNS forwarder for containers on the internal network.
# It forwards to the DNS servers in /etc/resolv.conf (Docker's upstream DNS).
sudo dnsmasq --no-daemon --log-queries --log-facility=/dev/stderr &

if [ -z "${1}" ]; then
  if [ ! -d /var/cache/squid/00 ]; then
    echo "Initializing cache..."
    $(which squid) -N -z
  fi
  echo "Starting squid..."
  exec $(which squid) -NYC
else
  exec "$@"
fi
