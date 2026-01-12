package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cybergarage/go-matter-pack/internal/commission"
	"github.com/cybergarage/go-matter-pack/internal/store"
	"github.com/cybergarage/go-matter/matter"
)

// CommissionService handles onboarding payload import and commissioning.
type CommissionService struct {
	commissioner matter.Commissioner
	store        store.Store
}

// NewCommissionService returns a new CommissionService.
func NewCommissionService(commissioner matter.Commissioner, store store.Store) *CommissionService {
	return &CommissionService{commissioner: commissioner, store: store}
}

// ImportPayload parses and saves an onboarding payload.
func (s *CommissionService) ImportPayload(ctx context.Context, nodeID uint64, payload string) (commission.State, error) {
	if s == nil {
		return commission.State{}, errors.New("commission service is nil")
	}
	if s.store == nil {
		return commission.State{}, errors.New("store is nil")
	}
	return commission.ImportPayload(ctx, s.store, nodeID, payload)
}

// Commission commissions a device and updates the commissioning result.
func (s *CommissionService) Commission(ctx context.Context, nodeID uint64, payload string) (commission.State, matter.Commissionee, error) {
	if s == nil {
		return commission.State{}, nil, errors.New("commission service is nil")
	}
	if s.store == nil {
		return commission.State{}, nil, errors.New("store is nil")
	}
	if s.commissioner == nil {
		return commission.State{}, nil, errors.New("commissioner is nil")
	}

	onboarding, _, err := commission.ParseOnboardingPayload(payload)
	if err != nil {
		return commission.State{}, nil, err
	}

	state, err := commission.ImportPayload(ctx, s.store, nodeID, payload)
	if err != nil {
		return commission.State{}, nil, err
	}

	commissionee, err := s.commissioner.Commission(ctx, onboarding)
	if err != nil {
		return state, nil, err
	}

	result := commission.ResultRecord{
		NodeID:             nodeID,
		VendorID:           uint16(commissionee.VendorID()),
		ProductID:          uint16(commissionee.ProductID()),
		Device:             commissionee.String(),
		CommissionedAt:     time.Now().UTC(),
		PayloadFingerprint: payloadFingerprint(state),
	}
	updated, err := commission.UpdateResult(ctx, s.store, result)
	if err != nil {
		return state, commissionee, fmt.Errorf("commissioned but failed to update state: %w", err)
	}
	return updated, commissionee, nil
}

func payloadFingerprint(state commission.State) string {
	if state.Payload == nil {
		return ""
	}
	return state.Payload.PayloadFingerprint
}
