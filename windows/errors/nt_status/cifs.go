package nt_status

// CIFS-specific NTSTATUS values.
//
// [MS-CIFS] extends the NTSTATUS set defined in [MS-ERREF] section 2.3 with
// 32-bit CIFS-specific error codes. Each one is wire-identical to the
// equivalent SMBSTATUS ErrorClass/ErrorCode pair, so the same 32-bit field can
// be read either way depending on what the client negotiated:
//
//	SMBSTATUS { UCHAR ErrorClass; UCHAR Reserved; USHORT ErrorCode; }
//
// The field is little-endian, so the value is ErrorClass | (Reserved << 8) |
// (ErrorCode << 16) — for example STATUS_SMB_BAD_TID is the ERRSRV class (0x02)
// with error code ERRinvtid (0x0005), giving 0x00050002.
//
// These are the values a server needs in order to answer a request at all, and
// all but one of them is absent from the [MS-ERREF] table. The exception is
// NT_STATUS_INVALID_SMB, which shares 0x00010002 with [MS-ERREF]'s
// NT_STATUS_DBG_CONTINUE; lookup documents which name wins and why.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/d3b37bec-a9da-460c-89b0-8a8e83e93534
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8f11e0f3-d545-46cc-97e6-f00569e3e1bc
const (
	// ERRDOS class (0x01)

	// NT_STATUS_SMB_BAD_FID is ERRDOS/ERRbadfid: the FID is invalid.
	NT_STATUS_SMB_BAD_FID NT_STATUS = 0x00060001
	// NT_STATUS_OS2_INVALID_ACCESS is ERRDOS/ERRbadaccess: invalid open mode.
	NT_STATUS_OS2_INVALID_ACCESS NT_STATUS = 0x000C0001
	// NT_STATUS_OS2_NO_MORE_SIDS is ERRDOS/ERROR_NO_MORE_SEARCH_HANDLES: the
	// maximum number of searches has been exhausted.
	NT_STATUS_OS2_NO_MORE_SIDS NT_STATUS = 0x00710001
	// NT_STATUS_OS2_INVALID_LEVEL is ERRDOS/ERRunknownlevel: invalid
	// information level.
	NT_STATUS_OS2_INVALID_LEVEL NT_STATUS = 0x007C0001
	// NT_STATUS_OS2_NEGATIVE_SEEK is ERRDOS/ERRinvalidseek: a seek to a
	// negative absolute offset was attempted.
	NT_STATUS_OS2_NEGATIVE_SEEK NT_STATUS = 0x00830001
	// NT_STATUS_OS2_CANCEL_VIOLATION is ERRDOS/ERROR_CANCEL_VIOLATION: no lock
	// request was outstanding for the supplied cancel region.
	NT_STATUS_OS2_CANCEL_VIOLATION NT_STATUS = 0x00AD0001
	// NT_STATUS_OS2_ATOMIC_LOCKS_NOT_SUPPORTED is
	// ERRDOS/ERROR_ATOMIC_LOCKS_NOT_SUPPORTED: the file system does not support
	// atomic changes to the lock type.
	NT_STATUS_OS2_ATOMIC_LOCKS_NOT_SUPPORTED NT_STATUS = 0x00AE0001
	// NT_STATUS_OS2_CANNOT_COPY is ERRDOS/ERROR_CANNOT_COPY: the copy functions
	// cannot be used.
	NT_STATUS_OS2_CANNOT_COPY NT_STATUS = 0x010A0001
	// NT_STATUS_OS2_EAS_DIDNT_FIT is ERRDOS/ERROR_EAS_DIDNT_FIT: the available
	// extended attributes did not fit into the response.
	NT_STATUS_OS2_EAS_DIDNT_FIT NT_STATUS = 0x01130001
	// NT_STATUS_OS2_EA_ACCESS_DENIED is ERRDOS/ERROR_EA_ACCESS_DENIED: access
	// to the extended attribute was denied.
	NT_STATUS_OS2_EA_ACCESS_DENIED NT_STATUS = 0x03E20001

	// ERRSRV class (0x02)

	// NT_STATUS_INVALID_SMB is ERRSRV/ERRerror: an unspecified server error,
	// used for a message the server cannot interpret as a valid SMB.
	NT_STATUS_INVALID_SMB NT_STATUS = 0x00010002
	// NT_STATUS_SMB_BAD_TID is ERRSRV/ERRinvtid: the TID in the request is not
	// a valid tree connect on this session. Earlier documentation calls this
	// error code ERRinvnid.
	NT_STATUS_SMB_BAD_TID NT_STATUS = 0x00050002
	// NT_STATUS_SMB_BAD_COMMAND is ERRSRV/ERRbadcmd: an unknown SMB command
	// code was received.
	NT_STATUS_SMB_BAD_COMMAND NT_STATUS = 0x00160002
	// NT_STATUS_SMB_BAD_UID is ERRSRV/ERRbaduid: the UID in the request is not
	// a valid session on this connection.
	NT_STATUS_SMB_BAD_UID NT_STATUS = 0x005B0002
	// NT_STATUS_SMB_USE_MPX is ERRSRV/ERRusempx: raw-mode transfers are
	// temporarily unavailable, use MPX mode.
	NT_STATUS_SMB_USE_MPX NT_STATUS = 0x00FA0002
	// NT_STATUS_SMB_USE_STANDARD is ERRSRV/ERRusestd: raw and MPX mode
	// transfers are temporarily unavailable, use standard read/write.
	NT_STATUS_SMB_USE_STANDARD NT_STATUS = 0x00FB0002
	// NT_STATUS_SMB_CONTINUE_MPX is ERRSRV/ERRcontmpx: continue in MPX mode.
	// Reserved for future use.
	NT_STATUS_SMB_CONTINUE_MPX NT_STATUS = 0x00FC0002
	// NT_STATUS_SMB_NO_SUPPORT is ERRSRV/ERRnosupport: the function is not
	// supported by the server.
	NT_STATUS_SMB_NO_SUPPORT NT_STATUS = 0xFFFF0002
)

