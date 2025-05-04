package nbtns

import (
	"net"
	"time"
)

// PacketHandler provides common packet handling methods for both TCP and UDP servers
type PacketHandler struct {
	nbtns *NetBIOSNameServer
}

// NewPacketHandler creates a new packet handler instance
func NewPacketHandler(nbtns *NetBIOSNameServer) *PacketHandler {
	return &PacketHandler{
		nbtns: nbtns,
	}
}

// handleNameQuery processes a name query request
func (h *PacketHandler) handleNameQuery(request *NBTNSPacket, response *NBTNSPacket) {
	for _, q := range request.Questions {
		owners, nameType, err := h.nbtns.QueryName(q.Name.Name)
		if err != nil {
			response.Header.Flags |= RcodeNameError
			return
		}

		// Create resource record for each owner
		for _, owner := range owners {
			rr := NBTNSResourceRecord{
				Name:     q.Name,
				Type:     q.Type,
				Class:    q.Class,
				TTL:      uint32(24 * time.Hour.Seconds()), // 24 hour TTL
				RDLength: uint16(len(owner)),
				RData:    owner,
			}
			response.Answers = append(response.Answers, rr)
		}

		response.Header.Answers = uint16(len(response.Answers))

		// Set group bit if this is a group name
		if nameType == Group {
			response.Header.Flags |= 0x0080 // Group name bit
		}
	}
}

// handleRegistration processes a name registration request
func (h *PacketHandler) handleRegistration(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		nameType := Unique
		if request.Header.Flags&0x0080 != 0 {
			nameType = Group
		}

		err := h.nbtns.RegisterName(
			rr.Name.Name,
			nameType,
			net.IP(rr.RData),
			time.Duration(rr.TTL)*time.Second,
		)

		if err != nil {
			response.Header.Flags |= RcodeConflict
			return
		}
	}
}

// handleRelease processes a name release request
func (h *PacketHandler) handleRelease(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		if err := h.nbtns.ReleaseName(rr.Name.Name, net.IP(rr.RData)); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}

// handleRefresh processes a name refresh request
func (h *PacketHandler) handleRefresh(request *NBTNSPacket, response *NBTNSPacket) {
	for _, rr := range request.Answers {
		if err := h.nbtns.RefreshName(rr.Name.Name, net.IP(rr.RData)); err != nil {
			response.Header.Flags |= RcodeServerError
			return
		}
	}
}
