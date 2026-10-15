package client

import (
	"context"
	"errors"
	"testing"
	"time"

	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/dataqueue"
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
)

func TestOpenDataQueue(t *testing.T) {
	ctx := context.Background()
	srv := newServer()
	c := newClient(t, srv)
	cred := credential(t, testPassword)
	defer cred.Destroy()
	if _, err := c.OpenDataQueue(ctx, cred); !errors.Is(err, ErrNotReady) {
		t.Fatalf("before Connect = %v", err)
	}
	if err := c.Connect(ctx, cred); err != nil {
		t.Fatal(err)
	}
	dq, err := c.OpenDataQueue(ctx, cred)
	if err != nil {
		t.Fatal(err)
	}
	if dq.Info().SystemCCSID != c.SessionInfo().SystemCCSID {
		t.Errorf("data queue CCSID %d, sign-on %d", dq.Info().SystemCCSID, c.SessionInfo().SystemCCSID)
	}
	q := dataqueue.Queue{Library: "QGPL", Name: "Q"}
	if err := dq.Create(ctx, q, dataqueue.CreateOptions{MaxEntryLength: 4, Kind: dataqueue.FIFO}); err != nil {
		t.Fatal(err)
	}
	// The data-queue session outlives the Client.
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	if err := dq.Write(ctx, q, []byte("hi")); err != nil {
		t.Fatalf("write after Client.Close: %v", err)
	}
	if err := dq.Close(); err != nil {
		t.Fatal(err)
	}
	if err := srv.WaitClosed(2 * time.Second); err != nil {
		t.Fatal(err)
	}
	seen := 0
	for _, r := range srv.Requests() {
		if r.ServerID == hostserver.ServerIDDataQueue {
			seen++
		}
	}
	if seen != 5 { // 7001, 7002, 0000, 0003, 0005
		t.Errorf("%d data-queue requests", seen)
	}
}

func TestOpenDataQueueDiscoveryFailure(t *testing.T) {
	srv := newServer()
	// No DataQueuePort: the lookup goes to a mapper nothing answers.
	c := newClient(t, srv, func(cfg *Config) { cfg.DataQueuePort = 0 })
	cred := credential(t, testPassword)
	defer cred.Destroy()
	if err := c.Connect(context.Background(), cred); err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if _, err := c.OpenDataQueue(context.Background(), cred); err == nil {
		t.Fatal("OpenDataQueue succeeded without a data-queue endpoint")
	}
	if _, err := c.OpenDataQueue(context.Background(), nil); err == nil {
		t.Error("nil credential accepted")
	}
	if _, err := New(Config{Host: "h", Transport: ibminet.TransportCleartextExplicit, DataQueuePort: 70000}); !errors.Is(err, ErrConfig) {
		t.Errorf("DataQueuePort 70000 = %v", err)
	}
}
