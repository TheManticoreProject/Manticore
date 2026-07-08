package msdnsp_test

import (
	"bytes"
	"testing"

	msdnsp "github.com/TheManticoreProject/Manticore/windows/protocols/ms-dnsp"
)

// buildProperty assembles a raw dnsProperty value: a 20-byte little-endian header, the Data
// field, and the trailing 1-byte Name field.
func buildProperty(id msdnsp.PropertyId, data []byte) []byte {
	b := make([]byte, 20)
	// DataLength
	b[0] = byte(len(data))
	b[1] = byte(len(data) >> 8)
	b[2] = byte(len(data) >> 16)
	b[3] = byte(len(data) >> 24)
	// NameLength = 1
	b[4] = 0x01
	// Flag = 0 (b[8:12])
	// Version = 1
	b[12] = 0x01
	// Id
	b[16] = byte(id)
	b[17] = byte(id >> 8)
	b[18] = byte(id >> 16)
	b[19] = byte(id >> 24)
	b = append(b, data...)
	b = append(b, 0x00) // Name
	return b
}

// TestAllowUpdateSingleByte decodes a DSPROPERTY_ZONE_ALLOW_UPDATE property stored in a single
// byte, as a Windows DNS server actually writes it (DataLength=1, Data=0x02), rather than the
// full DWORD described by [MS-DNSP].
func TestAllowUpdateSingleByte(t *testing.T) {
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_ALLOW_UPDATE, []byte{0x02})

	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if p.DataLength != 1 {
		t.Fatalf("DataLength = %d; want 1", p.DataLength)
	}
	v, err := p.AsUint32()
	if err != nil {
		t.Fatalf("AsUint32 on 1-byte data failed: %v", err)
	}
	if v != 2 {
		t.Errorf("AsUint32 = %d; want 2", v)
	}
	zu, err := p.AsZoneUpdate()
	if err != nil {
		t.Fatalf("AsZoneUpdate on 1-byte data failed: %v", err)
	}
	if zu != msdnsp.ZONE_UPDATE_SECURE {
		t.Errorf("ZoneUpdate = %s; want ZONE_UPDATE_SECURE", zu)
	}
}

// TestAsUint32EmptyData verifies AsUint32 rejects an empty Data field.
func TestAsUint32EmptyData(t *testing.T) {
	p := &msdnsp.DNS_PROPERTY{Id: msdnsp.DSPROPERTY_ZONE_TYPE, Data: nil}
	if _, err := p.AsUint32(); err == nil {
		t.Error("AsUint32 on empty Data = nil error; want error")
	}
}

// TestZoneTypeProperty decodes a DSPROPERTY_ZONE_TYPE property.
func TestZoneTypeProperty(t *testing.T) {
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_TYPE, []byte{0x01, 0x00, 0x00, 0x00})

	p := &msdnsp.DNS_PROPERTY{}
	read, err := p.Unmarshal(raw)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(raw) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(raw))
	}
	if p.Id != msdnsp.DSPROPERTY_ZONE_TYPE {
		t.Errorf("Id = %s; want DSPROPERTY_ZONE_TYPE", p.Id)
	}
	zt, err := p.AsZoneType()
	if err != nil {
		t.Fatalf("AsZoneType failed: %v", err)
	}
	if zt != msdnsp.DNS_ZONE_TYPE_PRIMARY {
		t.Errorf("ZoneType = %s; want DNS_ZONE_TYPE_PRIMARY", zt)
	}
}

// TestAllowUpdateProperty decodes a DSPROPERTY_ZONE_ALLOW_UPDATE property carrying the
// insecure-update policy.
func TestAllowUpdateProperty(t *testing.T) {
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_ALLOW_UPDATE, []byte{0x01, 0x00, 0x00, 0x00})

	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	zu, err := p.AsZoneUpdate()
	if err != nil {
		t.Fatalf("AsZoneUpdate failed: %v", err)
	}
	if zu != msdnsp.ZONE_UPDATE_UNSECURE {
		t.Errorf("ZoneUpdate = %s; want ZONE_UPDATE_UNSECURE", zu)
	}
}

