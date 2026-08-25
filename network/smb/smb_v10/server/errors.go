package server

import (
	"github.com/TheManticoreProject/Manticore/windows/nt_status"
)

// ErrorClass is the SMB error class carried in the low byte of the Status field
// when the client has not negotiated SMB_FLAGS2_NT_STATUS_ERROR_CODES.
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8f11e0f3-d545-46cc-97e6-f00569e3e1bc
type ErrorClass uint8

const (
	// ERRSUCCESS indicates the command completed with no error.
	ERRSUCCESS ErrorClass = 0x00
	// ERRDOS is the OS/2 (MS-DOS) error class.
	ERRDOS ErrorClass = 0x01
	// ERRSRV is the server error class, used for errors in the SMB protocol
	// itself rather than in the underlying file system operation.
	ERRSRV ErrorClass = 0x02
	// ERRHRD is the hardware error class.
	ERRHRD ErrorClass = 0x03
	// ERRCMD indicates the server received a message that was not in SMB
	// format. No error codes are defined for use with this class.
	ERRCMD ErrorClass = 0xFF
)

// String renders the error class by its [MS-CIFS] name.
func (c ErrorClass) String() string {
	switch c {
	case ERRSUCCESS:
		return "SUCCESS"
	case ERRDOS:
		return "ERRDOS"
	case ERRSRV:
		return "ERRSRV"
	case ERRHRD:
		return "ERRHRD"
	case ERRCMD:
		return "ERRCMD"
	}
	return "UNKNOWN"
}

// SMBStatus is the legacy SMBSTATUS form of an error: an error class paired with
// a class-scoped error code.
//
//	SMBSTATUS { UCHAR ErrorClass; UCHAR Reserved; USHORT ErrorCode; }
type SMBStatus struct {
	Class ErrorClass
	Code  uint16
}

// Encode renders the pair into the 32-bit Status header field. The field is
// little-endian and laid out ErrorClass(1) | Reserved(1) | ErrorCode(2), so the
// class occupies the low byte and the code the high half-word.
func (s SMBStatus) Encode() uint32 {
	return uint32(s.Class) | (uint32(s.Code) << 16)
}

