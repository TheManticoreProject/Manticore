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
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/client"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/interfaces"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v4/internal/ndr"
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
	r := ndr.NewReader(resp)
	status := r.U32()
	listening := r.U32()
	if err := r.Err(); err != nil {
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
	r := ndr.NewReader(data)

	var ids []IfID
	if vecRef := r.U32(); vecRef != 0 {
		_ = r.U32() // hoisted conformant maximum_count
		count := r.U32()
		if r.Err() != nil {
			return nil, 0, r.Err()
		}
		if count > maxIfIds {
			return nil, 0, fmt.Errorf("mgmt: inq_if_ids returned implausible count %d", count)
		}
		refs := make([]uint32, count)
		for i := range refs {
			refs[i] = r.U32()
		}
		for _, ref := range refs {
			if ref == 0 {
				continue
			}
			id := IfID{UUID: r.UUID()}
			id.VersionMajor = r.U16()
			id.VersionMinor = r.U16()
			r.Align(4)
			ids = append(ids, id)
		}
	}

	status := r.U32()
	if r.Err() != nil {
		return nil, 0, r.Err()
	}
	return ids, status, nil
}
