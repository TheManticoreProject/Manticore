package rpcpipe

import (
	"github.com/TheManticoreProject/Manticore/logger"
	srvsvc "github.com/TheManticoreProject/Manticore/network/dcerpc/interfaces/4b324fc8-1670-01d3-1278-5a47bf6ee188/3.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/ndr"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/syntax"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/rpcserver"
	"github.com/TheManticoreProject/Manticore/windows/errors/win32"
	mssrvs "github.com/TheManticoreProject/Manticore/windows/protocols/ms-srvs"
)

// srvsvcService answers the srvsvc interface ([MS-SRVS]).
//
// It serves NetrShareEnum and nothing else, because that is the method a client
// calls to list shares — "net view", "smbclient -L" and Explorer's network
// browser all reach a host through it. Every other opnum faults with
// nca_s_op_rng_error, which is what a client expects from a server that does not
// implement a method.
type srvsvcService struct {
	// shares supplies the shares to report, called afresh on each enumeration.
	shares func() []ShareEntry
}

// netrShareEnumRequest is the [in] parameter set of NetrShareEnum.
//
// It mirrors the client-side declaration in the interface's functions package,
// which is unexported there. The NDR tags are what matter and they are identical:
// the two have to agree on the wire or nothing decodes.
type netrShareEnumRequest struct {
	ServerName            *ndr.WSTR `ndr:"unique"`
	InfoStruct            mssrvs.SHARE_ENUM_STRUCT
	PreferedMaximumLength ndr.DWORD
	ResumeHandle          *ndr.DWORD `ndr:"unique"`
}

// netrShareEnumResponse is the [out] parameter set and return value of
// NetrShareEnum.
type netrShareEnumResponse struct {
	InfoStruct   mssrvs.SHARE_ENUM_STRUCT
	TotalEntries ndr.DWORD
	ResumeHandle *ndr.DWORD `ndr:"unique"`
	Status       ndr.DWORD  `ndr:"retval"`
}

// maxPreferredLength is MAX_PREFERRED_LENGTH, the value a client sends to mean
// "no limit, send everything" ([MS-SRVS] 2.2.2.1). It is what every ordinary
// share listing sends.
const maxPreferredLength = 0xFFFFFFFF

// AbstractSyntax identifies the srvsvc interface, version 3.0.
func (s *srvsvcService) AbstractSyntax() syntax.SyntaxID {
	return srvsvc.SyntaxID()
}

// Call runs one srvsvc method.
func (s *srvsvcService) Call(opnum uint16, stub []byte) ([]byte, error) {
	if opnum != srvsvc.OpnumNetrShareEnum {
		return nil, rpcserver.ErrUnknownOpnum
	}
	return s.netrShareEnum(stub)
}

// netrShareEnum answers NetrShareEnum, opnum 15 ([MS-SRVS] 3.1.4.8).
//
// Levels 0 and 1 are served. A level the union knows but this does not — 2, 501,
// 502 and 503, which describe a share's path, permissions and use counts — is
// answered with ERROR_INVALID_LEVEL rather than a fault, because an unsupported
// information level is a successful call reporting a refusal and a client falls
// back to a lower level when it sees one.
func (s *srvsvcService) netrShareEnum(stub []byte) ([]byte, error) {
	var request netrShareEnumRequest
	if err := ndr.Unmarshal(stub, &request); err != nil {
		logger.Debugf("rpcpipe: srvsvc NetrShareEnum stub would not decode: %v", err)
		return nil, rpcserver.ErrBadStub
	}

	level := uint32(request.InfoStruct.Level)
	if level != 0 && level != 1 {
		logger.Debugf("rpcpipe: srvsvc NetrShareEnum asked for level %d, which is not served", level)
		return ndr.Marshal(&netrShareEnumResponse{
			InfoStruct: mssrvs.SHARE_ENUM_STRUCT{
				Level:     ndr.DWORD(level),
				ShareInfo: mssrvs.SHARE_ENUM_UNION{Tag: ndr.DWORD(level)},
			},
			ResumeHandle: echoResumeHandle(request.ResumeHandle, 0),
			Status:       ndr.DWORD(win32.ERROR_INVALID_LEVEL),
		})
	}

	// A resume handle past the end is not an error: the enumeration has simply
	// finished, and answering an empty list is how a client learns that.
	all := s.shares()
	from := 0
	if request.ResumeHandle != nil {
		from = int(uint32(*request.ResumeHandle))
	}
	if from < 0 || from > len(all) {
		from = len(all)
	}
	remaining := all[from:]

	// The client's budget is a preference, not a limit on the wire: it asks for
	// an answer of about that size and takes a resume handle for the rest.
	// MAX_PREFERRED_LENGTH waives it, which is what an ordinary listing sends.
	sent := remaining
	truncated := false
	if budget := uint32(request.PreferedMaximumLength); budget != maxPreferredLength {
		sent, truncated = fitEntries(remaining, level, budget)
	}

	response := &netrShareEnumResponse{
		InfoStruct:   shareEnumStruct(level, sent),
		TotalEntries: ndr.DWORD(len(remaining)),
		Status:       ndr.DWORD(win32.NERR_Success),
	}
	if truncated {
		// ERROR_MORE_DATA with a resume handle is how the enumeration continues.
		// The handle is an index into the share list, so a share added or removed
		// between two calls shifts what a resumed enumeration sees; SMB offers no
		// way to hold an enumeration open across calls, so an index is the whole
		// of what a resume handle can be.
		response.Status = ndr.DWORD(win32.ERROR_MORE_DATA)
		response.ResumeHandle = echoResumeHandle(request.ResumeHandle, uint32(from+len(sent)))
	} else {
		response.ResumeHandle = echoResumeHandle(request.ResumeHandle, 0)
	}

	logger.Debugf("rpcpipe: srvsvc NetrShareEnum level %d reported %d of %d shares from %d",
		level, len(sent), len(remaining), from)
	return ndr.Marshal(response)
}

