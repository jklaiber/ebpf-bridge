//go:build e2e && linux

package e2e

import (
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestForwarding checks that every kind of frame is forwarded between the
// bridged hosts and mirrored to the monitor, in both directions, and never
// reflected back to the sender.
func TestForwarding(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo := newTopology(t)
	ep := openEndpoints(t, topo)
	svc.addBridge("fwd", topo, true)

	names := map[*packetSocket]string{ep.h1: "h1", ep.h2: "h2", ep.mon: "mon"}
	frames := append(append([]frame{}, controlFrames...), dataFrames...)
	for _, f := range frames {
		t.Run(f.name+"/h1->h2", func(t *testing.T) {
			expectFrame(t, f, ep.h1, map[*packetSocket]bool{ep.h2: true, ep.mon: true, ep.h1: false}, names)
		})
		t.Run(f.name+"/h2->h1", func(t *testing.T) {
			expectFrame(t, f, ep.h2, map[*packetSocket]bool{ep.h1: true, ep.mon: true, ep.h2: false}, names)
		})
	}
}

// TestMonitorIsReceiveOnly checks that frames sent from the monitor port are
// not injected into the bridge.
func TestMonitorIsReceiveOnly(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo := newTopology(t)
	ep := openEndpoints(t, topo)
	svc.addBridge("mon", topo, true)

	names := map[*packetSocket]string{ep.h1: "h1", ep.h2: "h2", ep.mon: "mon"}
	for _, f := range []frame{controlFrames[0], dataFrames[0]} {
		expectFrame(t, f, ep.mon, map[*packetSocket]bool{ep.h1: false, ep.h2: false}, names)
	}
}

// TestWithoutMonitor checks a bridge without monitor interface.
func TestWithoutMonitor(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo := newTopology(t)
	ep := openEndpoints(t, topo)
	svc.addBridge("nomon", topo, false)

	names := map[*packetSocket]string{ep.h1: "h1", ep.h2: "h2", ep.mon: "mon"}
	expectFrame(t, controlFrames[0], ep.h1, map[*packetSocket]bool{ep.h2: true, ep.mon: false}, names)
	expectFrame(t, dataFrames[0], ep.h2, map[*packetSocket]bool{ep.h1: true, ep.mon: false}, names)
}

// TestIPConnectivity checks ARP and IP traffic, including full-size packets.
func TestIPConnectivity(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo := newTopology(t)
	topo.setIP(topo.h1, "10.10.0.1/24")
	topo.setIP(topo.h2, "10.10.0.2/24")

	if topo.ping(topo.h1, "10.10.0.2") {
		t.Fatal("h1 reached h2 before a bridge was added")
	}
	svc.addBridge("ip", topo, true)
	if !topo.ping(topo.h1, "10.10.0.2") {
		t.Error("h1 cannot ping h2")
	}
	if !topo.ping(topo.h2, "10.10.0.1") {
		t.Error("h2 cannot ping h1")
	}
	// 1472 bytes ICMP payload + 28 bytes headers = 1500 bytes MTU, no fragmentation.
	if !topo.ping(topo.h1, "10.10.0.2", "-s", "1472", "-M", "do") {
		t.Error("full MTU ping from h1 to h2 failed")
	}
}

// TestLifecycle checks add, list and remove, and that traffic stops after
// removal and resumes after adding the bridge again.
func TestLifecycle(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo := newTopology(t)
	topo.setIP(topo.h1, "10.20.0.1/24")
	topo.setIP(topo.h2, "10.20.0.2/24")

	if out := svc.mustCLI("list"); !strings.Contains(out, "No bridges found") {
		t.Errorf("list on empty service = %q", out)
	}

	svc.addBridge("life", topo, true)
	out := svc.mustCLI("list")
	for _, want := range []string{"life", topo.portH1, topo.portH2, topo.portMon} {
		if !strings.Contains(out, want) {
			t.Errorf("list output does not contain %q:\n%s", want, out)
		}
	}
	if !topo.xdpAttached(topo.portH1) || !topo.xdpAttached(topo.portH2) {
		t.Error("XDP program not attached to the bridged ports")
	}
	if topo.xdpAttached(topo.portMon) {
		t.Error("XDP program attached to the monitor port")
	}
	if !topo.ping(topo.h1, "10.20.0.2") {
		t.Fatal("no connectivity with bridge")
	}

	svc.mustCLI("remove", "--name", "life")
	if topo.xdpAttached(topo.portH1) || topo.xdpAttached(topo.portH2) {
		t.Error("XDP program still attached after remove")
	}
	if topo.ping(topo.h1, "10.20.0.2") {
		t.Error("connectivity after remove")
	}
	if out := svc.mustCLI("list"); !strings.Contains(out, "No bridges found") {
		t.Errorf("list after remove = %q", out)
	}

	svc.addBridge("life", topo, false)
	if !topo.ping(topo.h1, "10.20.0.2") {
		t.Error("no connectivity after adding the bridge again")
	}
}

// TestMultipleBridgesAreIsolated checks that bridges managed by the same
// service do not leak frames into each other.
func TestMultipleBridgesAreIsolated(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	a, b := newTopology(t), newTopology(t)
	epA, epB := openEndpoints(t, a), openEndpoints(t, b)
	svc.addBridge("a", a, true)
	svc.addBridge("b", b, true)

	names := map[*packetSocket]string{
		epA.h1: "a/h1", epA.h2: "a/h2", epA.mon: "a/mon",
		epB.h1: "b/h1", epB.h2: "b/h2", epB.mon: "b/mon",
	}
	for _, f := range []frame{controlFrames[0], dataFrames[0]} {
		expectFrame(t, f, epA.h1, map[*packetSocket]bool{
			epA.h2: true, epA.mon: true, epB.h1: false, epB.h2: false, epB.mon: false,
		}, names)
		expectFrame(t, f, epB.h1, map[*packetSocket]bool{
			epB.h2: true, epB.mon: true, epA.h1: false, epA.h2: false, epA.mon: false,
		}, names)
	}

	// Removing one bridge must not affect the other.
	svc.mustCLI("remove", "--name", "a")
	expectFrame(t, dataFrames[0], epA.h1, map[*packetSocket]bool{epA.h2: false}, names)
	expectFrame(t, dataFrames[0], epB.h1, map[*packetSocket]bool{epB.h2: true}, names)
}

// TestInvalidRequests checks that bad requests fail with a useful error and a
// non-zero exit code, and leave existing bridges untouched.
func TestInvalidRequests(t *testing.T) {
	t.Parallel()
	svc := startService(t)
	topo, other := newTopology(t), newTopology(t)
	svc.addBridge("ok", topo, false)

	cases := []struct {
		name    string
		args    []string
		wantErr string
	}{
		{"unknown interface", []string{"add", "-n", "x", "--iface1", "does-not-exist", "--iface2", other.portH2}, "does-not-exist"},
		{"duplicate name", []string{"add", "-n", "ok", "--iface1", other.portH1, "--iface2", other.portH2}, "already exists"},
		{"interface in use", []string{"add", "-n", "x", "--iface1", topo.portH1, "--iface2", other.portH2}, "already used"},
		{"same interface twice", []string{"add", "-n", "x", "--iface1", other.portH1, "--iface2", other.portH1}, "only be used once"},
		{"monitor equals port", []string{"add", "-n", "x", "--iface1", other.portH1, "--iface2", other.portH2, "--monitor", other.portH1}, "only be used once"},
		{"missing flag", []string{"add", "-n", "x", "--iface1", other.portH1}, "iface2"},
		{"remove unknown", []string{"remove", "-n", "nope"}, "does not exist"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if stderr := svc.failCLI(tc.args...); !strings.Contains(stderr, tc.wantErr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tc.wantErr)
			}
		})
	}

	out := svc.mustCLI("list")
	if !strings.Contains(out, "ok") || strings.Contains(out, " x ") {
		t.Errorf("unexpected bridges after invalid requests:\n%s", out)
	}
}

