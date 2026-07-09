package kerberos

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strings"
)

// KDC discovery via DNS SRV records (RFC 4120 Section 7.2.3, RFC 2782).
//
// When no KDC host is configured for a realm, the client locates the realm's
// KDCs through the SRV records "_kerberos._tcp.<realm>" and "_kerberos._udp.
// <realm>". Records are ordered by RFC 2782 rules — ascending priority, then a
// weighted shuffle within each priority — and tried in that order with failover
// (see kdcSendEndpoints). The SRV target host and port are honoured, so a KDC
// on a non-standard port is reached correctly.

// endpointsForRealm returns the ordered list of KDC endpoints to contact for the
// given realm. Resolution order mirrors resolveKDCForRealm — explicit
// WithRealmKDC overrides, then the client's own configured KDC for its home
// realm, then a custom WithKDCResolver, then DNS-SRV discovery — but yields the
// full failover list rather than a single host. Nothing is hardcoded.
func (c *KerberosClient) endpointsForRealm(realm string) ([]kdcEndpoint, error) {
	realm = strings.ToUpper(realm)

	if host, ok := c.realmKDCs[realm]; ok && host != "" {
		return []kdcEndpoint{{host: host, port: defaultKDCPort}}, nil
	}
	if realm == strings.ToUpper(c.realm) && c.kdcHost != "" {
		return []kdcEndpoint{{host: c.kdcHost, port: defaultKDCPort}}, nil
	}
	if c.kdcResolver != nil {
		host, err := c.kdcResolver(realm)
		if err != nil {
			return nil, err
		}
		return []kdcEndpoint{{host: host, port: defaultKDCPort}}, nil
	}
	return c.discoverKDCEndpoints(realm)
}

// discoverKDCEndpoints resolves a realm's KDCs from DNS SRV records, querying
// both the TCP and UDP service names and merging the results into a single
// failover-ordered list (RFC 4120 Section 7.2.3). Both records normally point at
// the same KDCs; the merge de-duplicates by host:port while preserving the
// RFC 2782 ordering.
func (c *KerberosClient) discoverKDCEndpoints(realm string) ([]kdcEndpoint, error) {
	var records []*net.SRV
	var firstErr error
	for _, proto := range []string{"tcp", "udp"} {
		recs, err := c.lookupKerberosSRV(proto, realm)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		records = append(records, recs...)
	}
	if len(records) == 0 {
		if firstErr != nil {
			return nil, fmt.Errorf("kerberos: DNS SRV discovery for realm %q: %w", realm, firstErr)
		}
		return nil, fmt.Errorf("kerberos: no _kerberos SRV records for realm %q", realm)
	}
	endpoints := orderSRVEndpoints(records)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("kerberos: no usable _kerberos SRV records for realm %q", realm)
	}
	return endpoints, nil
}

// lookupKerberosSRV performs the "_kerberos._<proto>.<realm>" SRV lookup through
// the client's resolver (nil uses net.DefaultResolver). Splitting it out lets
// tests inject records without real DNS.
func (c *KerberosClient) lookupKerberosSRV(proto, realm string) ([]*net.SRV, error) {
	resolver := c.resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()
	_, addrs, err := resolver.LookupSRV(ctx, "kerberos", proto, realm)
	if err != nil {
		return nil, fmt.Errorf("_kerberos._%s.%s: %w", proto, realm, err)
	}
	return addrs, nil
}

// orderSRVEndpoints turns raw SRV records into an ordered, de-duplicated list of
// KDC endpoints following RFC 2782: records are grouped by priority (lowest
// first) and, within each priority group, ordered by a weighted random shuffle
// so higher-weight targets are more likely to be tried first while every target
// still appears exactly once (giving deterministic failover coverage). A target
// of "." (RFC 2782 "service decidedly not available") is dropped.
func orderSRVEndpoints(records []*net.SRV) []kdcEndpoint {
	return orderSRVEndpointsRand(records, rand.Float64)
}

// orderSRVEndpointsRand is orderSRVEndpoints with an injectable [0,1) random
// source, so the weighted selection can be exercised deterministically in tests.
func orderSRVEndpointsRand(records []*net.SRV, randFloat func() float64) []kdcEndpoint {
	// Work on a copy so the caller's slice is not reordered.
	recs := make([]*net.SRV, 0, len(records))
	for _, r := range records {
		if r == nil || r.Target == "." || r.Target == "" {
			continue
		}
		recs = append(recs, r)
	}

	// Stable primary sort by priority (ascending). The weighted shuffle below
	// only reorders within a priority group.
	sort.SliceStable(recs, func(i, j int) bool {
		return recs[i].Priority < recs[j].Priority
	})

	ordered := make([]*net.SRV, 0, len(recs))
	for i := 0; i < len(recs); {
		// Find the extent of the current priority group.
		j := i
		for j < len(recs) && recs[j].Priority == recs[i].Priority {
			j++
		}
		ordered = append(ordered, weightedOrder(recs[i:j], randFloat)...)
		i = j
	}

	// Map to endpoints, de-duplicating by host:port while preserving order.
	seen := make(map[string]bool, len(ordered))
	endpoints := make([]kdcEndpoint, 0, len(ordered))
	for _, r := range ordered {
		host := strings.TrimSuffix(r.Target, ".")
		ep := kdcEndpoint{host: host, port: int(r.Port)}
		key := ep.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		endpoints = append(endpoints, ep)
	}
	return endpoints
}

// weightedOrder orders one priority group by the RFC 2782 weighting algorithm:
// repeatedly pick a record with probability proportional to its weight, remove
// it, and repeat, producing a full ordering (not just a first choice) so the
// list doubles as a failover sequence. Records of weight 0 are still selectable
// (with low probability, and always eventually), per the RFC.
func weightedOrder(group []*net.SRV, randFloat func() float64) []*net.SRV {
	remaining := make([]*net.SRV, len(group))
	copy(remaining, group)

	result := make([]*net.SRV, 0, len(group))
	for len(remaining) > 0 {
		total := 0
		for _, r := range remaining {
			total += int(r.Weight)
		}

		var pick int
		if total == 0 {
			// All weights zero: pick uniformly at random.
			pick = int(randFloat() * float64(len(remaining)))
			if pick >= len(remaining) {
				pick = len(remaining) - 1
			}
		} else {
			// RFC 2782: pick the record whose running weight sum first meets or
			// exceeds a random point in [0,total).
			target := randFloat() * float64(total)
			running := 0.0
			pick = len(remaining) - 1
			for idx, r := range remaining {
				running += float64(r.Weight)
				if running >= target {
					pick = idx
					break
				}
			}
		}

		result = append(result, remaining[pick])
		remaining = append(remaining[:pick], remaining[pick+1:]...)
	}
	return result
}
