package targetinfo_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/avpair"
	"github.com/TheManticoreProject/Manticore/crypto/spnego/ntlm/targetinfo"
	"github.com/TheManticoreProject/Manticore/encoding/utf16"
)

// tlsServerEndPointToken is an RFC 5929 tls-server-end-point token over a fixed
// 32-byte certificate hash.
func tlsServerEndPointToken() []byte {
	hash := make([]byte, 32)
	for i := range hash {
		hash[i] = byte(i)
	}
	return append([]byte("tls-server-end-point:"), hash...)
}

// TestChannelBindingsHashKnownAnswer pins MsvAvChannelBindings to the MD5 of the
// RFC 2744 gss_channel_bindings_struct with empty addresses: four zero uint32s, the
// little-endian application-data length, then the data ([MS-NLMP] 2.2.2.1).
func TestChannelBindingsHashKnownAnswer(t *testing.T) {
	want, _ := hex.DecodeString("8f1214c9c9cab8dc3bf866da9aba57a7")
	if got := targetinfo.ChannelBindingsHash(tlsServerEndPointToken()); !bytes.Equal(got, want) {
		t.Fatalf("ChannelBindingsHash = %x, want %x", got, want)
	}
	if got := targetinfo.ChannelBindingsHash(nil); !bytes.Equal(got, make([]byte, 16)) {
		t.Fatalf("ChannelBindingsHash(nil) = %x, want 16 zero bytes", got)
	}
}

// TestBlobOptionsAssertTargetNameAndChannelBindings checks the caller's SPN replaces
// the default cifs/ one and the caller's bindings replace the all-zero value, and
// that the zero options still produce the defaults.
func TestBlobOptionsAssertTargetNameAndChannelBindings(t *testing.T) {
	serverInfo, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "dc01.tmp.local", "tmp.local", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}

	cases := []struct {
		name     string
		opts     targetinfo.BlobOptions
		wantName string
		wantCB   []byte
	}{
		{"defaults", targetinfo.BlobOptions{}, "cifs/dc01.tmp.local", make([]byte, 16)},
		{"ldap over TLS", targetinfo.BlobOptions{TargetName: "ldap/dc01.tmp.local", ChannelBindings: tlsServerEndPointToken()},
			"ldap/dc01.tmp.local", targetinfo.ChannelBindingsHash(tlsServerEndPointToken())},
		{"rpc", targetinfo.BlobOptions{TargetName: "host/dc01.tmp.local"}, "host/dc01.tmp.local", make([]byte, 16)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			blob := targetinfo.BuildBlobTargetInfoWithOptions(serverInfo, true, tc.opts)
			counts := countAvPairs(t, blob)
			for _, id := range []avpair.AvId{avpair.MsvAvTargetName, avpair.MsvAvChannelBindings} {
				if counts[id] != 1 {
					t.Errorf("blob carries %d %s AV_PAIRs, want 1", counts[id], id)
				}
			}
			_, values := avPairOffsets(t, blob)
			if got := values[avpair.MsvAvTargetName]; !bytes.Equal(got, utf16.EncodeUTF16LE(tc.wantName)) {
				t.Errorf("MsvAvTargetName = %q, want %q", utf16.DecodeUTF16LE(got), tc.wantName)
			}
			if got := values[avpair.MsvAvChannelBindings]; !bytes.Equal(got, tc.wantCB) {
				t.Errorf("MsvAvChannelBindings = %x, want %x", got, tc.wantCB)
			}
		})
	}

	if !bytes.Equal(targetinfo.BuildBlobTargetInfo(serverInfo, true),
		targetinfo.BuildBlobTargetInfoWithOptions(serverInfo, true, targetinfo.BlobOptions{})) {
		t.Error("BuildBlobTargetInfo differs from BuildBlobTargetInfoWithOptions with zero options")
	}
}

// TestBlobOptionsTargetNameWithoutDnsComputerName checks a caller-supplied SPN is
// sent even when the challenge has no DNS computer name to build the default from.
func TestBlobOptionsTargetNameWithoutDnsComputerName(t *testing.T) {
	serverInfo, err := targetinfo.BuildServerTargetInfo("DC01", "TMP", "", "", nil)
	if err != nil {
		t.Fatalf("BuildServerTargetInfo: %v", err)
	}
	_, values := avPairOffsets(t, targetinfo.BuildBlobTargetInfo(serverInfo, false))
	if _, ok := values[avpair.MsvAvTargetName]; ok {
		t.Fatal("default blob carries an MsvAvTargetName with no DNS computer name to build it from")
	}
	_, values = avPairOffsets(t, targetinfo.BuildBlobTargetInfoWithOptions(serverInfo, false, targetinfo.BlobOptions{TargetName: "host/dc01"}))
	if got := utf16.DecodeUTF16LE(values[avpair.MsvAvTargetName]); got != "host/dc01" {
		t.Fatalf("MsvAvTargetName = %q, want host/dc01", got)
	}
}
