package ip_test

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/network/ip"
)

// TestIPv6IsInSubnetUsesPrefix checks subnet membership compares only the
// subnet's leading MaskBits; it used to require the two addresses to be equal.
func TestIPv6IsInSubnetUsesPrefix(t *testing.T) {
	subnet := ip.NewIPv6FromString("2001:db8:abcd::/48")
	for _, tc := range []struct {
		addr string
		in   bool
	}{
		{"2001:db8:abcd::1", true},
		{"2001:db8:abcd:ffff:ffff:ffff:ffff:ffff", true},
		{"2001:db8:abce::", false},
		{"2001:db9::", false},
	} {
		if got := ip.NewIPv6FromString(tc.addr).IsInSubnet(subnet); got != tc.in {
			t.Errorf("%s in %s = %v, want %v", tc.addr, subnet.CIDRAddress(), got, tc.in)
		}
	}
	for _, tc := range []struct {
		subnet, addr string
		in           bool
	}{
		{"::/0", "ffff::1", true},
		{"fe80::1/128", "fe80::1", true},
		{"fe80::1/128", "fe80::2", false},
		{"2001:db8::/65", "2001:db8::7fff:0:0:1", true},
		{"2001:db8::/65", "2001:db8::8000:0:0:1", false},
	} {
		if got := ip.NewIPv6FromString(tc.addr).IsInSubnet(ip.NewIPv6FromString(tc.subnet)); got != tc.in {
			t.Errorf("%s in %s = %v, want %v", tc.addr, tc.subnet, got, tc.in)
		}
	}
	if ip.NewIPv6FromString("::1").IsInSubnet(&ip.IPv6{H: 1, MaskBits: 129}) {
		t.Error("a /129 subnet contained an address")
	}
}

func TestIPv6ComputeMaskAndCIDRRange(t *testing.T) {
	p := ip.NewIPv6FromString("2001:db8:1:2:3:4:5:6/64")
	if got := p.ComputeMask().CIDRAddress(); got != "2001:db8:1:2:0:0:0:0/64" {
		t.Errorf("ComputeMask = %s", got)
	}
	r := p.CIDRRange()
	if r.Start.String() != "2001:db8:1:2:0:0:0:0" || r.End.String() != "2001:db8:1:2:ffff:ffff:ffff:ffff" {
		t.Errorf("CIDRRange = %s", r)
	}
	if r := ip.NewIPv6FromString("::/0").CIDRRange(); r.Start.String() != "0:0:0:0:0:0:0:0" || r.End.String() != "ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff" {
		t.Errorf("/0 CIDRRange = %s", r)
	}
}

func TestIPv4CIDRRange(t *testing.T) {
	for _, tc := range []struct{ cidr, start, end string }{
		{"10.1.2.3/24", "10.1.2.0/32", "10.1.2.255/32"},
		{"10.1.2.3/32", "10.1.2.3/32", "10.1.2.3/32"},
		{"10.1.2.3/0", "0.0.0.0/32", "255.255.255.255/32"},
		{"192.168.7.200/30", "192.168.7.200/32", "192.168.7.203/32"},
	} {
		r := ip.NewIPv4FromString(tc.cidr).CIDRRange()
		if r.Start.String() != tc.start || r.End.String() != tc.end {
			t.Errorf("%s: CIDRRange = %s, want %s - %s", tc.cidr, r, tc.start, tc.end)
		}
	}
}

func TestRangeIteration(t *testing.T) {
	var got []string
	for a := range ip.NewIPv4FromString("10.0.0.254/31").CIDRRange().All() {
		got = append(got, a.String())
	}
	if len(got) != 2 || got[0] != "10.0.0.254/32" || got[1] != "10.0.0.255/32" {
		t.Errorf("IPv4 /31 = %v", got)
	}

	// The last address of the space ends the loop instead of wrapping to 0.0.0.0.
	n := 0
	for range (&ip.IPv4Range{Start: ip.NewIPv4(255, 255, 255, 254, 32), End: ip.NewIPv4(255, 255, 255, 255, 32)}).All() {
		n++
	}
	if n != 2 {
		t.Errorf("range ending at 255.255.255.255 yielded %d addresses, want 2", n)
	}

	// Early stop on a huge range, and an empty range.
	n = 0
	for range ip.NewIPv4FromString("0.0.0.0/0").CIDRRange().All() {
		if n++; n == 5 {
			break
		}
	}
	if n != 5 {
		t.Errorf("early stop yielded %d", n)
	}
	for range (&ip.IPv4Range{Start: ip.NewIPv4(10, 0, 0, 2, 32), End: ip.NewIPv4(10, 0, 0, 1, 32)}).All() {
		t.Error("an empty range yielded an address")
	}

	// IPv6: the low 64 bits carry into the high 64.
	r := &ip.IPv6Range{Start: ip.NewIPv6FromString("::ffff:ffff:ffff:fffe"), End: ip.NewIPv6FromString("0:0:0:1::1")}
	got = got[:0]
	for a := range r.All() {
		got = append(got, a.String())
	}
	want := []string{"0:0:0:0:ffff:ffff:ffff:fffe", "0:0:0:0:ffff:ffff:ffff:ffff", "0:0:0:1:0:0:0:0", "0:0:0:1:0:0:0:1"}
	if len(got) != len(want) {
		t.Fatalf("IPv6 carry range = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("IPv6 carry range[%d] = %s, want %s", i, got[i], want[i])
		}
	}
	n = 0
	for range ip.NewIPv6FromString("ffff:ffff:ffff:ffff:ffff:ffff:ffff:fffe/127").CIDRRange().All() {
		n++
	}
	if n != 2 {
		t.Errorf("range ending at the last IPv6 address yielded %d, want 2", n)
	}
}