// TestServiceUnavailable checks the CLI error when no service is running.
func TestServiceUnavailable(t *testing.T) {
	t.Parallel()
	svc := &service{t: t, socket: t.TempDir() + "/missing.sock"}
	if stderr := svc.failCLI("list"); !strings.Contains(stderr, "not reachable") {
		t.Errorf("stderr = %q, want a 'not reachable' error", stderr)
	}
}

// TestShutdownDetaches checks that no XDP program stays attached when the
// service stops, whether it shuts down cleanly or is killed.
func TestShutdownDetaches(t *testing.T) {
	t.Parallel()
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL} {
		t.Run(sig.String(), func(t *testing.T) {
			t.Parallel()
			svc := startService(t)
			topo := newTopology(t)
			topo.setIP(topo.h1, "10.30.0.1/24")
			topo.setIP(topo.h2, "10.30.0.2/24")
			svc.addBridge("down", topo, true)
			if !topo.ping(topo.h1, "10.30.0.2") {
				t.Fatal("no connectivity with bridge")
			}

			svc.stop(sig)
			// The kernel releases the XDP links of a killed process asynchronously.
			deadline := time.Now().Add(5 * time.Second)
			for topo.xdpAttached(topo.portH1) || topo.xdpAttached(topo.portH2) {
				if time.Now().After(deadline) {
					t.Fatal("XDP program still attached after the service stopped")
				}
				time.Sleep(50 * time.Millisecond)
			}
			if topo.ping(topo.h1, "10.30.0.2") {
				t.Error("connectivity after the service stopped")
			}
		})
	}
}

// TestServiceRestart checks that a service killed without cleanup can be
// restarted on the same socket and bridge the same interfaces again.
func TestServiceRestart(t *testing.T) {
	t.Parallel()
	topo := newTopology(t)
	topo.setIP(topo.h1, "10.40.0.1/24")
	topo.setIP(topo.h2, "10.40.0.2/24")

	first := startService(t)
	first.addBridge("again", topo, true)
	first.stop(syscall.SIGKILL) // leaves a stale socket behind

	restarted := startServiceAt(t, first.socket)
	restarted.addBridge("again", topo, true)
	if !topo.ping(topo.h1, "10.40.0.2") {
		t.Error("no connectivity after restart")
	}
}
