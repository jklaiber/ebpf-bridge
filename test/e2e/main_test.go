//go:build e2e && linux

// Package e2e runs ebpf-bridge end to end: it builds the binary, starts the
// service and pushes real frames through XDP bridges between network
// namespaces. It needs Linux and root, e.g. `sudo make e2e` or, on macOS,
// `make container-e2e`.
package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

// bridgeBin is the ebpf-bridge binary under test.
var bridgeBin string

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	if os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "e2e tests need root (try `sudo make e2e` or `make container-e2e`)")
		return 1
	}
	if _, err := exec.LookPath("ip"); err != nil {
		fmt.Fprintln(os.Stderr, "e2e tests need iproute2 (`ip`)")
		return 1
	}

	bridgeBin = os.Getenv("EBPF_BRIDGE_BIN")
	if bridgeBin == "" {
		dir, err := os.MkdirTemp("", "ebpf-bridge-e2e")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		defer func() { _ = os.RemoveAll(dir) }()
		bridgeBin = filepath.Join(dir, "ebpf-bridge")
		args := []string{"build", "-o", bridgeBin}
		if os.Getenv("GOCOVERDIR") != "" {
			// The service and CLI processes inherit GOCOVERDIR and write
			// their coverage there (see `make e2e-coverage`).
			args = append(args, "-cover", "-coverpkg=github.com/jklaiber/ebpf-bridge/...")
		}
		build := exec.Command("go", append(args, "../../ebpf-bridge")...)
		build.Env = append(os.Environ(), "CGO_ENABLED=0")
		build.Stdout, build.Stderr = os.Stderr, os.Stderr
		if err := build.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "building ebpf-bridge:", err)
			return 1
		}
	}
	return m.Run()
}

// service is a running `ebpf-bridge start-service` process with its own socket.
type service struct {
	t      *testing.T
	socket string
	cmd    *exec.Cmd
	logs   *syncBuffer
	exited chan error
}

// startService starts a dedicated service for the test. Every test gets its
// own service, so tests are independent and can run in parallel.
func startService(t *testing.T) *service {
	t.Helper()
	return startServiceAt(t, filepath.Join(t.TempDir(), "ebpf-bridge.sock"))
}

// startServiceAt starts a service listening on socket.
func startServiceAt(t *testing.T, socket string) *service {
	t.Helper()
	s := &service{
		t:      t,
		socket: socket,
		logs:   &syncBuffer{},
		exited: make(chan error, 1),
	}
	s.cmd = exec.Command(bridgeBin, "start-service", "--socket", s.socket, "--log-level", "debug")
	s.cmd.Stdout, s.cmd.Stderr = s.logs, s.logs
	if err := s.cmd.Start(); err != nil {
		t.Fatalf("starting service: %v", err)
	}
	go func() { s.exited <- s.cmd.Wait() }()

	deadline := time.Now().Add(10 * time.Second)
	for {
		if _, _, err := s.cli("list"); err == nil {
			break
		}
		select {
		case err := <-s.exited:
			t.Fatalf("service exited early: %v\n%s", err, s.logs)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("service did not create its socket\n%s", s.logs)
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Cleanup(func() {
		if s.cmd.ProcessState == nil {
			s.stop(syscall.SIGTERM)
		}
		if t.Failed() {
			t.Logf("service logs:\n%s", s.logs)
		}
	})
	return s
}

// stop sends sig to the service and waits for it to exit. SIGTERM and
// SIGINT must result in a clean shutdown with exit code 0.
func (s *service) stop(sig syscall.Signal) {
	s.t.Helper()
	_ = s.cmd.Process.Signal(sig)
	select {
	case err := <-s.exited:
		if sig != syscall.SIGKILL && err != nil {
			s.t.Errorf("service did not shut down cleanly after %v: %v", sig, err)
		}
	case <-time.After(10 * time.Second):
		_ = s.cmd.Process.Kill()
		err := <-s.exited
		s.t.Errorf("service did not stop after %v, killed it: %v", sig, err)
	}
}

// cli runs an ebpf-bridge client command against the service.
func (s *service) cli(args ...string) (stdout, stderr string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bridgeBin, append([]string{"--socket", s.socket}, args...)...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	err = cmd.Run()
	return out.String(), errOut.String(), err
}

// mustCLI runs a client command and fails the test if it does not succeed.
func (s *service) mustCLI(args ...string) string {
	s.t.Helper()
	out, errOut, err := s.cli(args...)
	if err != nil {
		s.t.Fatalf("ebpf-bridge %s: %v\nstdout: %s\nstderr: %s", strings.Join(args, " "), err, out, errOut)
	}
	return out
}

// failCLI runs a client command that must fail and returns its stderr.
func (s *service) failCLI(args ...string) string {
	s.t.Helper()
	out, errOut, err := s.cli(args...)
	if err == nil {
		s.t.Fatalf("ebpf-bridge %s succeeded, want failure\nstdout: %s", strings.Join(args, " "), out)
	}
	return errOut
}

// addBridge adds a bridge between the h1 and h2 ports of topo, optionally
// mirroring to the monitor port.
func (s *service) addBridge(name string, topo *topology, withMonitor bool) {
	s.t.Helper()
	args := []string{"add", "--name", name, "--iface1", topo.portH1, "--iface2", topo.portH2}
	if withMonitor {
		args = append(args, "--monitor", topo.portMon)
	}
	s.mustCLI(args...)
}
