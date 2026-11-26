---
name: Bug report
about: Report something that does not work as expected
title: ''
labels: bug
assignees: ''

---

**Describe the bug**
A clear and concise description of what the bug is.

**To reproduce**
The commands you ran (for example `ebpf-bridge add ...`), and their output.

```
```

**Expected behavior**
What you expected to happen instead.

**Service logs**
Output of `journalctl -u ebpf-bridge` or `docker logs <container>`.

```
```

**Environment**
 - ebpf-bridge version (`ebpf-bridge --version`):
 - Installation method (deb/rpm package, Docker, from source):
 - Distribution and version:
 - Kernel version (`uname -r`):
 - Interface types involved (physical NIC/driver, veth, VLAN, ...):

**Additional context**
Anything else that might help, for example the network topology.
