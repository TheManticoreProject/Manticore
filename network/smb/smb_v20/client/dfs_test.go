package client

import (
	"encoding/binary"
	"testing"
	"unicode/utf16"
)

func TestMarshalDfsReferralRequest(t *testing.T) {
	buf := marshalDfsReferralRequest(4, `\domain\share`)
	if len(buf) < 4 {
		t.Fatalf("buffer too short: %d", len(buf))
	}
	level := binary.LittleEndian.Uint16(buf[0:2])
	if level != 4 {
		t.Errorf("MaxReferralLevel = %d, want 4", level)
	}
	// Decode the UTF-16LE path (skip level, stop before null terminator).
	path := readUTF16NullTerm(buf[2:])
	if path != `\domain\share` {
		t.Errorf("path = %q, want %q", path, `\domain\share`)
	}
}

func writeUTF16LE(s string) []byte {
	runes := utf16.Encode([]rune(s))
	buf := make([]byte, len(runes)*2+2) // +2 for null terminator
	for i, r := range runes {
		binary.LittleEndian.PutUint16(buf[i*2:], r)
	}
	return buf
}

func buildDfsV3Referral(serverType, flags uint16, ttl uint32, dfsPath, altPath, netAddr string) []byte {
	pathBytes := writeUTF16LE(dfsPath)
	altBytes := writeUTF16LE(altPath)
	addrBytes := writeUTF16LE(netAddr)

	fixedSize := 34
	totalSize := fixedSize + len(pathBytes) + len(altBytes) + len(addrBytes)

	entry := make([]byte, totalSize)
	binary.LittleEndian.PutUint16(entry[0:2], 3)                   // Version
	binary.LittleEndian.PutUint16(entry[2:4], uint16(totalSize))    // Size
	binary.LittleEndian.PutUint16(entry[4:6], serverType)           // ServerType
	binary.LittleEndian.PutUint16(entry[6:8], flags)                // Flags
	binary.LittleEndian.PutUint32(entry[8:12], ttl)                 // TTL

	pathOff := fixedSize
	altOff := pathOff + len(pathBytes)
	addrOff := altOff + len(altBytes)

	binary.LittleEndian.PutUint16(entry[12:14], uint16(pathOff))    // DFSPathOffset
	binary.LittleEndian.PutUint16(entry[14:16], uint16(altOff))     // DFSAlternatePathOffset
	binary.LittleEndian.PutUint16(entry[16:18], uint16(addrOff))    // NetworkAddressOffset
	// entry[18:34] = ServiceSiteGuid (zeros)

	copy(entry[pathOff:], pathBytes)
	copy(entry[altOff:], altBytes)
	copy(entry[addrOff:], addrBytes)
	return entry
}

func buildDfsReferralResponse(pathConsumed uint16, headerFlags uint32, entries ...[]byte) []byte {
	header := make([]byte, 8)
	binary.LittleEndian.PutUint16(header[0:2], pathConsumed)
	binary.LittleEndian.PutUint16(header[2:4], uint16(len(entries)))
	binary.LittleEndian.PutUint32(header[4:8], headerFlags)

	buf := header
	for _, e := range entries {
		buf = append(buf, e...)
	}
	return buf
}

func TestParseDfsReferralResponseSingleV3(t *testing.T) {
	entry := buildDfsV3Referral(DFS_SERVER_ROOT, 0, 300,
		`\domain\share`, `\domain\share`, `\server1\share`)
	data := buildDfsReferralResponse(26, DFS_REFERRAL_HEADER_FLAG_REFERRAL_SERVERS|DFS_REFERRAL_HEADER_FLAG_STORAGE_SERVERS, entry)

	resp, err := parseDfsReferralResponse(data)
	if err != nil {
		t.Fatalf("parseDfsReferralResponse: %v", err)
	}
	if resp.PathConsumed != 26 {
		t.Errorf("PathConsumed = %d, want 26", resp.PathConsumed)
	}
	if resp.HeaderFlags&DFS_REFERRAL_HEADER_FLAG_REFERRAL_SERVERS == 0 {
		t.Error("expected REFERRAL_SERVERS flag")
	}
	if len(resp.ReferralEntries) != 1 {
		t.Fatalf("got %d entries, want 1", len(resp.ReferralEntries))
	}
	e := resp.ReferralEntries[0]
	if e.Version != 3 {
		t.Errorf("Version = %d, want 3", e.Version)
	}
	if !e.IsRootTarget() {
		t.Error("expected root target")
	}
	if e.TimeToLive != 300 {
		t.Errorf("TTL = %d, want 300", e.TimeToLive)
	}
	if e.DFSPath != `\domain\share` {
		t.Errorf("DFSPath = %q, want %q", e.DFSPath, `\domain\share`)
	}
	if e.NetworkAddress != `\server1\share` {
		t.Errorf("NetworkAddress = %q, want %q", e.NetworkAddress, `\server1\share`)
	}
}

func TestParseDfsReferralResponseMultiple(t *testing.T) {
	e1 := buildDfsV3Referral(DFS_SERVER_ROOT, 0, 600, `\domain\ns`, `\domain\ns`, `\dc1\ns`)
	e2 := buildDfsV3Referral(DFS_SERVER_ROOT, 0, 600, `\domain\ns`, `\domain\ns`, `\dc2\ns`)
	data := buildDfsReferralResponse(20, DFS_REFERRAL_HEADER_FLAG_STORAGE_SERVERS, e1, e2)

	resp, err := parseDfsReferralResponse(data)
	if err != nil {
		t.Fatalf("parseDfsReferralResponse: %v", err)
	}
	if len(resp.ReferralEntries) != 2 {
		t.Fatalf("got %d entries, want 2", len(resp.ReferralEntries))
	}
	if resp.ReferralEntries[0].NetworkAddress != `\dc1\ns` {
		t.Errorf("entry[0].NetworkAddress = %q, want %q", resp.ReferralEntries[0].NetworkAddress, `\dc1\ns`)
	}
	if resp.ReferralEntries[1].NetworkAddress != `\dc2\ns` {
		t.Errorf("entry[1].NetworkAddress = %q, want %q", resp.ReferralEntries[1].NetworkAddress, `\dc2\ns`)
	}
}

func TestParseDfsReferralResponseTooShort(t *testing.T) {
	_, err := parseDfsReferralResponse(make([]byte, 4))
	if err == nil {
		t.Error("expected error for short data")
	}
}

func TestReadUTF16NullTerm(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"empty", []byte{0, 0}, ""},
		{"hello", append(writeUTF16LE("hello")[:10], 0, 0), "hello"},
		{"odd byte", []byte{0x41}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readUTF16NullTerm(tt.data)
			if got != tt.want {
				t.Errorf("readUTF16NullTerm = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetDfsReferralRequiresTree(t *testing.T) {
	ft := &fakeTransport{connected: true}
	c := newTestClient(ft)

	_, err := c.GetDfsReferral(`\domain\share`, 4)
	if err == nil {
		t.Error("expected error when no tree connect is established")
	}
}
