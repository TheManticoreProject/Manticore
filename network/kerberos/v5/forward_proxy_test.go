package kerberos

import (
	"bytes"
	"testing"

	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

func TestForwardedAndProxyRequestShape(t *testing.T) {
	c := fakeTGTClient(t)
	addresses := []messages.HostAddress{{AddrType: 2, Address: []byte{192, 0, 2, 10}}}

	forwardedName := messages.PrincipalName{NameType: messages.NameTypeSRVInst, NameString: []string{"krbtgt", c.realm}}
	forwarded, err := c.buildForwardedProxyTGSReq(kdcOptionForwarded, forwardedName, addresses, 101)
	if err != nil {
		t.Fatal(err)
	}
	if forwarded.ReqBody.KDCOptions.At(kdcOptionForwarded) == 0 || forwarded.ReqBody.KDCOptions.At(kdcOptionProxy) != 0 {
		t.Fatalf("FORWARDED options = %08b", forwarded.ReqBody.KDCOptions.Bytes)
	}
	if len(forwarded.ReqBody.Addresses) != 1 || !bytes.Equal(forwarded.ReqBody.Addresses[0].Address, addresses[0].Address) {
		t.Fatalf("FORWARDED addresses = %#v", forwarded.ReqBody.Addresses)
	}

	proxyName, _ := parseSPN("cifs/host.corp.local", c.realm)
	proxy, err := c.buildForwardedProxyTGSReq(kdcOptionProxy, proxyName, addresses, 102)
	if err != nil {
		t.Fatal(err)
	}
	if proxy.ReqBody.KDCOptions.At(kdcOptionProxy) == 0 || proxy.ReqBody.KDCOptions.At(kdcOptionForwarded) != 0 {
		t.Fatalf("PROXY options = %08b", proxy.ReqBody.KDCOptions.Bytes)
	}
	if _, err := proxy.Marshal(); err != nil {
		t.Fatalf("marshal PROXY request: %v", err)
	}
}

func TestForwardedAndProxyRequireTGTAndAddresses(t *testing.T) {
	c := NewClient("alice", "CORP.LOCAL", "10.0.0.1")
	address := []messages.HostAddress{{AddrType: 2, Address: []byte{192, 0, 2, 10}}}
	if _, _, _, _, err := c.GetForwardedTGT(address); err == nil {
		t.Fatal("GetForwardedTGT without a TGT succeeded")
	}
	if _, _, _, _, err := c.GetProxyTicket("cifs/host", address); err == nil {
		t.Fatal("GetProxyTicket without a TGT succeeded")
	}
	c = fakeTGTClient(t)
	if _, err := c.buildForwardedProxyTGSReq(kdcOptionForwarded, c.tgtTicket.SName, nil, 1); err == nil {
		t.Fatal("addressless FORWARDED request was built")
	}
}
