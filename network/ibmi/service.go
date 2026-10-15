package ibmi

import (
	"context"
	"fmt"
	"net"
)

// Service identifies a Client Access host server.
type Service uint8

const (
	// ServiceSignon is the Sign-on host server (as-signon / as-signon-s).
	ServiceSignon Service = iota + 1
	// ServiceRemoteCommand is the Remote Command and distributed Program
	// Call host server (as-rmtcmd / as-rmtcmd-s).
	ServiceRemoteCommand
	// ServiceDataQueue is the Data Queue host server (as-dtaq / as-dtaq-s).
	ServiceDataQueue
	// ServiceFile is the File host server for the integrated file system
	// (as-file / as-file-s).
	ServiceFile
)

// Services lists every supported service in a stable order.
func Services() []Service {
	return []Service{ServiceSignon, ServiceRemoteCommand, ServiceDataQueue, ServiceFile}
}

func (s Service) String() string {
	switch s {
	case ServiceSignon:
		return "signon"
	case ServiceRemoteCommand:
		return "remote-command"
	case ServiceDataQueue:
		return "data-queue"
	case ServiceFile:
		return "file"
	}
	return fmt.Sprintf("Service(%d)", uint8(s))
}

// Valid reports whether s is a supported service.
func (s Service) Valid() bool { return s >= ServiceSignon && s <= ServiceFile }

// MapperName returns the IBM i service-mapper name of s: the "-s" form when
// tls is true, for example "as-signon-s". It returns "" for an invalid s.
func (s Service) MapperName(tls bool) string {
	var base string
	switch s {
	case ServiceSignon:
		base = "as-signon"
	case ServiceRemoteCommand:
		base = "as-rmtcmd"
	case ServiceDataQueue:
		base = "as-dtaq"
	case ServiceFile:
		base = "as-file"
	default:
		return ""
	}
	if tls {
		return base + "-s"
	}
	return base
}

// ServiceForMapperName returns the service and TLS flag of a supported
// mapper name. ok is false for any other name.
func ServiceForMapperName(name string) (service Service, tls bool, ok bool) {
	for _, s := range Services() {
		for _, t := range []bool{false, true} {
			if s.MapperName(t) == name {
				return s, t, true
			}
		}
	}
	return 0, false, false
}

// DialFunc opens a network connection. It has the signature of
// net.Dialer.DialContext so tests can inject in-memory or scripted peers.
type DialFunc func(ctx context.Context, network, address string) (net.Conn, error)
