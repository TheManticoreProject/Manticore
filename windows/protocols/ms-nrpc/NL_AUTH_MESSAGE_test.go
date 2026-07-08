package msnrpc

import (
	"encoding/hex"
	"testing"
)

// TestBuildClientNlAuthMessage pins both name forms byte-for-byte against impacket's
// getSSPType1 output.
func TestBuildClientNlAuthMessage(t *testing.T) {
	tests := []struct {
		name     string
		computer string
		domain   string
		want     string
	}{
		{
			name:     "DNS domain",
			computer: "MANTICORE1",
			domain:   "TMP-W-2016.local",
			want:     "00000000160000004d414e5449434f524531000a544d502d572d32303136056c6f63616c000a4d414e5449434f52453100",
		},
		{
			name:     "NetBIOS domain",
			computer: "MANTICORE1",
			domain:   "TMP-W-2016",
			want:     "0000000013000000544d502d572d32303136004d414e5449434f524531000a4d414e5449434f52453100",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := hex.EncodeToString(BuildClientNlAuthMessage(tc.computer, tc.domain).Marshal())
			if got != tc.want {
				t.Errorf("NL_AUTH_MESSAGE\n got  %s\n want %s", got, tc.want)
			}
		})
	}
}

// TestNLAuthMessageRoundTrip checks Unmarshal recovers the header of a marshalled token.
func TestNLAuthMessageRoundTrip(t *testing.T) {
	orig := BuildClientNlAuthMessage("HOST", "example.com")
	var got NL_AUTH_MESSAGE
	if err := got.Unmarshal(orig.Marshal()); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.MessageType != NlAuthMessageTypeRequest {
		t.Errorf("MessageType = %d, want %d", got.MessageType, NlAuthMessageTypeRequest)
	}
	if got.Flags != orig.Flags {
		t.Errorf("Flags = %#x, want %#x", got.Flags, orig.Flags)
	}
}
