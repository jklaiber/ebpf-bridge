package manager

import (
	"fmt"
	"testing"

	bridgeMock "github.com/jklaiber/ebpf-bridge/pkg/bridge/mocks"
	hostlinkMock "github.com/jklaiber/ebpf-bridge/pkg/hostlink/mocks"
	"go.uber.org/mock/gomock"
)

func TestBridgeManager_New(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	linkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
	bridgeFactory := bridgeMock.NewMockBridgeFactory(ctrl)

	bm := NewBridgeManager(linkFactory, bridgeFactory)
	if bm == nil {
		t.Error("NewBridgeManager returned nil")
	}
}

func TestBridgeManager_Add(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory)
		addCall  func(*hostlinkMock.MockLinkFactory, *bridgeMock.MockBridgeFactory) *BridgeManager
		testCall func(*BridgeManager) error
		wantErr  bool
	}{
		{
			"Test Add with iface1 error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, fmt.Errorf("error"))
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				return bm.Add("test", 1, 2, nil)
			},
			true,
		},
		{
			"Test Add with iface2 error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, fmt.Errorf("error"))
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				return bm.Add("test", 1, 2, nil)
			},
			true,
		},
		{
			"Test Add with monitorIface error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(3).Return(nil, fmt.Errorf("error"))
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				mI := 3
				return bm.Add("test", 1, 2, &mI)
			},
			true,
		},
		{
			"Test Add with monitorIface with bridge error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(3).Return(nil, nil)
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Add().Return(fmt.Errorf("error"))
				mockBridgeFactory.EXPECT().NewBridge("test", gomock.Any(), gomock.Any(), gomock.Any()).Return(mockBridge)
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				mI := 3
				return bm.Add("test", 1, 2, &mI)
			},
			true,
		},
		{
			"Test Add with monitorIface without error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(3).Return(nil, nil)
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Add().Return(nil)
				mockBridgeFactory.EXPECT().NewBridge("test", gomock.Any(), gomock.Any(), gomock.Any()).Return(mockBridge)
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				mI := 3
				return bm.Add("test", 1, 2, &mI)
			},
			false,
		},
		{
			"Test Add without monitorIface with error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, nil)
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Add().Return(fmt.Errorf("error"))
				mockBridgeFactory.EXPECT().NewBridge("test", gomock.Any(), gomock.Any(), nil).Return(mockBridge)
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				return bm.Add("test", 1, 2, nil)
			},
			true,
		},
		{
			"Test Add without monitorIface without error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) {
				mockLinkFactory.EXPECT().NewLinkWithIndex(1).Return(nil, nil)
				mockLinkFactory.EXPECT().NewLinkWithIndex(2).Return(nil, nil)
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Add().Return(nil)
				mockBridgeFactory.EXPECT().NewBridge("test", gomock.Any(), gomock.Any(), nil).Return(mockBridge)
			},
			func(mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				return NewBridgeManager(mockLinkFactory, mockBridgeFactory)
			},
			func(bm *BridgeManager) error {
				return bm.Add("test", 1, 2, nil)
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
			mockBridgeFactory := bridgeMock.NewMockBridgeFactory(ctrl)

			tt.mockCall(ctrl, mockLinkFactory, mockBridgeFactory)
			bm := tt.addCall(mockLinkFactory, mockBridgeFactory)

			err := tt.testCall(bm)
			if (err != nil) != tt.wantErr {
				t.Errorf("BridgeManager.Add() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridgeManager_Remove(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager
		wantErr  bool
	}{
		{
			"Test Remove bridge not found",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				bm := NewBridgeManager(mockLinkFactory, mockBridgeFactory)
				return bm
			},
			true,
		},
		{
			"Test Remove with bridge error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Remove().Return(fmt.Errorf("error"))
				bm := NewBridgeManager(mockLinkFactory, mockBridgeFactory)
				bm.bridges["test"] = mockBridge
				return bm
			},
			true,
		},
		{
			"Test Remove without error",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				mockBridge := bridgeMock.NewMockBridge(ctrl)
				mockBridge.EXPECT().Remove().Return(nil)
				bm := NewBridgeManager(mockLinkFactory, mockBridgeFactory)
				bm.bridges["test"] = mockBridge
				return bm
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
			mockBridgeFactory := bridgeMock.NewMockBridgeFactory(ctrl)

			bm := tt.mockCall(ctrl, mockLinkFactory, mockBridgeFactory)

			err := bm.Remove("test")
			if (err != nil) != tt.wantErr {
				t.Errorf("BridgeManager.Remove() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBridgeManager_List(t *testing.T) {
	tests := []struct {
		name     string
		mockCall func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager
		wantErr  bool
	}{
		{
			"Test List without monitorIface",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				bm := NewBridgeManager(mockLinkFactory, mockBridgeFactory)
				mockBridge1 := bridgeMock.NewMockBridge(ctrl)
				hostLink1 := hostlinkMock.NewMockLink(ctrl)
				hostLink1.EXPECT().Index().Return(1)
				mockBridge1.EXPECT().Interface1().Return(hostLink1)
				hostLink2 := hostlinkMock.NewMockLink(ctrl)
				hostLink2.EXPECT().Index().Return(2)
				mockBridge1.EXPECT().Interface2().Return(hostLink2)
				mockBridge1.EXPECT().MonitorInterface().Return(nil)
				bm.bridges["test1"] = mockBridge1

				return bm
			},
			false,
		},
		{
			"Test List with monitorIface",
			func(ctrl *gomock.Controller, mockLinkFactory *hostlinkMock.MockLinkFactory, mockBridgeFactory *bridgeMock.MockBridgeFactory) *BridgeManager {
				bm := NewBridgeManager(mockLinkFactory, mockBridgeFactory)
				mockBridge2 := bridgeMock.NewMockBridge(ctrl)
				hostLink3 := hostlinkMock.NewMockLink(ctrl)
				hostLink3.EXPECT().Index().Return(3)
				mockBridge2.EXPECT().Interface1().Return(hostLink3)
				hostLink4 := hostlinkMock.NewMockLink(ctrl)
				hostLink4.EXPECT().Index().Return(4)
				mockBridge2.EXPECT().Interface2().Return(hostLink4)
				monitorLink := hostlinkMock.NewMockLink(ctrl)
				monitorLink.EXPECT().Index().Return(5)
				mockBridge2.EXPECT().MonitorInterface().Return(monitorLink).Times(2)
				bm.bridges["test2"] = mockBridge2
				return bm
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLinkFactory := hostlinkMock.NewMockLinkFactory(ctrl)
			mockBridgeFactory := bridgeMock.NewMockBridgeFactory(ctrl)

			bm := tt.mockCall(ctrl, mockLinkFactory, mockBridgeFactory)

			bridges := bm.List()
			if bridges == nil {
				t.Error("BridgeManager.List() returned nil")
			}
		})
	}
}
