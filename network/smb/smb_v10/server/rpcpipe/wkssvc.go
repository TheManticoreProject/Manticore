package rpcpipe

import (
	"github.com/TheManticoreProject/Manticore/logger"
	wkssvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/6bffd098-a112-3610-9833-46c3f87e345a/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/rpcserver"
	mswkst "github.com/TheManticoreProject/Manticore/windows/protocols/ms-wkst"
)

// wkssvcService answers the wkssvc interface ([MS-WKST]).
//
// It serves NetrWkstaGetInfo and nothing else, because that is the method a
// client calls to identify a host: it is how a browser learns the computer name
// and the workgroup to file the host under. Every other opnum faults with
// nca_s_op_rng_error.
type wkssvcService struct {
	// serverName and domainName are what the workstation reports itself as.
	serverName string
	domainName string

	// versionMajor and versionMinor are the operating system version reported.
	versionMajor uint32
	versionMinor uint32
}

// netrWkstaGetInfoRequest is the [in] parameter set of NetrWkstaGetInfo.
//
// It mirrors the client-side declaration in the interface's functions package,
// which is unexported there.
type netrWkstaGetInfoRequest struct {
	ServerName *ndr.WSTR `ndr:"unique"`
	Level      ndr.DWORD
}

// netrWkstaGetInfoResponse is the [out] parameter and return value of
// NetrWkstaGetInfo.
type netrWkstaGetInfoResponse struct {
	WkstaInfo mswkst.WKSTA_INFO
	Status    ndr.DWORD `ndr:"retval"`
}

// platformIDNT is PLATFORM_ID_NT, the platform identifier every current host
// reports ([MS-WKST] 2.2.5.1). The older DOS, OS/2 and OSF values describe
// systems no client still expects to meet.
const platformIDNT = 500

// AbstractSyntax identifies the wkssvc interface, version 1.0.
func (s *wkssvcService) AbstractSyntax() syntax.SyntaxID {
	return wkssvc.SyntaxID()
}

// Call runs one wkssvc method.
func (s *wkssvcService) Call(opnum uint16, stub []byte) ([]byte, error) {
	if opnum != wkssvc.OpnumNetrWkstaGetInfo {
		return nil, rpcserver.ErrUnknownOpnum
	}
	return s.netrWkstaGetInfo(stub)
}

// netrWkstaGetInfo answers NetrWkstaGetInfo, opnum 0 ([MS-WKST] 3.2.4.1).
//
// Levels 100 and 101 are served. Level 100 is what an unprivileged caller gets
// and what a browser asks for; level 101 adds the LAN root, a path that has no
// meaning on a host that is not running the Windows redirector, and is reported
// as empty rather than invented.
//
// Every other level — 102 and the 502 and 1013-and-above configuration levels —
// is answered with ERROR_INVALID_LEVEL. Those describe a workstation's logged-on
// user count and redirector tuning, which this server has no equivalent of, and
// reporting zeroes for them would be describing a machine that does not exist.
func (s *wkssvcService) netrWkstaGetInfo(stub []byte) ([]byte, error) {
	var request netrWkstaGetInfoRequest
	if err := ndr.Unmarshal(stub, &request); err != nil {
		logger.Debugf("rpcpipe: wkssvc NetrWkstaGetInfo stub would not decode: %v", err)
		return nil, rpcserver.ErrBadStub
	}

	level := uint32(request.Level)
	response := &netrWkstaGetInfoResponse{
		WkstaInfo: mswkst.WKSTA_INFO{Tag: ndr.DWORD(level)},
		Status:    ndr.DWORD(wkssvc.StatusSuccess),
	}

	switch level {
	case 100:
		response.WkstaInfo.WkstaInfo100 = &mswkst.WKSTA_INFO_100{
			Wki100_platform_id:  platformIDNT,
			Wki100_computername: optionalWideString(s.serverName),
			Wki100_langroup:     optionalWideString(s.domainName),
			Wki100_ver_major:    ndr.DWORD(s.versionMajor),
			Wki100_ver_minor:    ndr.DWORD(s.versionMinor),
		}
	case 101:
		response.WkstaInfo.WkstaInfo101 = &mswkst.WKSTA_INFO_101{
			Wki101_platform_id:  platformIDNT,
			Wki101_computername: optionalWideString(s.serverName),
			Wki101_langroup:     optionalWideString(s.domainName),
			Wki101_ver_major:    ndr.DWORD(s.versionMajor),
			Wki101_ver_minor:    ndr.DWORD(s.versionMinor),
			Wki101_lanroot:      nil,
		}
	default:
		logger.Debugf("rpcpipe: wkssvc NetrWkstaGetInfo asked for level %d, which is not served", level)
		response.Status = ndr.DWORD(wkssvc.ErrorInvalidLevel)
	}

	return ndr.Marshal(response)
}

// optionalWideString returns a pointer to a wide string, or nil for the empty
// one.
//
// A name the server does not have is a null pointer and not an empty string: the
// field is [unique], so null is the representation for "not set", and a client
// that gets an empty string displays an empty name instead of falling back.
func optionalWideString(value string) *ndr.WSTR {
	if value == "" {
		return nil
	}
	wide := ndr.WSTR(value)
	return &wide
}
