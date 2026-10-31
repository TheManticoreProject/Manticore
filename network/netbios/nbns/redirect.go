package nbns

import (
	"net"
	"sync"
)

// RedirectInfo contains information about where to redirect a client
type RedirectInfo struct {
	ServerIP   net.IP
	ServerPort uint16
}

// RedirectManager handles NBNS redirection
type RedirectManager struct {
	mu          sync.RWMutex
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
	r.mu.Lock()
	defer r.mu.Unlock()
	r.redirectMap[scope] = RedirectInfo{
		ServerIP:   serverIP,
		ServerPort: port,
	}
}

// RemoveRedirect removes a redirect mapping
func (r *RedirectManager) RemoveRedirect(scope string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.redirectMap, scope)
}

// GetRedirect returns redirect information for a scope
func (r *RedirectManager) GetRedirect(scope string) (RedirectInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, exists := r.redirectMap[scope]
	return info, exists
}

// HandleRedirect turns response into a REDIRECT NAME QUERY RESPONSE (RFC 1002
// 4.2.15) when the first question's scope has a redirect configured. The packet
// is a query-opcode response with RD set and no answers: the authority section
// holds an NS record naming the NBNS to ask (NSD_NAME) and the additional
// section holds an A record for that same name whose 4-byte RDATA is the NBNS
// address (NSD_IP_ADDR) the requestor must reiterate its query to. The queried
// name stands in for both the redirected domain and NSD_NAME, so the additional
// RR_NAME matches NSD_NAME as the RFC requires. The packet has no port field, so
// RedirectInfo.ServerPort is not carried; a redirect to a server without an IPv4
// address is not applied.
func (r *RedirectManager) HandleRedirect(request *NBNSPacket, response *NBNSPacket) bool {
	// Only redirect name queries
	if request.Header.Flags&OpcodeMask != OpNameQuery {
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

	// NSD_IP_ADDR is a 4-byte IPv4 address.
	serverIP := info.ServerIP.To4()
	if serverIP == nil {
		return false
	}

	// NSD_NAME is carried as an encoded label sequence in the NS RDATA.
	name := request.Questions[0].Name
	nsdName, err := marshalName(name)
	if err != nil {
		return false
	}

	// Modify response for redirection: R=1, OPCODE=query, RD=1, RCODE=0.
	response.Header.Flags = FlagResponse | OpNameQuery | FlagRecursion
	response.Answers = nil
	response.Header.Answers = 0
	response.Authority = []NBNSResourceRecord{
		{
			Name:     name,
			Type:     0x0002, // NS record
			Class:    QuestionClassIn,
			TTL:      600, // 10 minutes
			RDLength: uint16(len(nsdName)),
			RData:    nsdName,
		},
	}
	response.Header.Authority = 1
	response.Additional = []NBNSResourceRecord{
		{
			Name:     name,
			Type:     0x0001, // A record
			Class:    QuestionClassIn,
			TTL:      600, // 10 minutes
			RDLength: uint16(len(serverIP)),
			RData:    append([]byte(nil), serverIP...),
		},
	}
	response.Header.Additional = 1

	return true
}
