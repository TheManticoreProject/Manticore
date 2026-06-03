package client_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
)

// TestTreeConnectWithoutSession verifies that calling TreeConnect before a
// session has been established returns an error rather than panicking with a
// nil-pointer dereference on c.Session.
func TestTreeConnectWithoutSession(t *testing.T) {
	c := &client.Client{}

	err := c.TreeConnect("share")
	if err == nil {
		t.Fatal("expected an error when TreeConnect is called without a session, got nil")
	}
}
