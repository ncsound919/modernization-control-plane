package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/models"
)

// Log is an in-memory, append-only, cryptographically chained audit log.
// Each entry's hash covers its own content plus the previous entry's hash,
// making the chain tamper-evident.
type Log struct {
	mu      sync.RWMutex
	entries []*models.AuditEntry
}

func NewLog() *Log {
	return &Log{}
}

// Append adds a new audit entry and computes its chained hash.
func (l *Log) Append(entry *models.AuditEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry.Timestamp = time.Now().UTC()

	prevHash := ""
	if len(l.entries) > 0 {
		prevHash = l.entries[len(l.entries)-1].Hash
	}
	entry.PrevHash = prevHash
	entry.Hash = computeHash(entry, prevHash)

	l.entries = append(l.entries, entry)
	return nil
}

// Entries returns a copy of all audit entries.
func (l *Log) Entries() []*models.AuditEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	out := make([]*models.AuditEntry, len(l.entries))
	copy(out, l.entries)
	return out
}

// Verify walks the chain and confirms every hash is consistent.
func (l *Log) Verify() error {
	l.mu.RLock()
	defer l.mu.RUnlock()

	for i, e := range l.entries {
		prevHash := ""
		if i > 0 {
			prevHash = l.entries[i-1].Hash
		}
		expected := computeHash(e, prevHash)
		if e.Hash != expected {
			return fmt.Errorf("chain broken at entry %s (index %d): hash mismatch", e.ID, i)
		}
	}
	return nil
}

// computeHash produces a SHA-256 digest over the entry's stable fields and prevHash.
func computeHash(e *models.AuditEntry, prevHash string) string {
	payload := struct {
		ID         string                 `json:"id"`
		Timestamp  time.Time              `json:"timestamp"`
		Actor      string                 `json:"actor"`
		Action     string                 `json:"action"`
		Resource   string                 `json:"resource"`
		TenantID   string                 `json:"tenant_id"`
		WorkflowID string                 `json:"workflow_id"`
		Decision   string                 `json:"decision"`
		Details    map[string]interface{} `json:"details"`
		PrevHash   string                 `json:"prev_hash"`
	}{
		ID:         e.ID,
		Timestamp:  e.Timestamp,
		Actor:      e.Actor,
		Action:     e.Action,
		Resource:   e.Resource,
		TenantID:   e.TenantID,
		WorkflowID: e.WorkflowID,
		Decision:   e.Decision,
		Details:    e.Details,
		PrevHash:   prevHash,
	}

	data, _ := json.Marshal(payload)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
