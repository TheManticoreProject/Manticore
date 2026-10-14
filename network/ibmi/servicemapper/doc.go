// Package servicemapper queries the IBM i service mapper (as-svrmap, TCP
// 449) for host-server ports.
//
// The mapper is an IBM i Access discovery protocol, not ONC RPC portmap, DNS
// SRV or a host-server frame. Each lookup uses one short-lived connection:
// the client writes the raw ASCII service name, and a successful reply is
// exactly '+' followed by a big-endian 32-bit port.
//
// The mapper is cleartext and unauthenticated, so its answer is only a
// port. A lookup for a TLS ("-s") name always yields a TLS endpoint, and TLS
// then authenticates the real server; discovery never substitutes a
// cleartext service when a secure lookup fails.
package servicemapper
