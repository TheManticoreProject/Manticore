package targetinfo

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// MaxTargetInfoLength is the largest TargetInfo that can be advertised. The
// CHALLENGE_MESSAGE carries its length in a 16-bit field, so a longer list
// cannot be described on the wire.
const MaxTargetInfoLength = 0xFFFF

// Build marshals an AV_PAIR list into the TargetInfo form a CHALLENGE_MESSAGE
// carries, appending the MsvAvEOL terminator that ends the list.
//
// This package could previously only parse TargetInfo. An acceptor has to compose
// one: the pairs it advertises are what the client folds into its NTLMv2 blob, so
// they end up covered by the NTProofStr and cannot be added after the fact.
//
// Each pair's AvLen is taken from its value rather than from the AvLen field, so
// a caller cannot produce a list whose declared lengths disagree with its
// contents. A caller-supplied MsvAvEOL is ignored, since the terminator is
// appended here.
//
// Parameters:
//   - pairs: the AV_PAIRs to advertise, in order
//
// Returns:
//   - The marshalled TargetInfo, EOL-terminated
//   - An error if the list cannot be described on the wire
func Build(pairs []avpair.AvPair) ([]byte, error) {
	marshalled := []byte{}

	for i := range pairs {
		if pairs[i].AvID == avpair.MsvAvEOL {
			// The terminator is appended below; one supplied mid-list would
			// truncate the list for every reader.
			continue
		}
		if len(pairs[i].AvData) > 0xFFFF {
			return nil, fmt.Errorf("AV_PAIR %s value is %d bytes, which does not fit the 16-bit AvLen field",
				pairs[i].AvID, len(pairs[i].AvData))
		}

		// Take the length from the value so the two cannot disagree.
		pair := avpair.AvPair{
			AvID:   pairs[i].AvID,
			AvLen:  uint16(len(pairs[i].AvData)),
			AvData: pairs[i].AvData,
		}
		encoded, err := pair.Marshal()
		if err != nil {
			return nil, fmt.Errorf("failed to marshal AV_PAIR %s: %v", pair.AvID, err)
		}
		marshalled = append(marshalled, encoded...)
	}

	// MsvAvEOL terminates the list and carries no value.
	terminator := avpair.AvPair{AvID: avpair.MsvAvEOL, AvLen: 0, AvData: nil}
	encoded, err := terminator.Marshal()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal the MsvAvEOL terminator: %v", err)
	}
	marshalled = append(marshalled, encoded...)

	if len(marshalled) > MaxTargetInfoLength {
		return nil, fmt.Errorf("TargetInfo is %d bytes, which does not fit the 16-bit length field", len(marshalled))
	}

	return marshalled, nil
}

// BuildServerTargetInfo composes the TargetInfo an acceptor advertises: the
// NetBIOS and DNS computer and domain names, and a timestamp.
//
// The names are encoded UTF-16LE, which is how they appear on the wire. A name
// left empty is omitted rather than advertised as a zero-length pair. The
// timestamp is included only when non-empty; when it is present a client is
// required to carry a MIC in its AUTHENTICATE ([MS-NLMP] 3.1.5.1.2), so an
// acceptor that supplies one must be prepared to verify it.
//
// Parameters:
//   - netBIOSComputerName: the server's NetBIOS computer name
//   - netBIOSDomainName: the server's NetBIOS domain name
//   - dnsComputerName: the server's fully qualified computer name
//   - dnsDomainName: the server's fully qualified domain name
//   - timestamp: an 8-byte Windows FILETIME, or nil to omit it
//
// Returns:
//   - The marshalled TargetInfo, EOL-terminated
//   - An error if the list cannot be described on the wire
func BuildServerTargetInfo(netBIOSComputerName, netBIOSDomainName, dnsComputerName, dnsDomainName string, timestamp []byte) ([]byte, error) {
	pairs := []avpair.AvPair{}

	appendName := func(id avpair.AvId, name string) {
		if name == "" {
			return
		}
		pairs = append(pairs, avpair.AvPair{AvID: id, AvData: utf16.EncodeUTF16LE(name)})
	}

	// The order matches what a Windows server sends.
	appendName(avpair.MsvAvNbDomainName, netBIOSDomainName)
	appendName(avpair.MsvAvNbComputerName, netBIOSComputerName)
	appendName(avpair.MsvAvDnsDomainName, dnsDomainName)
	appendName(avpair.MsvAvDnsComputerName, dnsComputerName)

	if len(timestamp) > 0 {
		if len(timestamp) != 8 {
			return nil, fmt.Errorf("MsvAvTimestamp must be an 8-byte FILETIME, got %d bytes", len(timestamp))
		}
		pairs = append(pairs, avpair.AvPair{AvID: avpair.MsvAvTimestamp, AvData: timestamp})
	}

	return Build(pairs)
}
