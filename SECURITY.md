# Security Policy

## Supported versions

Security fixes are made for the latest release. Please upgrade to the newest
version before reporting an issue.

## Reporting a vulnerability

Please do **not** open a public issue for security problems.

Report vulnerabilities privately through
[GitHub private vulnerability reporting](https://github.com/jklaiber/ebpf-bridge/security/advisories/new).
You will get a response as soon as possible, and we will coordinate a fix and
disclosure with you.

## Security model

The `ebpf-bridge` service runs as root because loading and attaching XDP
programs requires it. It is controlled through the unix socket
`/run/ebpf-bridge.sock`, which is only accessible to root and members of the
socket's group (mode `0660`). Anyone who can write to the socket can bridge
arbitrary interfaces of the host, so do not relax these permissions.
