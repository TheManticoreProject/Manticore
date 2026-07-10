package nbns

import (
	"encoding/binary"
	"net"
	"time"
)

const (
	// DefaultWACKWait is the time-to-wait (in seconds) advertised in a WAIT FOR
	// ACKNOWLEDGEMENT (WACK) RESPONSE while the server runs a name challenge
	// before answering a registration (RFC 1002 4.2.16). It comfortably covers
	// the challenger's retry budget (ChallengeRetries * ChallengeTimeout).
	DefaultWACKWait = 8 * time.Second
)

// PacketHandler provides common packet handling methods for both TCP and UDP servers
type PacketHandler struct {
	nbns *NetBIOSNameServer

	// nodeStatusEnabled gates the NODE STATUS responder. It is off by default so
	// an NBSTAT (0x0021) query falls through to the historical behaviour; a
	// server opts in through EnableNodeStatus.
	nodeStatusEnabled bool
	// nodeStatusMAC is the adapter unit id reported in the STATISTICS block of a
	// NODE STATUS RESPONSE. It may be nil, in which case the UNIT_ID is zeroed.
	nodeStatusMAC net.HardwareAddr

	// challenger, when non-nil, defends an owned unique name against a
	// conflicting registration by challenging the current owner (RFC 1002
	// 4.2.10). Nil by default, so registrations behave exactly as before.
	challenger *NameChallenger
	// redirect, when non-nil, may turn a NAME QUERY into a REDIRECT NAME QUERY
	// RESPONSE for a configured scope (RFC 1002 4.2.14). Nil by default.
	redirect *RedirectManager
}

// NewPacketHandler creates a new packet handler instance
func NewPacketHandler(nbns *NetBIOSNameServer) *PacketHandler {
	return &PacketHandler{
		nbns: nbns,
	}
}

// EnableNodeStatus turns on the NODE STATUS responder and sets the adapter MAC
// reported as the STATISTICS UNIT_ID. A nil or non-6-byte mac reports a zeroed
// UNIT_ID.
func (h *PacketHandler) EnableNodeStatus(mac net.HardwareAddr) {
	h.nodeStatusEnabled = true
	h.nodeStatusMAC = mac
}

// SetChallenger installs (or, with nil, removes) the name challenger used to
// defend owned unique names on a conflicting registration.
func (h *PacketHandler) SetChallenger(c *NameChallenger) { h.challenger = c }

// SetRedirectManager installs (or, with nil, removes) the redirect manager
// consulted on the name-query path.
func (h *PacketHandler) SetRedirectManager(r *RedirectManager) { h.redirect = r }

// isNodeStatusQuery reports whether request is a NODE STATUS REQUEST: an OPCODE
// query (0x0000, the same opcode as a name query) whose first question asks for
// the NBSTAT (0x0021) type (RFC 1002 4.2.17). Node status is distinguished from
// a name query only by that question type.
func (h *PacketHandler) isNodeStatusQuery(request *NBNSPacket) bool {
	return request.Header.Flags&OpcodeMask == OpNameQuery &&
		len(request.Questions) > 0 &&
		request.Questions[0].Type == QuestionTypeNBSTAT
}

// handleNodeStatus answers a NODE STATUS REQUEST with the local name table (RFC
// 1002 4.2.18): a single NBSTAT answer record whose RDATA lists every registered
// name as a NODE_NAME entry (with NAME_FLAGS reflecting group/active/conflict)
// followed by the STATISTICS block carrying the configured adapter MAC. The
// answer echoes the request's question name (the reserved "*" wildcard), which
// the label-sequence marshaller encodes directly.
func (h *PacketHandler) handleNodeStatus(request *NBNSPacket, response *NBNSPacket) {
	response.Header.Flags = FlagResponse | FlagAuthoritative

	var name *NetBIOSName
	if len(request.Questions) > 0 {
		name = request.Questions[0].Name
	} else {
		name = &NetBIOSName{Name: "*"}
	}

	rdata := buildNodeStatusRData(h.nbns.NameTable(), h.nodeStatusMAC)
	response.Answers = append(response.Answers, NBNSResourceRecord{
		Name:     name,
		Type:     QuestionTypeNBSTAT,
		Class:    QuestionClassIn,
		TTL:      0,
		RDLength: uint16(len(rdata)),
		RData:    rdata,
	})
	response.Header.Answers = uint16(len(response.Answers))
}

