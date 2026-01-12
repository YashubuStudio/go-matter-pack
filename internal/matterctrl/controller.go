package matterctrl

import "context"

// Controller abstracts Matter operational connections.
type Controller interface {
	// Ping checks connectivity to an operational node.
	Ping(ctx context.Context, nodeID uint64) error

	// ReadAttribute reads an attribute value by numeric identifiers.
	ReadAttribute(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32) (any, error)
	// WriteAttribute writes an attribute value by numeric identifiers.
	WriteAttribute(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, attrID uint32, value any) error
	// InvokeCommand invokes a command by numeric identifiers.
	InvokeCommand(ctx context.Context, nodeID uint64, endpoint uint16, clusterID uint32, cmdID uint32, payload any) (any, error)
}