// TestIP4ArrayProperty decodes a scavenging-servers property carrying two IPv4 addresses.
func TestIP4ArrayProperty(t *testing.T) {
	data := []byte{
		0x02, 0x00, 0x00, 0x00, // AddrCount = 2
		192, 0, 2, 1, // 192.0.2.1 (network byte order)
		10, 20, 30, 40, // 10.20.30.40
	}
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_SCAVENGING_SERVERS, data)

	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	addrs, err := p.AsIP4Array()
	if err != nil {
		t.Fatalf("AsIP4Array failed: %v", err)
	}
	if len(addrs) != 2 {
		t.Fatalf("AsIP4Array returned %d addresses; want 2", len(addrs))
	}
	if addrs[0].String() != "192.0.2.1" || addrs[1].String() != "10.20.30.40" {
		t.Errorf("addresses = %v; want [192.0.2.1 10.20.30.40]", addrs)
	}
}

// TestEmptyIP4Array decodes an empty IP4_ARRAY (AddrCount 0).
func TestEmptyIP4Array(t *testing.T) {
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_MASTER_SERVERS, []byte{0x00, 0x00, 0x00, 0x00})
	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	addrs, err := p.AsIP4Array()
	if err != nil {
		t.Fatalf("AsIP4Array failed: %v", err)
	}
	if len(addrs) != 0 {
		t.Errorf("AsIP4Array returned %d addresses; want 0", len(addrs))
	}
}

// TestDeletedFromHostnameProperty decodes the null-terminated UTF-16 hostname property.
func TestDeletedFromHostnameProperty(t *testing.T) {
	// "DC1" in UTF-16LE with a NUL terminator.
	data := []byte{'D', 0, 'C', 0, '1', 0, 0, 0}
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_DELETED_FROM_HOSTNAME, data)

	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(raw); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if got := p.AsUTF16String(); got != "DC1" {
		t.Errorf("AsUTF16String = %q; want %q", got, "DC1")
	}
}

// TestRoundTrip marshals a property built with the constructor and checks the bytes round-trip.
func TestRoundTrip(t *testing.T) {
	p := msdnsp.NewDNS_PROPERTY()
	p.Id = msdnsp.DSPROPERTY_ZONE_REFRESH_INTERVAL
	p.Data = []byte{0xA8, 0x00, 0x00, 0x00} // 168 hours

	marshalled, err := p.Marshal()
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	out := &msdnsp.DNS_PROPERTY{}
	read, err := out.Unmarshal(marshalled)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if read != len(marshalled) {
		t.Errorf("Unmarshal read %d bytes; want %d", read, len(marshalled))
	}
	if out.DataLength != 4 || out.Version != msdnsp.DNSPropertyVersion {
		t.Errorf("header mismatch: DataLength=%d Version=%d", out.DataLength, out.Version)
	}
	v, err := out.AsUint32()
	if err != nil {
		t.Fatalf("AsUint32 failed: %v", err)
	}
	if v != 168 {
		t.Errorf("AsUint32 = %d; want 168", v)
	}

	remarshalled, err := out.Marshal()
	if err != nil {
		t.Fatalf("re-Marshal failed: %v", err)
	}
	if !bytes.Equal(marshalled, remarshalled) {
		t.Errorf("round-trip mismatch:\n got % x\nwant % x", remarshalled, marshalled)
	}
}

// TestUnmarshalTruncated verifies that a value too short for the declared Data plus Name field
// is rejected.
func TestUnmarshalTruncated(t *testing.T) {
	raw := buildProperty(msdnsp.DSPROPERTY_ZONE_TYPE, []byte{0x01, 0x00, 0x00, 0x00})
	// Drop the trailing Name byte so the buffer is one byte short.
	truncated := raw[:len(raw)-1]

	p := &msdnsp.DNS_PROPERTY{}
	if _, err := p.Unmarshal(truncated); err == nil {
		t.Errorf("expected error unmarshalling truncated property, got nil")
	}
}

// TestStringFallback verifies String renders a byte-count fallback for an undecoded property.
func TestStringFallback(t *testing.T) {
	p := &msdnsp.DNS_PROPERTY{
		Id:   msdnsp.DSPROPERTY_ZONE_SECURE_TIME,
		Data: []byte{1, 2, 3, 4, 5, 6, 7, 8},
	}
	if got := p.String(); got != "DSPROPERTY_ZONE_SECURE_TIME = <8 bytes>" {
		t.Errorf("String = %q; want fallback rendering", got)
	}
}
