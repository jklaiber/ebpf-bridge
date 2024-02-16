#!/bin/sh
systemctl daemon-reload
systemctl enable ebpf-bridge.service
systemctl start ebpf-bridge.service