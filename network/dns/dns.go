package dns

import (
	"context"
	"net"
	"strings"
	"time"
)

// DNSLookup performs a DNS lookup of a hostname (FQDN or hostname) to a specified DNS server.
//
// Parameters:
// - hostname: A string representing the hostname or FQDN to be resolved.
// - dnsServer: A string representing the DNS server to use for the lookup.
//
// Returns:
// - A slice of strings containing the IP addresses of the DNS server.
func DNSLookup(hostname string, dnsServer string) []string {
	if !strings.Contains(dnsServer, ":") {
		dnsServer = dnsServer + ":53"
	}

	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(10000),
			}
			return d.DialContext(ctx, network, dnsServer)
		},
	}
	ip, _ := r.LookupHost(context.Background(), hostname)

	return ip
}
