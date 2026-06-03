package pdu

import (
	"encoding/binary"
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
)

// Presentation context negotiation result codes (p_cont_def_result_t), as returned in
// the result field of each p_result_t in a bind_ack PDU.
//
// References:
//   - [C706] section 12.6.3.1 ("The bind_ack PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.4 Presentation Context and Transfer Syntax Negotiation:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/ca4d3552-4647-4f40-830b-fd2090adec8f
//   - [MS-RPCE] negotiate_ack member of p_cont_def_result_t:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/8df5c4d4-364d-468c-81fe-ec94c1b40917
const (
	ResultAcceptance        uint16 = 0
	ResultUserRejection     uint16 = 1
	ResultProviderRejection uint16 = 2
	ResultNegotiateAck      uint16 = 3 // [MS-RPCE] bind-time feature negotiation
)

// Provider reject reasons (p_provider_reason_t / p_reject_reason_t).
//
// Reference: [C706] section 12.6.3.1:
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
const (
	ReasonNotSpecified                         uint16 = 0
	ReasonAbstractSyntaxNotSupported           uint16 = 1
	ReasonProposedTransferSyntaxesNotSupported uint16 = 2
	ReasonLocalLimitExceeded                   uint16 = 3
)

// align4 returns the number of padding bytes needed to align offset to a 4-byte
// boundary.
func align4(offset int) int { return (4 - (offset % 4)) % 4 }

// ContextElement is a single presentation context proposed in a bind (or
// alter_context) PDU: a context id, an abstract syntax (the target interface), and
// one or more candidate transfer syntaxes (p_cont_elem_t).
//
// References:
//   - [C706] section 12.6.3.1 ("The bind PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.4 Presentation Context and Transfer Syntax Negotiation:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/ca4d3552-4647-4f40-830b-fd2090adec8f
type ContextElement struct {
	ContextID        uint16
	AbstractSyntax   syntax.SyntaxID
	TransferSyntaxes []syntax.SyntaxID
}

// Marshal serializes a presentation context element.
func (c *ContextElement) Marshal() ([]byte, error) {
	if len(c.TransferSyntaxes) == 0 {
		return nil, fmt.Errorf("context element %d has no transfer syntaxes", c.ContextID)
	}
	if len(c.TransferSyntaxes) > 0xFF {
		return nil, fmt.Errorf("context element %d has too many transfer syntaxes (%d)", c.ContextID, len(c.TransferSyntaxes))
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], c.ContextID)
	buf[2] = byte(len(c.TransferSyntaxes))
	buf[3] = 0 // reserved

	abs, err := c.AbstractSyntax.Marshal()
	if err != nil {
		return nil, fmt.Errorf("context element %d: %w", c.ContextID, err)
	}
	buf = append(buf, abs...)

	for i := range c.TransferSyntaxes {
		ts, err := c.TransferSyntaxes[i].Marshal()
		if err != nil {
			return nil, fmt.Errorf("context element %d transfer syntax %d: %w", c.ContextID, i, err)
		}
		buf = append(buf, ts...)
	}
	return buf, nil
}

// Unmarshal parses a presentation context element and returns the bytes consumed.
func (c *ContextElement) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("context element header truncated")
	}
	c.ContextID = binary.LittleEndian.Uint16(data[0:2])
	n := int(data[2])
	pos := 4

	read, err := c.AbstractSyntax.Unmarshal(data[pos:])
	if err != nil {
		return 0, fmt.Errorf("context element %d abstract syntax: %w", c.ContextID, err)
	}
	pos += read

	c.TransferSyntaxes = make([]syntax.SyntaxID, 0, n)
	for i := 0; i < n; i++ {
		var ts syntax.SyntaxID
		read, err := ts.Unmarshal(data[pos:])
		if err != nil {
			return 0, fmt.Errorf("context element %d transfer syntax %d: %w", c.ContextID, i, err)
		}
		pos += read
		c.TransferSyntaxes = append(c.TransferSyntaxes, ts)
	}
	return pos, nil
}

