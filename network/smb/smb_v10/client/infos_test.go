package client_test

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/client"
	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/types"
	msdtyp "github.com/TheManticoreProject/Manticore/windows/ms-dtyp"
)

func TestGetRemoteServerTimeReturnsNegotiatedTime(t *testing.T) {
	want := time.Date(2021, time.June, 15, 12, 30, 45, 0, time.UTC)
	ft := msdtyp.NewFILETIMEFromTime(want)

	c := &client.Client{
		Connection: &client.Connection{
			Server: &client.Server{SystemTime: types.SMB_TIME(*ft)},
		},
	}

	got, err := c.GetRemoteServerTime()
	if err != nil {
		t.Fatalf("GetRemoteServerTime returned error: %v", err)
	}
	if !got.Equal(want) {
		t.Errorf("GetRemoteServerTime = %s, want %s", got, want)
	}
}

func TestGetRemoteServerTimeUnavailable(t *testing.T) {
	// Zero SystemTime (negotiation has not populated it) -> error.
	c := &client.Client{Connection: &client.Connection{Server: &client.Server{}}}
	if _, err := c.GetRemoteServerTime(); err == nil {
		t.Error("expected an error when server time is unavailable, got nil")
	}

	// No connection -> error.
	empty := &client.Client{}
	if _, err := empty.GetRemoteServerTime(); err == nil {
		t.Error("expected an error when no connection exists, got nil")
	}
}
