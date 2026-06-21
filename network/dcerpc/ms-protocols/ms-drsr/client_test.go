package msdrsr

import "testing"

// TestClientZeroState checks the accessors and Close are safe on a Client that has not
// connected — Close must not panic or dereference a nil RPC client, and the handle must
// read as the null context handle.
func TestClientZeroState(t *testing.T) {
	c := New("dc.example.test", nil)

	if !c.Handle().IsNull() {
		t.Error("handle should be null before Connect")
	}
	if c.SessionKey() != nil {
		t.Error("session key should be nil before Connect")
	}
	if c.ServerExtensions() != nil {
		t.Error("server extensions should be nil before Connect")
	}
	if c.RPC() != nil {
		t.Error("RPC client should be nil before Connect")
	}
	if err := c.Close(); err != nil {
		t.Errorf("Close on an unconnected client should be a no-op, got %v", err)
	}
}
