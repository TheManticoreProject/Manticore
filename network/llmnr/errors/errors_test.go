package errors_test

import (
	stderrors "errors"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/llmnr/errors"
)

// TestSentinelMessages asserts each sentinel error carries its documented message.
func TestSentinelMessages(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{errors.ErrInvalidDomainName, "invalid domain name"},
		{errors.ErrInvalidMessage, "invalid message format"},
		{errors.ErrInvalidHeader, "invalid header format"},
		{errors.ErrInvalidQuestion, "invalid question format"},
		{errors.ErrInvalidAnswer, "invalid answer format"},
		{errors.ErrInvalidAuthority, "invalid authority format"},
		{errors.ErrInvalidAdditional, "invalid additional format"},
		{errors.ErrInvalidResourceRecord, "invalid resource record format"},
		{errors.ErrInvalidResourceRecordType, "invalid resource record type"},
		{errors.ErrNameTooLong, "domain name too long"},
		{errors.ErrLabelTooLong, "label too long"},
	}

	for _, c := range cases {
		if c.err == nil {
			t.Errorf("sentinel error for %q is nil", c.want)
			continue
		}
		if c.err.Error() != c.want {
			t.Errorf("error message = %q, want %q", c.err.Error(), c.want)
		}
	}
}

// TestSentinelIdentity confirms the sentinels are distinct comparable values and
// are matchable through errors.Is even after wrapping with %w.
func TestSentinelIdentity(t *testing.T) {
	sentinels := []error{
		errors.ErrInvalidDomainName,
		errors.ErrInvalidMessage,
		errors.ErrInvalidHeader,
		errors.ErrInvalidQuestion,
		errors.ErrInvalidAnswer,
		errors.ErrInvalidAuthority,
		errors.ErrInvalidAdditional,
		errors.ErrInvalidResourceRecord,
		errors.ErrInvalidResourceRecordType,
		errors.ErrNameTooLong,
		errors.ErrLabelTooLong,
	}

	// No two sentinels should share identity (each is a unique errors.New value).
	for i := range sentinels {
		for j := i + 1; j < len(sentinels); j++ {
			if sentinels[i] == sentinels[j] {
				t.Errorf("sentinels %d and %d share identity: %v", i, j, sentinels[i])
			}
		}
	}

	// A wrapped sentinel must remain matchable by errors.Is.
	wrapped := stderrors.Join(errors.ErrNameTooLong, stderrors.New("context"))
	if !stderrors.Is(wrapped, errors.ErrNameTooLong) {
		t.Error("errors.Is failed to match a wrapped ErrNameTooLong sentinel")
	}
	if stderrors.Is(wrapped, errors.ErrLabelTooLong) {
		t.Error("errors.Is matched the wrong sentinel")
	}
}
