package kerberos

import (
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// TestApplyClockSkewForward verifies that when the KDC clock is ahead of the
// client, the computed offset shifts the client's now() to match the KDC within
// a small tolerance.
func TestApplyClockSkewForward(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")

	// KDC reports a time 10 minutes ahead of ours.
	serverTime := time.Now().UTC().Add(10 * time.Minute)
	if !c.applyClockSkew(messages.KRBError{STime: serverTime}) {
		t.Fatal("applyClockSkew returned false for a valid server time")
	}

	drift := c.now().Sub(serverTime)
	if drift < -2*time.Second || drift > 2*time.Second {
		t.Errorf("after skew correction now()=%v drifts %v from server time %v, want ~0",
			c.now(), drift, serverTime)
	}
	// The offset must be roughly +10 minutes.
	if c.clockOffset < 9*time.Minute || c.clockOffset > 11*time.Minute {
		t.Errorf("clockOffset = %v, want ~10m", c.clockOffset)
	}
}

// TestApplyClockSkewBackward covers a KDC that is behind the client.
func TestApplyClockSkewBackward(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")
	serverTime := time.Now().UTC().Add(-7 * time.Minute)
	if !c.applyClockSkew(messages.KRBError{STime: serverTime}) {
		t.Fatal("applyClockSkew returned false")
	}
	if c.clockOffset > -6*time.Minute || c.clockOffset < -8*time.Minute {
		t.Errorf("clockOffset = %v, want ~-7m", c.clockOffset)
	}
	drift := c.now().Sub(serverTime)
	if drift < -2*time.Second || drift > 2*time.Second {
		t.Errorf("now() drifts %v from server time, want ~0", drift)
	}
}

// TestApplyClockSkewNoServerTime verifies a KRB-ERROR without a server time is
// not actionable (no retry warranted).
func TestApplyClockSkewNoServerTime(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")
	if c.applyClockSkew(messages.KRBError{}) {
		t.Error("applyClockSkew should return false when STime is zero")
	}
	if c.clockOffset != 0 {
		t.Errorf("clockOffset = %v, want 0 (unchanged)", c.clockOffset)
	}
}

// TestClockSkewNotCompounded verifies applying skew twice does not accumulate:
// the offset is always measured from the raw system clock, so a second identical
// correction yields the same offset.
func TestClockSkewNotCompounded(t *testing.T) {
	c := NewClient("alice", "corp.local", "10.0.0.1")
	serverTime := time.Now().UTC().Add(5 * time.Minute)

	c.applyClockSkew(messages.KRBError{STime: serverTime})
	first := c.clockOffset
	c.applyClockSkew(messages.KRBError{STime: serverTime})
	second := c.clockOffset

	if diff := first - second; diff < -2*time.Second || diff > 2*time.Second {
		t.Errorf("offset compounded: first=%v second=%v", first, second)
	}
}