// shareEnumStruct builds the level's container around a set of entries.
func shareEnumStruct(level uint32, entries []ShareEntry) mssrvs.SHARE_ENUM_STRUCT {
	info := mssrvs.SHARE_ENUM_UNION{Tag: ndr.DWORD(level)}

	switch level {
	case 0:
		buffer := make([]mssrvs.SHARE_INFO_0, 0, len(entries))
		for _, entry := range entries {
			buffer = append(buffer, mssrvs.SHARE_INFO_0{Shi0Netname: ndr.WSTR(entry.Name)})
		}
		info.Level0 = &mssrvs.SHARE_INFO_0_CONTAINER{
			EntriesRead: ndr.DWORD(len(buffer)),
			Buffer:      buffer,
		}
	case 1:
		buffer := make([]mssrvs.SHARE_INFO_1, 0, len(entries))
		for _, entry := range entries {
			buffer = append(buffer, mssrvs.SHARE_INFO_1{
				Shi1Netname: ndr.WSTR(entry.Name),
				Shi1Type:    ndr.DWORD(entry.Type),
				Shi1Remark:  ndr.WSTR(entry.Comment),
			})
		}
		info.Level1 = &mssrvs.SHARE_INFO_1_CONTAINER{
			EntriesRead: ndr.DWORD(len(buffer)),
			Buffer:      buffer,
		}
	}

	return mssrvs.SHARE_ENUM_STRUCT{Level: ndr.DWORD(level), ShareInfo: info}
}

// fitEntries returns the longest prefix of entries whose reported size stays
// within budget, and whether anything was left out.
//
// At least one entry is always returned when there is one to return: a budget
// too small for a single share would otherwise make the enumeration unable to
// advance, and the client would ask again for the same nothing forever.
func fitEntries(entries []ShareEntry, level uint32, budget uint32) ([]ShareEntry, bool) {
	used := uint32(0)
	for i, entry := range entries {
		size := entrySize(entry, level)
		if i > 0 && used+size > budget {
			return entries[:i], true
		}
		used += size
	}
	return entries, false
}

// entrySize is the number of bytes an entry contributes to an enumeration.
//
// It counts the entry's own NDR representation: the fixed part of the level's
// structure, plus the conformant-varying wide strings its members point at. The
// container's own header is not counted, since the budget the client sends
// describes the entries it wants.
func entrySize(entry ShareEntry, level uint32) uint32 {
	switch level {
	case 0:
		// A referent identifier for the name.
		return 4 + wideStringSize(entry.Name)
	default:
		// Referent identifiers for the name and the remark, and the type between
		// them.
		return 12 + wideStringSize(entry.Name) + wideStringSize(entry.Comment)
	}
}

// wideStringSize is the wire size of a [string] wchar_t* referent under NDR20:
// the maximum count, offset and actual count, then the characters and the
// terminator, padded to the next four-byte boundary.
func wideStringSize(value string) uint32 {
	characters := uint32(len([]rune(value)) + 1)
	body := 2 * characters
	if remainder := body % 4; remainder != 0 {
		body += 4 - remainder
	}
	return 12 + body
}

// echoResumeHandle returns a resume handle carrying value when the client sent
// one, and nil when it did not.
//
// A client that passed no handle is not asking to resume, and answering with one
// would be describing a continuation it never requested.
func echoResumeHandle(requested *ndr.DWORD, value uint32) *ndr.DWORD {
	if requested == nil {
		return nil
	}
	handle := ndr.DWORD(value)
	return &handle
}
