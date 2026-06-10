//go:build integration

package mssrvs_test

import "testing"

func TestIntegration_GetServerInfo(t *testing.T) {
	c, cleanup := liveClient(t)
	defer cleanup()
	info, err := c.GetServerInfo()
	if err != nil {
		t.Fatalf("[WIRE FAIL] GetServerInfo: %v", err)
	}
	t.Logf("[ok] server %q version %d.%d platform=%d type=0x%08x %q",
		info.Name, info.VersionMajor, info.VersionMinor, info.PlatformID, info.Type, info.Comment)
}
