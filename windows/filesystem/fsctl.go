package filesystem

// FSCTL control codes carried in an SMB2 IOCTL request (with the
// SMB2_0_IOCTL_IS_FSCTL flag) or issued locally via DeviceIoControl.
//
// [MS-FSCC] 2.3 FSCTL Structures:
// https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-fscc/9d34c7c6-c0f6-46b4-9c00-c8b78f3b3041
const (
	// FSCTL_PIPE_TRANSCEIVE writes a message to a named pipe and reads the reply in
	// a single round-trip — the basis of DCE/RPC over SMB2 (ncacn_np).
	FSCTL_PIPE_TRANSCEIVE uint32 = 0x0011C017
	// FSCTL_PIPE_WAIT waits for a named-pipe instance to become available.
	FSCTL_PIPE_WAIT uint32 = 0x00110018
	// FSCTL_PIPE_PEEK reads data from a named pipe without removing it.
	FSCTL_PIPE_PEEK uint32 = 0x0011400C

	// FSCTL_DFS_GET_REFERRALS requests the DFS referrals for a path.
	FSCTL_DFS_GET_REFERRALS    uint32 = 0x00060194
	FSCTL_DFS_GET_REFERRALS_EX uint32 = 0x000601B0

	// FSCTL_VALIDATE_NEGOTIATE_INFO confirms the negotiated dialect/capabilities
	// were not tampered with (SMB 3.x).
	FSCTL_VALIDATE_NEGOTIATE_INFO uint32 = 0x00140204

	// Server-side copy.
	FSCTL_SRV_REQUEST_RESUME_KEY  uint32 = 0x00140078
	FSCTL_SRV_COPYCHUNK           uint32 = 0x001440F2
	FSCTL_SRV_COPYCHUNK_WRITE     uint32 = 0x001480F2
	FSCTL_SRV_ENUMERATE_SNAPSHOTS uint32 = 0x00144064

	// Reparse points.
	FSCTL_GET_REPARSE_POINT uint32 = 0x000900A8
	FSCTL_SET_REPARSE_POINT uint32 = 0x000900A4
)
