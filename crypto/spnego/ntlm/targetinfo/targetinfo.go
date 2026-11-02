package targetinfo

import (
	"crypto/md5"
	"encoding/binary"
	"errors"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// ParseTargetInfo parses the target info from a challenge message
func ParseTargetInfo(targetInfo []byte) (map[avpair.AvId][]byte, error) {
	result := make(map[avpair.AvId][]byte)

	offset := 0
	for offset < len(targetInfo) {
		// Need at least 4 bytes for the AV_PAIR header
		if offset+4 > len(targetInfo) {
			return nil, errors.New("target info truncated")
		}

		avId := avpair.AvId(binary.LittleEndian.Uint16(targetInfo[offset : offset+2]))
		offset += 2

		avLen := binary.LittleEndian.Uint16(targetInfo[offset : offset+2])
		offset += 2

		// Check if we have enough bytes for the value
		if offset+int(avLen) > len(targetInfo) {
			return nil, errors.New("target info value truncated")
		}

		// Extract the value
		if avId != avpair.MsvAvEOL {
			result[avId] = targetInfo[offset : offset+int(avLen)]
		}

		offset += int(avLen)

		// If we reached the end of list marker, we're done
		if avId == avpair.MsvAvEOL {
			break
		}
	}

	return result, nil
}

// HasTimestamp reports whether MsvAvTimestamp is present in the TargetInfo.
func HasTimestamp(targetInfo []byte) bool {
	pairs, err := ParseTargetInfo(targetInfo)
	if err != nil {
		return false
	}
	_, ok := pairs[avpair.MsvAvTimestamp]
	return ok
}

// GetTimestamp returns the raw 8-byte Windows FILETIME from TargetInfo, or nil if absent.
func GetTimestamp(targetInfo []byte) []byte {
	pairs, err := ParseTargetInfo(targetInfo)
	if err != nil {
		return nil
	}
	ts, ok := pairs[avpair.MsvAvTimestamp]
	if !ok {
		return nil
	}
	return ts
}

// channelBindingsLength is the size of an MsvAvChannelBindings value: the MD5 hash
// of a gss_channel_bindings_struct, or all zeroes when the application supplied no
// bindings ([MS-NLMP] 2.2.2.1).
const channelBindingsLength = 16

// BlobOptions are the client assertions BuildBlobTargetInfoWithOptions writes into
// the NTLMv2 blob. The zero value reproduces BuildBlobTargetInfo.
type BlobOptions struct {
	// TargetName is the SPN of the service being authenticated to, e.g.
	// "ldap/dc01.corp.local" or "host/dc01.corp.local", sent as MsvAvTargetName.
	// Empty selects "cifs/<DnsComputerName>" from the challenge, which is what a
	// Windows SMB client sends.
	TargetName string

	// ChannelBindings is the application data of the gss_channel_bindings_struct
	// the authentication is bound to, e.g. the RFC 5929 "tls-server-end-point:"
	// token of a TLS channel. MsvAvChannelBindings carries the MD5 of the
	// serialized structure (see ChannelBindingsHash). Nil sends the all-zero
	// value of a client with no bindings.
	ChannelBindings []byte
}

// ChannelBindingsHash returns the MsvAvChannelBindings value for the given channel
// binding application data: the MD5 hash of the gss_channel_bindings_struct
// ([MS-NLMP] 2.2.2.1) holding it, with empty initiator and acceptor addresses. The
// structure is serialized as RFC 2744 3.11 lays it out, with little-endian
// address-type and length fields, as Windows SSPI and MIT krb5 hash it. Empty
// application data means no bindings, which is the all-zero value.
func ChannelBindingsHash(applicationData []byte) []byte {
	if len(applicationData) == 0 {
		return make([]byte, channelBindingsLength)
	}
	// initiator_addrtype, initiator_address.length, acceptor_addrtype and
	// acceptor_address.length are all zero; application_data.length follows.
	structure := make([]byte, 16, 20+len(applicationData))
	structure = binary.LittleEndian.AppendUint32(structure, uint32(len(applicationData)))
	structure = append(structure, applicationData...)
	sum := md5.Sum(structure)
	return sum[:]
}

// BuildBlobTargetInfo constructs the modified TargetInfo to embed in the NTLMv2 blob,
// with the default SPN and no channel bindings (see BuildBlobTargetInfoWithOptions).
func BuildBlobTargetInfo(targetInfo []byte, micPresent bool) []byte {
	return BuildBlobTargetInfoWithOptions(targetInfo, micPresent, BlobOptions{})
}

// BuildBlobTargetInfoWithOptions constructs the modified TargetInfo to embed in the
// NTLMv2 blob.
//
// It copies all AVPairs from the challenge TargetInfo and inserts an MsvAvTargetName
// AVPair before the EOL marker: opts.TargetName when set, otherwise the SMB service
// principal name "cifs/<DnsComputerName>" when the challenge carries a DNS computer
// name (MsvAvDnsComputerName). This mirrors what the Windows client sends; modern
// Windows servers require the SPN in the AUTHENTICATE's NTLMv2 AVPairs and reject the
// authentication (STATUS_INVALID_PARAMETER) when it is absent.
//
// MsvAvChannelBindings is always inserted, holding ChannelBindingsHash of
// opts.ChannelBindings, so all zeroes when the caller supplied no bindings.
//
// When the AUTHENTICATE will carry a MIC and the challenge TargetInfo already holds an
// MsvAvFlags AVPair, the MIC-present bit is set in that pair rather than a second one
// being added (MS-NLMP 3.1.5.1.2). The MsvAvChannelBindings and MsvAvTargetName pairs
// are the client's own assertions, so any copy of them in the challenge TargetInfo is
// replaced rather than duplicated.
func BuildBlobTargetInfoWithOptions(targetInfo []byte, micPresent bool, opts BlobOptions) []byte {
	// The SPN is UTF-16LE on the wire. The default is the SMB SPN built from the
	// DNS computer name AVPair, which is already UTF-16LE.
	var targetName []byte
	if opts.TargetName != "" {
		targetName = utf16.EncodeUTF16LE(opts.TargetName)
	} else if pairs, err := ParseTargetInfo(targetInfo); err == nil {
		if dnsComputer, ok := pairs[avpair.MsvAvDnsComputerName]; ok && len(dnsComputer) > 0 {
			targetName = append(utf16.EncodeUTF16LE("cifs/"), dnsComputer...)
		}
	}
	channelBindings := ChannelBindingsHash(opts.ChannelBindings)

	result := make([]byte, 0, len(targetInfo)+4+len(targetName))
	flagsWritten := false
	i := 0
	for i+4 <= len(targetInfo) {
		currentID := avpair.AvId(binary.LittleEndian.Uint16(targetInfo[i : i+2]))
		avLen := binary.LittleEndian.Uint16(targetInfo[i+2 : i+4])

		if currentID == avpair.MsvAvEOL {
			// Insert MsvAvFlags before EOL when the AUTHENTICATE will carry a MIC:
			// that bit is how a server detects one (MS-NLMP 3.2.5.1.2). It is added
			// before the NT response is computed, so it is covered by the NTProofStr
			// as well as by the MIC itself.
			if micPresent && !flagsWritten {
				hdr := make([]byte, 4)
				binary.LittleEndian.PutUint16(hdr[0:2], uint16(avpair.MsvAvFlags))
				binary.LittleEndian.PutUint16(hdr[2:4], 4)
				value := make([]byte, 4)
				binary.LittleEndian.PutUint32(value, avpair.MsvAvFlagMICPresent)
				result = append(result, hdr...)
				result = append(result, value...)
			}

			// Insert MsvAvChannelBindings before EOL. A client with no application-
			// supplied channel bindings sends the pair with an all-zero 16-byte
			// value, which is what a Windows client emits; omitting the pair
			// entirely is what a server enforcing channel binding (EPA) rejects
			// (MS-NLMP 3.2.5.1.2).
			hdr := make([]byte, 4)
			binary.LittleEndian.PutUint16(hdr[0:2], uint16(avpair.MsvAvChannelBindings))
			binary.LittleEndian.PutUint16(hdr[2:4], channelBindingsLength)
			result = append(result, hdr...)
			result = append(result, channelBindings...)

			// Insert MsvAvTargetName (SPN) before EOL, then append EOL.
			if len(targetName) > 0 {
				hdr := make([]byte, 4)
				binary.LittleEndian.PutUint16(hdr[0:2], uint16(avpair.MsvAvTargetName))
				binary.LittleEndian.PutUint16(hdr[2:4], uint16(len(targetName)))
				result = append(result, hdr...)
				result = append(result, targetName...)
			}
			result = append(result, targetInfo[i:i+4]...)
			return result
		}

		// avLen is server-controlled; ensure the declared value bytes are present
		// before slicing, so a length running past the buffer cannot panic. A
		// malformed (truncated) tail stops processing rather than crashing.
		if i+4+int(avLen) > len(targetInfo) {
			break
		}
		switch currentID {
		case avpair.MsvAvFlags:
			// Keep a single MsvAvFlags, the first well-formed one, with the
			// MIC-present bit set in it when a MIC will be sent.
			if avLen == 4 && !flagsWritten {
				value := binary.LittleEndian.Uint32(targetInfo[i+4 : i+8])
				if micPresent {
					value |= avpair.MsvAvFlagMICPresent
				}
				pair := append([]byte(nil), targetInfo[i:i+4]...)
				pair = binary.LittleEndian.AppendUint32(pair, value)
				result = append(result, pair...)
				flagsWritten = true
			}
		case avpair.MsvAvChannelBindings, avpair.MsvAvTargetName:
			// Written by the client before EOL; drop the challenge's copy.
		default:
			result = append(result, targetInfo[i:i+4+int(avLen)]...)
		}
		i += 4 + int(avLen)
	}

	return result
}