// dosErrors maps an NTSTATUS onto the SMBSTATUS class/code pair a server must
// send when the client did not negotiate NT status codes.
//
// The mapping is the one tabulated in [MS-CIFS] 2.2.2.4. That table is written
// in the other direction — one SMBSTATUS pair to the several NTSTATUS values
// that reduce to it — so a few NTSTATUS values legitimately appear against more
// than one pair there; where that happens the entry below takes the pair whose
// description matches the condition a server is actually reporting (for example
// STATUS_DIRECTORY_NOT_EMPTY is listed under both ERRDOS/ERRnoaccess and
// ERRDOS/ERRremcd, and ERRremcd is the one that names the condition).
//
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/8f11e0f3-d545-46cc-97e6-f00569e3e1bc
var dosErrors = map[nt_status.NT_STATUS]SMBStatus{
	nt_status.NT_STATUS_SUCCESS: {ERRSUCCESS, 0x0000},

	// ERRDOS class.
	nt_status.NT_STATUS_NOT_IMPLEMENTED:          {ERRDOS, 0x0001}, // ERRbadfunc
	nt_status.NT_STATUS_INVALID_DEVICE_REQUEST:   {ERRDOS, 0x0001}, // ERRbadfunc
	nt_status.NT_STATUS_NO_SUCH_FILE:             {ERRDOS, 0x0002}, // ERRbadfile
	nt_status.NT_STATUS_NO_SUCH_DEVICE:           {ERRDOS, 0x0002}, // ERRbadfile
	nt_status.NT_STATUS_OBJECT_NAME_NOT_FOUND:    {ERRDOS, 0x0002}, // ERRbadfile
	nt_status.NT_STATUS_OBJECT_PATH_INVALID:      {ERRDOS, 0x0003}, // ERRbadpath
	nt_status.NT_STATUS_OBJECT_PATH_NOT_FOUND:    {ERRDOS, 0x0003}, // ERRbadpath
	nt_status.NT_STATUS_OBJECT_PATH_SYNTAX_BAD:   {ERRDOS, 0x0003}, // ERRbadpath
	nt_status.NT_STATUS_TOO_MANY_OPENED_FILES:    {ERRDOS, 0x0004}, // ERRnofids
	nt_status.NT_STATUS_ACCESS_DENIED:            {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_DELETE_PENDING:           {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_PRIVILEGE_NOT_HELD:       {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_LOGON_FAILURE:            {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_FILE_IS_A_DIRECTORY:      {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_CANNOT_DELETE:            {ERRDOS, 0x0005}, // ERRnoaccess
	nt_status.NT_STATUS_INVALID_HANDLE:           {ERRDOS, 0x0006}, // ERRbadfid
	nt_status.NT_STATUS_OBJECT_TYPE_MISMATCH:     {ERRDOS, 0x0006}, // ERRbadfid
	nt_status.NT_STATUS_FILE_CLOSED:              {ERRDOS, 0x0006}, // ERRbadfid
	nt_status.NT_STATUS_INSUFF_SERVER_RESOURCES:  {ERRDOS, 0x0008}, // ERRnomem
	nt_status.NT_STATUS_DATA_ERROR:               {ERRDOS, 0x000D}, // ERRbaddata
	nt_status.NT_STATUS_DIRECTORY_NOT_EMPTY:      {ERRDOS, 0x0010}, // ERRremcd
	nt_status.NT_STATUS_NOT_SAME_DEVICE:          {ERRDOS, 0x0011}, // ERRdiffdevice
	nt_status.NT_STATUS_NO_MORE_FILES:            {ERRDOS, 0x0012}, // ERRnofiles
	nt_status.NT_STATUS_UNSUCCESSFUL:             {ERRDOS, 0x001F}, // ERRgeneral
	nt_status.NT_STATUS_SHARING_VIOLATION:        {ERRDOS, 0x0020}, // ERRbadshare
	nt_status.NT_STATUS_FILE_LOCK_CONFLICT:       {ERRDOS, 0x0021}, // ERRlock
	nt_status.NT_STATUS_LOCK_NOT_GRANTED:         {ERRDOS, 0x0021}, // ERRlock
	nt_status.NT_STATUS_END_OF_FILE:              {ERRDOS, 0x0026}, // ERReof
	nt_status.NT_STATUS_NOT_SUPPORTED:            {ERRDOS, 0x0032}, // ERRunsup
	nt_status.NT_STATUS_OBJECT_NAME_COLLISION:    {ERRDOS, 0x0050}, // ERRfilexists
	nt_status.NT_STATUS_INVALID_PARAMETER:        {ERRDOS, 0x0057}, // ERRinvalidparam
	nt_status.NT_STATUS_RANGE_NOT_LOCKED:         {ERRDOS, 0x009E}, // ERROR_NOT_LOCKED
	nt_status.NT_STATUS_INVALID_INFO_CLASS:       {ERRDOS, 0x00E6}, // ERRbadpipe
	nt_status.NT_STATUS_INVALID_PIPE_STATE:       {ERRDOS, 0x00E6}, // ERRbadpipe
	nt_status.NT_STATUS_PIPE_BUSY:                {ERRDOS, 0x00E7}, // ERRpipebusy
	nt_status.NT_STATUS_PIPE_CLOSING:             {ERRDOS, 0x00E8}, // ERRpipeclosing
	nt_status.NT_STATUS_PIPE_DISCONNECTED:        {ERRDOS, 0x00E9}, // ERRnotconnected
	nt_status.NT_STATUS_BUFFER_OVERFLOW:          {ERRDOS, 0x00EA}, // ERRmoredata
	nt_status.NT_STATUS_MORE_PROCESSING_REQUIRED: {ERRDOS, 0x00EA}, // ERRmoredata
	nt_status.NT_STATUS_EA_TOO_LARGE:             {ERRDOS, 0x0113}, // ERROR_EAS_DIDNT_FIT
	nt_status.NT_STATUS_EAS_NOT_SUPPORTED:        {ERRDOS, 0x011A}, // ERROR_EAS_NOT_SUPPORTED

	// ERRSRV class.
	nt_status.NT_STATUS_WRONG_PASSWORD:           {ERRSRV, 0x0002}, // ERRbadpw
	nt_status.NT_STATUS_PATH_NOT_COVERED:         {ERRSRV, 0x0003}, // ERRbadpath
	nt_status.NT_STATUS_NETWORK_ACCESS_DENIED:    {ERRSRV, 0x0004}, // ERRaccess
	nt_status.NT_STATUS_NETWORK_NAME_DELETED:     {ERRSRV, 0x0005}, // ERRinvtid
	nt_status.NT_STATUS_BAD_NETWORK_NAME:         {ERRSRV, 0x0006}, // ERRinvnetname
	nt_status.NT_STATUS_BAD_DEVICE_TYPE:          {ERRSRV, 0x0007}, // ERRinvdevice
	nt_status.NT_STATUS_PRINT_QUEUE_FULL:         {ERRSRV, 0x0031}, // ERRqfull
	nt_status.NT_STATUS_NO_SPOOL_SPACE:           {ERRSRV, 0x0032}, // ERRqtoobig
	nt_status.NT_STATUS_UNEXPECTED_NETWORK_ERROR: {ERRSRV, 0x0041}, // ERRsrverror
	nt_status.NT_STATUS_IO_TIMEOUT:               {ERRSRV, 0x0058}, // ERRtimeout
	nt_status.NT_STATUS_REQUEST_NOT_ACCEPTED:     {ERRSRV, 0x0059}, // ERRnoresource
	nt_status.NT_STATUS_TOO_MANY_SESSIONS:        {ERRSRV, 0x005A}, // ERRtoomanyuids
	nt_status.NT_STATUS_ACCOUNT_DISABLED:         {ERRSRV, 0x08BF}, // ERRaccountExpired
	nt_status.NT_STATUS_ACCOUNT_EXPIRED:          {ERRSRV, 0x08BF}, // ERRaccountExpired
	nt_status.NT_STATUS_INVALID_WORKSTATION:      {ERRSRV, 0x08C0}, // ERRbadClient
	nt_status.NT_STATUS_INVALID_LOGON_HOURS:      {ERRSRV, 0x08C1}, // ERRbadLogonTime
	nt_status.NT_STATUS_PASSWORD_EXPIRED:         {ERRSRV, 0x08C2}, // ERRpasswordExpired
	nt_status.NT_STATUS_PASSWORD_MUST_CHANGE:     {ERRSRV, 0x08C2}, // ERRpasswordExpired

	// ERRHRD class.
	nt_status.NT_STATUS_MEDIA_WRITE_PROTECTED: {ERRHRD, 0x0013}, // ERRnowrite
	nt_status.NT_STATUS_NO_MEDIA_IN_DEVICE:    {ERRHRD, 0x0015}, // ERRnotready
	nt_status.NT_STATUS_CRC_ERROR:             {ERRHRD, 0x0017}, // ERRdata
	nt_status.NT_STATUS_DISK_CORRUPT_ERROR:    {ERRHRD, 0x001A}, // ERRbadmedia
	nt_status.NT_STATUS_NONEXISTENT_SECTOR:    {ERRHRD, 0x001B}, // ERRbadsector
	nt_status.NT_STATUS_WRONG_VOLUME:          {ERRHRD, 0x0022}, // ERRwrongdisk
	nt_status.NT_STATUS_DISK_FULL:             {ERRHRD, 0x0027}, // ERRdiskfull
}

// unmappedError is the pair sent for an NTSTATUS with no tabulated SMBSTATUS
// equivalent. ERRSRV/ERRsrverror ("Internal server error") is the honest answer:
// the condition exists but cannot be expressed in the legacy encoding.
var unmappedError = SMBStatus{ERRSRV, 0x0041}

// DOSError returns the legacy SMBSTATUS class/code pair for an NTSTATUS, and
// whether the mapping was tabulated. An untabulated status returns
// ERRSRV/ERRsrverror with false, which is still a valid thing to send.
//
// Parameters:
//   - status: the NTSTATUS the server wants to report
//
// Returns:
//   - The SMBSTATUS pair to send
//   - Whether a tabulated mapping was found
func DOSError(status nt_status.NT_STATUS) (SMBStatus, bool) {
	if mapped, ok := dosErrors[status]; ok {
		return mapped, true
	}
	// A CIFS-specific NTSTATUS is already wire-identical to its SMBSTATUS pair,
	// so it can be decomposed rather than looked up.
	if class := ErrorClass(uint32(status) & 0xFF); class != ERRSUCCESS && uint32(status)>>24 == 0 {
		if code := uint16(uint32(status) >> 16); code != 0 {
			return SMBStatus{class, code}, true
		}
	}
	return unmappedError, false
}

// EncodeStatus renders an NTSTATUS into the 32-bit Status header field in
// whichever form the client selected. A client that set
// SMB_FLAGS2_NT_STATUS_ERROR_CODES in its request receives the NTSTATUS
// unchanged; any other client receives the legacy SMBSTATUS encoding.
//
// Parameters:
//   - status: the NTSTATUS the server wants to report
//   - ntStatusCodes: whether the client negotiated NT status codes
//
// Returns:
//   - The value to place in the response header's Status field
func EncodeStatus(status nt_status.NT_STATUS, ntStatusCodes bool) uint32 {
	if ntStatusCodes {
		return uint32(status)
	}
	if status == nt_status.NT_STATUS_SUCCESS {
		return 0
	}
	pair, _ := DOSError(status)
	return pair.Encode()
}

// statusName renders an NTSTATUS for a log line, by name where one is known and
// as a bare hex value otherwise.
func statusName(status nt_status.NT_STATUS) string {
	if name, ok := nt_status.NTStatusToStringName[status]; ok {
		return name
	}
	return "0x" + hex32(uint32(status))
}

// hex32 formats a 32-bit value as eight lowercase hex digits, without pulling in
// a format call on the per-request logging path.
func hex32(v uint32) string {
	const digits = "0123456789abcdef"
	out := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		out[i] = digits[v&0xF]
		v >>= 4
	}
	return string(out)
}
