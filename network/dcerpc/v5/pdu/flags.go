package pdu

import "strings"

// PFCFlags is the set of PFC_* flags carried in the pfc_flags field of the common
// header.
//
// References:
//   - [C706] section 12.6.3.1 (pfc_flags):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.3 PFC_SUPPORT_HEADER_SIGN and related flags:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/4886f349-2a73-4f9e-9262-a8404462c7e9
type PFCFlags uint8

const (
	// PFCFirstFrag marks the first fragment of a multi-fragment PDU.
	PFCFirstFrag PFCFlags = 0x01
	// PFCLastFrag marks the last fragment of a multi-fragment PDU.
	PFCLastFrag PFCFlags = 0x02
	// PFCPendingCancel signals a pending cancel (request) or a cancel was pending
	// (response).
	PFCPendingCancel PFCFlags = 0x04
	// PFCSupportHeaderSign requests or acknowledges header signing. It occupies the
	// same bit as PFCPendingCancel (0x04) but carries this meaning only on bind,
	// bind_ack, alter_context and alter_context_resp PDUs ([MS-RPCE] 2.2.2.3): the
	// client sets it on the bind to offer header signing and the server echoes it on
	// the bind_ack to accept. On request/response PDUs the same bit is
	// PFCPendingCancel, so this name must only be used on the context-negotiation
	// PDUs.
	PFCSupportHeaderSign PFCFlags = 0x04
	// PFCReserved is reserved.
	PFCReserved PFCFlags = 0x08
	// PFCConcMpx indicates support for concurrent multiplexing on the connection.
	PFCConcMpx PFCFlags = 0x10
	// PFCDidNotExecute indicates the server did not execute the call (response/fault).
	PFCDidNotExecute PFCFlags = 0x20
	// PFCMaybe indicates a "maybe" call with no response expected.
	PFCMaybe PFCFlags = 0x40
	// PFCObjectUuid indicates an object UUID is present in the PDU body.
	PFCObjectUuid PFCFlags = 0x80
)

// Has reports whether all bits in flag are set.
func (f PFCFlags) Has(flag PFCFlags) bool { return f&flag == flag }

// String returns a "|"-joined list of set flag names.
func (f PFCFlags) String() string {
	if f == 0 {
		return "none"
	}
	names := []struct {
		flag PFCFlags
		name string
	}{
		{PFCFirstFrag, "first_frag"},
		{PFCLastFrag, "last_frag"},
		{PFCPendingCancel, "pending_cancel"},
		{PFCReserved, "reserved"},
		{PFCConcMpx, "conc_mpx"},
		{PFCDidNotExecute, "did_not_execute"},
		{PFCMaybe, "maybe"},
		{PFCObjectUuid, "object_uuid"},
	}
	var set []string
	for _, n := range names {
		if f.Has(n.flag) {
			set = append(set, n.name)
		}
	}
	return strings.Join(set, "|")
}