// handleNameQuery processes a name query request
func (h *PacketHandler) handleNameQuery(request *NBNSPacket, response *NBNSPacket) {
	for _, q := range request.Questions {
		owners, nameType, ttl, err := h.nbns.QueryName(q.Name.Name, q.Name.ScopeID)
		if err != nil {
			response.Header.Flags |= RcodeNameError
			return
		}

		ttlSeconds := uint32(ttl.Seconds())

		// The Group (G) bit lives in the resource record's NB_FLAGS, not in the
		// header (RFC 1002 4.2.1.3).
		var nbFlags uint16
		if nameType == Group {
			nbFlags |= NBFlagGroup
		}

		// Create resource record for each owner
		for _, ip := range owners {
			owner := ADDR_ENTRY{
				Address: binary.BigEndian.Uint32(ip.To4()),
				Flags:   nbFlags,
			}
			rr := NBNSResourceRecord{
				Name:     q.Name,
				Type:     q.Type,
				Class:    q.Class,
				TTL:      ttlSeconds,
				RDLength: uint16(owner.Length()),
				RData:    owner.Marshal(),
			}
			response.Answers = append(response.Answers, rr)
		}

		response.Header.Answers = uint16(len(response.Answers))
	}
}

// handleRegistration processes a name registration request
func (h *PacketHandler) handleRegistration(request *NBNSPacket, response *NBNSPacket) {
	for _, rr := range request.Answers {
		var entry ADDR_ENTRY
		if err := entry.Unmarshal(rr.RData); err != nil {
			response.Header.Flags |= RcodeFormatError
			return
		}
		ip := entry.IP()

		// The unique/group distinction is carried by the Group (G) bit of the
		// resource record's NB_FLAGS, not by the header (RFC 1002 4.2.1.3).
		nameType := Unique
		if entry.Flags&NBFlagGroup != 0 {
			nameType = Group
		}

		err := h.nbns.RegisterName(
			rr.Name.Name,
			rr.Name.ScopeID,
			nameType,
			ip,
			time.Duration(rr.TTL)*time.Second,
		)

		if err != nil {
			response.Header.Flags |= RcodeConflict
			return
		}
	}
}

