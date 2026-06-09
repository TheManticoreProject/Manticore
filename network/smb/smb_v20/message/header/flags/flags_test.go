package flags_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/header/flags"
)

func TestFlagsPredicates(t *testing.T) {
	f := flags.SMB2_FLAGS_SERVER_TO_REDIR | flags.SMB2_FLAGS_SIGNED

	if !f.IsServerToRedir() {
		t.Errorf("IsServerToRedir() = false, want true")
	}
	if !f.IsSigned() {
		t.Errorf("IsSigned() = false, want true")
	}
	if f.IsAsync() {
		t.Errorf("IsAsync() = true, want false")
	}
	if f.IsRelatedOperations() || f.IsDfsOperations() || f.IsReplayOperation() {
		t.Errorf("unset predicates returned true for flags %s", f)
	}
}

func TestFlagsPriority(t *testing.T) {
	// Priority occupies bits 4..6 (mask 0x70). A priority of 5 encodes as 5<<4 = 0x50.
	f := flags.Flags(0x50)
	if got := f.Priority(); got != 5 {
		t.Errorf("Priority() = %d, want 5", got)
	}
}

func TestFlagsString(t *testing.T) {
	if got := flags.Flags(0).String(); got != "NONE" {
		t.Errorf("Flags(0).String() = %q, want %q", got, "NONE")
	}

	f := flags.SMB2_FLAGS_SIGNED | flags.SMB2_FLAGS_ASYNC_COMMAND
	// Alphabetical order: ASYNC_COMMAND before SIGNED.
	if got := f.String(); got != "ASYNC_COMMAND|SIGNED" {
		t.Errorf("String() = %q, want %q", got, "ASYNC_COMMAND|SIGNED")
	}
}
