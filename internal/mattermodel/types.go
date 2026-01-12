package mattermodel

// BridgedDevice represents a bridged endpoint attached to a hub node.
type BridgedDevice struct {
	NodeID    uint64
	Endpoint  uint16
	UniqueID  string
	NodeLabel string
	Reachable *bool
}
