package commands_test

import (
	"strings"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/message/commands"
)

// Test_NegotiateResponse_Unmarshal_NoDialectSupported decodes the reply of a
// server that supports none of the offered dialects: WordCount 0x01 carrying
// only DialectIndex 0xFFFF, and ByteCount 0 ([MS-CIFS] 2.2.4.52.2). Samba with
// SMB1 disabled sends exactly this.
func Test_NegotiateResponse_Unmarshal_NoDialectSupported(t *testing.T) {
	raw := []byte{0x01, 0xFF, 0xFF, 0x00, 0x00}

	resp := commands.NewNegotiateResponse()
	if _, err := resp.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if resp.DialectIndex != commands.NoDialectSupported {
		t.Fatalf("DialectIndex = %#04x, want 0xFFFF", uint16(resp.DialectIndex))
	}

	_, err := resp.GetSelectedDialect(dialects.Dialects{Dialects: []string{"NT LM 0.12"}})
	if err == nil || !strings.Contains(err.Error(), "none of the offered dialects") {
		t.Errorf("GetSelectedDialect error = %v, want 'none of the offered dialects'", err)
	}
}
