#!/bin/sh
systemctl disable --now ebpf-bridge.service || true
systemctl daemon-reload
