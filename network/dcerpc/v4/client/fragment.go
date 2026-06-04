package client

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/pdu"
)

// fragmentRequest splits a request stub into one or more request PDUs whose
// marshalled size does not exceed maxPDU bytes. The fragments share tmpl as their
// header (activity, interface, sequence number, object, and any caller flags such as
// idempotent are taken from it); each fragment gets its own fragnum and the
// frag/lastfrag flags set per [C706] section 12.6.3.2:
//
//   - a single, unfragmented request sets neither frag nor lastfrag;
//   - a multi-fragment request sets frag on every fragment and additionally sets
//     lastfrag on the final one.
//
// References:
//   - [C706] section 12.6.3.2 (flags1) and chapter 10 (fragment transmission):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap10.htm
func fragmentRequest(tmpl pdu.Header, stub []byte, maxPDU int) ([]pdu.PDU, error) {
	maxBody := maxPDU - pdu.HeaderSize
	if maxBody <= 0 {
		return nil, fmt.Errorf("dcerpc(cl): max PDU size %d too small for an %d-byte header", maxPDU, pdu.HeaderSize)
	}

	var frags []pdu.PDU
	for off := 0; off < len(stub); off += maxBody {
		end := min(off+maxBody, len(stub))
		h := tmpl
		h.PacketType = pdu.PacketTypeRequest
		h.FragmentNumber = uint16(len(frags))
		frags = append(frags, pdu.PDU{Header: h, Body: stub[off:end]})
	}
	// An empty stub is still one request PDU carrying a zero-length body.
	if len(frags) == 0 {
		h := tmpl
		h.PacketType = pdu.PacketTypeRequest
		h.FragmentNumber = 0
		frags = append(frags, pdu.PDU{Header: h})
	}

	if len(frags) > 1 {
		if len(frags) > 0xFFFF {
			return nil, fmt.Errorf("dcerpc(cl): request needs %d fragments, exceeds the 16-bit fragnum space", len(frags))
		}
		for i := range frags {
			frags[i].Header.Flags1 |= pdu.Flags1Frag
		}
		frags[len(frags)-1].Header.Flags1 |= pdu.Flags1LastFrag
	}
	return frags, nil
}

// responseReassembler accumulates response PDU fragments and reports when the full
// response has arrived. A single, unfragmented response (neither frag nor lastfrag
// set) is complete on its own; a fragmented response is complete once the lastfrag
// PDU and every fragnum from 0 through it have been seen.
type responseReassembler struct {
	frags    map[uint16][]byte
	haveLast bool
	lastNum  uint16
}

// add records a response fragment. Duplicate fragnums overwrite, which is harmless
// since retransmitted fragments carry identical data.
func (r *responseReassembler) add(h pdu.Header, body []byte) {
	if r.frags == nil {
		r.frags = make(map[uint16][]byte)
	}
	cp := append([]byte(nil), body...)
	if !h.Flags1.Has(pdu.Flags1Frag) && !h.Flags1.Has(pdu.Flags1LastFrag) {
		// Single, unfragmented response.
		r.frags[0] = cp
		r.haveLast = true
		r.lastNum = 0
		return
	}
	r.frags[h.FragmentNumber] = cp
	if h.Flags1.Has(pdu.Flags1LastFrag) {
		r.haveLast = true
		r.lastNum = h.FragmentNumber
	}
}

// complete reports whether every fragment up to and including the last has arrived.
func (r *responseReassembler) complete() bool {
	if !r.haveLast {
		return false
	}
	for i := uint16(0); i <= r.lastNum; i++ {
		if _, ok := r.frags[i]; !ok {
			return false
		}
		if i == 0xFFFF { // guard against uint16 wrap when lastNum is 0xFFFF
			break
		}
	}
	return true
}

// assemble concatenates the fragments in fragnum order. It is only meaningful once
// complete returns true.
func (r *responseReassembler) assemble() []byte {
	var out []byte
	for i := uint16(0); i <= r.lastNum; i++ {
		out = append(out, r.frags[i]...)
		if i == 0xFFFF {
			break
		}
	}
	return out
}
