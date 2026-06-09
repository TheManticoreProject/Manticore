package dialects_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
)

func TestDialectString(t *testing.T) {
	if got := dialects.SMB2_DIALECT_2_0_2.String(); got != "SMB 2.0.2" {
		t.Errorf("SMB2_DIALECT_2_0_2.String() = %q, want %q", got, "SMB 2.0.2")
	}
	if got := dialects.SMB2_DIALECT_3_1_1.String(); got != "SMB 3.1.1" {
		t.Errorf("SMB2_DIALECT_3_1_1.String() = %q, want %q", got, "SMB 3.1.1")
	}
	if got := dialects.Dialect(0x1234).String(); got != "Dialect(0x1234)" {
		t.Errorf("unknown dialect String() = %q, want %q", got, "Dialect(0x1234)")
	}
}
