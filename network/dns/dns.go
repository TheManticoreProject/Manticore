package dns

import (
	"context"
	"net"
	"strings"
	"time"
)

// DNSLookup performs a DNS lookup of a hostname (FQDN or hostname) to a specified DNS server.
//
// A failed lookup returns no addresses, the same as a name with none; use
// DNSLookupWithError to tell the two apart.
//
// Parameters:
// - hostname: A string representing the hostname or FQDN to be resolved.
// - dnsServer: A string representing the DNS server to use for the lookup.
//
// Returns:
// - A slice of strings containing the IP addresses of the DNS server.
func DNSLookup(hostname string, dnsServer string) []string {
	ip, _ := DNSLookupWithError(hostname, dnsServer)
	return ip
}

// DNSLookupWithError is DNSLookup that also returns the resolver's error, so a
// failure (no answer from the server, NXDOMAIN, SERVFAIL) can be told apart from
// a successful lookup. The error is a *net.DNSError when the resolver produced
// one.
//
// Parameters:
// - hostname: A string representing the hostname or FQDN to be resolved.
// - dnsServer: A string representing the DNS server to use for the lookup.
//
// Returns:
// - A slice of strings containing the resolved IP addresses.
// - The resolver's error, or nil.
func DNSLookupWithError(hostname string, dnsServer string) ([]string, error) {
	dnsServer = dnsServerAddress(dnsServer)

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}
	return r.LookupHost(context.Background(), hostname)
}

// dnsServerAddress returns the dial address for a DNS server, appending the
// default port 53 when the server has no port. Bare IPv6 addresses such as
// "::1" or bracketed ones such as "[::1]" are bracketed once, and addresses
// with a port such as "[::1]:5353" are kept as-is.
func dnsServerAddress(dnsServer string) string {
	if _, _, err := net.SplitHostPort(dnsServer); err == nil {
		return dnsServer
	}
	if strings.HasPrefix(dnsServer, "[") && strings.HasSuffix(dnsServer, "]") {
		dnsServer = dnsServer[1 : len(dnsServer)-1]
	}
	return net.JoinHostPort(dnsServer, "53")
}
