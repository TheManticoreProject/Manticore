package commands

import (
	"encoding/binary"
	"testing"
)

func TestNewNetnameContext(t *testing.T) {
	ctx := NewNetnameContext("DC01")
	if ctx.ContextType != SMB2_NETNAME_NEGOTIATE_CONTEXT_ID {
		t.Errorf("ContextType = 0x%04x, want 0x%04x", ctx.ContextType, SMB2_NETNAME_NEGOTIATE_CONTEXT_ID)
	}
	// "DC01" → 4 UTF-16LE code units + null terminator = 10 bytes
	if len(ctx.Data) != 10 {
		t.Fatalf("Data length = %d, want 10", len(ctx.Data))
	}
	want := []byte{'D', 0, 'C', 0, '0', 0, '1', 0, 0, 0}
	for i, b := range ctx.Data {
		if b != want[i] {
			t.Errorf("Data[%d] = 0x%02x, want 0x%02x", i, b, want[i])
		}
	}
}

func TestNewNetnameContextEmpty(t *testing.T) {
	ctx := NewNetnameContext("")
	if ctx.ContextType != SMB2_NETNAME_NEGOTIATE_CONTEXT_ID {
		t.Errorf("ContextType = 0x%04x, want 0x%04x", ctx.ContextType, SMB2_NETNAME_NEGOTIATE_CONTEXT_ID)
	}
	// Empty name still carries the UTF-16 null terminator.
	if len(ctx.Data) != 2 {
		t.Errorf("Data length = %d, want 2 (null terminator only)", len(ctx.Data))
	}
}

func TestNewNetnameContextIPv4(t *testing.T) {
	ctx := NewNetnameContext("10.0.0.1")
	// "10.0.0.1" → 8 UTF-16LE code units + null terminator = 18 bytes
	if len(ctx.Data) != 18 {
		t.Errorf("Data length = %d, want 18", len(ctx.Data))
	}
}

func TestNewTransportCapabilitiesContext(t *testing.T) {
	ctx := NewTransportCapabilitiesContext(SMB2_ACCEPT_TRANSPORT_LEVEL_SECURITY)
	if ctx.ContextType != SMB2_TRANSPORT_CAPABILITIES {
		t.Errorf("ContextType = 0x%04x, want 0x%04x", ctx.ContextType, SMB2_TRANSPORT_CAPABILITIES)
	}
	if len(ctx.Data) != 4 {
		t.Fatalf("Data length = %d, want 4", len(ctx.Data))
	}
	got := binary.LittleEndian.Uint32(ctx.Data)
	if got != SMB2_ACCEPT_TRANSPORT_LEVEL_SECURITY {
		t.Errorf("flags = 0x%08x, want 0x%08x", got, SMB2_ACCEPT_TRANSPORT_LEVEL_SECURITY)
	}
}

func TestNewTransportCapabilitiesContextZero(t *testing.T) {
	ctx := NewTransportCapabilitiesContext(0)
	if binary.LittleEndian.Uint32(ctx.Data) != 0 {
		t.Error("expected zero flags")
	}
}

func TestNewSigningCapabilitiesContext(t *testing.T) {
	ctx := NewSigningCapabilitiesContext([]uint16{SMB2_SIGNING_ALG_AES_GMAC, SMB2_SIGNING_ALG_AES_CMAC})
	if ctx.ContextType != SMB2_SIGNING_CAPABILITIES {
		t.Errorf("ContextType = 0x%04x, want 0x%04x", ctx.ContextType, SMB2_SIGNING_CAPABILITIES)
	}
	// 2-byte count + 2 algorithms × 2 bytes = 6 bytes
	if len(ctx.Data) != 6 {
		t.Fatalf("Data length = %d, want 6", len(ctx.Data))
	}
	count := binary.LittleEndian.Uint16(ctx.Data[0:2])
	if count != 2 {
		t.Errorf("algorithm count = %d, want 2", count)
	}
	alg0 := binary.LittleEndian.Uint16(ctx.Data[2:4])
	alg1 := binary.LittleEndian.Uint16(ctx.Data[4:6])
	if alg0 != SMB2_SIGNING_ALG_AES_GMAC {
		t.Errorf("algorithm[0] = 0x%04x, want AES-GMAC", alg0)
	}
	if alg1 != SMB2_SIGNING_ALG_AES_CMAC {
		t.Errorf("algorithm[1] = 0x%04x, want AES-CMAC", alg1)
	}
}

func TestSelectedSigningAlgorithm(t *testing.T) {
	// Server responds with a single algorithm
	ctx := NewSigningCapabilitiesContext([]uint16{SMB2_SIGNING_ALG_AES_GMAC})
	got := SelectedSigningAlgorithm([]*NegotiateContext{ctx})
	if got != SMB2_SIGNING_ALG_AES_GMAC {
		t.Errorf("SelectedSigningAlgorithm = %d, want %d", got, SMB2_SIGNING_ALG_AES_GMAC)
	}

	// No signing capabilities context → -1
	got = SelectedSigningAlgorithm([]*NegotiateContext{})
	if got != -1 {
		t.Errorf("SelectedSigningAlgorithm on empty = %d, want -1", got)
	}
}

func TestNetnameContextRoundTrip(t *testing.T) {
	original := []*NegotiateContext{
		NewPreauthIntegrityContext(make([]byte, 32)),
		NewEncryptionContext([]uint16{0x0002, 0x0001}),
		NewNetnameContext("SERVER01"),
		NewTransportCapabilitiesContext(SMB2_ACCEPT_TRANSPORT_LEVEL_SECURITY),
	}

	bodyOffset := 36 + 10 // 36-byte fixed negotiate body + 5 dialects × 2 bytes
	ctxBytes, first := marshalNegotiateContexts(bodyOffset, original)

	// Build the full body: fixed fields (zeroed) + dialect slots + context bytes.
	body := make([]byte, bodyOffset)
	body = append(body, ctxBytes...)

	parsed, err := parseNegotiateContexts(body, first, len(original))
	if err != nil {
		t.Fatalf("parseNegotiateContexts: %v", err)
	}
	if len(parsed) != len(original) {
		t.Fatalf("parsed %d contexts, want %d", len(parsed), len(original))
	}
	for i, ctx := range parsed {
		if ctx.ContextType != original[i].ContextType {
			t.Errorf("context %d: type 0x%04x, want 0x%04x", i, ctx.ContextType, original[i].ContextType)
		}
		if len(ctx.Data) != len(original[i].Data) {
			t.Errorf("context %d: data length %d, want %d", i, len(ctx.Data), len(original[i].Data))
		}
	}
}
