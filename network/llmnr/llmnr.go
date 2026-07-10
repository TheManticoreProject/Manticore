// Package llmnr implements a Link-Local Multicast Name Resolution (LLMNR)
// client and the message types defined by RFC 4795. LLMNR reuses the DNS
// message format (RFC 1035) to resolve single-label, link-local names over the
// multicast group 224.0.0.252 (or FF02::1:3) on UDP/TCP port 5355.
//
// The Client type (see client.go) exposes the low-level Query primitive, which
// sends a single question and returns the raw *message.Message it was answered
// with. This file layers the high-level resolver API most callers actually
// want on top of it: Resolve, ResolveA and ResolveAAAA turn a name into a slice
// of net.IP, and LookupRecords returns the typed resource records answering a
// name for a given type. Each helper reuses the typed RDATA accessors on
// resourcerecord.ResourceRecord (AsA/AsAAAA/…) to decode answers rather than
// walking the wire format by hand.
//
// No-answer semantics: when a responder replies but its answer section carries
// no record matching the queried name and type, the resolver helpers return an
// empty (non-nil) slice and a nil error, i.e. "resolved to nothing" is not an
// error. A genuine failure to obtain any response (the query timing out because
// no host on the link owns the name, or the caller's context being cancelled)
// is surfaced as the error returned by Client.Query.
//
// Usage example:
//
//	client, err := llmnr.NewClient()
//	if err != nil {
//	    log.Fatalf("failed to create client: %v", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//
//	ips, err := client.Resolve(ctx, "wpad")
//	if err != nil {
//	    log.Fatalf("resolve failed: %v", err)
//	}
//	for _, ip := range ips {
//	    fmt.Println(ip)
//	}
package llmnr

import (
	"context"
	"net"
	"strings"

	"github.com/TheManticoreProject/Manticore/network/llmnr/llmnr_type"
	"github.com/TheManticoreProject/Manticore/network/llmnr/resourcerecord"
)

// LookupRecords sends an LLMNR query for name and qtype and returns the answer
// records that actually answer it: only records whose owner name matches name
// (compared case-insensitively, as DNS names are case-insensitive) and whose
// type equals qtype are returned. This filtering discards any unrelated records
// a responder may have bundled into the answer section.
//
// The returned slice is always non-nil; a response that carries no matching
// record yields an empty slice and a nil error (see the package-level no-answer
// semantics). An error is returned only when the underlying query fails to
// obtain a response (timeout or context cancellation).
func (c *Client) LookupRecords(ctx context.Context, name string, qtype llmnr_type.Type) ([]resourcerecord.ResourceRecord, error) {
	resp, err := c.Query(ctx, name, qtype)
	if err != nil {
		return nil, err
	}

	records := make([]resourcerecord.ResourceRecord, 0, len(resp.Answers))
	for _, rr := range resp.Answers {
		if rr.Type != qtype {
			continue
		}
		if !strings.EqualFold(string(rr.Name), name) {
			continue
		}
		records = append(records, rr)
	}
	return records, nil
}

// ResolveA resolves name to its IPv4 addresses by issuing a Type A LLMNR query
// and decoding every matching A answer with resourcerecord.AsA. The returned
// slice is always non-nil and holds one net.IP per A record answering name (an
// empty slice when the responder returned none). An error is returned only when
// the query itself fails (timeout or context cancellation); a malformed A record
// is skipped rather than failing the whole resolution.
func (c *Client) ResolveA(ctx context.Context, name string) ([]net.IP, error) {
	records, err := c.LookupRecords(ctx, name, llmnr_type.TypeA)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, 0, len(records))
	for i := range records {
		ip, err := records[i].AsA()
		if err != nil {
			continue
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

// ResolveAAAA resolves name to its IPv6 addresses by issuing a Type AAAA LLMNR
// query and decoding every matching AAAA answer with resourcerecord.AsAAAA. The
// returned slice is always non-nil and holds one net.IP per AAAA record
// answering name (an empty slice when the responder returned none). An error is
// returned only when the query itself fails (timeout or context cancellation); a
// malformed AAAA record is skipped rather than failing the whole resolution.
func (c *Client) ResolveAAAA(ctx context.Context, name string) ([]net.IP, error) {
	records, err := c.LookupRecords(ctx, name, llmnr_type.TypeAAAA)
	if err != nil {
		return nil, err
	}

	ips := make([]net.IP, 0, len(records))
	for i := range records {
		ip, err := records[i].AsAAAA()
		if err != nil {
			continue
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

// Resolve resolves name to all of its IPv4 and IPv6 addresses by issuing both a
// Type A and a Type AAAA query concurrently and concatenating their results
// (A addresses first, then AAAA). The two queries run in parallel so the
// combined lookup is not slower than a single one.
//
// Because a host commonly owns only one address family, Resolve is deliberately
// lenient: it returns the addresses from whichever query succeeded and only
// propagates an error when both queries failed, in which case the A query's
// error is returned. The returned slice is always non-nil; it is empty when the
// name resolved to no addresses at all.
func (c *Client) Resolve(ctx context.Context, name string) ([]net.IP, error) {
	type result struct {
		ips []net.IP
		err error
	}

	aChan := make(chan result, 1)
	aaaaChan := make(chan result, 1)

	go func() {
		ips, err := c.ResolveA(ctx, name)
		aChan <- result{ips: ips, err: err}
	}()
	go func() {
		ips, err := c.ResolveAAAA(ctx, name)
		aaaaChan <- result{ips: ips, err: err}
	}()

	a := <-aChan
	aaaa := <-aaaaChan

	// Only fail when neither address family could be obtained; a host owning
	// just A (or just AAAA) records must still resolve successfully.
	if a.err != nil && aaaa.err != nil {
		return nil, a.err
	}

	ips := make([]net.IP, 0, len(a.ips)+len(aaaa.ips))
	ips = append(ips, a.ips...)
	ips = append(ips, aaaa.ips...)
	return ips, nil
}
