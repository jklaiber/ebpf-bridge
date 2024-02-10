package bridge

import (
	"testing"
)

func TestBridgeManager_New(t *testing.T) {
	m := NewBridgeManager()
	if m == nil {
		t.Errorf("NewBridgeManager() returned nil")
	}
}

func TestBridgeManager_Add(t *testing.T) {
	type args struct {
		name         string
		iface1       int
		iface2       int
		monitorIface *int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Add bridge with monitor interface",
			args: args{
				name:         "test",
				iface1:       1,
				iface2:       2,
				monitorIface: new(int),
			},
			wantErr: false,
		},
		{
			name: "Add bridge without monitor interface",
			args: args{
				name:         "test",
				iface1:       1,
				iface2:       2,
				monitorIface: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		m := NewBridgeManager()
		t.Run(tt.name, func(t *testing.T) {
			if err := m.Add(tt.args.name, tt.args.iface1, tt.args.iface2, tt.args.monitorIface); (err != nil) != tt.wantErr {
				t.Errorf("BridgeManager.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridgeManager_Remove(t *testing.T) {
}

func TestBridgeManager_List(t *testing.T) {
}
