// Package mgmt implements a client for the DCE/RPC remote management (mgmt)
// interface over the connectionless (v4) transport ([C706] Appendix Q).
//
// The mgmt interface is a standard, well-known interface every RPC server exposes,
// which makes it a useful reconnaissance and liveness target: inq_if_ids enumerates
// the interfaces a server has registered, and is_server_listening probes whether the
// server is accepting calls. Both operations take only the binding handle as input
// (so the request stub is empty) and differ only in how their response is decoded,
// which makes them a clean demonstration of the connectionless binding pattern.
//
// References:
//   - [C706] Appendix Q (remote management interface):
//     https://pubs.opengroup.org/onlinepubs/9629399/apdxq.htm
package mgmt

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/interfaces"
	"github.com/TheManticoreProject/Manticore/windows/guid"
)

// Remote management (mgmt) interface identity ([C706] Appendix Q).
var (
	// Interface is the mgmt interface UUID, afa8bd80-7d8a-11c9-bef4-08002b102989.
	Interface = guid.GUID{A: 0xafa8bd80, B: 0x7d8a, C: 0x11c9, D: 0xbef4, E: 0x08002b102989}
)

const (
	// InterfaceMajorVersion is the mgmt interface major version (1.0).
	InterfaceMajorVersion = 1
	// InterfaceMinorVersion is the mgmt interface minor version.
	InterfaceMinorVersion = 0

	// OpnumInqIfIds is rpc__mgmt_inq_if_ids.
	OpnumInqIfIds = 0
	// OpnumIsServerListening is rpc__mgmt_is_server_listening.
	OpnumIsServerListening = 2
)

// maxIfIds bounds the interface-id count accepted from a response before allocating,
// guarding against a malformed count.
const maxIfIds = 4096

// IfID is an interface identifier returned by inq_if_ids: an interface UUID and its
// major/minor version (rpc_if_id_t).
type IfID struct {
	UUID         guid.GUID
	VersionMajor uint16
	VersionMinor uint16
}

// String renders the interface id as uuid vMAJOR.MINOR.
func (id IfID) String() string {
	return fmt.Sprintf("%s v%d.%d", id.UUID.ToFormatD(), id.VersionMajor, id.VersionMinor)
}

// Client invokes the mgmt interface over a connectionless RPC client.
type Client struct {
	binding *interfaces.Binding
}

// New returns a mgmt client bound over rpc, which should target the server's mgmt
// endpoint (the well-known RPC endpoint, or one resolved via the endpoint mapper).
func New(rpc *client.Client) *Client {
	return &Client{binding: interfaces.NewBinding(rpc, Interface, InterfaceMajorVersion, InterfaceMinorVersion)}
}

// IsServerListening calls rpc__mgmt_is_server_listening and reports whether the
// server is listening. A non-zero comm status from the operation is returned as an
// error.
func (c *Client) IsServerListening() (bool, error) {
	resp, err := c.binding.Invoke(OpnumIsServerListening, nil)
	if err != nil {
		return false, err
	}
	// Response stub: [out] error_status_t status, then the boolean32 return value
	// (the return value is marshalled after the out parameters).
	r := newReader(resp)
	status := r.u32()
	listening := r.u32()
	if err := r.err(); err != nil {
		return false, err
	}
	if status != 0 {
		return false, fmt.Errorf("mgmt: is_server_listening status 0x%08x", status)
	}
	return listening != 0, nil
}

// InqIfIds calls rpc__mgmt_inq_if_ids and returns the interface identifiers the
// server has registered.
func (c *Client) InqIfIds() ([]IfID, error) {
	resp, err := c.binding.Invoke(OpnumInqIfIds, nil)
	if err != nil {
		return nil, err
	}
	ids, status, err := parseInqIfIdsResponse(resp)
	if err != nil {
		return nil, err
	}
	if status != 0 {
		return nil, fmt.Errorf("mgmt: inq_if_ids status 0x%08x", status)
	}
	return ids, nil
}

// parseInqIfIdsResponse decodes the inq_if_ids [out] parameters:
//
//	rpc_if_id_vector_p_t if_id_vector  (full pointer to rpc_if_id_vector_t)
//	error_status_t       status
//
// rpc_if_id_vector_t is { unsigned32 count; [size_is(count)] rpc_if_id_p_t if_id[] };
// as a conformant struct its array maximum_count is hoisted to the front, the array
// elements are full pointers (referent ids), and each rpc_if_id_t pointee is
// { uuid_t uuid; unsigned16 vers_major; unsigned16 vers_minor } (20 octets).
func parseInqIfIdsResponse(data []byte) ([]IfID, uint32, error) {
	r := newReader(data)

	var ids []IfID
	if vecRef := r.u32(); vecRef != 0 {
		_ = r.u32() // hoisted conformant maximum_count
		count := r.u32()
		if r.err() != nil {
			return nil, 0, r.err()
		}
		if count > maxIfIds {
			return nil, 0, fmt.Errorf("mgmt: inq_if_ids returned implausible count %d", count)
		}
		refs := make([]uint32, count)
		for i := range refs {
			refs[i] = r.u32()
		}
		for _, ref := range refs {
			if ref == 0 {
				continue
			}
			id := IfID{UUID: r.uuid()}
			id.VersionMajor = r.u16()
			id.VersionMinor = r.u16()
			r.align(4)
			ids = append(ids, id)
		}
	}

	status := r.u32()
	if r.err() != nil {
		return nil, 0, r.err()
	}
	return ids, status, nil
}

// reader is a minimal bounds-checked little-endian NDR cursor; the first error is
// sticky and reported by err().
type reader struct {
	data []byte
	off  int
	fail error
}

func newReader(b []byte) *reader { return &reader{data: b} }

func (r *reader) err() error { return r.fail }

func (r *reader) take(n int) []byte {
	if r.fail != nil {
		return nil
	}
	if n < 0 || r.off+n > len(r.data) {
		r.fail = fmt.Errorf("mgmt: NDR underrun reading %d bytes at offset %d", n, r.off)
		return nil
	}
	b := r.data[r.off : r.off+n]
	r.off += n
	return b
}

func (r *reader) u16() uint16 {
	b := r.take(2)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint16(b)
}

func (r *reader) u32() uint32 {
	b := r.take(4)
	if b == nil {
		return 0
	}
	return binary.LittleEndian.Uint32(b)
}

func (r *reader) uuid() guid.GUID {
	var g guid.GUID
	if b := r.take(16); b != nil {
		g.FromRawBytes(b)
	}
	return g
}

func (r *reader) align(n int) {
	if r.fail != nil {
		return
	}
	for r.off%n != 0 {
		if r.off >= len(r.data) {
			r.fail = fmt.Errorf("mgmt: NDR underrun aligning to %d at offset %d", n, r.off)
			return
		}
		r.off++
	}
}
