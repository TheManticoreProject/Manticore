package client

import (
	"github.com/TheManticoreProject/Manticore/ibmi"
	ibminet "github.com/TheManticoreProject/Manticore/network/ibmi"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd"
)

// SessionInfo is an immutable, non-secret snapshot of a connection. It never
// contains seeds, substitutes, raw authentication frames or the credential,
// and its slices do not alias client state.
type SessionInfo struct {
	Connected      bool
	SignonEndpoint ibminet.Endpoint
	RemoteEndpoint ibminet.Endpoint
	// ServerVersion is the raw Sign-on VRM integer; VRM is its decoding.
	ServerVersion  uint32
	VRM            ibmi.VRM
	PasswordLevel  byte
	PasswordFamily ibminet.PasswordFamily
	// SystemCCSID comes from Sign-on; JobCCSID is the Remote Command job
	// CCSID in use and NegotiatedJobCCSID the value the server reported.
	SystemCCSID        uint32
	JobCCSID           uint32
	NegotiatedJobCCSID uint32
	RemoteCommandLevel uint16
	Capabilities       rmtcmd.Capabilities
	ServerJob          string
	Warnings           []string
}

func (s SessionInfo) clone() SessionInfo {
	s.Warnings = append([]string(nil), s.Warnings...)
	return s
}