// cifsTable is the [MS-CIFS] extension to the [MS-ERREF] table, consulted first
// by lookup so these values render under their SMB names. The descriptions are
// the ones [MS-CIFS] gives for the equivalent ErrorClass/ErrorCode pair.
var cifsTable = map[NT_STATUS]Entry{
	NT_STATUS_SMB_BAD_FID: {
		Name:        "NT_STATUS_SMB_BAD_FID",
		Description: "Invalid FID.",
	},
	NT_STATUS_OS2_INVALID_ACCESS: {
		Name:        "NT_STATUS_OS2_INVALID_ACCESS",
		Description: "Invalid open mode.",
	},
	NT_STATUS_OS2_NO_MORE_SIDS: {
		Name:        "NT_STATUS_OS2_NO_MORE_SIDS",
		Description: "Maximum number of searches has been exhausted.",
	},
	NT_STATUS_OS2_INVALID_LEVEL: {
		Name:        "NT_STATUS_OS2_INVALID_LEVEL",
		Description: "Invalid information level.",
	},
	NT_STATUS_OS2_NEGATIVE_SEEK: {
		Name:        "NT_STATUS_OS2_NEGATIVE_SEEK",
		Description: "An attempt was made to seek to a negative absolute offset within a file.",
	},
	NT_STATUS_OS2_CANCEL_VIOLATION: {
		Name:        "NT_STATUS_OS2_CANCEL_VIOLATION",
		Description: "No lock request was outstanding for the supplied cancel region.",
	},
	NT_STATUS_OS2_ATOMIC_LOCKS_NOT_SUPPORTED: {
		Name:        "NT_STATUS_OS2_ATOMIC_LOCKS_NOT_SUPPORTED",
		Description: "The file system does not support atomic changes to the lock type.",
	},
	NT_STATUS_OS2_CANNOT_COPY: {
		Name:        "NT_STATUS_OS2_CANNOT_COPY",
		Description: "The copy functions cannot be used.",
	},
	NT_STATUS_OS2_EAS_DIDNT_FIT: {
		Name:        "NT_STATUS_OS2_EAS_DIDNT_FIT",
		Description: "Either there are no extended attributes, or the available extended attributes did not fit into the response.",
	},
	NT_STATUS_OS2_EA_ACCESS_DENIED: {
		Name:        "NT_STATUS_OS2_EA_ACCESS_DENIED",
		Description: "Access to the extended attribute was denied.",
	},
	NT_STATUS_INVALID_SMB: {
		Name:        "NT_STATUS_INVALID_SMB",
		Description: "Unspecified server error.",
	},
	NT_STATUS_SMB_BAD_TID: {
		Name:        "NT_STATUS_SMB_BAD_TID",
		Description: "The TID specified in the command was invalid.",
	},
	NT_STATUS_SMB_BAD_COMMAND: {
		Name:        "NT_STATUS_SMB_BAD_COMMAND",
		Description: "An unknown SMB command code was received by the server.",
	},
	NT_STATUS_SMB_BAD_UID: {
		Name:        "NT_STATUS_SMB_BAD_UID",
		Description: "The UID specified is not known as a valid ID on this server session.",
	},
	NT_STATUS_SMB_USE_MPX: {
		Name:        "NT_STATUS_SMB_USE_MPX",
		Description: "Temporarily unable to support Raw mode transfers. Use MPX mode.",
	},
	NT_STATUS_SMB_USE_STANDARD: {
		Name:        "NT_STATUS_SMB_USE_STANDARD",
		Description: "Temporarily unable to support Raw or MPX mode transfers. Use standard read/write.",
	},
	NT_STATUS_SMB_CONTINUE_MPX: {
		Name:        "NT_STATUS_SMB_CONTINUE_MPX",
		Description: "Continue in MPX mode. This error code is reserved for future use.",
	},
	NT_STATUS_SMB_NO_SUPPORT: {
		Name:        "NT_STATUS_SMB_NO_SUPPORT",
		Description: "Function not supported by the server.",
	},
}

// cifsNameToCode resolves the CIFS-specific names, consulted first by FromName
// for the same reason cifsTable is consulted first by lookup.
var cifsNameToCode = func() map[string]NT_STATUS {
	names := make(map[string]NT_STATUS, len(cifsTable))
	for status, entry := range cifsTable {
		names[entry.Name] = status
	}
	return names
}()
