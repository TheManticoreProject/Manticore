package targetinfo_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// appendAvPair appends one AV_PAIR to a TargetInfo under construction.
func appendAvPair(info []byte, id avpair.AvId, value []byte) []byte {
	info = binary.LittleEndian.AppendUint16(info, uint16(id))
	info = binary.LittleEndian.AppendUint16(info, uint16(len(value)))
	return append(info, value...)
}

// countAvPairs returns how many AV_PAIRs with each AvId precede the EOL marker.
func countAvPairs(t *testing.T, info []byte) map[avpair.AvId]int {
	t.Helper()

	counts := make(map[avpair.AvId]int)
	for i := 0; i+4 <= len(info); {
		id := avpair.AvId(binary.LittleEndian.Uint16(info[i : i+2]))
		avLen := int(binary.LittleEndian.Uint16(info[i+2 : i+4]))
		if i+4+avLen > len(info) {
			t.Fatalf("AV_PAIR %s at offset %d declares %d bytes, past the end of the buffer", id, i, avLen)
		}
		if id == avpair.MsvAvEOL {
			break
		}
		counts[id]++
		i += 4 + avLen
	}
	return counts
}

// TestBlobMergesExistingMsvAvFlags checks a challenge TargetInfo that already carries
// MsvAvFlags (here with the constrained-authentication bit) gets the MIC-present bit
// set in that pair instead of a second MsvAvFlags being appended (MS-NLMP 3.1.5.1.2),
// and that client-asserted pairs echoed by the server are not duplicated.
func TestBlobMergesExistingMsvAvFlags(t *testing.T) {
	serverFlags := binary.LittleEndian.AppendUint32(nil, avpair.MsvAvFlagConstrainedAuth)
	var info []byte
	info = appendAvPair(info, avpair.MsvAvNbDomainName, utf16.EncodeUTF16LE("TMP"))
	info = appendAvPair(info, avpair.MsvAvDnsComputerName, utf16.EncodeUTF16LE("dc01.tmp.local"))
	info = appendAvPair(info, avpair.MsvAvFlags, serverFlags)
	info = appendAvPair(info, avpair.MsvAvTimestamp, make([]byte, 8))
	info = appendAvPair(info, avpair.MsvAvChannelBindings, bytes.Repeat([]byte{0xAA}, 16))
	info = appendAvPair(info, avpair.MsvAvTargetName, utf16.EncodeUTF16LE("bogus/name"))
	info = appendAvPair(info, avpair.MsvAvEOL, nil)

	for _, micPresent := range []bool{true, false} {
		blob := targetinfo.BuildBlobTargetInfo(info, micPresent)

		counts := countAvPairs(t, blob)
		for _, id := range []avpair.AvId{avpair.MsvAvFlags, avpair.MsvAvChannelBindings, avpair.MsvAvTargetName} {
			if counts[id] != 1 {
				t.Errorf("micPresent=%v: blob carries %d %s AV_PAIRs, want 1", micPresent, counts[id], id)
			}
		}

		_, values := avPairOffsets(t, blob)
		want := avpair.MsvAvFlagConstrainedAuth
		if micPresent {
			want |= avpair.MsvAvFlagMICPresent
		}
		if got := binary.LittleEndian.Uint32(values[avpair.MsvAvFlags]); got != want {
			t.Errorf("micPresent=%v: MsvAvFlags = 0x%08x, want 0x%08x", micPresent, got, want)
		}
		if !bytes.Equal(values[avpair.MsvAvChannelBindings], make([]byte, 16)) {
			t.Errorf("micPresent=%v: MsvAvChannelBindings = % x, want the client's all-zero value", micPresent, values[avpair.MsvAvChannelBindings])
		}
		if wantName := utf16.EncodeUTF16LE("cifs/dc01.tmp.local"); !bytes.Equal(values[avpair.MsvAvTargetName], wantName) {
			t.Errorf("micPresent=%v: MsvAvTargetName = % x, want the client's SPN % x", micPresent, values[avpair.MsvAvTargetName], wantName)
		}
	}
}
