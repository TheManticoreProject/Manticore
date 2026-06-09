package client

import "testing"

func TestSessionSetupNilCredentials(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	if err := c.SessionSetup(nil); err == nil {
		t.Errorf("expected SessionSetup(nil) to error, got nil")
	}
}

func TestFormatNTStatus(t *testing.T) {
	// A success code formats as a bare hex value.
	if got := formatNTStatus(0x00000000); got == "" {
		t.Errorf("formatNTStatus returned empty string")
	}
	// MORE_PROCESSING_REQUIRED should at least include its hex form.
	got := formatNTStatus(ntStatusMoreProcessingRequired)
	if got == "" {
		t.Errorf("formatNTStatus returned empty string for MORE_PROCESSING_REQUIRED")
	}
}
