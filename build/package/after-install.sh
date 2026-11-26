#!/bin/sh
set -e
systemctl daemon-reload
systemctl enable --now ebpf-bridge.service
