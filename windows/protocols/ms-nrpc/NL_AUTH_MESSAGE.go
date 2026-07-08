package msnrpc

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// NL_AUTH_MESSAGE message types and flag bits ([MS-NRPC] 2.2.1.3.1). The client sends a
// request message in the RPC bind auth_value; the server replies with a response message in
// the bind_ack. The flags describe which name fields the Buffer carries, in the fixed order
// NetBIOS domain, NetBIOS host, DNS domain, DNS host, then (last) the UTF-8 NetBIOS host.
const (
	NlAuthMessageNetbiosDomain   uint32 = 0x00000001
	NlAuthMessageNetbiosHost     uint32 = 0x00000002
	NlAuthMessageDNSDomain       uint32 = 0x00000004
	NlAuthMessageDNSHost         uint32 = 0x00000008
	NlAuthMessageNetbiosHostUTF8 uint32 = 0x00000010

	NlAuthMessageTypeRequest  uint32 = 0x00000000
	NlAuthMessageTypeResponse uint32 = 0x00000001
)

// NL_AUTH_MESSAGE ([MS-NRPC] 2.2.1.3.1) is the token carried in the RPC bind/bind_ack
// auth_value when Netlogon is its own security provider (RPC_C_AUTHN_NETLOGON). It is a
// fixed 8-byte header (MessageType, Flags) followed by a variable Buffer of null-terminated
// name strings selected by Flags. It is not NDR — it carries its own Marshal/Unmarshal.
type NL_AUTH_MESSAGE struct {
	MessageType uint32
	Flags       uint32
	Buffer      []byte
}

// Marshal serializes the token: the two little-endian header DWORDs then the Buffer.
func (m *NL_AUTH_MESSAGE) Marshal() []byte {
	out := make([]byte, 8+len(m.Buffer))
	binary.LittleEndian.PutUint32(out[0:4], m.MessageType)
	binary.LittleEndian.PutUint32(out[4:8], m.Flags)
	copy(out[8:], m.Buffer)
	return out
}

// Unmarshal parses a token from data. The Buffer is the remainder after the 8-byte header;
// its interpretation follows Flags, but a client only needs to confirm the server returned a
// response, so the fields are not further decoded here.
func (m *NL_AUTH_MESSAGE) Unmarshal(data []byte) error {
	if len(data) < 8 {
		return fmt.Errorf("NL_AUTH_MESSAGE truncated: have %d bytes, need at least 8", len(data))
	}
	m.MessageType = binary.LittleEndian.Uint32(data[0:4])
	m.Flags = binary.LittleEndian.Uint32(data[4:8])
	m.Buffer = append([]byte(nil), data[8:]...)
	return nil
}

// compressedUTF8DNS encodes a DNS domain as length-prefixed labels terminated by a zero
// octet ([MS-NRPC] 2.2.1.3.1, the DNS-domain field uses the [RFC1035] label form without
// pointer compression), e.g. "corp.example.com" -> 04 'corp' 07 'example' 03 'com' 00.
func compressedUTF8DNS(domain string) []byte {
	var out []byte
	for _, label := range strings.Split(domain, ".") {
		out = append(out, byte(len(label)))
		out = append(out, label...)
	}
	return append(out, 0x00)
}

// BuildClientNlAuthMessage builds the client request NL_AUTH_MESSAGE for the RPC bind
// ([MS-NRPC] 3.1.4.1). computerName is the client NetBIOS host name (no trailing '$'). A
// domain containing a '.' is treated as a DNS domain (NetBIOS host + DNS domain fields);
// otherwise it is treated as a NetBIOS domain (NetBIOS domain + host fields). Both forms
// additionally append the UTF-8 NetBIOS host field, matching what Windows and impacket send.
func BuildClientNlAuthMessage(computerName, domain string) *NL_AUTH_MESSAGE {
	m := &NL_AUTH_MESSAGE{MessageType: NlAuthMessageTypeRequest}
	if strings.Contains(domain, ".") {
		m.Flags = NlAuthMessageNetbiosHost | NlAuthMessageDNSDomain
		m.Buffer = append(m.Buffer, computerName...)
		m.Buffer = append(m.Buffer, 0x00)
		m.Buffer = append(m.Buffer, compressedUTF8DNS(domain)...)
	} else {
		m.Flags = NlAuthMessageNetbiosDomain | NlAuthMessageNetbiosHost
		m.Buffer = append(m.Buffer, domain...)
		m.Buffer = append(m.Buffer, 0x00)
		m.Buffer = append(m.Buffer, computerName...)
		m.Buffer = append(m.Buffer, 0x00)
	}
	// The UTF-8 NetBIOS host is always appended last: a length octet, the name, and a NUL.
	m.Flags |= NlAuthMessageNetbiosHostUTF8
	m.Buffer = append(m.Buffer, byte(len(computerName)))
	m.Buffer = append(m.Buffer, computerName...)
	m.Buffer = append(m.Buffer, 0x00)
	return m
}
