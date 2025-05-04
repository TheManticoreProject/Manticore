package transport_test

import (
	"net"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/smb/smb_v10/transport"
)

// MockTransport implements the Transport interface for testing
type MockTransport struct {
	connected bool
	data      []byte
	sendErr   error
	recvErr   error
}

func (m *MockTransport) Connect(ipaddr net.IP, port int) error {
	m.connected = true
	return nil
}

func (m *MockTransport) Close() error {
	m.connected = false
	return nil
}

func (m *MockTransport) Send(data []byte) (int, error) {
	if !m.connected {
		return 0, nil
	}
	if m.sendErr != nil {
		return 0, m.sendErr
	}
	m.data = data
	return len(data), nil
}

func (m *MockTransport) Receive() ([]byte, error) {
	if !m.connected {
		return nil, nil
	}
	if m.recvErr != nil {
		return nil, m.recvErr
	}
	return m.data, nil
}

func (m *MockTransport) IsConnected() bool {
	return m.connected
}

func TestNewTransport(t *testing.T) {
	tests := []struct {
		name          string
		transportType string
		wantNil       bool
	}{
		{
			name:          "NBT transport",
			transportType: "nbt",
			wantNil:       false,
		},
		{
			name:          "NBT transport uppercase",
			transportType: "NBT",
			wantNil:       false,
		},
		{
			name:          "Unsupported transport",
			transportType: "unsupported",
			wantNil:       true,
		},
		{
			name:          "Empty transport type",
			transportType: "",
			wantNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := transport.NewTransport(tt.transportType)
			if (tr == nil) != tt.wantNil {
				t.Errorf("NewTransport() = %v, wantNil = %v", tr, tt.wantNil)
			}
		})
	}
}

func TestTransportInterface(t *testing.T) {
	// Create a mock transport
	mock := &MockTransport{}

	// Test Connect
	err := mock.Connect(net.ParseIP("127.0.0.1"), 445)
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}
	if !mock.IsConnected() {
		t.Errorf("IsConnected() = false, want true after Connect()")
	}

	// Test Send
	testData := []byte("test data")
	n, err := mock.Send(testData)
	if err != nil {
		t.Errorf("Send() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Send() = %v, want %v", n, len(testData))
	}

	// Test Receive
	received, err := mock.Receive()
	if err != nil {
		t.Errorf("Receive() error = %v", err)
	}
	if string(received) != string(testData) {
		t.Errorf("Receive() = %v, want %v", string(received), string(testData))
	}

	// Test Close
	err = mock.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
	if mock.IsConnected() {
		t.Errorf("IsConnected() = true, want false after Close()")
	}
}
