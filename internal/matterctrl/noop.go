package matterctrl

import (
	"context"
	"errors"
)

// ErrControllerUnavailable is returned when no operational controller is configured.
var ErrControllerUnavailable = errors.New("operational controller is unavailable")

// NoopController returns ErrControllerUnavailable for all operations.
type NoopController struct{}

var _ Controller = (*NoopController)(nil)

// NewNoopController returns a controller that always fails with ErrControllerUnavailable.
func NewNoopController() *NoopController {
	return &NoopController{}
}

// Ping checks connectivity to an operational node.
func (c *NoopController) Ping(_ context.Context, _ uint64) error {
	return ErrControllerUnavailable
}

// ReadAttribute reads an attribute value by numeric identifiers.
func (c *NoopController) ReadAttribute(_ context.Context, _ uint64, _ uint16, _ uint32, _ uint32) (any, error) {
	return nil, ErrControllerUnavailable
}

// WriteAttribute writes an attribute value by numeric identifiers.
func (c *NoopController) WriteAttribute(_ context.Context, _ uint64, _ uint16, _ uint32, _ uint32, _ any) error {
	return ErrControllerUnavailable
}

// InvokeCommand invokes a command by numeric identifiers.
func (c *NoopController) InvokeCommand(_ context.Context, _ uint64, _ uint16, _ uint32, _ uint32, _ any) (any, error) {
	return nil, ErrControllerUnavailable
}
