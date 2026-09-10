package targetinfo_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
)

// avPairOffsets walks a TargetInfo and records the byte offset of each AvId, so a
// test can assert ordering as well as presence.
func avPairOffsets(t *testing.T, info []byte) (map[avpair.AvId]int, map[avpair.AvId][]byte) {
	t.Helper()

	offsets := make(map[avpair.AvId]int)
	values := make(map[avpair.AvId][]byte)
	for i := 0; i+4 <= len(info); {
		id := avpair.AvId(binary.LittleEndian.Uint16(info[i : i+2]))
		avLen := int(binary.LittleEndian.Uint16(info[i+2 : i+4]))
		if i+4+avLen > len(info) {
			t.Fatalf("AV_PAIR %s at offset %d declares %d bytes, past the end of the buffer", id, i, avLen)
		}
		offsets[id] = i
		values[id] = info[i+4 : i+4+avLen]
		if id == avpair.MsvAvEOL {
			break
		}
		i += 4 + avLen
	}
	return offsets, values
}

// TestBlobCarriesChannelBindings guards the defect: MsvAvChannelBindings was defined
// but had no writer, and a server enforcing channel binding returns
// GSS_S_BAD_BINDINGS when the pair is absent (MS-NLMP 3.2.5.1.2).
func TestBlobCarriesChannelBindings(t *testing.T) {
	serverInfo, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}

	blob := targetinfo.BuildBlobTargetInfo(serverInfo, false)
	offsets, values := avPairOffsets(t, blob)

	cb, ok := values[avpair.MsvAvChannelBindings]
	if !ok {
		t.Fatal("blob TargetInfo carries no MsvAvChannelBindings AV_PAIR")
	}
	if len(cb) != 16 {
		t.Errorf("MsvAvChannelBindings value is %d bytes, want 16", len(cb))
	}
	if !bytes.Equal(cb, make([]byte, 16)) {
		t.Errorf("MsvAvChannelBindings = % x, want 16 zero bytes when no bindings were supplied", cb)
	}

	// The EOL marker must still terminate the list, after the inserted pairs.
	eol, ok := offsets[avpair.MsvAvEOL]
	if !ok {
		t.Fatal("blob TargetInfo has no MsvAvEOL terminator")
	}
	if offsets[avpair.MsvAvChannelBindings] >= eol {
		t.Error("MsvAvChannelBindings appears at or after the EOL marker")
	}

	// The server's own pairs must survive unchanged alongside the insertions.
	for _, id := range []avpair.AvId{
		avpair.MsvAvNbDomainName,
		avpair.MsvAvNbComputerName,
		avpair.MsvAvDnsDomainName,
		avpair.MsvAvDnsComputerName,
	} {
		if _, ok := values[id]; !ok {
			t.Errorf("server AV_PAIR %s was dropped from the blob", id)
		}
	}
}

// TestBlobChannelBindingsPrecedeTargetName pins the ordering a Windows client uses:
// the channel-bindings pair comes before the SPN pair.
func TestBlobChannelBindingsPrecedeTargetName(t *testing.T) {
	serverInfo, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}

	offsets, _ := avPairOffsets(t, targetinfo.BuildBlobTargetInfo(serverInfo, false))

	spn, ok := offsets[avpair.MsvAvTargetName]
	if !ok {
		t.Skip("no MsvAvTargetName inserted, so there is no ordering to check")
	}
	if offsets[avpair.MsvAvChannelBindings] >= spn {
		t.Errorf("MsvAvChannelBindings at %d, want before MsvAvTargetName at %d",
			offsets[avpair.MsvAvChannelBindings], spn)
	}
}
