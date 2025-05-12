package ntlm

// NTLM Negotiate Flags
// During NTLM authentication, each of the following flags is a possible value of
// the NegotiateFlags field of the NEGOTIATE_MESSAGE, CHALLENGE_MESSAGE, and
// AUTHENTICATE_MESSAGE, unless otherwise noted. These flags define client or server
// NTLM capabilities supported by the sender.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-nlmp/99d90ff4-957f-4c8a-80e4-5bfe5a9a9832
const (
	// A (1 bit): If set, requests Unicode character set encoding. An alternate name for this field is NTLMSSP_NEGOTIATE_UNICODE.
	// The A and B bits are evaluated together as follows:
	// A==1: The choice of character set encoding MUST be Unicode.
	// A==0 and B==1: The choice of character set encoding MUST be OEM.
	// A==0 and B==0: The protocol MUST return SEC_E_INVALID_TOKEN.
	NTLMSSP_NEGOTIATE_UNICODE uint32 = 0x00000001

	// B (1 bit): If set, requests OEM character set encoding. An alternate name for this field is NTLM_NEGOTIATE_OEM. See bit A for details.
	NTLMSSP_NEGOTIATE_OEM uint32 = 0x00000002

	// C (1 bit): If set, a TargetName field of the CHALLENGE_MESSAGE (section 2.2.1.2) MUST be supplied. An alternate name for this field is NTLMSSP_REQUEST_TARGET.
	NTLMSSP_REQUEST_TARGET uint32 = 0x00000004

	// r10 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R10 uint32 = 0x00000008

	// D (1 bit): If set, requests session key negotiation for message signatures. If the client sends NTLMSSP_NEGOTIATE_SIGN to the server in the NEGOTIATE_MESSAGE, the server MUST return NTLMSSP_NEGOTIATE_SIGN to the client in the CHALLENGE_MESSAGE. An alternate name for this field is NTLMSSP_NEGOTIATE_SIGN.
	NTLMSSP_NEGOTIATE_SIGN uint32 = 0x00000010

	// E (1 bit): If set, requests session key negotiation for message confidentiality. If the client sends NTLMSSP_NEGOTIATE_SEAL to the server in the NEGOTIATE_MESSAGE, the server MUST return NTLMSSP_NEGOTIATE_SEAL to the client in the CHALLENGE_MESSAGE. Clients and servers that set NTLMSSP_NEGOTIATE_SEAL SHOULD always set NTLMSSP_NEGOTIATE_56 and NTLMSSP_NEGOTIATE_128, if they are supported. An alternate name for this field is NTLMSSP_NEGOTIATE_SEAL.
	NTLMSSP_NEGOTIATE_SEAL uint32 = 0x00000020

	// F (1 bit): If set, requests connectionless authentication. If NTLMSSP_NEGOTIATE_DATAGRAM is set, then NTLMSSP_NEGOTIATE_KEY_EXCH MUST always be set in the AUTHENTICATE_MESSAGE to the server and the CHALLENGE_MESSAGE to the client. An alternate name for this field is NTLMSSP_NEGOTIATE_DATAGRAM.
	NTLMSSP_NEGOTIATE_DATAGRAM uint32 = 0x00000040

	// G (1 bit): If set, requests LAN Manager (LM) session key computation. NTLMSSP_NEGOTIATE_LM_KEY and NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY are mutually exclusive. If both NTLMSSP_NEGOTIATE_LM_KEY and NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY are requested, NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY alone MUST be returned to the client. NTLM v2 authentication session key generation MUST be supported by both the client and the DC in order to be used, and extended session security signing and sealing requires support from the client and the server to be used. An alternate name for this field is NTLMSSP_NEGOTIATE_LM_KEY.
	NTLMSSP_NEGOTIATE_LM_KEY uint32 = 0x00000080

	// r9 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R9 uint32 = 0x00000100

	// H (1 bit): If set, requests usage of the NTLM v1 session security protocol. NTLMSSP_NEGOTIATE_NTLM MUST be set in the NEGOTIATE_MESSAGE to the server and the CHALLENGE_MESSAGE to the client. An alternate name for this field is NTLMSSP_NEGOTIATE_NTLM.
	NTLMSSP_NEGOTIATE_NTLM uint32 = 0x00000200

	// r8 (1 bit): This bit is unused and SHOULD be zero.
	NTLMSSP_REQUEST_R8 uint32 = 0x00000400

	// J (1 bit): If set, the connection SHOULD be anonymous.
	NTLMSSP_NEGOTIATE_ANONYMOUS uint32 = 0x00000800

	// K (1 bit): If set, the domain name is provided (section 2.2.1.1). An alternate name for this field is NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED.
	NTLMSSP_NEGOTIATE_OEM_DOMAIN_SUPPLIED uint32 = 0x00001000

	// L (1 bit):  This flag indicates whether the Workstation field is present. If this flag is not set, the Workstation field MUST be ignored. If this flag is set, the length of the Workstation field specifies whether the workstation name is nonempty or not. An alternate name for this field is NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED.
	NTLMSSP_NEGOTIATE_OEM_WORKSTATION_SUPPLIED uint32 = 0x00002000

	// r7 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R7 uint32 = 0x00004000

	// M (1 bit): If set, a session key is generated regardless of the states of NTLMSSP_NEGOTIATE_SIGN and NTLMSSP_NEGOTIATE_SEAL. A session key MUST always exist to generate the MIC (section 3.1.5.1.2) in the authenticate message. NTLMSSP_NEGOTIATE_ALWAYS_SIGN MUST be set in the NEGOTIATE_MESSAGE to the server and the CHALLENGE_MESSAGE to the client. NTLMSSP_NEGOTIATE_ALWAYS_SIGN is overridden by NTLMSSP_NEGOTIATE_SIGN and NTLMSSP_NEGOTIATE_SEAL, if they are supported. An alternate name for this field is NTLMSSP_NEGOTIATE_ALWAYS_SIGN.
	NTLMSSP_NEGOTIATE_ALWAYS_SIGN uint32 = 0x00008000

	// N (1 bit): If set, TargetName MUST be a domain name. The data corresponding to this flag is provided by the server in the TargetName field of the CHALLENGE_MESSAGE. If set, then NTLMSSP_TARGET_TYPE_SERVER MUST NOT be set. This flag MUST be ignored in the NEGOTIATE_MESSAGE and the AUTHENTICATE_MESSAGE. An alternate name for this field is NTLMSSP_TARGET_TYPE_DOMAIN.
	NTLMSSP_TARGET_TYPE_DOMAIN uint32 = 0x00010000

	// O (1 bit): If set, TargetName MUST be a server name. The data corresponding to this flag is provided by the server in the TargetName field of the CHALLENGE_MESSAGE. If this bit is set, then NTLMSSP_TARGET_TYPE_DOMAIN MUST NOT be set. This flag MUST be ignored in the NEGOTIATE_MESSAGE and the AUTHENTICATE_MESSAGE. An alternate name for this field is NTLMSSP_TARGET_TYPE_SERVER.
	NTLMSSP_TARGET_TYPE_SERVER uint32 = 0x00020000

	// r6 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R6 uint32 = 0x00040000

	// P (1 bit): If set, requests usage of the NTLM v2 session security. NTLM v2 session security is a misnomer because it is not NTLM v2. It is NTLM v1 using the extended session security that is also in NTLM v2. NTLMSSP_NEGOTIATE_LM_KEY and NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY are mutually exclusive. If both NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY and NTLMSSP_NEGOTIATE_LM_KEY are requested, NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY alone MUST be returned to the client. NTLM v2 authentication session key generation MUST be supported by both the client and the DC in order to be used, and extended session security signing and sealing requires support from the client and the server in order to be used. An alternate name for this field is NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY.
	NTLMSSP_NEGOTIATE_EXTENDED_SESSIONSECURITY uint32 = 0x00080000

	// Q (1 bit): If set, requests an identify level token. An alternate name for this field is NTLMSSP_NEGOTIATE_IDENTIFY.
	NTLMSSP_NEGOTIATE_IDENTIFY uint32 = 0x00100000

	// r5 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R5 uint32 = 0x00200000

	// R (1 bit): If set, requests the usage of the LMOWF. An alternate name for this field is NTLMSSP_REQUEST_NON_NT_SESSION_KEY.
	NTLMSSP_REQUEST_NON_NT_SESSION_KEY uint32 = 0x00400000

	// S (1 bit): If set, indicates that the TargetInfo fields in the CHALLENGE_MESSAGE (section 2.2.1.2) are populated. An alternate name for this field is NTLMSSP_NEGOTIATE_TARGET_INFO.
	NTLMSSP_NEGOTIATE_TARGET_INFO uint32 = 0x00800000

	// r4 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R4 uint32 = 0x01000000

	// T (1 bit): If set, requests the protocol version number. The data corresponding to this flag is provided in the Version field of the NEGOTIATE_MESSAGE, the CHALLENGE_MESSAGE, and the AUTHENTICATE_MESSAGE. An alternate name for this field is NTLMSSP_NEGOTIATE_VERSION.
	NTLMSSP_NEGOTIATE_VERSION uint32 = 0x02000000

	// r1 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R1 uint32 = 0x04000000

	// r2 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R2 uint32 = 0x08000000

	// r3 (1 bit): This bit is unused and MUST be zero.
	NTLMSSP_REQUEST_R3 uint32 = 0x10000000

	// U (1 bit): If set, requests 128-bit session key negotiation. An alternate name for this field is NTLMSSP_NEGOTIATE_128. If the client sends NTLMSSP_NEGOTIATE_128 to the server in the NEGOTIATE_MESSAGE, the server MUST return NTLMSSP_NEGOTIATE_128 to the client in the CHALLENGE_MESSAGE only if the client sets NTLMSSP_NEGOTIATE_SEAL or NTLMSSP_NEGOTIATE_SIGN. Otherwise it is ignored. If both NTLMSSP_NEGOTIATE_56 and NTLMSSP_NEGOTIATE_128 are requested and supported by the client and server, NTLMSSP_NEGOTIATE_56 and NTLMSSP_NEGOTIATE_128 will both be returned to the client. Clients and servers that set NTLMSSP_NEGOTIATE_SEAL SHOULD set NTLMSSP_NEGOTIATE_128 if it is supported. An alternate name for this field is NTLMSSP_NEGOTIATE_128.
	NTLMSSP_NEGOTIATE_128 uint32 = 0x20000000

	// V (1 bit): If set, requests an explicit key exchange. This capability SHOULD be used because it improves security for message integrity or confidentiality. See sections 3.2.5.1.2, 3.2.5.2.1, and 3.2.5.2.2 for details. An alternate name for this field is NTLMSSP_NEGOTIATE_KEY_EXCH.
	NTLMSSP_NEGOTIATE_KEY_EXCH uint32 = 0x40000000

	// W (1 bit): If set, requests 56-bit encryption. If the client sends NTLMSSP_NEGOTIATE_SEAL or NTLMSSP_NEGOTIATE_SIGN with NTLMSSP_NEGOTIATE_56 to the server in the NEGOTIATE_MESSAGE, the server MUST return NTLMSSP_NEGOTIATE_56 to the client in the CHALLENGE_MESSAGE.   Otherwise it is ignored. If both NTLMSSP_NEGOTIATE_56 and NTLMSSP_NEGOTIATE_128 are requested and supported by the client and server, NTLMSSP_NEGOTIATE_56 and NTLMSSP_NEGOTIATE_128 will both be returned to the client. Clients and servers that set NTLMSSP_NEGOTIATE_SEAL SHOULD set NTLMSSP_NEGOTIATE_56 if it is supported. An alternate name for this field is NTLMSSP_NEGOTIATE_56.
	NTLMSSP_NEGOTIATE_56 uint32 = 0x80000000
)
