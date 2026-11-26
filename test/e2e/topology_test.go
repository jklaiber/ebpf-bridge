//go:build e2e && linux

package e2e

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"os/exec"
	"strings"
	"sync"
	"testing"
)

// topology is three hosts, each in its own network namespace, connected by
// veth pairs to ports in the root namespace where ebpf-bridge runs:
//
//	h1  eth0 <-> portH1 \
//	                      ebpf-bridge
//	h2  eth0 <-> portH2 /     |
//	mon eth0 <-> portMon -----+ (monitor copy)
type topology struct {
	t                       *testing.T
	h1, h2, mon             string // namespace names
	portH1, portH2, portMon string // bridge-side interface names
}

func newTopology(t *testing.T) *topology {
	t.Helper()
	id := randomID()
	topo := &topology{
		t:       t,
		h1:      "e2e-" + id + "-h1",
		h2:      "e2e-" + id + "-h2",
		mon:     "e2e-" + id + "-mon",
		portH1:  "eb" + id + "a",
		portH2:  "eb" + id + "b",
		portMon: "eb" + id + "m",
	}
	t.Cleanup(topo.destroy)
	topo.addHost(topo.h1, topo.portH1)
	topo.addHost(topo.h2, topo.portH2)
	topo.addHost(topo.mon, topo.portMon)
	return topo
}

func (topo *topology) addHost(ns, port string) {
	topo.t.Helper()
	ip(topo.t, "netns", "add", ns)
	ip(topo.t, "link", "add", port, "type", "veth", "peer", "name", "eth0", "netns", ns)
	// Keep the test traffic free of IPv6 autoconfiguration noise.
	sysctl(topo.t, "", "net.ipv6.conf."+port+".disable_ipv6=1")
	sysctl(topo.t, ns, "net.ipv6.conf.all.disable_ipv6=1")
	ip(topo.t, "link", "set", port, "up")
	ip(topo.t, "-n", ns, "link", "set", "lo", "up")
	ip(topo.t, "-n", ns, "link", "set", "eth0", "up")
}

// setIP assigns addr (CIDR) to eth0 in namespace ns.
func (topo *topology) setIP(ns, addr string) {
	topo.t.Helper()
	ip(topo.t, "-n", ns, "addr", "add", addr, "dev", "eth0")
}

// ping sends a few ICMP echo requests from ns to dst and reports success.
func (topo *topology) ping(ns, dst string, extraArgs ...string) bool {
	args := append([]string{"netns", "exec", ns, "ping", "-c", "3", "-i", "0.2", "-W", "1"}, extraArgs...)
	return exec.Command("ip", append(args, dst)...).Run() == nil
}

// xdpAttached reports whether an XDP program is attached to port.
func (topo *topology) xdpAttached(port string) bool {
	out := ip(topo.t, "link", "show", "dev", port)
	return strings.Contains(out, "xdp")
}

func (topo *topology) destroy() {
	for _, port := range []string{topo.portH1, topo.portH2, topo.portMon} {
		_ = exec.Command("ip", "link", "del", port).Run()
	}
	for _, ns := range []string{topo.h1, topo.h2, topo.mon} {
		_ = exec.Command("ip", "netns", "del", ns).Run()
	}
}

func ip(t *testing.T, args ...string) string {
	t.Helper()
	var out, errOut bytes.Buffer
	cmd := exec.Command("ip", args...)
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		t.Fatalf("ip %s: %v: %s", strings.Join(args, " "), err, errOut.String())
	}
	return out.String()
}

func sysctl(t *testing.T, ns, setting string) {
	t.Helper()
	args := []string{"sysctl", "-qw", setting}
	if ns != "" {
		args = append([]string{"netns", "exec", ns}, args...)
		_ = exec.Command("ip", args...).Run()
		return
	}
	_ = exec.Command(args[0], args[1:]...).Run()
}

func randomID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// syncBuffer is a bytes.Buffer that is safe for concurrent writes.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
