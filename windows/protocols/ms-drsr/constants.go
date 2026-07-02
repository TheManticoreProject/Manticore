package msdrsr

// DS_NAME_FORMAT selects how a name is expressed in IDL_DRSCrackNames'
// formatOffered/formatDesired ([MS-DRSR] 4.1.4.1.3). DCSync resolves an account name by
// offering DS_NT4_ACCOUNT_NAME (or DS_USER_PRINCIPAL_NAME / DS_SID_OR_SID_HISTORY_NAME)
// and requesting DS_UNIQUE_ID_NAME, whose result is the objectGUID in "{...}" form.
const (
	DS_UNKNOWN_NAME            uint32 = 0
	DS_FQDN_1779_NAME          uint32 = 1
	DS_NT4_ACCOUNT_NAME        uint32 = 2
	DS_DISPLAY_NAME            uint32 = 3
	DS_UNIQUE_ID_NAME          uint32 = 6
	DS_CANONICAL_NAME          uint32 = 7
	DS_USER_PRINCIPAL_NAME     uint32 = 8
	DS_CANONICAL_NAME_EX       uint32 = 9
	DS_SERVICE_PRINCIPAL_NAME  uint32 = 10
	DS_SID_OR_SID_HISTORY_NAME uint32 = 11
	DS_DNS_DOMAIN_NAME         uint32 = 12
)

// DS_NAME_ERROR is the per-item status in DS_NAME_RESULT_ITEMW.Status
// ([MS-DRSR] 4.1.4.1.2). DS_NAME_NO_ERROR (0) means the item resolved.
const (
	DS_NAME_NO_ERROR                     uint32 = 0
	DS_NAME_ERROR_RESOLVING              uint32 = 1
	DS_NAME_ERROR_NOT_FOUND              uint32 = 2
	DS_NAME_ERROR_NOT_UNIQUE             uint32 = 3
	DS_NAME_ERROR_NO_MAPPING             uint32 = 4
	DS_NAME_ERROR_DOMAIN_ONLY            uint32 = 5
	DS_NAME_ERROR_NO_SYNTACTICAL_MAPPING uint32 = 6
	DS_NAME_ERROR_TRUST_REFERRAL         uint32 = 7
)

// EXOP is the extended operation in DRS_MSG_GETCHGREQ_V8.UlExtendedOp ([MS-DRSR] 5.40,
// "EXOP_REQ Codes"). EXOP_REPL_OBJ replicates a single object by GUID (the DCSync path);
// EXOP_REPL_SECRETS additionally forces secret values even without recent changes.
const (
	EXOP_FSMO_REQ_ROLE      uint32 = 0x00000001
	EXOP_FSMO_REQ_RID_ALLOC uint32 = 0x00000002
	EXOP_FSMO_RID_REQ_ROLE  uint32 = 0x00000003
	EXOP_FSMO_REQ_PDC       uint32 = 0x00000004
	EXOP_FSMO_ABANDON_ROLE  uint32 = 0x00000005
	EXOP_REPL_OBJ           uint32 = 0x00000006
	EXOP_REPL_SECRETS       uint32 = 0x00000007
)

// DS_REPL_INFO_TYPE selects which replication information IDL_DRSGetReplInfo returns
// ([MS-DRSR] 4.1.13.3); it is both the request InfoType and the reply union discriminant.
const (
	DS_REPL_INFO_NEIGHBORS                 uint32 = 0
	DS_REPL_INFO_CURSORS_FOR_NC            uint32 = 1
	DS_REPL_INFO_METADATA_FOR_OBJ          uint32 = 2
	DS_REPL_INFO_KCC_DSA_CONNECT_FAILURES  uint32 = 3
	DS_REPL_INFO_KCC_DSA_LINK_FAILURES     uint32 = 4
	DS_REPL_INFO_PENDING_OPS               uint32 = 5
	DS_REPL_INFO_METADATA_FOR_ATTR_VALUE   uint32 = 6
	DS_REPL_INFO_CURSORS_2_FOR_NC          uint32 = 7
	DS_REPL_INFO_CURSORS_3_FOR_NC          uint32 = 8
	DS_REPL_INFO_METADATA_2_FOR_OBJ        uint32 = 9
	DS_REPL_INFO_METADATA_2_FOR_ATTR_VALUE uint32 = 10
)

// DRS_OPTIONS are the ulFlags bits of IDL_DRSGetNCChanges and related calls
// ([MS-DRSR] 5.41). A single-object replication uses DRS_INIT_SYNC | DRS_WRIT_REP.
const (
	DRS_ASYNC_OP                  uint32 = 0x00000001
	DRS_WRIT_REP                  uint32 = 0x00000010
	DRS_INIT_SYNC                 uint32 = 0x00000020
	DRS_PER_SYNC                  uint32 = 0x00000040
	DRS_GET_ANC                   uint32 = 0x00000800
	DRS_GET_NC_SIZE               uint32 = 0x00001000
	DRS_FULL_SYNC_NOW             uint32 = 0x00008000
	DRS_SYNC_URGENT               uint32 = 0x00080000
	DRS_NEVER_SYNCED              uint32 = 0x00200000
	DRS_SPECIAL_SECRET_PROCESSING uint32 = 0x00400000
)
