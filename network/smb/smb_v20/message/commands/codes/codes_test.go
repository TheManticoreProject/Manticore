package codes_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands/codes"
)

func TestCommandCode_String(t *testing.T) {
	tests := []struct {
		code codes.CommandCode
		want string
	}{
		{codes.SMB2_NEGOTIATE, "NEGOTIATE"},
		{codes.SMB2_SESSION_SETUP, "SESSION_SETUP"},
		{codes.SMB2_OPLOCK_BREAK, "OPLOCK_BREAK"},
		{codes.CommandCode(0x1234), "CommandCode(0x1234)"},
	}

	for _, tt := range tests {
		if got := tt.code.String(); got != tt.want {
			t.Errorf("CommandCode(0x%04x).String() = %q, want %q", uint16(tt.code), got, tt.want)
		}
	}
}

func TestCommandCode_IsValid(t *testing.T) {
	// All 19 defined commands span the contiguous range 0x0000..0x0012.
	for c := codes.CommandCode(0x0000); c <= codes.SMB2_OPLOCK_BREAK; c++ {
		if !c.IsValid() {
			t.Errorf("expected command code 0x%04x to be valid", uint16(c))
		}
	}

	if codes.CommandCode(0x0013).IsValid() {
		t.Errorf("expected command code 0x0013 to be invalid")
	}

	if len(codes.CommandCodeNames) != 19 {
		t.Errorf("expected 19 defined SMB2 commands, got %d", len(codes.CommandCodeNames))
	}
}
