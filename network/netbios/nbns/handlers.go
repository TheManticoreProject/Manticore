package nbns

import (
	"encoding/binary"
	"time"
)

// PacketHandler provides common packet handling methods for both TCP and UDP servers
type PacketHandler struct {
	nbns *NetBIOSNameServer
}

// NewPacketHandler creates a new packet handler instance
func NewPacketHandler(nbns *NetBIOSNameServer) *PacketHandler {
	return &PacketHandler{
		nbns: nbns,
	}
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
