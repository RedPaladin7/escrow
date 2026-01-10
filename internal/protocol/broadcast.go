package protocol

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	Data []byte   // The message data to broadcast
	To   []string // Target client IDs (empty means broadcast to all)
}

// NewBroadcast creates a new broadcast message for specific targets
func NewBroadcast(data []byte, targets ...string) *BroadcastMessage {
	return &BroadcastMessage{
		Data: data,
		To:   targets,
	}
}

// NewGlobalBroadcast creates a broadcast message for all clients
func NewGlobalBroadcast(data []byte) *BroadcastMessage {
	return &BroadcastMessage{
		Data: data,
		To:   []string{},
	}
}

// IsGlobal returns true if this is a global broadcast
func (bm *BroadcastMessage) IsGlobal() bool {
	return len(bm.To) == 0
}

// HasTarget checks if a specific client ID is targeted
func (bm *BroadcastMessage) HasTarget(clientID string) bool {
	if bm.IsGlobal() {
		return true
	}

	for _, target := range bm.To {
		if target == clientID {
			return true
		}
	}
	return false
}

// TargetCount returns the number of targets
func (bm *BroadcastMessage) TargetCount() int {
	if bm.IsGlobal() {
		return -1 // Indicates all
	}
	return len(bm.To)
}
