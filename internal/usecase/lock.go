package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/cybergarage/go-matter/internal/matterctrl"
	"github.com/cybergarage/go-matter/internal/store"
)

const (
	doorLockClusterID   uint32 = 0x0101
	lockDoorCommandID   uint32 = 0x00
	unlockDoorCommandID uint32 = 0x01
)

// DoorLockPayload represents the optional payload for door lock commands.
type DoorLockPayload struct {
	PINCode string `json:"PINCode,omitempty"`
}

// LockService executes lock and unlock commands against bridged devices.
type LockService struct {
	ctrl  matterctrl.Controller
	store store.Store
}

// NewLockService returns a new LockService.
func NewLockService(ctrl matterctrl.Controller, store store.Store) *LockService {
	return &LockService{ctrl: ctrl, store: store}
}

// Lock sends the LockDoor command to the target device.
func (s *LockService) Lock(ctx context.Context, uniqueID string) error {
	return s.invoke(ctx, uniqueID, "", lockDoorCommandID)
}

// Unlock sends the UnlockDoor command to the target device.
func (s *LockService) Unlock(ctx context.Context, uniqueID, pin string) error {
	return s.invoke(ctx, uniqueID, pin, unlockDoorCommandID)
}

func (s *LockService) invoke(ctx context.Context, uniqueID, pin string, cmdID uint32) error {
	if s == nil {
		return errors.New("lock service is nil")
	}
	if s.ctrl == nil {
		return errors.New("controller is nil")
	}
	if s.store == nil {
		return errors.New("store is nil")
	}
	if uniqueID == "" {
		return errors.New("unique id is required")
	}

	var registry store.Registry
	if err := s.store.Load(ctx, &registry); err != nil {
		return err
	}

	record, ok := registry.Find(uniqueID)
	if !ok {
		return fmt.Errorf("device not found for unique_id %s", uniqueID)
	}
	if record.Missing {
		return fmt.Errorf("device %s is marked missing", uniqueID)
	}

	nodeID := record.NodeID
	if nodeID == 0 {
		nodeID = registry.HubNodeID
	}
	if nodeID == 0 {
		return errors.New("hub node id is not set")
	}
	if record.Endpoint == 0 {
		return errors.New("device endpoint is not set")
	}

	payload := DoorLockPayload{}
	if pin != "" {
		payload.PINCode = pin
	}
	_, err := s.ctrl.InvokeCommand(ctx, nodeID, record.Endpoint, doorLockClusterID, cmdID, payload)
	return err
}
