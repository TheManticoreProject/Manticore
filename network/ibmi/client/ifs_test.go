package client

import (
	"context"
	"errors"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func TestOpenIFS(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	srv.Files = map[string][]byte{"/home/f.txt": []byte("content")}
	c := newClient(t, srv)
	cred := credential(t, testPassword)
	defer cred.Destroy()
	if _, err := c.OpenIFS(ctx, cred); !errors.Is(err, ErrNotReady) {
		t.Fatalf("before Connect = %v", err)
	}
	if err := c.Connect(ctx, cred); err != nil {
		t.Fatal(err)
	}
	s, err := c.OpenIFS(ctx, cred)
	if err != nil {
		t.Fatal(err)
	}
	if s.Info().ServerJob == "" || s.Info().PathCCSID != 1200 {
		t.Errorf("info = %+v", s.Info())
	}
	// The file session outlives the Client.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if data, err := s.ReadFile(ctx, "/home/f.txt", 100); err != nil || string(data) != "content" {
		t.Fatalf("ReadFile after Client.Close = %q, %v", data, err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, r := range srv.Requests() {
		if r.ServerID == hostserver.ServerIDFile {
			n++
		}
	}
	if n != 7 { // 7001, 7002, 0016, 0002, 0003, 0003 (EOF), 0009
		t.Errorf("%d file-server requests", n)
	}
}

func TestOpenIFSFailures(t *testing.T) {
	srv := newServer()
	c := newClient(t, srv, func(cfg *Config) { cfg.FilePort = 0 })
	cred := credential(t, testPassword)
	defer cred.Destroy()
	if err := c.Connect(context.Background(), cred); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.OpenIFS(context.Background(), cred); err == nil {
		t.Fatal("OpenIFS succeeded without a file-server endpoint")
	}
	if _, err := c.OpenIFS(context.Background(), nil); err == nil {
		t.Error("nil credential accepted")
	}
	if _, err := New(Config{Host: "h", Transport: ibminet.TransportCleartextExplicit, FilePort: -1}); !errors.Is(err, ErrConfig) {
		t.Errorf("FilePort -1 = %v", err)
	}
}
