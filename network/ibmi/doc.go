// Package ibmi is the IBM i Client Access host-server network protocol
// boundary of Manticore.
//
// It owns service discovery (TCP/449 service mapper), the common 20-byte
// host-server frame and LL/CP fields, transport policy, authentication, and
// the Sign-on and Remote Command sessions. Client Access is an independent
// protocol stack: it is not DCE/RPC, does not use NDR, does not traverse SMB
// and is not a 5250 terminal session, so none of those packages are reused.
//
// IBM i platform definitions and program-call contracts live under the root
// ibmi tree. Packages here may import only the root ibmi, ibmi/ccsid and
// ibmi/rpc packages from it, never a concrete program wrapper or an ibmi/api
// feature package. Lower layers never import network/ibmi/client, and only
// _test.go files may import the network/ibmi/internal/hosttest fake servers.
// These rules are enforced by ibmi/architecture_test.go.
//
// Examples alias this package as ibminet, since the platform package is also
// named ibmi.
package ibmi
