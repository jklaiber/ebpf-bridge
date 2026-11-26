package service

import (
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/jklaiber/ebpf-bridge/pkg/manager/mocks"
)

type fakeServer struct {
	startErr         error
	started, stopped bool
}

func (s *fakeServer) Start() error { s.started = true; return s.startErr }
func (s *fakeServer) Stop()        { s.stopped = true }

func TestStartStop(t *testing.T) {
	m := mocks.NewMockManager(gomock.NewController(t))
	server := &fakeServer{}
	svc := NewEbpfBridgeService(m, server)
	if err := svc.Start(); err != nil || !server.started {
		t.Fatalf("Start() = %v, started = %v", err, server.started)
	}

	// Stopping removes all bridges after the server stopped accepting requests.
	m.EXPECT().RemoveAll().DoAndReturn(func() error {
		if !server.stopped {
			t.Error("bridges removed before the server stopped")
		}
		return nil
	})
	if err := svc.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
}

func TestStartError(t *testing.T) {
	svc := NewEbpfBridgeService(mocks.NewMockManager(gomock.NewController(t)), &fakeServer{startErr: errors.New("listen")})
	if err := svc.Start(); err == nil {
		t.Error("Start() succeeded, want error")
	}
}

func TestStopReportsRemoveErrors(t *testing.T) {
	m := mocks.NewMockManager(gomock.NewController(t))
	m.EXPECT().RemoveAll().Return(errors.New("detach failed"))
	if err := NewEbpfBridgeService(m, &fakeServer{}).Stop(); err == nil {
		t.Error("Stop() succeeded, want error")
	}
}
