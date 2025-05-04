package nbtns

import (
	"net"
)

// RedirectInfo contains information about where to redirect a client
type RedirectInfo struct {
	ServerIP   net.IP
	ServerPort uint16
}

// RedirectManager handles NBNS redirection
type RedirectManager struct {
	redirectMap map[string]RedirectInfo // Maps scope to redirect info
}

// NewRedirectManager creates a new redirect manager
func NewRedirectManager() *RedirectManager {
	return &RedirectManager{
		redirectMap: make(map[string]RedirectInfo),
	}
}

// AddRedirect adds or updates a redirect mapping
func (r *RedirectManager) AddRedirect(scope string, serverIP net.IP, port uint16) {
	r.redirectMap[scope] = RedirectInfo{
		ServerIP:   serverIP,
		ServerPort: port,
	}
}

// RemoveRedirect removes a redirect mapping
func (r *RedirectManager) RemoveRedirect(scope string) {
	delete(r.redirectMap, scope)
}

// GetRedirect returns redirect information for a scope
func (r *RedirectManager) GetRedirect(scope string) (RedirectInfo, bool) {
	info, exists := r.redirectMap[scope]
	return info, exists
}

// HandleRedirect modifies a response packet for redirection if needed
func (r *RedirectManager) HandleRedirect(request *NBTNSPacket, response *NBTNSPacket) bool {
	// Only redirect name queries
	if request.Header.Flags&0xF000 != OpNameQuery {
		return false
	}

	// Check if we have any questions
	if len(request.Questions) == 0 {
		return false
	}

	// Get scope from first question
	scope := request.Questions[0].Name.ScopeID

	// Look up redirect information
	info, exists := r.GetRedirect(scope)
	if !exists {
		return false
	}

	// Modify response for redirection
	response.Header.Flags = FlagResponse | OpRedirect
	response.Additional = []NBTNSResourceRecord{
		{
			Name:     request.Questions[0].Name,
			Type:     0x20,                           // NB record
			Class:    1,                              // IN class
			TTL:      600,                            // 10 minutes
			RDLength: uint16(len(info.ServerIP) + 2), // IP + port
			RData:    append(info.ServerIP, byte(info.ServerPort>>8), byte(info.ServerPort)),
		},
	}
	response.Header.Additional = 1

	return true
}
