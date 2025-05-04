package nbtns

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// NetBIOSNameServer represents a NetBIOS Name Server
type NetBIOSNameServer struct {
	mu      sync.RWMutex
	names   map[string]*NameRecord
	secured bool // Whether this is a secured NetBIOSNameServer
}

// NewNetBIOSNameServer creates a new NetBIOS Name Server instance
func NewNetBIOSNameServer(secured bool) *NetBIOSNameServer {
	return &NetBIOSNameServer{
		names:   make(map[string]*NameRecord),
		secured: secured,
	}
}

// RegisterName attempts to register a name with the name server
func (n *NetBIOSNameServer) RegisterName(name string, nameType NameType, owner net.IP, ttl time.Duration) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Check if name exists
	if record, exists := n.names[name]; exists {
		// Handle group name registration
		if record.Type == Group && nameType == Group {
			// Add new owner to group
			for _, ip := range record.Owners {
				if ip.Equal(owner) {
					return nil // Already registered
				}
			}
			record.Owners = append(record.Owners, owner)
			record.TTL = time.Now().Add(ttl)
			return nil
		}

		// Handle unique name conflicts
		if record.Type == Unique || nameType == Unique {
			return fmt.Errorf("name conflict: %s is already registered", name)
		}
	}

	// Create new record
	n.names[name] = &NameRecord{
		Name:            name,
		Type:            nameType,
		Status:          Active,
		Owners:          []net.IP{owner},
		TTL:             time.Now().Add(ttl),
		RefreshInterval: ttl,
	}

	return nil
}

// QueryName looks up a name and returns its owners
func (n *NetBIOSNameServer) QueryName(name string) ([]net.IP, NameType, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	record, exists := n.names[name]
	if !exists || record.Status != Active {
		return nil, Unique, fmt.Errorf("name not found: %s", name)
	}

	// Make copy of owners slice to prevent external modification
	owners := make([]net.IP, len(record.Owners))
	copy(owners, record.Owners)

	return owners, record.Type, nil
}

// ReleaseName removes a name registration for a specific owner
func (n *NetBIOSNameServer) ReleaseName(name string, owner net.IP) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	record, exists := n.names[name]
	if !exists {
		return fmt.Errorf("name not found: %s", name)
	}

	// For group names, remove only the specified owner
	if record.Type == Group {
		for i, ip := range record.Owners {
			if ip.Equal(owner) {
				record.Owners = append(record.Owners[:i], record.Owners[i+1:]...)
				// Remove record if no owners remain
				if len(record.Owners) == 0 {
					delete(n.names, name)
				}
				return nil
			}
		}
		return fmt.Errorf("owner not found for name: %s", name)
	}

	// For unique names, verify owner and remove record
	if !record.Owners[0].Equal(owner) {
		return fmt.Errorf("owner mismatch for name: %s", name)
	}
	delete(n.names, name)
	return nil
}

// RefreshName updates the TTL for a name registration
func (n *NetBIOSNameServer) RefreshName(name string, owner net.IP) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	record, exists := n.names[name]
	if !exists {
		return fmt.Errorf("name not found: %s", name)
	}

	// Verify ownership
	ownerFound := false
	for _, ip := range record.Owners {
		if ip.Equal(owner) {
			ownerFound = true
			break
		}
	}
	if !ownerFound {
		return fmt.Errorf("owner not found for name: %s", name)
	}

	// Update TTL
	record.TTL = time.Now().Add(record.RefreshInterval)
	return nil
}

// MarkNameConflict marks a name as being in conflict
func (n *NetBIOSNameServer) MarkNameConflict(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	record, exists := n.names[name]
	if !exists {
		return fmt.Errorf("name not found: %s", name)
	}

	record.Status = Conflict
	return nil
}

// CleanExpiredNames removes names that have exceeded their TTL
func (n *NetBIOSNameServer) CleanExpiredNames() {
	n.mu.Lock()
	defer n.mu.Unlock()

	now := time.Now()
	for name, record := range n.names {
		if now.After(record.TTL) {
			delete(n.names, name)
		}
	}
}
