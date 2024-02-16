#!/bin/sh
systemctl stop ebpf-bridge.service
systemctl disable ebpf-bridge.service
systemctl daemon-reload