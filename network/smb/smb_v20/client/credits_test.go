package client

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/dialects"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v20/message/commands"
)

// TestCreditsFor pins the charge formula from MS-SMB2 3.1.5.2:
// (payload - 1) / 65536 + 1, with an empty payload still costing one credit.
func TestCreditsFor(t *testing.T) {
	tests := []struct {
		payload uint32
		want    uint16
	}{
		{0, 1},
		{1, 1},
		{65535, 1},
		{65536, 1},
		{65537, 2},
		{131072, 2},
		{131073, 3},
		{1 << 20, 16},          // 1 MiB
		{8 * 1024 * 1024, 128}, // the 8 MiB a Windows server advertises
	}
	for _, tt := range tests {
		if got := creditsFor(tt.payload); got != tt.want {
			t.Errorf("creditsFor(%d) = %d, want %d", tt.payload, got, tt.want)
		}
	}
}

// TestRequestCreditCharge covers each command whose charge depends on its payload
// — the charge must reflect the larger of what is sent and what is expected back.
func TestRequestCreditCharge(t *testing.T) {
	read := commands.NewReadRequest()
	read.Length = 1 << 20 // expects 1 MiB back
	if got := requestCreditCharge(read); got != 16 {
		t.Errorf("READ of 1 MiB charges %d, want 16", got)
	}

	write := commands.NewWriteRequest()
	write.Data = make([]byte, 200000)
	if got := requestCreditCharge(write); got != 4 {
		t.Errorf("WRITE of 200000 bytes charges %d, want 4", got)
	}

	qd := commands.NewQueryDirectoryRequest()
	qd.OutputBufferLength = 8 * 1024 * 1024
	if got := requestCreditCharge(qd); got != 128 {
		t.Errorf("QUERY_DIRECTORY expecting 8 MiB charges %d, want 128", got)
	}

	qi := commands.NewQueryInfoRequest()
	qi.OutputBufferLength = 65536
	if got := requestCreditCharge(qi); got != 1 {
		t.Errorf("QUERY_INFO expecting 64 KiB charges %d, want 1", got)
	}

	ioctl := commands.NewIoctlRequest()
	ioctl.MaxOutputResponse = 131072
	if got := requestCreditCharge(ioctl); got != 2 {
		t.Errorf("IOCTL expecting 128 KiB charges %d, want 2", got)
	}

	// A command with no variable payload costs a single credit.
	if got := requestCreditCharge(commands.NewCloseRequest()); got != 1 {
		t.Errorf("CLOSE charges %d, want 1", got)
	}
}

// TestNewRequestChargesAndConsumesSequenceNumbers is the regression guard for the
// defect: a multi-credit request must declare its charge and must consume one
// sequence number per credit charged (MS-SMB2 3.2.4.1.3). Advancing MessageId by
// one reuses identifiers the server has retired, and it drops the connection.
func TestNewRequestChargesAndConsumesSequenceNumbers(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1
	c.Connection.MessageId = 4
	c.Connection.Credits = 200

	read := commands.NewReadRequest()
	read.Length = 1 << 20 // 16 credits

	m := c.newRequest(read)
	if m.Header.CreditCharge != 16 {
		t.Errorf("CreditCharge = %d, want 16", m.Header.CreditCharge)
	}
	if m.Header.MessageId != 4 {
		t.Errorf("MessageId = %d, want 4", m.Header.MessageId)
	}
	if c.Connection.MessageId != 20 {
		t.Errorf("next MessageId = %d, want 20 (4 + the 16 charged)", c.Connection.MessageId)
	}
	if c.Connection.Credits != 184 {
		t.Errorf("credits after spending = %d, want 184", c.Connection.Credits)
	}
}

// TestNewRequestSMB202HasNoCreditModel checks the SMB 2.0.2 exception: CreditCharge
// MUST be 0, one sequence number is consumed, and no credit window is requested.
func TestNewRequestSMB202HasNoCreditModel(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Connection.Dialect = dialects.SMB2_DIALECT_2_0_2

	read := commands.NewReadRequest()
	read.Length = 1 << 20

	m := c.newRequest(read)
	if m.Header.CreditCharge != 0 {
		t.Errorf("CreditCharge = %d, want 0 for SMB 2.0.2", m.Header.CreditCharge)
	}
	if m.Header.Credit != 1 {
		t.Errorf("CreditRequest = %d, want 1 for SMB 2.0.2", m.Header.Credit)
	}
	if c.Connection.MessageId != 1 {
		t.Errorf("next MessageId = %d, want 1", c.Connection.MessageId)
	}
}

// TestCreditRequestGrowsTheWindow checks that the client asks for credits while its
// window is below target — without asking, the single credit granted alongside
// NEGOTIATE is all it would ever hold.
func TestCreditRequestGrowsTheWindow(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1

	c.Connection.Credits = 1
	if got := c.creditRequest(1); got != targetCredits-1 {
		t.Errorf("creditRequest with 1 credit held = %d, want %d", got, targetCredits-1)
	}

	// At target, ask only to replace what is spent.
	c.Connection.Credits = targetCredits
	if got := c.creditRequest(4); got != 4 {
		t.Errorf("creditRequest at target = %d, want 4", got)
	}
}

// TestMaxPayloadForRequest checks that a request is sized by the credits actually
// held, not by the server's advertised ceiling.
func TestMaxPayloadForRequest(t *testing.T) {
	c := newTestClient(&fakeTransport{})
	c.Connection.Dialect = dialects.SMB2_DIALECT_3_1_1

	// One credit held against an 8 MiB ceiling: the credit window wins.
	c.Connection.Credits = 1
	if got := c.maxPayloadForRequest(8 * 1024 * 1024); got != creditGranularity {
		t.Errorf("with 1 credit = %d, want %d", got, creditGranularity)
	}

	// Plenty of credits: the server's ceiling wins.
	c.Connection.Credits = 512
	if got := c.maxPayloadForRequest(1 << 20); got != 1<<20 {
		t.Errorf("with 512 credits = %d, want %d", got, 1<<20)
	}

	// SMB 2.0.2 has no credit model: one credit's worth regardless.
	c.Connection.Dialect = dialects.SMB2_DIALECT_2_0_2
	c.Connection.Credits = 512
	if got := c.maxPayloadForRequest(8 * 1024 * 1024); got != creditGranularity {
		t.Errorf("SMB 2.0.2 = %d, want %d", got, creditGranularity)
	}
}

// TestGrantCreditsAccumulates checks that a response's Credit field adds to the
// window rather than replacing it: the field is what that response grants, not the
// size of the window.
func TestGrantCreditsAccumulates(t *testing.T) {
	c := newTestClient(&fakeTransport{})

	c.Connection.Credits = 0
	c.grantCredits(31)
	c.grantCredits(31)
	if c.Connection.Credits != 62 {
		t.Errorf("credits = %d, want 62", c.Connection.Credits)
	}

	// Spending more than is held floors at zero rather than wrapping.
	c.spendCredits(100)
	if c.Connection.Credits != 0 {
		t.Errorf("credits after overspend = %d, want 0", c.Connection.Credits)
	}

	// Granting cannot overflow the 16-bit field.
	c.Connection.Credits = 0xFFF0
	c.grantCredits(0xFF)
	if c.Connection.Credits != 0xFFFF {
		t.Errorf("credits = %#x, want 0xFFFF", c.Connection.Credits)
	}
}
