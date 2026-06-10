package nbt_test

import (
	"net"
	"testing"
	"time"

	"github.com/TheManticoreProject/Manticore/network/netbios/nbt"
)

func TestNBTTransport_Connect(t *testing.T) {
	tests := []struct {
		name    string
		ip      net.IP
		port    int
		wantErr bool
	}{
		{
			name:    "Valid IPv4 connection with default port",
			ip:      net.ParseIP("127.0.0.1"),
			port:    0,
			wantErr: true, // Will fail since no server is running
		},
		{
			name:    "Valid IPv4 connection with custom port",
			ip:      net.ParseIP("127.0.0.1"),
			port:    139,
			wantErr: true,
		},
		{
			name:    "Valid IPv6 connection",
			ip:      net.ParseIP("::1"),
			port:    139,
			wantErr: true,
		},
		{
			name:    "Invalid IP",
			ip:      nil,
			port:    139,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := nbt.NewNBTTransport()
			err := tr.Connect(tt.ip, tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("NBTTransport.Connect() error = %v, wantErr %v", err, tt.wantErr)
			}
			tr.Close()
		})
	}
}

func TestNBTTransport_Send(t *testing.T) {
	tr := nbt.NewNBTTransport()

	// Test sending without connection
	_, err := tr.Send([]byte("test"))
	if err == nil {
		t.Error("NBTTransport.Send() should error when not connected")
	}

	tr.Close()
}

func TestNBTTransport_ReceiveTimesOutOnSilentServer(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test server: %v", err)
	}
	defer ln.Close()

	// Accept the connection but never write anything.
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		select {}
	}()

	addr := ln.Addr().(*net.TCPAddr)

	tr := nbt.NewNBTTransport()
	tr.SetTimeout(100 * time.Millisecond)
	if err := tr.Connect(addr.IP, addr.Port); err != nil {
		t.Fatalf("NBTTransport.Connect() error = %v", err)
	}
	defer tr.Close()

	start := time.Now()
	_, err = tr.Receive()
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("NBTTransport.Receive() should time out on a silent server, got nil error")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("NBTTransport.Receive() took %v to fail, want a bounded timeout", elapsed)
	}
}
