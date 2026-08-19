package kerberos

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func allZero(b []byte) bool {
	return bytes.Equal(b, make([]byte, len(b)))
}

func TestDestroyClearsAllKerberosMaterial(t *testing.T) {
	src := fakeServiceTicketClient(t)
	blob, err := src.ExportTGTKirbi()
	if err != nil {
		t.Fatal(err)
	}
	st, err := LoadServiceTicketFromKirbiBytes(blob, "cifs/host.corp.local")
	if err != nil {
		t.Fatal(err)
	}

	c := NewClient("", "", "10.0.0.1")
	if err := c.LoadServiceTicket(st); err != nil {
		t.Fatal(err)
	}
	c.pkinitReplyKey = bytes.Repeat([]byte{0x31}, 32)
	c.pkinitReplyEType = messages.ETypeAES256CTSHMACSHA196
	armorInput := bytes.Repeat([]byte{0x42}, 32)
	c.WithFASTArmor("", "", messages.Ticket{}, nil, armorInput, messages.ETypeAES256CTSHMACSHA196)

	preloadedKey := c.preloadedTGS[normalizeSPN("cifs/host.corp.local")].sessionKey
	cachedKey := c.serviceTickets[normalizeSPN("cifs/host.corp.local")].credInfo.Key.KeyValue
	pkinitKey := c.pkinitReplyKey
	armorKey := c.fast.sessionKey

	if _, _, _, _, err := c.GetTGS("cifs/host.corp.local", true); err != nil {
		t.Fatalf("precondition: loaded service ticket is unusable: %v", err)
	}
	if _, err := c.ExportServiceTicketKirbi("cifs/host.corp.local"); err != nil {
		t.Fatalf("precondition: cached service ticket is not exportable: %v", err)
	}

	c.Destroy()

	for name, key := range map[string][]byte{
		"preloaded service key": preloadedKey,
		"cached service key":    cachedKey,
		"PKINIT reply key":      pkinitKey,
		"FAST armor key":        armorKey,
	} {
		if !allZero(key) {
			t.Errorf("%s was not zeroed: %x", name, key)
		}
	}
	if c.cred != nil || c.hasTGT || c.fast != nil || c.preloadedTGS != nil || c.serviceTickets != nil {
		t.Fatalf("Destroy retained authentication state: %+v", c)
	}
	if key, etype := c.PKINITReplyKey(); key != nil || etype != 0 {
		t.Errorf("PKINITReplyKey after Destroy = (%x, %d)", key, etype)
	}
	if allZero(st.SessionKey) {
		t.Error("Destroy zeroed the caller-owned service-ticket key")
	}
	if allZero(armorInput) {
		t.Error("Destroy zeroed the caller-owned FAST armor key")
	}
	if _, _, _, _, err := c.GetTGS("cifs/host.corp.local", true); err == nil {
		t.Error("GetTGS returned a preloaded ticket after Destroy")
	}
	if _, err := c.ExportServiceTicketKirbi("cifs/host.corp.local"); err == nil {
		t.Error("ExportServiceTicketKirbi succeeded after Destroy")
	}
}
