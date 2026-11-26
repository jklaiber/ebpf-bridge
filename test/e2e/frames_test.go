//go:build e2e && linux

package e2e

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/vishvananda/netns"
	"golang.org/x/sys/unix"
)

// frame describes an Ethernet frame to send through a bridge.
type frame struct {
	name string
	dst  net.HardwareAddr
	// etherType is the EtherType, or the 802.3 length if < 0x0600.
	etherType uint16
	// vlan adds an 802.1Q tag with this VLAN ID when non-zero.
	vlan uint16
	// size pads the payload so the frame (without FCS) has this length.
	size int
}

func mac(s string) net.HardwareAddr {
	m, err := net.ParseMAC(s)
	if err != nil {
		panic(err)
	}
	return m
}

var (
	srcMAC   = mac("02:00:00:00:00:01")
	expEther = uint16(0x88b5) // IEEE 802 local experimental EtherType
)

// controlFrames are frames a standard Linux bridge does not forward (or
// only forwards with special configuration). ebpf-bridge must pass them all.
var controlFrames = []frame{
	{name: "LLDP", dst: mac("01:80:c2:00:00:0e"), etherType: 0x88cc},
	{name: "STP BPDU", dst: mac("01:80:c2:00:00:00"), etherType: 38}, // 802.3 length + LLC
	{name: "LACP", dst: mac("01:80:c2:00:00:02"), etherType: 0x8809},
	{name: "802.1X EAPOL", dst: mac("01:80:c2:00:00:03"), etherType: 0x888e},
	{name: "LLDP nearest non-TPMR bridge", dst: mac("01:80:c2:00:00:03"), etherType: 0x88cc},
	{name: "Cisco CDP", dst: mac("01:00:0c:cc:cc:cc"), etherType: 100},
}

var dataFrames = []frame{
	{name: "broadcast", dst: mac("ff:ff:ff:ff:ff:ff"), etherType: expEther},
	{name: "unicast", dst: mac("02:00:00:00:00:99"), etherType: expEther},
	{name: "IPv4 multicast", dst: mac("01:00:5e:00:00:fb"), etherType: expEther},
	{name: "802.1Q tagged", dst: mac("02:00:00:00:00:99"), etherType: expEther, vlan: 42},
	{name: "802.1Q tagged LLDP", dst: mac("01:80:c2:00:00:0e"), etherType: 0x88cc, vlan: 7},
	{name: "full MTU", dst: mac("02:00:00:00:00:99"), etherType: expEther, size: 1514},
}

// build returns the raw frame bytes carrying marker in its payload.
func (f frame) build(marker []byte) []byte {
	var b bytes.Buffer
	b.Write(f.dst)
	b.Write(srcMAC)
	if f.vlan != 0 {
		_ = binary.Write(&b, binary.BigEndian, uint16(0x8100))
		_ = binary.Write(&b, binary.BigEndian, f.vlan)
	}
	_ = binary.Write(&b, binary.BigEndian, f.etherType)
	b.Write(marker)
	size := max(f.size, 60)
	for b.Len() < size {
		b.WriteByte(0)
	}
	return b.Bytes()
}

// packetSocket is a raw AF_PACKET socket bound to an interface inside a
// network namespace.
type packetSocket struct {
	fd int
}

// openPacketSocket opens a raw socket on iface in namespace ns. The socket
// stays bound to that namespace after the thread switches back.
func openPacketSocket(t *testing.T, ns, iface string) *packetSocket {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	orig, err := netns.Get()
	if err != nil {
		t.Fatalf("get current netns: %v", err)
	}
	defer func() { _ = orig.Close() }()
	target, err := netns.GetFromName(ns)
	if err != nil {
		t.Fatalf("get netns %s: %v", ns, err)
	}
	defer func() { _ = target.Close() }()
	if err := netns.Set(target); err != nil {
		t.Fatalf("enter netns %s: %v", ns, err)
	}
	defer func() {
		if err := netns.Set(orig); err != nil {
			panic("cannot restore network namespace: " + err.Error())
		}
	}()

	ifi, err := net.InterfaceByName(iface)
	if err != nil {
		t.Fatalf("interface %s in %s: %v", iface, ns, err)
	}
	proto := htons(unix.ETH_P_ALL)
	fd, err := unix.Socket(unix.AF_PACKET, unix.SOCK_RAW|unix.SOCK_CLOEXEC, int(proto))
	if err != nil {
		t.Fatalf("open packet socket: %v", err)
	}
	if err := unix.Bind(fd, &unix.SockaddrLinklayer{Protocol: proto, Ifindex: ifi.Index}); err != nil {
		_ = unix.Close(fd)
		t.Fatalf("bind packet socket to %s: %v", iface, err)
	}
	s := &packetSocket{fd: fd}
	t.Cleanup(func() { _ = unix.Close(fd) })
	return s
}

func (s *packetSocket) send(t *testing.T, data []byte) {
	t.Helper()
	if _, err := unix.Write(s.fd, data); err != nil {
		t.Fatalf("send frame: %v", err)
	}
}

// received reports whether a frame containing marker arrived within timeout.
// Frames sent by this host itself are ignored.
func (s *packetSocket) received(t *testing.T, marker []byte, timeout time.Duration) bool {
	t.Helper()
	buf := make([]byte, 9000)
	deadline := time.Now().Add(timeout)
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return false
		}
		tv := unix.NsecToTimeval(remaining.Nanoseconds())
		if err := unix.SetsockoptTimeval(s.fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &tv); err != nil {
			t.Fatalf("set receive timeout: %v", err)
		}
		n, from, err := unix.Recvfrom(s.fd, buf, 0)
		if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EINTR) {
			continue
		}
		if err != nil {
			t.Fatalf("receive frame: %v", err)
		}
		if ll, ok := from.(*unix.SockaddrLinklayer); ok && ll.Pkttype == unix.PACKET_OUTGOING {
			continue
		}
		if bytes.Contains(buf[:n], marker) {
			return true
		}
	}
}

func newMarker() []byte {
	m := make([]byte, 16)
	_, _ = rand.Read(m)
	return m
}

func htons(v uint16) uint16 {
	return v<<8 | v>>8
}

// Timeouts for frames that must arrive and for frames that must not.
const (
	arriveTimeout  = 2 * time.Second
	silenceTimeout = 300 * time.Millisecond
)

// endpoints are open packet sockets on eth0 of every host of a topology.
type endpoints struct {
	h1, h2, mon *packetSocket
}

func openEndpoints(t *testing.T, topo *topology) endpoints {
	return endpoints{
		h1:  openPacketSocket(t, topo.h1, "eth0"),
		h2:  openPacketSocket(t, topo.h2, "eth0"),
		mon: openPacketSocket(t, topo.mon, "eth0"),
	}
}

// expectFrame sends f from src and checks on which hosts it shows up.
func expectFrame(t *testing.T, f frame, src *packetSocket, want map[*packetSocket]bool, names map[*packetSocket]string) {
	t.Helper()
	marker := newMarker()
	src.send(t, f.build(marker))
	for sock, shouldArrive := range want {
		timeout := silenceTimeout
		if shouldArrive {
			timeout = arriveTimeout
		}
		if got := sock.received(t, marker, timeout); got != shouldArrive {
			t.Errorf("%s frame on %s: received=%v, want %v", f.name, names[sock], got, shouldArrive)
		}
	}
}
