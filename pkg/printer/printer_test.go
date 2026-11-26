package printer

import (
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/api"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink"
	"github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
)

type fakeLink struct {
	index int
	name  string
}

func (l fakeLink) Index() int   { return l.index }
func (l fakeLink) Name() string { return l.name }

func newPrinter(t *testing.T) *PrettyPrinter {
	f := mocks.NewMockLinkFactory(gomock.NewController(t))
	f.EXPECT().NewLinkWithIndex(gomock.Any()).DoAndReturn(func(i int) (hostlink.Link, error) {
		names := map[int]string{1: "eth1", 2: "eth2", 3: "eth3"}
		if n, ok := names[i]; ok {
			return fakeLink{i, n}, nil
		}
		return nil, errors.New("Link not found")
	}).AnyTimes()
	return NewPrettyPrinter(f)
}

func ptr[T any](v T) *T { return &v }

func TestPrintEmpty(t *testing.T) {
	if got := newPrinter(t).PrintBridgeDescriptions(nil); got != NoBridges {
		t.Errorf("got %q, want %q", got, NoBridges)
	}
}

func TestPrintBridges(t *testing.T) {
	out := newPrinter(t).PrintBridgeDescriptions([]*api.BridgeDescription{
		{Name: "with-monitor", Iface1: 1, Iface2: 2, Monitor: ptr[int32](3)},
		{Name: "gone", Iface1: 1, Iface2: 42},
	})
	lines := strings.Split(out, "\n")
	find := func(name string) string {
		for _, l := range lines {
			if strings.Contains(l, name) {
				return l
			}
		}
		t.Fatalf("no row for %s in:\n%s", name, out)
		return ""
	}
	if row := find("with-monitor"); !strings.Contains(row, "eth1") || !strings.Contains(row, "eth2") || !strings.Contains(row, "eth3") {
		t.Errorf("row = %q", row)
	}
	// Interfaces that cannot be resolved are shown by index.
	if row := find("gone"); !strings.Contains(row, "#42") {
		t.Errorf("row = %q, want #42", row)
	}
	if !strings.Contains(strings.ToUpper(out), "MONITOR-IFACE") {
		t.Errorf("missing header:\n%s", out)
	}
}
