#!/bin/sh

mkdir -p /tmp/tailscale
echo "nameserver 100.100.100.100" >/etc/resolv.conf
/var/runtime/tailscaled --tun=userspace-networking --socks5-server=localhost:1055 &
/var/runtime/tailscale up --authkey=${TS_KEY} --hostname=aws-lambda-app
echo Tailscale started
ALL_PROXY=socks5://localhost:1055/ /var/runtime/main
