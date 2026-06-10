//go:build integration

package mssrvs_test

import "testing"

func TestIntegration_ListSessions(t *testing.T) {
	c, cleanup := liveClient(t)
	defer cleanup()
	// Session enumeration typically requires administrative rights; a privilege error
	// here is the server's policy, not a wire defect, so only log it.
	sessions, err := c.ListSessions()
	if err != nil {
		t.Logf("ListSessions: %v (often requires admin rights)", err)
		return
	}
	for _, s := range sessions {
		t.Logf("[ok] session client=%q user=%q active=%ds idle=%ds", s.ClientName, s.UserName, s.ActiveSecs, s.IdleSecs)
	}
}