// Bind is a bind PDU: it proposes the maximum fragment sizes, an association group,
// and a list of presentation contexts to negotiate.
//
// References:
//   - [C706] section 12.6.3.1 ("The bind PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.4 Presentation Context and Transfer Syntax Negotiation:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/ca4d3552-4647-4f40-830b-fd2090adec8f
type Bind struct {
	Header       Header
	MaxXmitFrag  uint16
	MaxRecvFrag  uint16
	AssocGroupID uint32
	ContextList  []ContextElement
}

// Marshal serializes the complete bind PDU, filling in the header's packet type and
// frag_length.
func (b *Bind) Marshal() ([]byte, error) {
	if len(b.ContextList) == 0 {
		return nil, fmt.Errorf("bind PDU has no presentation contexts")
	}
	if len(b.ContextList) > 0xFF {
		return nil, fmt.Errorf("bind PDU has too many presentation contexts (%d)", len(b.ContextList))
	}

	body := make([]byte, 8)
	binary.LittleEndian.PutUint16(body[0:2], b.MaxXmitFrag)
	binary.LittleEndian.PutUint16(body[2:4], b.MaxRecvFrag)
	binary.LittleEndian.PutUint32(body[4:8], b.AssocGroupID)

	cl := make([]byte, 4)
	cl[0] = byte(len(b.ContextList)) // n_context_elem
	// cl[1] reserved, cl[2:4] reserved2
	for i := range b.ContextList {
		elem, err := b.ContextList[i].Marshal()
		if err != nil {
			return nil, err
		}
		cl = append(cl, elem...)
	}
	body = append(body, cl...)

	if b.Header.RPCVersion == 0 && b.Header.DataRepresentation == ([4]byte{}) {
		b.Header = NewHeader(PacketTypeBind, b.Header.PacketFlags, b.Header.CallID)
	}
	b.Header.PacketType = PacketTypeBind
	b.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := b.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete bind PDU and returns the bytes consumed.
func (b *Bind) Unmarshal(data []byte) (int, error) {
	pos, err := b.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if b.Header.PacketType != PacketTypeBind {
		return 0, fmt.Errorf("not a bind PDU: packet type is %s", b.Header.PacketType)
	}
	if len(data) < pos+12 {
		return 0, fmt.Errorf("bind PDU truncated")
	}
	b.MaxXmitFrag = binary.LittleEndian.Uint16(data[pos : pos+2])
	b.MaxRecvFrag = binary.LittleEndian.Uint16(data[pos+2 : pos+4])
	b.AssocGroupID = binary.LittleEndian.Uint32(data[pos+4 : pos+8])
	n := int(data[pos+8]) // n_context_elem
	pos += 12             // skip n_context_elem + reserved + reserved2

	b.ContextList = make([]ContextElement, 0, n)
	for i := 0; i < n; i++ {
		var elem ContextElement
		read, err := elem.Unmarshal(data[pos:])
		if err != nil {
			return 0, err
		}
		pos += read
		b.ContextList = append(b.ContextList, elem)
	}
	return pos, nil
}

// String returns a one-line summary.
func (b *Bind) String() string {
	return fmt.Sprintf("bind xmit=%d recv=%d assoc_group=%d contexts=%d", b.MaxXmitFrag, b.MaxRecvFrag, b.AssocGroupID, len(b.ContextList))
}

// PresentationResult is one server response in a bind_ack p_result_list (p_result_t):
// the negotiation result, an optional reason, and the accepted transfer syntax.
//
// Reference: [C706] section 12.6.3.1 ("The bind_ack PDU"):
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type PresentationResult struct {
	Result         uint16
	Reason         uint16
	TransferSyntax syntax.SyntaxID
}

// Marshal serializes a single presentation result (24 bytes).
func (r *PresentationResult) Marshal() ([]byte, error) {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint16(buf[0:2], r.Result)
	binary.LittleEndian.PutUint16(buf[2:4], r.Reason)
	ts, err := r.TransferSyntax.Marshal()
	if err != nil {
		return nil, err
	}
	return append(buf, ts...), nil
}

// Unmarshal parses a single presentation result and returns the bytes consumed.
func (r *PresentationResult) Unmarshal(data []byte) (int, error) {
	if len(data) < 4 {
		return 0, fmt.Errorf("presentation result truncated")
	}
	r.Result = binary.LittleEndian.Uint16(data[0:2])
	r.Reason = binary.LittleEndian.Uint16(data[2:4])
	read, err := r.TransferSyntax.Unmarshal(data[4:])
	if err != nil {
		return 0, err
	}
	return 4 + read, nil
}

// BindAck is a bind_ack PDU: the server's response to a bind, echoing the negotiated
// fragment sizes and reporting per-context negotiation results.
//
// References:
//   - [C706] section 12.6.3.1 ("The bind_ack PDU"):
//     https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
//   - [MS-RPCE] 2.2.2.4 Presentation Context and Transfer Syntax Negotiation:
//     https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-rpce/ca4d3552-4647-4f40-830b-fd2090adec8f
type BindAck struct {
	Header           Header
	MaxXmitFrag      uint16
	MaxRecvFrag      uint16
	AssocGroupID     uint32
	SecondaryAddress string // port_spec, without the trailing NUL
	Results          []PresentationResult
}

// Marshal serializes the complete bind_ack PDU, including the 4-byte alignment pad
// after the secondary address.
func (b *BindAck) Marshal() ([]byte, error) {
	body := make([]byte, 8)
	binary.LittleEndian.PutUint16(body[0:2], b.MaxXmitFrag)
	binary.LittleEndian.PutUint16(body[2:4], b.MaxRecvFrag)
	binary.LittleEndian.PutUint32(body[4:8], b.AssocGroupID)

	// sec_addr (port_any_t): a length followed by a NUL-terminated string. A length
	// of 0 means no string at all.
	if b.SecondaryAddress == "" {
		body = append(body, 0x00, 0x00)
	} else {
		addr := append([]byte(b.SecondaryAddress), 0x00)
		lenBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(lenBuf, uint16(len(addr)))
		body = append(body, lenBuf...)
		body = append(body, addr...)
	}

	// Pad to a 4-byte boundary measured from the start of the PDU.
	pad := align4(HeaderSize + len(body))
	body = append(body, make([]byte, pad)...)

	rl := make([]byte, 4)
	rl[0] = byte(len(b.Results)) // n_results
	for i := range b.Results {
		r, err := b.Results[i].Marshal()
		if err != nil {
			return nil, err
		}
		rl = append(rl, r...)
	}
	body = append(body, rl...)

	if b.Header.RPCVersion == 0 && b.Header.DataRepresentation == ([4]byte{}) {
		b.Header = NewHeader(PacketTypeBindAck, b.Header.PacketFlags, b.Header.CallID)
	}
	b.Header.PacketType = PacketTypeBindAck
	b.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := b.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete bind_ack PDU and returns the bytes consumed.
func (b *BindAck) Unmarshal(data []byte) (int, error) {
	pos, err := b.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if b.Header.PacketType != PacketTypeBindAck {
		return 0, fmt.Errorf("not a bind_ack PDU: packet type is %s", b.Header.PacketType)
	}
	if len(data) < pos+10 {
		return 0, fmt.Errorf("bind_ack PDU truncated")
	}
	b.MaxXmitFrag = binary.LittleEndian.Uint16(data[pos : pos+2])
	b.MaxRecvFrag = binary.LittleEndian.Uint16(data[pos+2 : pos+4])
	b.AssocGroupID = binary.LittleEndian.Uint32(data[pos+4 : pos+8])
	pos += 8

	addrLen := int(binary.LittleEndian.Uint16(data[pos : pos+2]))
	pos += 2
	if len(data) < pos+addrLen {
		return 0, fmt.Errorf("bind_ack secondary address truncated")
	}
	if addrLen > 0 {
		addr := data[pos : pos+addrLen]
		// Strip the trailing NUL if present.
		if n := len(addr); n > 0 && addr[n-1] == 0x00 {
			addr = addr[:n-1]
		}
		b.SecondaryAddress = string(addr)
	}
	pos += addrLen

	// Skip the 4-byte alignment pad.
	pos += align4(pos)

	if len(data) < pos+4 {
		return 0, fmt.Errorf("bind_ack result list truncated")
	}
	n := int(data[pos]) // n_results
	pos += 4            // skip n_results + reserved + reserved2

	b.Results = make([]PresentationResult, 0, n)
	for i := 0; i < n; i++ {
		var r PresentationResult
		read, err := r.Unmarshal(data[pos:])
		if err != nil {
			return 0, err
		}
		pos += read
		b.Results = append(b.Results, r)
	}
	return pos, nil
}

// Accepted reports whether at least one presentation context was accepted.
func (b *BindAck) Accepted() bool {
	for _, r := range b.Results {
		if r.Result == ResultAcceptance {
			return true
		}
	}
	return false
}

// String returns a one-line summary.
func (b *BindAck) String() string {
	return fmt.Sprintf("bind_ack xmit=%d recv=%d assoc_group=%d sec_addr=%q results=%d", b.MaxXmitFrag, b.MaxRecvFrag, b.AssocGroupID, b.SecondaryAddress, len(b.Results))
}

// ProtocolVersion is one supported RPC protocol version (p_rt_version_t).
type ProtocolVersion struct {
	Major uint8
	Minor uint8
}

// BindNak is a bind_nak PDU: the server's rejection of a bind, with a reason and the
// list of RPC protocol versions it supports.
//
// Reference: [C706] section 12.6.3.1 ("The bind_nak PDU"):
// https://pubs.opengroup.org/onlinepubs/9629399/chap12.htm
type BindNak struct {
	Header       Header
	RejectReason uint16
	Versions     []ProtocolVersion
}

// Marshal serializes the complete bind_nak PDU.
func (b *BindNak) Marshal() ([]byte, error) {
	if len(b.Versions) > 0xFF {
		return nil, fmt.Errorf("bind_nak has too many protocol versions (%d)", len(b.Versions))
	}
	body := make([]byte, 2)
	binary.LittleEndian.PutUint16(body[0:2], b.RejectReason)
	body = append(body, byte(len(b.Versions)))
	for _, v := range b.Versions {
		body = append(body, v.Major, v.Minor)
	}

	if b.Header.RPCVersion == 0 && b.Header.DataRepresentation == ([4]byte{}) {
		b.Header = NewHeader(PacketTypeBindNak, b.Header.PacketFlags, b.Header.CallID)
	}
	b.Header.PacketType = PacketTypeBindNak
	b.Header.FragLength = uint16(HeaderSize + len(body))

	hdr, err := b.Header.Marshal()
	if err != nil {
		return nil, err
	}
	return append(hdr, body...), nil
}

// Unmarshal parses a complete bind_nak PDU and returns the bytes consumed.
func (b *BindNak) Unmarshal(data []byte) (int, error) {
	pos, err := b.Header.Unmarshal(data)
	if err != nil {
		return 0, err
	}
	if b.Header.PacketType != PacketTypeBindNak {
		return 0, fmt.Errorf("not a bind_nak PDU: packet type is %s", b.Header.PacketType)
	}
	if len(data) < pos+3 {
		return 0, fmt.Errorf("bind_nak PDU truncated")
	}
	b.RejectReason = binary.LittleEndian.Uint16(data[pos : pos+2])
	n := int(data[pos+2])
	pos += 3
	if len(data) < pos+2*n {
		return 0, fmt.Errorf("bind_nak versions truncated")
	}
	b.Versions = make([]ProtocolVersion, 0, n)
	for i := 0; i < n; i++ {
		b.Versions = append(b.Versions, ProtocolVersion{Major: data[pos], Minor: data[pos+1]})
		pos += 2
	}
	return pos, nil
}

// String returns a one-line summary.
func (b *BindNak) String() string {
	return fmt.Sprintf("bind_nak reject_reason=%d versions=%d", b.RejectReason, len(b.Versions))
}
