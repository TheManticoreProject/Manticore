package nbtns

import (
	"net"
	"testing"
	"time"
)

func TestNBTNSIntegration(t *testing.T) {
	// Create NBTNS server
	nbtns := NewNetBIOSNameServer(true)
	// Create UDP and TCP servers
	udpServer, err := NewUDPServer(":0", nbtns)
	if err != nil {
		t.Fatalf("Failed to create UDP server: %v", err)
	}

	tcpServer, err := NewTCPServer(":0", nbtns)
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
		err := nbtns.RegisterName(name, Unique, ip, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register name: %v", err)
		}

		// Verify registration
		owners, nameType, err := nbtns.QueryName(name)
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
		err := nbtns.RegisterName(name, Unique, ip1, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register first name: %v", err)
		}

		// Try to register same name
		err = nbtns.RegisterName(name, Unique, ip2, 24*time.Hour)
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
		err := nbtns.RegisterName(name, Group, ip1, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register first group member: %v", err)
		}

		err = nbtns.RegisterName(name, Group, ip2, 24*time.Hour)
		if err != nil {
			t.Errorf("Failed to register second group member: %v", err)
		}

		// Verify group
		owners, nameType, err := nbtns.QueryName(name)
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
