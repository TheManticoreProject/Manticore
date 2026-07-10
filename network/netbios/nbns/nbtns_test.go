package nbns

import (
	"net"
	"testing"
	"time"
)

func TestNBNSIntegration(t *testing.T) {
	// Create NBNS server
	nbns := NewNetBIOSNameServer(true)
	// Create UDP and TCP servers
	udpServer, err := NewUDPServer(":0", nbns)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	tcpServer, err := NewTCPServer(":0", nbns)
	if err != nil {
		t.Fatalf("Failed to create TCP server: %v", err)
	}

	// Start servers
	if err := udpServer.Start(); err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer udpServer.Stop()

	if err := tcpServer.Start(); err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer tcpServer.Stop()

	// Test name registration
	t.Run("NameRegistration", func(t *testing.T) {
		name := "TESTNAME"
		ip := net.ParseIP("192.168.1.1")
		err := nbns.RegisterName(name, "", Unique, ip, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register name: %v", err)
		}

		// Verify registration
		owners, nameType, _, err := nbns.QueryName(name, "")
		if err != nil {
			t.Errorf("Failed to query name: %v", err)
		}
		if nameType != Unique {
			t.Errorf("Wrong name type: got %v, want %v", nameType, Unique)
		}
		if len(owners) != 1 || !owners[0].Equal(ip) {
			t.Errorf("Wrong owners: got %v, want [%v]", owners, ip)
		}
	})

	// Test name conflict
	t.Run("NameConflict", func(t *testing.T) {
		name := "CONFLICT"
		ip1 := net.ParseIP("192.168.1.2")
		ip2 := net.ParseIP("192.168.1.3")

		// Register first name
		err := nbns.RegisterName(name, "", Unique, ip1, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register first name: %v", err)
		}

		// Try to register same name
		err = nbns.RegisterName(name, "", Unique, ip2, 24*time.Hour)
		if err == nil {
			t.Error("Expected conflict error, got nil")
		}
	})

	// Test group name
	t.Run("GroupName", func(t *testing.T) {
		name := "GROUP"
		ip1 := net.ParseIP("192.168.1.4")
		ip2 := net.ParseIP("192.168.1.5")

		// Register group members
		err := nbns.RegisterName(name, "", Group, ip1, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register first group member: %v", err)
		}

		err = nbns.RegisterName(name, "", Group, ip2, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register second group member: %v", err)
		}

		// Verify group
		owners, nameType, _, err := nbns.QueryName(name, "")
		if err != nil {
			t.Errorf("Failed to query group: %v", err)
		}
		if nameType != Group {
			t.Errorf("Wrong name type: got %v, want %v", nameType, Group)
		}
		if len(owners) != 2 {
			t.Errorf("Wrong number of owners: got %d, want 2", len(owners))
		}
	})
}

// TestFlagBitLayout asserts the corrected NBNS header flag and RR NB_FLAGS bit
// layout per RFC 1002 4.2.1.1 and 4.2.1.3.
func TestFlagBitLayout(t *testing.T) {
	// RA (recursion available) is header bit 8 == 0x0080; it is not a group bit.
	if FlagRecursionAvailable != 0x0080 {
		t.Errorf("FlagRecursionAvailable: got 0x%04X, want 0x0080", FlagRecursionAvailable)
	}
	// The Group (G) bit is the MSB of the 16-bit RR NB_FLAGS field.
	if NBFlagGroup != 0x8000 {
		t.Errorf("NBFlagGroup: got 0x%04X, want 0x8000", NBFlagGroup)
	}
	if RcodeMask != 0x000F {
		t.Errorf("RcodeMask: got 0x%04X, want 0x000F", RcodeMask)
	}
}

// TestRcodeNameErrorMatch ensures the RCODE comparison classifies only RCODE 3
// as NAME_ERROR and does not falsely match RCODE 1 (FMT_ERR) or 2 (SRV_ERR).
func TestRcodeNameErrorMatch(t *testing.T) {
	isNameError := func(flags uint16) bool {
		return flags&RcodeMask == RcodeNameError
	}
	// Response with unrelated high bits set plus the RCODE nibble.
	base := FlagResponse | FlagAuthoritative | FlagRecursionAvailable
	for rcode := uint16(0); rcode <= 0x0007; rcode++ {
		got := isNameError(base | rcode)
		want := rcode == RcodeNameError
		if got != want {
			t.Errorf("RCODE %d: isNameError=%v, want %v", rcode, got, want)
		}
	}
}

// TestGroupBitInResourceRecord verifies a group name query response carries the
// Group bit in the RR NB_FLAGS (0x8000) and leaves the header RA bit (0x0080)
// clear, and that a group registration is read from the RR NB_FLAGS.
func TestGroupBitInResourceRecord(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	h := NewPacketHandler(nbns)

	name := "GRPNAME"
	ip := net.ParseIP("192.168.1.10")
	if err := nbns.RegisterName(name, "", Group, ip, time.Hour); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}

	request := &NBNSPacket{
		Header:    NBNSHeader{Flags: OpNameQuery, Questions: 1},
		Questions: []NBNSQuestion{{Name: &NetBIOSName{Name: name}, Type: QuestionTypeNB, Class: QuestionClassIn}},
	}
	response := &NBNSPacket{}
	h.handleNameQuery(request, response)

	if len(response.Answers) != 1 {
		t.Fatalf("expected 1 answer, got %d", len(response.Answers))
	}
	var entry ADDR_ENTRY
	if err := entry.Unmarshal(response.Answers[0].RData); err != nil {
		t.Fatalf("Unmarshal ADDR_ENTRY: %v", err)
	}
	if entry.Flags&NBFlagGroup == 0 {
		t.Errorf("group query answer: G bit not set in RR NB_FLAGS (0x%04X)", entry.Flags)
	}
	if response.Header.Flags&0x0080 != 0 {
		t.Errorf("group query answer: header 0x0080 (RA) bit must not be used as a group bit (flags=0x%04X)", response.Header.Flags)
	}

	// Registration: the G bit in the RR NB_FLAGS drives the unique/group type.
	nbns2 := NewNetBIOSNameServer(false)
	h2 := NewPacketHandler(nbns2)
	groupEntry := ADDR_ENTRY{Flags: NBFlagGroup, Address: 0xC0A8010B}
	regReq := &NBNSPacket{
		Header: NBNSHeader{Flags: OpRegistration}, // header 0x0080 intentionally clear
		Answers: []NBNSResourceRecord{{
			Name:     &NetBIOSName{Name: "REGGRP"},
			Type:     QuestionTypeNB,
			Class:    QuestionClassIn,
			RDLength: 6,
			RData:    groupEntry.Marshal(),
		}},
	}
	h2.handleRegistration(regReq, &NBNSPacket{})
	_, nameType, _, err := nbns2.QueryName("REGGRP", "")
	if err != nil {
		t.Fatalf("QueryName after registration: %v", err)
	}
	if nameType != Group {
		t.Errorf("registration name type: got %v, want Group", nameType)
	}
}