// handleRelease processes a name release request
func (h *PacketHandler) handleRelease(request *NBNSPacket, response *NBNSPacket) {
	for _, rr := range request.Answers {
		ip, err := ParseIPFromRData(rr.RData)
		if err != nil {
			response.Header.Flags |= RcodeFormatError
			return
		}
		if err := h.nbns.ReleaseName(rr.Name.Name, rr.Name.ScopeID, ip); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}

// handleRefresh processes a name refresh request
func (h *PacketHandler) handleRefresh(request *NBNSPacket, response *NBNSPacket) {
	for _, rr := range request.Answers {
		ip, err := ParseIPFromRData(rr.RData)
		if err != nil {
			response.Header.Flags |= RcodeFormatError
			return
		}
		if err := h.nbns.RefreshName(rr.Name.Name, rr.Name.ScopeID, ip); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}

// handleNameQueryWithRedirect answers a name query, first giving a configured
// redirect manager the chance to turn it into a REDIRECT NAME QUERY RESPONSE
// (RFC 1002 4.2.14) for a matching scope. When no redirect applies (or none is
// configured) it falls back to the authoritative name-table lookup.
func (h *PacketHandler) handleNameQueryWithRedirect(request *NBNSPacket, response *NBNSPacket) {
	if h.redirect != nil && h.redirect.HandleRedirect(request, response) {
		return
	}
	h.handleNameQuery(request, response)
}

// handleRegistrationWithChallenge processes a name registration when a name
// challenger is configured. Each proposed name is first registered optimistically;
// on a conflict with an already-owned name the server defers the decision per RFC
// 1002 4.2.10: it asks the transport to emit a WACK (so the requestor waits
// longer) through sendWACK, then challenges the current owner(s). If any owner
// still defends the name the registration is refused with RCODE CFT_ERR;
// otherwise the stale record is released and the new registration is accepted.
// sendWACK may be nil when the transport cannot emit an intermediate datagram, in
// which case the WACK is skipped but the challenge still runs.
func (h *PacketHandler) handleRegistrationWithChallenge(request *NBNSPacket, response *NBNSPacket, sendWACK func(*NBNSPacket)) {
	for _, rr := range request.Answers {
		var entry ADDR_ENTRY
		if err := entry.Unmarshal(rr.RData); err != nil {
			response.Header.Flags |= RcodeFormatError
			return
		}
		ip := entry.IP()
		ttl := time.Duration(rr.TTL) * time.Second

		// The unique/group distinction is carried by the RR NB_FLAGS G bit.
		nameType := Unique
		if entry.Flags&NBFlagGroup != 0 {
			nameType = Group
		}

		if err := h.nbns.RegisterName(rr.Name.Name, rr.Name.ScopeID, nameType, ip, ttl); err == nil {
			continue // no conflict: registered directly
		}

		// Conflict with an existing owner: tell the requestor to wait, then
		// challenge the current owner(s) before deciding.
		if sendWACK != nil {
			sendWACK(buildWACK(request, DefaultWACKWait))
		}

		owners, _, _, qerr := h.nbns.QueryName(rr.Name.Name, rr.Name.ScopeID)
		if qerr != nil {
			// The record disappeared between the failed registration and the
			// lookup; try once more to claim the now-free name.
			if err := h.nbns.RegisterName(rr.Name.Name, rr.Name.ScopeID, nameType, ip, ttl); err != nil {
				response.Header.Flags |= RcodeConflict
				return
			}
			continue
		}

		defended := false
		for _, owner := range owners {
			if ok, _ := h.challenger.ChallengeOwnership(rr.Name.Name, owner); ok {
				defended = true
				break
			}
		}

		if defended {
			// A current owner still answers: refuse the registration.
			response.Header.Flags |= RcodeConflict
			return
		}

		// No owner defended the name: release the stale record(s) and let the
		// new node take ownership.
		for _, owner := range owners {
			_ = h.nbns.ReleaseName(rr.Name.Name, rr.Name.ScopeID, owner)
		}
		if err := h.nbns.RegisterName(rr.Name.Name, rr.Name.ScopeID, nameType, ip, ttl); err != nil {
			response.Header.Flags |= RcodeConflict
			return
		}
	}
}

// buildWACK constructs a WAIT FOR ACKNOWLEDGEMENT (WACK) RESPONSE (RFC 1002
// 4.2.16) echoing request's transaction ID and name. The TTL is the number of
// seconds the requestor should wait for the pending operation to complete, and
// the 2-byte RDATA carries the OPCODE and NM_FLAGS of the request (its header
// flags with the response bit and RCODE cleared) so the requestor can correlate
// the wait with its outstanding request.
func buildWACK(request *NBNSPacket, wait time.Duration) *NBNSPacket {
	name := &NetBIOSName{}
	switch {
	case len(request.Answers) > 0:
		name = request.Answers[0].Name
	case len(request.Questions) > 0:
		name = request.Questions[0].Name
	}

	rdata := make([]byte, 2)
	binary.BigEndian.PutUint16(rdata, request.Header.Flags&^FlagResponse&^RcodeMask)

	return &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: request.Header.TransactionID,
			Flags:         FlagResponse | OpWACK | FlagAuthoritative,
			Answers:       1,
		},
		Answers: []NBNSResourceRecord{{
			Name:     name,
			Type:     0x0020, // RR_TYPE NULL, per the RFC 1002 4.2.16 WACK diagram
			Class:    QuestionClassIn,
			TTL:      uint32(wait.Seconds()),
			RDLength: uint16(len(rdata)),
			RData:    rdata,
		}},
	}
}

// buildNameConflictDemand constructs a NAME CONFLICT DEMAND (RFC 1002 4.2.15): a
// registration-opcode response carrying RCODE CFT_ERR that tells owner the unique
// name is claimed by more than one node. It is byte-identical to a NEGATIVE NAME
// REGISTRATION RESPONSE with RCODE = CFT_ERR, so the single NB answer record
// carries the owner's ADDR_ENTRY. It returns nil when owner has no IPv4 address
// to place in the ADDR_ENTRY.
func buildNameConflictDemand(name *NetBIOSName, owner net.IP) *NBNSPacket {
	v4 := owner.To4()
	if v4 == nil {
		return nil
	}
	entry := ADDR_ENTRY{Address: binary.BigEndian.Uint32(v4)}

	return &NBNSPacket{
		Header: NBNSHeader{
			TransactionID: generateTransactionID(),
			// R=1, OPCODE=registration, AA|RD|RA set, RCODE=CFT_ERR
			// (RFC 1002 4.2.15 packet diagram).
			Flags:   FlagResponse | OpConflict | FlagAuthoritative | FlagRecursion | FlagRecursionAvailable | RcodeConflict,
			Answers: 1,
		},
		Answers: []NBNSResourceRecord{{
			Name:     name,
			Type:     QuestionTypeNB,
			Class:    QuestionClassIn,
			TTL:      0,
			RDLength: uint16(entry.Length()),
			RData:    entry.Marshal(),
		}},
	}
}
