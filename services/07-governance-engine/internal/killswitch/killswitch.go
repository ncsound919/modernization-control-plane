package killswitch

import (
	"fmt"
	"sync"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/models"
)

// Manager holds the runtime state of all kill switches.
type Manager struct {
	mu      sync.RWMutex
	switches map[string]*models.KillSwitch
}

func NewManager() *Manager {
	m := &Manager{
		switches: make(map[string]*models.KillSwitch),
	}
	m.seed()
	return m
}

// seed pre-populates well-known kill switches in inactive state.
func (m *Manager) seed() {
	defaults := []*models.KillSwitch{
		{
			Name:        "emergency-readonly",
			Description: "Global read-only lockdown — all write operations denied",
			Active:      false,
			Scope:       models.ScopeEmergency,
		},
		{
			Name:        "tenant-lockdown",
			Description: "Halt all agent and write operations for a specific tenant",
			Active:      false,
			Scope:       models.ScopeTenant,
		},
		{
			Name:        "workflow-pause",
			Description: "Pause a specific modernization workflow pipeline",
			Active:      false,
			Scope:       models.ScopeWorkflow,
		},
	}
	for _, ks := range defaults {
		m.switches[ks.Name] = ks
	}
}

// List returns all kill switches.
func (m *Manager) List() []*models.KillSwitch {
	m.mu.RLock()
	defer m.mu.RUnlock()

	out := make([]*models.KillSwitch, 0, len(m.switches))
	for _, ks := range m.switches {
		cp := *ks
		out = append(out, &cp)
	}
	return out
}

// Activate sets a kill switch to active.
func (m *Manager) Activate(name, actor, reason string) (*models.KillSwitch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ks, ok := m.switches[name]
	if !ok {
		return nil, fmt.Errorf("kill switch %q not found", name)
	}

	now := time.Now().UTC()
	ks.Active = true
	ks.ActivatedBy = actor
	ks.ActivatedAt = &now
	ks.Reason = reason

	cp := *ks
	return &cp, nil
}

// Deactivate clears a kill switch.
func (m *Manager) Deactivate(name, actor string) (*models.KillSwitch, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ks, ok := m.switches[name]
	if !ok {
		return nil, fmt.Errorf("kill switch %q not found", name)
	}

	ks.Active = false
	ks.ActivatedBy = ""
	ks.ActivatedAt = nil
	ks.Reason = ""

	cp := *ks
	return &cp, nil
}

// IsBlocked returns true if any kill switch that applies to the given context
// is currently active.
func (m *Manager) IsBlocked(tenantID, workflowID string) (bool, string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Emergency / global switches block everything.
	for _, ks := range m.switches {
		if !ks.Active {
			continue
		}
		if ks.Scope == models.ScopeEmergency || ks.Scope == models.ScopeGlobal {
			return true, fmt.Sprintf("kill switch %q is active: %s", ks.Name, ks.Reason)
		}
		if ks.Scope == models.ScopeTenant && tenantID != "" && ks.TenantID == tenantID {
			return true, fmt.Sprintf("tenant kill switch %q is active: %s", ks.Name, ks.Reason)
		}
		if ks.Scope == models.ScopeWorkflow && workflowID != "" && ks.WorkflowID == workflowID {
			return true, fmt.Sprintf("workflow kill switch %q is active: %s", ks.Name, ks.Reason)
		}
	}
	return false, ""
}
