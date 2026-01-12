package commission

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cybergarage/go-matter-pack/internal/store"
	"github.com/cybergarage/go-matter/matter/encoding"
)

// PayloadRecord stores onboarding payload details for commissioning.
type PayloadRecord struct {
	NodeID             uint64    `json:"node_id"`
	VendorID           uint16    `json:"vendor_id"`
	ProductID          uint16    `json:"product_id"`
	CommissioningFlow  string    `json:"commissioning_flow"`
	Discriminator      uint16    `json:"discriminator"`
	Passcode           uint32    `json:"passcode,omitempty"`
	QRCode             string    `json:"qr_code,omitempty"`
	PairingCode        string    `json:"pairing_code,omitempty"`
	ImportedAt         time.Time `json:"imported_at"`
	PayloadFingerprint string    `json:"payload_fingerprint"`
}

// Bundle stores imported operational credentials for later reuse.
type Bundle struct {
	NodeID           uint64    `json:"node_id"`
	FabricID         uint64    `json:"fabric_id,omitempty"`
	RootCert         string    `json:"root_cert,omitempty"`
	IntermediateCert string    `json:"intermediate_cert,omitempty"`
	OperationalCert  string    `json:"operational_cert,omitempty"`
	OperationalKey   string    `json:"operational_key,omitempty"`
	IPK              string    `json:"ipk,omitempty"`
	Source           string    `json:"source,omitempty"`
	ImportedAt       time.Time `json:"imported_at"`
}

// State keeps commissioning-related data for a single hub.
type State struct {
	Payload *PayloadRecord `json:"payload,omitempty"`
	Bundle  *Bundle        `json:"bundle,omitempty"`
}

// LoadState loads commissioning state from the store.
func LoadState(ctx context.Context, s store.Store) (State, error) {
	var state State
	if err := s.Load(ctx, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

// SaveState persists commissioning state to the store.
func SaveState(ctx context.Context, s store.Store, state State) error {
	return s.Save(ctx, &state)
}

// ImportPayload records onboarding payload details (QR or manual pairing code).
func ImportPayload(ctx context.Context, s store.Store, nodeID uint64, payload string) (State, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return State{}, fmt.Errorf("payload is required")
	}

	record := PayloadRecord{NodeID: nodeID, ImportedAt: time.Now()}
	if strings.HasPrefix(payload, encoding.QRPayloadPrefix) {
		qrPayload, err := encoding.NewQRPayloadFromString(payload)
		if err != nil {
			return State{}, err
		}
		record.QRCode = payload
		record.VendorID = uint16(qrPayload.VendorID())
		record.ProductID = uint16(qrPayload.ProductID())
		record.CommissioningFlow = qrPayload.CommissioningFlow().String()
		record.Discriminator = uint16(qrPayload.Discriminator())
		record.Passcode = uint32(qrPayload.Passcode())
	} else {
		pairingCode, err := encoding.NewPairingCodeFromString(payload)
		if err != nil {
			return State{}, err
		}
		record.PairingCode = payload
		record.VendorID = uint16(pairingCode.VendorID())
		record.ProductID = uint16(pairingCode.ProductID())
		record.CommissioningFlow = pairingCode.CommissioningFlow().String()
		record.Discriminator = uint16(pairingCode.Discriminator())
		record.Passcode = uint32(pairingCode.Passcode())
	}

	record.PayloadFingerprint = fingerprintPayload(record)

	state, err := LoadState(ctx, s)
	if err != nil {
		return State{}, err
	}
	state.Payload = &record
	if err := SaveState(ctx, s, state); err != nil {
		return State{}, err
	}
	return state, nil
}

// ImportBundle saves a commissioning bundle for later operational reuse.
func ImportBundle(ctx context.Context, s store.Store, bundle Bundle) (State, error) {
	if bundle.ImportedAt.IsZero() {
		bundle.ImportedAt = time.Now()
	}
	state, err := LoadState(ctx, s)
	if err != nil {
		return State{}, err
	}
	state.Bundle = &bundle
	if err := SaveState(ctx, s, state); err != nil {
		return State{}, err
	}
	return state, nil
}

func fingerprintPayload(record PayloadRecord) string {
	parts := []string{
		fmt.Sprintf("node=%d", record.NodeID),
		fmt.Sprintf("vendor=%d", record.VendorID),
		fmt.Sprintf("product=%d", record.ProductID),
		fmt.Sprintf("disc=%d", record.Discriminator),
	}
	if record.QRCode != "" {
		parts = append(parts, "qr")
	}
	if record.PairingCode != "" {
		parts = append(parts, "pairing")
	}
	return strings.Join(parts, ";")
}
