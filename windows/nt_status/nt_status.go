package nt_status

import (
	"errors"
	"fmt"
)

type NT_STATUS uint32

// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-erref/596a1078-e883-4972-9bbc-49e60bebca55
const (
	NT_STATUS_SUCCESS                                                     NT_STATUS = 0x00000000
	NT_STATUS_WAIT_0                                                      NT_STATUS = 0x00000000
	NT_STATUS_WAIT_1                                                      NT_STATUS = 0x00000001
	NT_STATUS_WAIT_2                                                      NT_STATUS = 0x00000002
	NT_STATUS_WAIT_3                                                      NT_STATUS = 0x00000003
	NT_STATUS_WAIT_63                                                     NT_STATUS = 0x0000003F
	NT_STATUS_ABANDONED                                                   NT_STATUS = 0x00000080
	NT_STATUS_ABANDONED_WAIT_0                                            NT_STATUS = 0x00000080
	NT_STATUS_ABANDONED_WAIT_63                                           NT_STATUS = 0x000000BF
	NT_STATUS_USER_APC                                                    NT_STATUS = 0x000000C0
	NT_STATUS_ALERTED                                                     NT_STATUS = 0x00000101
	NT_STATUS_TIMEOUT                                                     NT_STATUS = 0x00000102
	NT_STATUS_PENDING                                                     NT_STATUS = 0x00000103
	NT_STATUS_REPARSE                                                     NT_STATUS = 0x00000104
	NT_STATUS_MORE_ENTRIES                                                NT_STATUS = 0x00000105
	NT_STATUS_NOT_ALL_ASSIGNED                                            NT_STATUS = 0x00000106
	NT_STATUS_SOME_NOT_MAPPED                                             NT_STATUS = 0x00000107
	NT_STATUS_OPLOCK_BREAK_IN_PROGRESS                                    NT_STATUS = 0x00000108
	NT_STATUS_VOLUME_MOUNTED                                              NT_STATUS = 0x00000109
	NT_STATUS_RXACT_COMMITTED                                             NT_STATUS = 0x0000010A
	NT_STATUS_NOTIFY_CLEANUP                                              NT_STATUS = 0x0000010B
	NT_STATUS_NOTIFY_ENUM_DIR                                             NT_STATUS = 0x0000010C
	NT_STATUS_NO_QUOTAS_FOR_ACCOUNT                                       NT_STATUS = 0x0000010D
	NT_STATUS_PRIMARY_TRANSPORT_CONNECT_FAILED                            NT_STATUS = 0x0000010E
	NT_STATUS_PAGE_FAULT_TRANSITION                                       NT_STATUS = 0x00000110
	NT_STATUS_PAGE_FAULT_DEMAND_ZERO                                      NT_STATUS = 0x00000111
	NT_STATUS_PAGE_FAULT_COPY_ON_WRITE                                    NT_STATUS = 0x00000112
	NT_STATUS_PAGE_FAULT_GUARD_PAGE                                       NT_STATUS = 0x00000113
	NT_STATUS_PAGE_FAULT_PAGING_FILE                                      NT_STATUS = 0x00000114
	NT_STATUS_CACHE_PAGE_LOCKED                                           NT_STATUS = 0x00000115
	NT_STATUS_CRASH_DUMP                                                  NT_STATUS = 0x00000116
	NT_STATUS_BUFFER_ALL_ZEROS                                            NT_STATUS = 0x00000117
	NT_STATUS_REPARSE_OBJECT                                              NT_STATUS = 0x00000118
	NT_STATUS_RESOURCE_REQUIREMENTS_CHANGED                               NT_STATUS = 0x00000119
	NT_STATUS_TRANSLATION_COMPLETE                                        NT_STATUS = 0x00000120
	NT_STATUS_DS_MEMBERSHIP_EVALUATED_LOCALLY                             NT_STATUS = 0x00000121
	NT_STATUS_NOTHING_TO_TERMINATE                                        NT_STATUS = 0x00000122
	NT_STATUS_PROCESS_NOT_IN_JOB                                          NT_STATUS = 0x00000123
	NT_STATUS_PROCESS_IN_JOB                                              NT_STATUS = 0x00000124
	NT_STATUS_VOLSNAP_HIBERNATE_READY                                     NT_STATUS = 0x00000125
	NT_STATUS_FSFILTER_OP_COMPLETED_SUCCESSFULLY                          NT_STATUS = 0x00000126
	NT_STATUS_INTERRUPT_VECTOR_ALREADY_CONNECTED                          NT_STATUS = 0x00000127
	NT_STATUS_INTERRUPT_STILL_CONNECTED                                   NT_STATUS = 0x00000128
	NT_STATUS_PROCESS_CLONED                                              NT_STATUS = 0x00000129
	NT_STATUS_FILE_LOCKED_WITH_ONLY_READERS                               NT_STATUS = 0x0000012A
	NT_STATUS_FILE_LOCKED_WITH_WRITERS                                    NT_STATUS = 0x0000012B
	NT_STATUS_RESOURCEMANAGER_READ_ONLY                                   NT_STATUS = 0x00000202
	NT_STATUS_WAIT_FOR_OPLOCK                                             NT_STATUS = 0x00000367
	NT_STATUS_DBG_EXCEPTION_HANDLED                                       NT_STATUS = 0x00010001
	NT_STATUS_DBG_CONTINUE                                                NT_STATUS = 0x00010002
	NT_STATUS_FLT_IO_COMPLETE                                             NT_STATUS = 0x001C0001
	NT_STATUS_FILE_NOT_AVAILABLE                                          NT_STATUS = 0xC0000467
	NT_STATUS_SHARE_UNAVAILABLE                                           NT_STATUS = 0xC0000480
	NT_STATUS_CALLBACK_RETURNED_THREAD_AFFINITY                           NT_STATUS = 0xC0000721
	NT_STATUS_OBJECT_NAME_EXISTS                                          NT_STATUS = 0x40000000
	NT_STATUS_THREAD_WAS_SUSPENDED                                        NT_STATUS = 0x40000001
	NT_STATUS_WORKING_SET_LIMIT_RANGE                                     NT_STATUS = 0x40000002
	NT_STATUS_IMAGE_NOT_AT_BASE                                           NT_STATUS = 0x40000003
	NT_STATUS_RXACT_STATE_CREATED                                         NT_STATUS = 0x40000004
	NT_STATUS_SEGMENT_NOTIFICATION                                        NT_STATUS = 0x40000005
	NT_STATUS_LOCAL_USER_SESSION_KEY                                      NT_STATUS = 0x40000006
	NT_STATUS_BAD_CURRENT_DIRECTORY                                       NT_STATUS = 0x40000007
	NT_STATUS_SERIAL_MORE_WRITES                                          NT_STATUS = 0x40000008
	NT_STATUS_REGISTRY_RECOVERED                                          NT_STATUS = 0x40000009
	NT_STATUS_FT_READ_RECOVERY_FROM_BACKUP                                NT_STATUS = 0x4000000A
	NT_STATUS_FT_WRITE_RECOVERY                                           NT_STATUS = 0x4000000B
	NT_STATUS_SERIAL_COUNTER_TIMEOUT                                      NT_STATUS = 0x4000000C
	NT_STATUS_NULL_LM_PASSWORD                                            NT_STATUS = 0x4000000D
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH                                 NT_STATUS = 0x4000000E
	NT_STATUS_RECEIVE_PARTIAL                                             NT_STATUS = 0x4000000F
	NT_STATUS_RECEIVE_EXPEDITED                                           NT_STATUS = 0x40000010
	NT_STATUS_RECEIVE_PARTIAL_EXPEDITED                                   NT_STATUS = 0x40000011
	NT_STATUS_EVENT_DONE                                                  NT_STATUS = 0x40000012
	NT_STATUS_EVENT_PENDING                                               NT_STATUS = 0x40000013
	NT_STATUS_CHECKING_FILE_SYSTEM                                        NT_STATUS = 0x40000014
	NT_STATUS_FATAL_APP_EXIT                                              NT_STATUS = 0x40000015
	NT_STATUS_PREDEFINED_HANDLE                                           NT_STATUS = 0x40000016
	NT_STATUS_WAS_UNLOCKED                                                NT_STATUS = 0x40000017
	NT_STATUS_SERVICE_NOTIFICATION                                        NT_STATUS = 0x40000018
	NT_STATUS_WAS_LOCKED                                                  NT_STATUS = 0x40000019
	NT_STATUS_LOG_HARD_ERROR                                              NT_STATUS = 0x4000001A
	NT_STATUS_ALREADY_WIN32                                               NT_STATUS = 0x4000001B
	NT_STATUS_WX86_UNSIMULATE                                             NT_STATUS = 0x4000001C
	NT_STATUS_WX86_CONTINUE                                               NT_STATUS = 0x4000001D
	NT_STATUS_WX86_SINGLE_STEP                                            NT_STATUS = 0x4000001E
	NT_STATUS_WX86_BREAKPOINT                                             NT_STATUS = 0x4000001F
	NT_STATUS_WX86_EXCEPTION_CONTINUE                                     NT_STATUS = 0x40000020
	NT_STATUS_WX86_EXCEPTION_LASTCHANCE                                   NT_STATUS = 0x40000021
	NT_STATUS_WX86_EXCEPTION_CHAIN                                        NT_STATUS = 0x40000022
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH_EXE                             NT_STATUS = 0x40000023
	NT_STATUS_NO_YIELD_PERFORMED                                          NT_STATUS = 0x40000024
	NT_STATUS_TIMER_RESUME_IGNORED                                        NT_STATUS = 0x40000025
	NT_STATUS_ARBITRATION_UNHANDLED                                       NT_STATUS = 0x40000026
	NT_STATUS_CARDBUS_NOT_SUPPORTED                                       NT_STATUS = 0x40000027
	NT_STATUS_WX86_CREATEWX86TIB                                          NT_STATUS = 0x40000028
	NT_STATUS_MP_PROCESSOR_MISMATCH                                       NT_STATUS = 0x40000029
	NT_STATUS_HIBERNATED                                                  NT_STATUS = 0x4000002A
	NT_STATUS_RESUME_HIBERNATION                                          NT_STATUS = 0x4000002B
	NT_STATUS_FIRMWARE_UPDATED                                            NT_STATUS = 0x4000002C
	NT_STATUS_DRIVERS_LEAKING_LOCKED_PAGES                                NT_STATUS = 0x4000002D
	NT_STATUS_MESSAGE_RETRIEVED                                           NT_STATUS = 0x4000002E
	NT_STATUS_SYSTEM_POWERSTATE_TRANSITION                                NT_STATUS = 0x4000002F
	NT_STATUS_ALPC_CHECK_COMPLETION_LIST                                  NT_STATUS = 0x40000030
	NT_STATUS_SYSTEM_POWERSTATE_COMPLEX_TRANSITION                        NT_STATUS = 0x40000031
	NT_STATUS_ACCESS_AUDIT_BY_POLICY                                      NT_STATUS = 0x40000032
	NT_STATUS_ABANDON_HIBERFILE                                           NT_STATUS = 0x40000033
	NT_STATUS_BIZRULES_NOT_ENABLED                                        NT_STATUS = 0x40000034
	NT_STATUS_WAKE_SYSTEM                                                 NT_STATUS = 0x40000294
	NT_STATUS_DS_SHUTTING_DOWN                                            NT_STATUS = 0x40000370
	NT_STATUS_DBG_REPLY_LATER                                             NT_STATUS = 0x40010001
	NT_STATUS_DBG_UNABLE_TO_PROVIDE_HANDLE                                NT_STATUS = 0x40010002
	NT_STATUS_DBG_TERMINATE_THREAD                                        NT_STATUS = 0x40010003
	NT_STATUS_DBG_TERMINATE_PROCESS                                       NT_STATUS = 0x40010004
	NT_STATUS_DBG_CONTROL_C                                               NT_STATUS = 0x40010005
	NT_STATUS_DBG_PRINTEXCEPTION_C                                        NT_STATUS = 0x40010006
	NT_STATUS_DBG_RIPEXCEPTION                                            NT_STATUS = 0x40010007
	NT_STATUS_DBG_CONTROL_BREAK                                           NT_STATUS = 0x40010008
	NT_STATUS_DBG_COMMAND_EXCEPTION                                       NT_STATUS = 0x40010009
	NT_STATUS_RPC_NT_UUID_LOCAL_ONLY                                      NT_STATUS = 0x40020056
	NT_STATUS_RPC_NT_SEND_INCOMPLETE                                      NT_STATUS = 0x400200AF
	NT_STATUS_CTX_CDM_CONNECT                                             NT_STATUS = 0x400A0004
	NT_STATUS_CTX_CDM_DISCONNECT                                          NT_STATUS = 0x400A0005
	NT_STATUS_SXS_RELEASE_ACTIVATION_CONTEXT                              NT_STATUS = 0x4015000D
	NT_STATUS_RECOVERY_NOT_NEEDED                                         NT_STATUS = 0x40190034
	NT_STATUS_RM_ALREADY_STARTED                                          NT_STATUS = 0x40190035
	NT_STATUS_LOG_NO_RESTART                                              NT_STATUS = 0x401A000C
	NT_STATUS_VIDEO_DRIVER_DEBUG_REPORT_REQUEST                           NT_STATUS = 0x401B00EC
	NT_STATUS_GRAPHICS_PARTIAL_DATA_POPULATED                             NT_STATUS = 0x401E000A
	NT_STATUS_GRAPHICS_DRIVER_MISMATCH                                    NT_STATUS = 0x401E0117
	NT_STATUS_GRAPHICS_MODE_NOT_PINNED                                    NT_STATUS = 0x401E0307
	NT_STATUS_GRAPHICS_NO_PREFERRED_MODE                                  NT_STATUS = 0x401E031E
	NT_STATUS_GRAPHICS_DATASET_IS_EMPTY                                   NT_STATUS = 0x401E034B
	NT_STATUS_GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET                        NT_STATUS = 0x401E034C
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED    NT_STATUS = 0x401E0351
	NT_STATUS_GRAPHICS_UNKNOWN_CHILD_STATUS                               NT_STATUS = 0x401E042F
	NT_STATUS_GRAPHICS_LEADLINK_START_DEFERRED                            NT_STATUS = 0x401E0437
	NT_STATUS_GRAPHICS_POLLING_TOO_FREQUENTLY                             NT_STATUS = 0x401E0439
	NT_STATUS_GRAPHICS_START_DEFERRED                                     NT_STATUS = 0x401E043A
	NT_STATUS_NDIS_INDICATION_REQUIRED                                    NT_STATUS = 0x40230001
	NT_STATUS_GUARD_PAGE_VIOLATION                                        NT_STATUS = 0x80000001
	NT_STATUS_DATATYPE_MISALIGNMENT                                       NT_STATUS = 0x80000002
	NT_STATUS_BREAKPOINT                                                  NT_STATUS = 0x80000003
	NT_STATUS_SINGLE_STEP                                                 NT_STATUS = 0x80000004
	NT_STATUS_BUFFER_OVERFLOW                                             NT_STATUS = 0x80000005
	NT_STATUS_NO_MORE_FILES                                               NT_STATUS = 0x80000006
	NT_STATUS_WAKE_SYSTEM_DEBUGGER                                        NT_STATUS = 0x80000007
	NT_STATUS_HANDLES_CLOSED                                              NT_STATUS = 0x8000000A
	NT_STATUS_NO_INHERITANCE                                              NT_STATUS = 0x8000000B
	NT_STATUS_GUID_SUBSTITUTION_MADE                                      NT_STATUS = 0x8000000C
	NT_STATUS_PARTIAL_COPY                                                NT_STATUS = 0x8000000D
	NT_STATUS_DEVICE_PAPER_EMPTY                                          NT_STATUS = 0x8000000E
	NT_STATUS_DEVICE_POWERED_OFF                                          NT_STATUS = 0x8000000F
	NT_STATUS_DEVICE_OFF_LINE                                             NT_STATUS = 0x80000010
	NT_STATUS_DEVICE_BUSY                                                 NT_STATUS = 0x80000011
	NT_STATUS_NO_MORE_EAS                                                 NT_STATUS = 0x80000012
	NT_STATUS_INVALID_EA_NAME                                             NT_STATUS = 0x80000013
	NT_STATUS_EA_LIST_INCONSISTENT                                        NT_STATUS = 0x80000014
	NT_STATUS_INVALID_EA_FLAG                                             NT_STATUS = 0x80000015
	NT_STATUS_VERIFY_REQUIRED                                             NT_STATUS = 0x80000016
	NT_STATUS_EXTRANEOUS_INFORMATION                                      NT_STATUS = 0x80000017
	NT_STATUS_RXACT_COMMIT_NECESSARY                                      NT_STATUS = 0x80000018
	NT_STATUS_NO_MORE_ENTRIES                                             NT_STATUS = 0x8000001A
	NT_STATUS_FILEMARK_DETECTED                                           NT_STATUS = 0x8000001B
	NT_STATUS_MEDIA_CHANGED                                               NT_STATUS = 0x8000001C
	NT_STATUS_BUS_RESET                                                   NT_STATUS = 0x8000001D
	NT_STATUS_END_OF_MEDIA                                                NT_STATUS = 0x8000001E
	NT_STATUS_BEGINNING_OF_MEDIA                                          NT_STATUS = 0x8000001F
	NT_STATUS_MEDIA_CHECK                                                 NT_STATUS = 0x80000020
	NT_STATUS_SETMARK_DETECTED                                            NT_STATUS = 0x80000021
	NT_STATUS_NO_DATA_DETECTED                                            NT_STATUS = 0x80000022
	NT_STATUS_REDIRECTOR_HAS_OPEN_HANDLES                                 NT_STATUS = 0x80000023
	NT_STATUS_SERVER_HAS_OPEN_HANDLES                                     NT_STATUS = 0x80000024
	NT_STATUS_ALREADY_DISCONNECTED                                        NT_STATUS = 0x80000025
	NT_STATUS_LONGJUMP                                                    NT_STATUS = 0x80000026
	NT_STATUS_CLEANER_CARTRIDGE_INSTALLED                                 NT_STATUS = 0x80000027
	NT_STATUS_PLUGPLAY_QUERY_VETOED                                       NT_STATUS = 0x80000028
	NT_STATUS_UNWIND_CONSOLIDATE                                          NT_STATUS = 0x80000029
	NT_STATUS_REGISTRY_HIVE_RECOVERED                                     NT_STATUS = 0x8000002A
	NT_STATUS_DLL_MIGHT_BE_INSECURE                                       NT_STATUS = 0x8000002B
	NT_STATUS_DLL_MIGHT_BE_INCOMPATIBLE                                   NT_STATUS = 0x8000002C
	NT_STATUS_STOPPED_ON_SYMLINK                                          NT_STATUS = 0x8000002D
	NT_STATUS_DEVICE_REQUIRES_CLEANING                                    NT_STATUS = 0x80000288
	NT_STATUS_DEVICE_DOOR_OPEN                                            NT_STATUS = 0x80000289
	NT_STATUS_DATA_LOST_REPAIR                                            NT_STATUS = 0x80000803
	NT_STATUS_DBG_EXCEPTION_NOT_HANDLED                                   NT_STATUS = 0x80010001
	NT_STATUS_CLUSTER_NODE_ALREADY_UP                                     NT_STATUS = 0x80130001
	NT_STATUS_CLUSTER_NODE_ALREADY_DOWN                                   NT_STATUS = 0x80130002
	NT_STATUS_CLUSTER_NETWORK_ALREADY_ONLINE                              NT_STATUS = 0x80130003
	NT_STATUS_CLUSTER_NETWORK_ALREADY_OFFLINE                             NT_STATUS = 0x80130004
	NT_STATUS_CLUSTER_NODE_ALREADY_MEMBER                                 NT_STATUS = 0x80130005
	NT_STATUS_COULD_NOT_RESIZE_LOG                                        NT_STATUS = 0x80190009
	NT_STATUS_NO_TXF_METADATA                                             NT_STATUS = 0x80190029
	NT_STATUS_CANT_RECOVER_WITH_HANDLE_OPEN                               NT_STATUS = 0x80190031
	NT_STATUS_TXF_METADATA_ALREADY_PRESENT                                NT_STATUS = 0x80190041
	NT_STATUS_TRANSACTION_SCOPE_CALLBACKS_NOT_SET                         NT_STATUS = 0x80190042
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED                  NT_STATUS = 0x801B00EB
	NT_STATUS_FLT_BUFFER_TOO_SMALL                                        NT_STATUS = 0x801C0001
	NT_STATUS_FVE_PARTIAL_METADATA                                        NT_STATUS = 0x80210001
	NT_STATUS_FVE_TRANSIENT_STATE                                         NT_STATUS = 0x80210002
	NT_STATUS_UNSUCCESSFUL                                                NT_STATUS = 0xC0000001
	NT_STATUS_NOT_IMPLEMENTED                                             NT_STATUS = 0xC0000002
	NT_STATUS_INVALID_INFO_CLASS                                          NT_STATUS = 0xC0000003
	NT_STATUS_INFO_LENGTH_MISMATCH                                        NT_STATUS = 0xC0000004
	NT_STATUS_ACCESS_VIOLATION                                            NT_STATUS = 0xC0000005
	NT_STATUS_IN_PAGE_ERROR                                               NT_STATUS = 0xC0000006
	NT_STATUS_PAGEFILE_QUOTA                                              NT_STATUS = 0xC0000007
	NT_STATUS_INVALID_HANDLE                                              NT_STATUS = 0xC0000008
	NT_STATUS_BAD_INITIAL_STACK                                           NT_STATUS = 0xC0000009
	NT_STATUS_BAD_INITIAL_PC                                              NT_STATUS = 0xC000000A
	NT_STATUS_INVALID_CID                                                 NT_STATUS = 0xC000000B
	NT_STATUS_TIMER_NOT_CANCELED                                          NT_STATUS = 0xC000000C
	NT_STATUS_INVALID_PARAMETER                                           NT_STATUS = 0xC000000D
	NT_STATUS_NO_SUCH_DEVICE                                              NT_STATUS = 0xC000000E
	NT_STATUS_NO_SUCH_FILE                                                NT_STATUS = 0xC000000F
	NT_STATUS_INVALID_DEVICE_REQUEST                                      NT_STATUS = 0xC0000010
	NT_STATUS_END_OF_FILE                                                 NT_STATUS = 0xC0000011
	NT_STATUS_WRONG_VOLUME                                                NT_STATUS = 0xC0000012
	NT_STATUS_NO_MEDIA_IN_DEVICE                                          NT_STATUS = 0xC0000013
	NT_STATUS_UNRECOGNIZED_MEDIA                                          NT_STATUS = 0xC0000014
	NT_STATUS_NONEXISTENT_SECTOR                                          NT_STATUS = 0xC0000015
	NT_STATUS_MORE_PROCESSING_REQUIRED                                    NT_STATUS = 0xC0000016
	NT_STATUS_NO_MEMORY                                                   NT_STATUS = 0xC0000017
	NT_STATUS_CONFLICTING_ADDRESSES                                       NT_STATUS = 0xC0000018
	NT_STATUS_NOT_MAPPED_VIEW                                             NT_STATUS = 0xC0000019
	NT_STATUS_UNABLE_TO_FREE_VM                                           NT_STATUS = 0xC000001A
	NT_STATUS_UNABLE_TO_DELETE_SECTION                                    NT_STATUS = 0xC000001B
	NT_STATUS_INVALID_SYSTEM_SERVICE                                      NT_STATUS = 0xC000001C
	NT_STATUS_ILLEGAL_INSTRUCTION                                         NT_STATUS = 0xC000001D
	NT_STATUS_INVALID_LOCK_SEQUENCE                                       NT_STATUS = 0xC000001E
	NT_STATUS_INVALID_VIEW_SIZE                                           NT_STATUS = 0xC000001F
	NT_STATUS_INVALID_FILE_FOR_SECTION                                    NT_STATUS = 0xC0000020
	NT_STATUS_ALREADY_COMMITTED                                           NT_STATUS = 0xC0000021
	NT_STATUS_ACCESS_DENIED                                               NT_STATUS = 0xC0000022
	NT_STATUS_BUFFER_TOO_SMALL                                            NT_STATUS = 0xC0000023
	NT_STATUS_OBJECT_TYPE_MISMATCH                                        NT_STATUS = 0xC0000024
	NT_STATUS_NONCONTINUABLE_EXCEPTION                                    NT_STATUS = 0xC0000025
	NT_STATUS_INVALID_DISPOSITION                                         NT_STATUS = 0xC0000026
	NT_STATUS_UNWIND                                                      NT_STATUS = 0xC0000027
	NT_STATUS_BAD_STACK                                                   NT_STATUS = 0xC0000028
	NT_STATUS_INVALID_UNWIND_TARGET                                       NT_STATUS = 0xC0000029
	NT_STATUS_NOT_LOCKED                                                  NT_STATUS = 0xC000002A
	NT_STATUS_PARITY_ERROR                                                NT_STATUS = 0xC000002B
	NT_STATUS_UNABLE_TO_DECOMMIT_VM                                       NT_STATUS = 0xC000002C
	NT_STATUS_NOT_COMMITTED                                               NT_STATUS = 0xC000002D
	NT_STATUS_INVALID_PORT_ATTRIBUTES                                     NT_STATUS = 0xC000002E
	NT_STATUS_PORT_MESSAGE_TOO_LONG                                       NT_STATUS = 0xC000002F
	NT_STATUS_INVALID_PARAMETER_MIX                                       NT_STATUS = 0xC0000030
	NT_STATUS_INVALID_QUOTA_LOWER                                         NT_STATUS = 0xC0000031
	NT_STATUS_DISK_CORRUPT_ERROR                                          NT_STATUS = 0xC0000032
	NT_STATUS_OBJECT_NAME_INVALID                                         NT_STATUS = 0xC0000033
	NT_STATUS_OBJECT_NAME_NOT_FOUND                                       NT_STATUS = 0xC0000034
	NT_STATUS_OBJECT_NAME_COLLISION                                       NT_STATUS = 0xC0000035
	NT_STATUS_PORT_DISCONNECTED                                           NT_STATUS = 0xC0000037
	NT_STATUS_DEVICE_ALREADY_ATTACHED                                     NT_STATUS = 0xC0000038
	NT_STATUS_OBJECT_PATH_INVALID                                         NT_STATUS = 0xC0000039
	NT_STATUS_OBJECT_PATH_NOT_FOUND                                       NT_STATUS = 0xC000003A
	NT_STATUS_OBJECT_PATH_SYNTAX_BAD                                      NT_STATUS = 0xC000003B
	NT_STATUS_DATA_OVERRUN                                                NT_STATUS = 0xC000003C
	NT_STATUS_DATA_LATE_ERROR                                             NT_STATUS = 0xC000003D
	NT_STATUS_DATA_ERROR                                                  NT_STATUS = 0xC000003E
	NT_STATUS_CRC_ERROR                                                   NT_STATUS = 0xC000003F
	NT_STATUS_SECTION_TOO_BIG                                             NT_STATUS = 0xC0000040
	NT_STATUS_PORT_CONNECTION_REFUSED                                     NT_STATUS = 0xC0000041
	NT_STATUS_INVALID_PORT_HANDLE                                         NT_STATUS = 0xC0000042
	NT_STATUS_SHARING_VIOLATION                                           NT_STATUS = 0xC0000043
	NT_STATUS_QUOTA_EXCEEDED                                              NT_STATUS = 0xC0000044
	NT_STATUS_INVALID_PAGE_PROTECTION                                     NT_STATUS = 0xC0000045
	NT_STATUS_MUTANT_NOT_OWNED                                            NT_STATUS = 0xC0000046
	NT_STATUS_SEMAPHORE_LIMIT_EXCEEDED                                    NT_STATUS = 0xC0000047
	NT_STATUS_PORT_ALREADY_SET                                            NT_STATUS = 0xC0000048
	NT_STATUS_SECTION_NOT_IMAGE                                           NT_STATUS = 0xC0000049
	NT_STATUS_SUSPEND_COUNT_EXCEEDED                                      NT_STATUS = 0xC000004A
	NT_STATUS_THREAD_IS_TERMINATING                                       NT_STATUS = 0xC000004B
	NT_STATUS_BAD_WORKING_SET_LIMIT                                       NT_STATUS = 0xC000004C
	NT_STATUS_INCOMPATIBLE_FILE_MAP                                       NT_STATUS = 0xC000004D
	NT_STATUS_SECTION_PROTECTION                                          NT_STATUS = 0xC000004E
	NT_STATUS_EAS_NOT_SUPPORTED                                           NT_STATUS = 0xC000004F
	NT_STATUS_EA_TOO_LARGE                                                NT_STATUS = 0xC0000050
	NT_STATUS_NONEXISTENT_EA_ENTRY                                        NT_STATUS = 0xC0000051
	NT_STATUS_NO_EAS_ON_FILE                                              NT_STATUS = 0xC0000052
	NT_STATUS_EA_CORRUPT_ERROR                                            NT_STATUS = 0xC0000053
	NT_STATUS_FILE_LOCK_CONFLICT                                          NT_STATUS = 0xC0000054
	NT_STATUS_LOCK_NOT_GRANTED                                            NT_STATUS = 0xC0000055
	NT_STATUS_DELETE_PENDING                                              NT_STATUS = 0xC0000056
	NT_STATUS_CTL_FILE_NOT_SUPPORTED                                      NT_STATUS = 0xC0000057
	NT_STATUS_UNKNOWN_REVISION                                            NT_STATUS = 0xC0000058
	NT_STATUS_REVISION_MISMATCH                                           NT_STATUS = 0xC0000059
	NT_STATUS_INVALID_OWNER                                               NT_STATUS = 0xC000005A
	NT_STATUS_INVALID_PRIMARY_GROUP                                       NT_STATUS = 0xC000005B
	NT_STATUS_NO_IMPERSONATION_TOKEN                                      NT_STATUS = 0xC000005C
	NT_STATUS_CANT_DISABLE_MANDATORY                                      NT_STATUS = 0xC000005D
	NT_STATUS_NO_LOGON_SERVERS                                            NT_STATUS = 0xC000005E
	NT_STATUS_NO_SUCH_LOGON_SESSION                                       NT_STATUS = 0xC000005F
	NT_STATUS_NO_SUCH_PRIVILEGE                                           NT_STATUS = 0xC0000060
	NT_STATUS_PRIVILEGE_NOT_HELD                                          NT_STATUS = 0xC0000061
	NT_STATUS_INVALID_ACCOUNT_NAME                                        NT_STATUS = 0xC0000062
	NT_STATUS_USER_EXISTS                                                 NT_STATUS = 0xC0000063
	NT_STATUS_NO_SUCH_USER                                                NT_STATUS = 0xC0000064
	NT_STATUS_GROUP_EXISTS                                                NT_STATUS = 0xC0000065
	NT_STATUS_NO_SUCH_GROUP                                               NT_STATUS = 0xC0000066
	NT_STATUS_MEMBER_IN_GROUP                                             NT_STATUS = 0xC0000067
	NT_STATUS_MEMBER_NOT_IN_GROUP                                         NT_STATUS = 0xC0000068
	NT_STATUS_LAST_ADMIN                                                  NT_STATUS = 0xC0000069
	NT_STATUS_WRONG_PASSWORD                                              NT_STATUS = 0xC000006A
	NT_STATUS_ILL_FORMED_PASSWORD                                         NT_STATUS = 0xC000006B
	NT_STATUS_PASSWORD_RESTRICTION                                        NT_STATUS = 0xC000006C
	NT_STATUS_LOGON_FAILURE                                               NT_STATUS = 0xC000006D
	NT_STATUS_ACCOUNT_RESTRICTION                                         NT_STATUS = 0xC000006E
	NT_STATUS_INVALID_LOGON_HOURS                                         NT_STATUS = 0xC000006F
	NT_STATUS_INVALID_WORKSTATION                                         NT_STATUS = 0xC0000070
	NT_STATUS_PASSWORD_EXPIRED                                            NT_STATUS = 0xC0000071
	NT_STATUS_ACCOUNT_DISABLED                                            NT_STATUS = 0xC0000072
	NT_STATUS_NONE_MAPPED                                                 NT_STATUS = 0xC0000073
	NT_STATUS_TOO_MANY_LUIDS_REQUESTED                                    NT_STATUS = 0xC0000074
	NT_STATUS_LUIDS_EXHAUSTED                                             NT_STATUS = 0xC0000075
	NT_STATUS_INVALID_SUB_AUTHORITY                                       NT_STATUS = 0xC0000076
	NT_STATUS_INVALID_ACL                                                 NT_STATUS = 0xC0000077
	NT_STATUS_INVALID_SID                                                 NT_STATUS = 0xC0000078
	NT_STATUS_INVALID_SECURITY_DESCR                                      NT_STATUS = 0xC0000079
	NT_STATUS_PROCEDURE_NOT_FOUND                                         NT_STATUS = 0xC000007A
	NT_STATUS_INVALID_IMAGE_FORMAT                                        NT_STATUS = 0xC000007B
	NT_STATUS_NO_TOKEN                                                    NT_STATUS = 0xC000007C
	NT_STATUS_BAD_INHERITANCE_ACL                                         NT_STATUS = 0xC000007D
	NT_STATUS_RANGE_NOT_LOCKED                                            NT_STATUS = 0xC000007E
	NT_STATUS_DISK_FULL                                                   NT_STATUS = 0xC000007F
	NT_STATUS_SERVER_DISABLED                                             NT_STATUS = 0xC0000080
	NT_STATUS_SERVER_NOT_DISABLED                                         NT_STATUS = 0xC0000081
	NT_STATUS_TOO_MANY_GUIDS_REQUESTED                                    NT_STATUS = 0xC0000082
	NT_STATUS_GUIDS_EXHAUSTED                                             NT_STATUS = 0xC0000083
	NT_STATUS_INVALID_ID_AUTHORITY                                        NT_STATUS = 0xC0000084
	NT_STATUS_AGENTS_EXHAUSTED                                            NT_STATUS = 0xC0000085
	NT_STATUS_INVALID_VOLUME_LABEL                                        NT_STATUS = 0xC0000086
	NT_STATUS_SECTION_NOT_EXTENDED                                        NT_STATUS = 0xC0000087
	NT_STATUS_NOT_MAPPED_DATA                                             NT_STATUS = 0xC0000088
	NT_STATUS_RESOURCE_DATA_NOT_FOUND                                     NT_STATUS = 0xC0000089
	NT_STATUS_RESOURCE_TYPE_NOT_FOUND                                     NT_STATUS = 0xC000008A
	NT_STATUS_RESOURCE_NAME_NOT_FOUND                                     NT_STATUS = 0xC000008B
	NT_STATUS_ARRAY_BOUNDS_EXCEEDED                                       NT_STATUS = 0xC000008C
	NT_STATUS_FLOAT_DENORMAL_OPERAND                                      NT_STATUS = 0xC000008D
	NT_STATUS_FLOAT_DIVIDE_BY_ZERO                                        NT_STATUS = 0xC000008E
	NT_STATUS_FLOAT_INEXACT_RESULT                                        NT_STATUS = 0xC000008F
	NT_STATUS_FLOAT_INVALID_OPERATION                                     NT_STATUS = 0xC0000090
	NT_STATUS_FLOAT_OVERFLOW                                              NT_STATUS = 0xC0000091
	NT_STATUS_FLOAT_STACK_CHECK                                           NT_STATUS = 0xC0000092
	NT_STATUS_FLOAT_UNDERFLOW                                             NT_STATUS = 0xC0000093
	NT_STATUS_INTEGER_DIVIDE_BY_ZERO                                      NT_STATUS = 0xC0000094
	NT_STATUS_INTEGER_OVERFLOW                                            NT_STATUS = 0xC0000095
	NT_STATUS_PRIVILEGED_INSTRUCTION                                      NT_STATUS = 0xC0000096
	NT_STATUS_TOO_MANY_PAGING_FILES                                       NT_STATUS = 0xC0000097
	NT_STATUS_FILE_INVALID                                                NT_STATUS = 0xC0000098
	NT_STATUS_ALLOTTED_SPACE_EXCEEDED                                     NT_STATUS = 0xC0000099
	NT_STATUS_INSUFFICIENT_RESOURCES                                      NT_STATUS = 0xC000009A
	NT_STATUS_DFS_EXIT_PATH_FOUND                                         NT_STATUS = 0xC000009B
	NT_STATUS_DEVICE_DATA_ERROR                                           NT_STATUS = 0xC000009C
	NT_STATUS_DEVICE_NOT_CONNECTED                                        NT_STATUS = 0xC000009D
	NT_STATUS_FREE_VM_NOT_AT_BASE                                         NT_STATUS = 0xC000009F
	NT_STATUS_MEMORY_NOT_ALLOCATED                                        NT_STATUS = 0xC00000A0
	NT_STATUS_WORKING_SET_QUOTA                                           NT_STATUS = 0xC00000A1
	NT_STATUS_MEDIA_WRITE_PROTECTED                                       NT_STATUS = 0xC00000A2
	NT_STATUS_DEVICE_NOT_READY                                            NT_STATUS = 0xC00000A3
	NT_STATUS_INVALID_GROUP_ATTRIBUTES                                    NT_STATUS = 0xC00000A4
	NT_STATUS_BAD_IMPERSONATION_LEVEL                                     NT_STATUS = 0xC00000A5
	NT_STATUS_CANT_OPEN_ANONYMOUS                                         NT_STATUS = 0xC00000A6
	NT_STATUS_BAD_VALIDATION_CLASS                                        NT_STATUS = 0xC00000A7
	NT_STATUS_BAD_TOKEN_TYPE                                              NT_STATUS = 0xC00000A8
	NT_STATUS_BAD_MASTER_BOOT_RECORD                                      NT_STATUS = 0xC00000A9
	NT_STATUS_INSTRUCTION_MISALIGNMENT                                    NT_STATUS = 0xC00000AA
	NT_STATUS_INSTANCE_NOT_AVAILABLE                                      NT_STATUS = 0xC00000AB
	NT_STATUS_PIPE_NOT_AVAILABLE                                          NT_STATUS = 0xC00000AC
	NT_STATUS_INVALID_PIPE_STATE                                          NT_STATUS = 0xC00000AD
	NT_STATUS_PIPE_BUSY                                                   NT_STATUS = 0xC00000AE
	NT_STATUS_ILLEGAL_FUNCTION                                            NT_STATUS = 0xC00000AF
	NT_STATUS_PIPE_DISCONNECTED                                           NT_STATUS = 0xC00000B0
	NT_STATUS_PIPE_CLOSING                                                NT_STATUS = 0xC00000B1
	NT_STATUS_PIPE_CONNECTED                                              NT_STATUS = 0xC00000B2
	NT_STATUS_PIPE_LISTENING                                              NT_STATUS = 0xC00000B3
	NT_STATUS_INVALID_READ_MODE                                           NT_STATUS = 0xC00000B4
	NT_STATUS_IO_TIMEOUT                                                  NT_STATUS = 0xC00000B5
	NT_STATUS_FILE_FORCED_CLOSED                                          NT_STATUS = 0xC00000B6
	NT_STATUS_PROFILING_NOT_STARTED                                       NT_STATUS = 0xC00000B7
	NT_STATUS_PROFILING_NOT_STOPPED                                       NT_STATUS = 0xC00000B8
	NT_STATUS_COULD_NOT_INTERPRET                                         NT_STATUS = 0xC00000B9
	NT_STATUS_FILE_IS_A_DIRECTORY                                         NT_STATUS = 0xC00000BA
	NT_STATUS_NOT_SUPPORTED                                               NT_STATUS = 0xC00000BB
	NT_STATUS_REMOTE_NOT_LISTENING                                        NT_STATUS = 0xC00000BC
	NT_STATUS_DUPLICATE_NAME                                              NT_STATUS = 0xC00000BD
	NT_STATUS_BAD_NETWORK_PATH                                            NT_STATUS = 0xC00000BE
	NT_STATUS_NETWORK_BUSY                                                NT_STATUS = 0xC00000BF
	NT_STATUS_DEVICE_DOES_NOT_EXIST                                       NT_STATUS = 0xC00000C0
	NT_STATUS_TOO_MANY_COMMANDS                                           NT_STATUS = 0xC00000C1
	NT_STATUS_ADAPTER_HARDWARE_ERROR                                      NT_STATUS = 0xC00000C2
	NT_STATUS_INVALID_NETWORK_RESPONSE                                    NT_STATUS = 0xC00000C3
	NT_STATUS_UNEXPECTED_NETWORK_ERROR                                    NT_STATUS = 0xC00000C4
	NT_STATUS_BAD_REMOTE_ADAPTER                                          NT_STATUS = 0xC00000C5
	NT_STATUS_PRINT_QUEUE_FULL                                            NT_STATUS = 0xC00000C6
	NT_STATUS_NO_SPOOL_SPACE                                              NT_STATUS = 0xC00000C7
	NT_STATUS_PRINT_CANCELLED                                             NT_STATUS = 0xC00000C8
	NT_STATUS_NETWORK_NAME_DELETED                                        NT_STATUS = 0xC00000C9
	NT_STATUS_NETWORK_ACCESS_DENIED                                       NT_STATUS = 0xC00000CA
	NT_STATUS_BAD_DEVICE_TYPE                                             NT_STATUS = 0xC00000CB
	NT_STATUS_BAD_NETWORK_NAME                                            NT_STATUS = 0xC00000CC
	NT_STATUS_TOO_MANY_NAMES                                              NT_STATUS = 0xC00000CD
	NT_STATUS_TOO_MANY_SESSIONS                                           NT_STATUS = 0xC00000CE
	NT_STATUS_SHARING_PAUSED                                              NT_STATUS = 0xC00000CF
	NT_STATUS_REQUEST_NOT_ACCEPTED                                        NT_STATUS = 0xC00000D0
	NT_STATUS_REDIRECTOR_PAUSED                                           NT_STATUS = 0xC00000D1
	NT_STATUS_NET_WRITE_FAULT                                             NT_STATUS = 0xC00000D2
	NT_STATUS_PROFILING_AT_LIMIT                                          NT_STATUS = 0xC00000D3
	NT_STATUS_NOT_SAME_DEVICE                                             NT_STATUS = 0xC00000D4
	NT_STATUS_FILE_RENAMED                                                NT_STATUS = 0xC00000D5
	NT_STATUS_VIRTUAL_CIRCUIT_CLOSED                                      NT_STATUS = 0xC00000D6
	NT_STATUS_NO_SECURITY_ON_OBJECT                                       NT_STATUS = 0xC00000D7
	NT_STATUS_CANT_WAIT                                                   NT_STATUS = 0xC00000D8
	NT_STATUS_PIPE_EMPTY                                                  NT_STATUS = 0xC00000D9
	NT_STATUS_CANT_ACCESS_DOMAIN_INFO                                     NT_STATUS = 0xC00000DA
	NT_STATUS_CANT_TERMINATE_SELF                                         NT_STATUS = 0xC00000DB
	NT_STATUS_INVALID_SERVER_STATE                                        NT_STATUS = 0xC00000DC
	NT_STATUS_INVALID_DOMAIN_STATE                                        NT_STATUS = 0xC00000DD
	NT_STATUS_INVALID_DOMAIN_ROLE                                         NT_STATUS = 0xC00000DE
	NT_STATUS_NO_SUCH_DOMAIN                                              NT_STATUS = 0xC00000DF
	NT_STATUS_DOMAIN_EXISTS                                               NT_STATUS = 0xC00000E0
	NT_STATUS_DOMAIN_LIMIT_EXCEEDED                                       NT_STATUS = 0xC00000E1
	NT_STATUS_OPLOCK_NOT_GRANTED                                          NT_STATUS = 0xC00000E2
	NT_STATUS_INVALID_OPLOCK_PROTOCOL                                     NT_STATUS = 0xC00000E3
	NT_STATUS_INTERNAL_DB_CORRUPTION                                      NT_STATUS = 0xC00000E4
	NT_STATUS_INTERNAL_ERROR                                              NT_STATUS = 0xC00000E5
	NT_STATUS_GENERIC_NOT_MAPPED                                          NT_STATUS = 0xC00000E6
	NT_STATUS_BAD_DESCRIPTOR_FORMAT                                       NT_STATUS = 0xC00000E7
	NT_STATUS_INVALID_USER_BUFFER                                         NT_STATUS = 0xC00000E8
	NT_STATUS_UNEXPECTED_IO_ERROR                                         NT_STATUS = 0xC00000E9
	NT_STATUS_UNEXPECTED_MM_CREATE_ERR                                    NT_STATUS = 0xC00000EA
	NT_STATUS_UNEXPECTED_MM_MAP_ERROR                                     NT_STATUS = 0xC00000EB
	NT_STATUS_UNEXPECTED_MM_EXTEND_ERR                                    NT_STATUS = 0xC00000EC
	NT_STATUS_NOT_LOGON_PROCESS                                           NT_STATUS = 0xC00000ED
	NT_STATUS_LOGON_SESSION_EXISTS                                        NT_STATUS = 0xC00000EE
	NT_STATUS_INVALID_PARAMETER_1                                         NT_STATUS = 0xC00000EF
	NT_STATUS_INVALID_PARAMETER_2                                         NT_STATUS = 0xC00000F0
	NT_STATUS_INVALID_PARAMETER_3                                         NT_STATUS = 0xC00000F1
	NT_STATUS_INVALID_PARAMETER_4                                         NT_STATUS = 0xC00000F2
	NT_STATUS_INVALID_PARAMETER_5                                         NT_STATUS = 0xC00000F3
	NT_STATUS_INVALID_PARAMETER_6                                         NT_STATUS = 0xC00000F4
	NT_STATUS_INVALID_PARAMETER_7                                         NT_STATUS = 0xC00000F5
	NT_STATUS_INVALID_PARAMETER_8                                         NT_STATUS = 0xC00000F6
	NT_STATUS_INVALID_PARAMETER_9                                         NT_STATUS = 0xC00000F7
	NT_STATUS_INVALID_PARAMETER_10                                        NT_STATUS = 0xC00000F8
	NT_STATUS_INVALID_PARAMETER_11                                        NT_STATUS = 0xC00000F9
	NT_STATUS_INVALID_PARAMETER_12                                        NT_STATUS = 0xC00000FA
	NT_STATUS_REDIRECTOR_NOT_STARTED                                      NT_STATUS = 0xC00000FB
	NT_STATUS_REDIRECTOR_STARTED                                          NT_STATUS = 0xC00000FC
	NT_STATUS_STACK_OVERFLOW                                              NT_STATUS = 0xC00000FD
	NT_STATUS_NO_SUCH_PACKAGE                                             NT_STATUS = 0xC00000FE
	NT_STATUS_BAD_FUNCTION_TABLE                                          NT_STATUS = 0xC00000FF
	NT_STATUS_VARIABLE_NOT_FOUND                                          NT_STATUS = 0xC0000100
	NT_STATUS_DIRECTORY_NOT_EMPTY                                         NT_STATUS = 0xC0000101
	NT_STATUS_FILE_CORRUPT_ERROR                                          NT_STATUS = 0xC0000102
	NT_STATUS_NOT_A_DIRECTORY                                             NT_STATUS = 0xC0000103
	NT_STATUS_BAD_LOGON_SESSION_STATE                                     NT_STATUS = 0xC0000104
	NT_STATUS_LOGON_SESSION_COLLISION                                     NT_STATUS = 0xC0000105
	NT_STATUS_NAME_TOO_LONG                                               NT_STATUS = 0xC0000106
	NT_STATUS_FILES_OPEN                                                  NT_STATUS = 0xC0000107
	NT_STATUS_CONNECTION_IN_USE                                           NT_STATUS = 0xC0000108
	NT_STATUS_MESSAGE_NOT_FOUND                                           NT_STATUS = 0xC0000109
	NT_STATUS_PROCESS_IS_TERMINATING                                      NT_STATUS = 0xC000010A
	NT_STATUS_INVALID_LOGON_TYPE                                          NT_STATUS = 0xC000010B
	NT_STATUS_NO_GUID_TRANSLATION                                         NT_STATUS = 0xC000010C
	NT_STATUS_CANNOT_IMPERSONATE                                          NT_STATUS = 0xC000010D
	NT_STATUS_IMAGE_ALREADY_LOADED                                        NT_STATUS = 0xC000010E
	NT_STATUS_NO_LDT                                                      NT_STATUS = 0xC0000117
	NT_STATUS_INVALID_LDT_SIZE                                            NT_STATUS = 0xC0000118
	NT_STATUS_INVALID_LDT_OFFSET                                          NT_STATUS = 0xC0000119
	NT_STATUS_INVALID_LDT_DESCRIPTOR                                      NT_STATUS = 0xC000011A
	NT_STATUS_INVALID_IMAGE_NE_FORMAT                                     NT_STATUS = 0xC000011B
	NT_STATUS_RXACT_INVALID_STATE                                         NT_STATUS = 0xC000011C
	NT_STATUS_RXACT_COMMIT_FAILURE                                        NT_STATUS = 0xC000011D
	NT_STATUS_MAPPED_FILE_SIZE_ZERO                                       NT_STATUS = 0xC000011E
	NT_STATUS_TOO_MANY_OPENED_FILES                                       NT_STATUS = 0xC000011F
	NT_STATUS_CANCELLED                                                   NT_STATUS = 0xC0000120
	NT_STATUS_CANNOT_DELETE                                               NT_STATUS = 0xC0000121
	NT_STATUS_INVALID_COMPUTER_NAME                                       NT_STATUS = 0xC0000122
	NT_STATUS_FILE_DELETED                                                NT_STATUS = 0xC0000123
	NT_STATUS_SPECIAL_ACCOUNT                                             NT_STATUS = 0xC0000124
	NT_STATUS_SPECIAL_GROUP                                               NT_STATUS = 0xC0000125
	NT_STATUS_SPECIAL_USER                                                NT_STATUS = 0xC0000126
	NT_STATUS_MEMBERS_PRIMARY_GROUP                                       NT_STATUS = 0xC0000127
	NT_STATUS_FILE_CLOSED                                                 NT_STATUS = 0xC0000128
	NT_STATUS_TOO_MANY_THREADS                                            NT_STATUS = 0xC0000129
	NT_STATUS_THREAD_NOT_IN_PROCESS                                       NT_STATUS = 0xC000012A
	NT_STATUS_TOKEN_ALREADY_IN_USE                                        NT_STATUS = 0xC000012B
	NT_STATUS_PAGEFILE_QUOTA_EXCEEDED                                     NT_STATUS = 0xC000012C
	NT_STATUS_COMMITMENT_LIMIT                                            NT_STATUS = 0xC000012D
	NT_STATUS_INVALID_IMAGE_LE_FORMAT                                     NT_STATUS = 0xC000012E
	NT_STATUS_INVALID_IMAGE_NOT_MZ                                        NT_STATUS = 0xC000012F
	NT_STATUS_INVALID_IMAGE_PROTECT                                       NT_STATUS = 0xC0000130
	NT_STATUS_INVALID_IMAGE_WIN_16                                        NT_STATUS = 0xC0000131
	NT_STATUS_LOGON_SERVER_CONFLICT                                       NT_STATUS = 0xC0000132
	NT_STATUS_TIME_DIFFERENCE_AT_DC                                       NT_STATUS = 0xC0000133
	NT_STATUS_SYNCHRONIZATION_REQUIRED                                    NT_STATUS = 0xC0000134
	NT_STATUS_DLL_NOT_FOUND                                               NT_STATUS = 0xC0000135
	NT_STATUS_OPEN_FAILED                                                 NT_STATUS = 0xC0000136
	NT_STATUS_IO_PRIVILEGE_FAILED                                         NT_STATUS = 0xC0000137
	NT_STATUS_ORDINAL_NOT_FOUND                                           NT_STATUS = 0xC0000138
	NT_STATUS_ENTRYPOINT_NOT_FOUND                                        NT_STATUS = 0xC0000139
	NT_STATUS_CONTROL_C_EXIT                                              NT_STATUS = 0xC000013A
	NT_STATUS_LOCAL_DISCONNECT                                            NT_STATUS = 0xC000013B
	NT_STATUS_REMOTE_DISCONNECT                                           NT_STATUS = 0xC000013C
	NT_STATUS_REMOTE_RESOURCES                                            NT_STATUS = 0xC000013D
	NT_STATUS_LINK_FAILED                                                 NT_STATUS = 0xC000013E
	NT_STATUS_LINK_TIMEOUT                                                NT_STATUS = 0xC000013F
	NT_STATUS_INVALID_CONNECTION                                          NT_STATUS = 0xC0000140
	NT_STATUS_INVALID_ADDRESS                                             NT_STATUS = 0xC0000141
	NT_STATUS_DLL_INIT_FAILED                                             NT_STATUS = 0xC0000142
	NT_STATUS_MISSING_SYSTEMFILE                                          NT_STATUS = 0xC0000143
	NT_STATUS_UNHANDLED_EXCEPTION                                         NT_STATUS = 0xC0000144
	NT_STATUS_APP_INIT_FAILURE                                            NT_STATUS = 0xC0000145
	NT_STATUS_PAGEFILE_CREATE_FAILED                                      NT_STATUS = 0xC0000146
	NT_STATUS_NO_PAGEFILE                                                 NT_STATUS = 0xC0000147
	NT_STATUS_INVALID_LEVEL                                               NT_STATUS = 0xC0000148
	NT_STATUS_WRONG_PASSWORD_CORE                                         NT_STATUS = 0xC0000149
	NT_STATUS_ILLEGAL_FLOAT_CONTEXT                                       NT_STATUS = 0xC000014A
	NT_STATUS_PIPE_BROKEN                                                 NT_STATUS = 0xC000014B
	NT_STATUS_REGISTRY_CORRUPT                                            NT_STATUS = 0xC000014C
	NT_STATUS_REGISTRY_IO_FAILED                                          NT_STATUS = 0xC000014D
	NT_STATUS_NO_EVENT_PAIR                                               NT_STATUS = 0xC000014E
	NT_STATUS_UNRECOGNIZED_VOLUME                                         NT_STATUS = 0xC000014F
	NT_STATUS_SERIAL_NO_DEVICE_INITED                                     NT_STATUS = 0xC0000150
	NT_STATUS_NO_SUCH_ALIAS                                               NT_STATUS = 0xC0000151
	NT_STATUS_MEMBER_NOT_IN_ALIAS                                         NT_STATUS = 0xC0000152
	NT_STATUS_MEMBER_IN_ALIAS                                             NT_STATUS = 0xC0000153
	NT_STATUS_ALIAS_EXISTS                                                NT_STATUS = 0xC0000154
	NT_STATUS_LOGON_NOT_GRANTED                                           NT_STATUS = 0xC0000155
	NT_STATUS_TOO_MANY_SECRETS                                            NT_STATUS = 0xC0000156
	NT_STATUS_SECRET_TOO_LONG                                             NT_STATUS = 0xC0000157
	NT_STATUS_INTERNAL_DB_ERROR                                           NT_STATUS = 0xC0000158
	NT_STATUS_FULLSCREEN_MODE                                             NT_STATUS = 0xC0000159
	NT_STATUS_TOO_MANY_CONTEXT_IDS                                        NT_STATUS = 0xC000015A
	NT_STATUS_LOGON_TYPE_NOT_GRANTED                                      NT_STATUS = 0xC000015B
	NT_STATUS_NOT_REGISTRY_FILE                                           NT_STATUS = 0xC000015C
	NT_STATUS_NT_CROSS_ENCRYPTION_REQUIRED                                NT_STATUS = 0xC000015D
	NT_STATUS_DOMAIN_CTRLR_CONFIG_ERROR                                   NT_STATUS = 0xC000015E
	NT_STATUS_FT_MISSING_MEMBER                                           NT_STATUS = 0xC000015F
	NT_STATUS_ILL_FORMED_SERVICE_ENTRY                                    NT_STATUS = 0xC0000160
	NT_STATUS_ILLEGAL_CHARACTER                                           NT_STATUS = 0xC0000161
	NT_STATUS_UNMAPPABLE_CHARACTER                                        NT_STATUS = 0xC0000162
	NT_STATUS_UNDEFINED_CHARACTER                                         NT_STATUS = 0xC0000163
	NT_STATUS_FLOPPY_VOLUME                                               NT_STATUS = 0xC0000164
	NT_STATUS_FLOPPY_ID_MARK_NOT_FOUND                                    NT_STATUS = 0xC0000165
	NT_STATUS_FLOPPY_WRONG_CYLINDER                                       NT_STATUS = 0xC0000166
	NT_STATUS_FLOPPY_UNKNOWN_ERROR                                        NT_STATUS = 0xC0000167
	NT_STATUS_FLOPPY_BAD_REGISTERS                                        NT_STATUS = 0xC0000168
	NT_STATUS_DISK_RECALIBRATE_FAILED                                     NT_STATUS = 0xC0000169
	NT_STATUS_DISK_OPERATION_FAILED                                       NT_STATUS = 0xC000016A
	NT_STATUS_DISK_RESET_FAILED                                           NT_STATUS = 0xC000016B
	NT_STATUS_SHARED_IRQ_BUSY                                             NT_STATUS = 0xC000016C
	NT_STATUS_FT_ORPHANING                                                NT_STATUS = 0xC000016D
	NT_STATUS_BIOS_FAILED_TO_CONNECT_INTERRUPT                            NT_STATUS = 0xC000016E
	NT_STATUS_PARTITION_FAILURE                                           NT_STATUS = 0xC0000172
	NT_STATUS_INVALID_BLOCK_LENGTH                                        NT_STATUS = 0xC0000173
	NT_STATUS_DEVICE_NOT_PARTITIONED                                      NT_STATUS = 0xC0000174
	NT_STATUS_UNABLE_TO_LOCK_MEDIA                                        NT_STATUS = 0xC0000175
	NT_STATUS_UNABLE_TO_UNLOAD_MEDIA                                      NT_STATUS = 0xC0000176
	NT_STATUS_EOM_OVERFLOW                                                NT_STATUS = 0xC0000177
	NT_STATUS_NO_MEDIA                                                    NT_STATUS = 0xC0000178
	NT_STATUS_NO_SUCH_MEMBER                                              NT_STATUS = 0xC000017A
	NT_STATUS_INVALID_MEMBER                                              NT_STATUS = 0xC000017B
	NT_STATUS_KEY_DELETED                                                 NT_STATUS = 0xC000017C
	NT_STATUS_NO_LOG_SPACE                                                NT_STATUS = 0xC000017D
	NT_STATUS_TOO_MANY_SIDS                                               NT_STATUS = 0xC000017E
	NT_STATUS_LM_CROSS_ENCRYPTION_REQUIRED                                NT_STATUS = 0xC000017F
	NT_STATUS_KEY_HAS_CHILDREN                                            NT_STATUS = 0xC0000180
	NT_STATUS_CHILD_MUST_BE_VOLATILE                                      NT_STATUS = 0xC0000181
	NT_STATUS_DEVICE_CONFIGURATION_ERROR                                  NT_STATUS = 0xC0000182
	NT_STATUS_DRIVER_INTERNAL_ERROR                                       NT_STATUS = 0xC0000183
	NT_STATUS_INVALID_DEVICE_STATE                                        NT_STATUS = 0xC0000184
	NT_STATUS_IO_DEVICE_ERROR                                             NT_STATUS = 0xC0000185
	NT_STATUS_DEVICE_PROTOCOL_ERROR                                       NT_STATUS = 0xC0000186
	NT_STATUS_BACKUP_CONTROLLER                                           NT_STATUS = 0xC0000187
	NT_STATUS_LOG_FILE_FULL                                               NT_STATUS = 0xC0000188
	NT_STATUS_TOO_LATE                                                    NT_STATUS = 0xC0000189
	NT_STATUS_NO_TRUST_LSA_SECRET                                         NT_STATUS = 0xC000018A
	NT_STATUS_NO_TRUST_SAM_ACCOUNT                                        NT_STATUS = 0xC000018B
	NT_STATUS_TRUSTED_DOMAIN_FAILURE                                      NT_STATUS = 0xC000018C
	NT_STATUS_TRUSTED_RELATIONSHIP_FAILURE                                NT_STATUS = 0xC000018D
	NT_STATUS_EVENTLOG_FILE_CORRUPT                                       NT_STATUS = 0xC000018E
	NT_STATUS_EVENTLOG_CANT_START                                         NT_STATUS = 0xC000018F
	NT_STATUS_TRUST_FAILURE                                               NT_STATUS = 0xC0000190
	NT_STATUS_MUTANT_LIMIT_EXCEEDED                                       NT_STATUS = 0xC0000191
	NT_STATUS_NETLOGON_NOT_STARTED                                        NT_STATUS = 0xC0000192
	NT_STATUS_ACCOUNT_EXPIRED                                             NT_STATUS = 0xC0000193
	NT_STATUS_POSSIBLE_DEADLOCK                                           NT_STATUS = 0xC0000194
	NT_STATUS_NETWORK_CREDENTIAL_CONFLICT                                 NT_STATUS = 0xC0000195
	NT_STATUS_REMOTE_SESSION_LIMIT                                        NT_STATUS = 0xC0000196
	NT_STATUS_EVENTLOG_FILE_CHANGED                                       NT_STATUS = 0xC0000197
	NT_STATUS_NOLOGON_INTERDOMAIN_TRUST_ACCOUNT                           NT_STATUS = 0xC0000198
	NT_STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT                           NT_STATUS = 0xC0000199
	NT_STATUS_NOLOGON_SERVER_TRUST_ACCOUNT                                NT_STATUS = 0xC000019A
	NT_STATUS_DOMAIN_TRUST_INCONSISTENT                                   NT_STATUS = 0xC000019B
	NT_STATUS_FS_DRIVER_REQUIRED                                          NT_STATUS = 0xC000019C
	NT_STATUS_IMAGE_ALREADY_LOADED_AS_DLL                                 NT_STATUS = 0xC000019D
	NT_STATUS_INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING        NT_STATUS = 0xC000019E
	NT_STATUS_SHORT_NAMES_NOT_ENABLED_ON_VOLUME                           NT_STATUS = 0xC000019F
	NT_STATUS_SECURITY_STREAM_IS_INCONSISTENT                             NT_STATUS = 0xC00001A0
	NT_STATUS_INVALID_LOCK_RANGE                                          NT_STATUS = 0xC00001A1
	NT_STATUS_INVALID_ACE_CONDITION                                       NT_STATUS = 0xC00001A2
	NT_STATUS_IMAGE_SUBSYSTEM_NOT_PRESENT                                 NT_STATUS = 0xC00001A3
	NT_STATUS_NOTIFICATION_GUID_ALREADY_DEFINED                           NT_STATUS = 0xC00001A4
	NT_STATUS_NETWORK_OPEN_RESTRICTION                                    NT_STATUS = 0xC0000201
	NT_STATUS_NO_USER_SESSION_KEY                                         NT_STATUS = 0xC0000202
	NT_STATUS_USER_SESSION_DELETED                                        NT_STATUS = 0xC0000203
	NT_STATUS_RESOURCE_LANG_NOT_FOUND                                     NT_STATUS = 0xC0000204
	NT_STATUS_INSUFF_SERVER_RESOURCES                                     NT_STATUS = 0xC0000205
	NT_STATUS_INVALID_BUFFER_SIZE                                         NT_STATUS = 0xC0000206
	NT_STATUS_INVALID_ADDRESS_COMPONENT                                   NT_STATUS = 0xC0000207
	NT_STATUS_INVALID_ADDRESS_WILDCARD                                    NT_STATUS = 0xC0000208
	NT_STATUS_TOO_MANY_ADDRESSES                                          NT_STATUS = 0xC0000209
	NT_STATUS_ADDRESS_ALREADY_EXISTS                                      NT_STATUS = 0xC000020A
	NT_STATUS_ADDRESS_CLOSED                                              NT_STATUS = 0xC000020B
	NT_STATUS_CONNECTION_DISCONNECTED                                     NT_STATUS = 0xC000020C
	NT_STATUS_CONNECTION_RESET                                            NT_STATUS = 0xC000020D
	NT_STATUS_TOO_MANY_NODES                                              NT_STATUS = 0xC000020E
	NT_STATUS_TRANSACTION_ABORTED                                         NT_STATUS = 0xC000020F
	NT_STATUS_TRANSACTION_TIMED_OUT                                       NT_STATUS = 0xC0000210
	NT_STATUS_TRANSACTION_NO_RELEASE                                      NT_STATUS = 0xC0000211
	NT_STATUS_TRANSACTION_NO_MATCH                                        NT_STATUS = 0xC0000212
	NT_STATUS_TRANSACTION_RESPONDED                                       NT_STATUS = 0xC0000213
	NT_STATUS_TRANSACTION_INVALID_ID                                      NT_STATUS = 0xC0000214
	NT_STATUS_TRANSACTION_INVALID_TYPE                                    NT_STATUS = 0xC0000215
	NT_STATUS_NOT_SERVER_SESSION                                          NT_STATUS = 0xC0000216
	NT_STATUS_NOT_CLIENT_SESSION                                          NT_STATUS = 0xC0000217
	NT_STATUS_CANNOT_LOAD_REGISTRY_FILE                                   NT_STATUS = 0xC0000218
	NT_STATUS_DEBUG_ATTACH_FAILED                                         NT_STATUS = 0xC0000219
	NT_STATUS_SYSTEM_PROCESS_TERMINATED                                   NT_STATUS = 0xC000021A
	NT_STATUS_DATA_NOT_ACCEPTED                                           NT_STATUS = 0xC000021B
	NT_STATUS_NO_BROWSER_SERVERS_FOUND                                    NT_STATUS = 0xC000021C
	NT_STATUS_VDM_HARD_ERROR                                              NT_STATUS = 0xC000021D
	NT_STATUS_DRIVER_CANCEL_TIMEOUT                                       NT_STATUS = 0xC000021E
	NT_STATUS_REPLY_MESSAGE_MISMATCH                                      NT_STATUS = 0xC000021F
	NT_STATUS_MAPPED_ALIGNMENT                                            NT_STATUS = 0xC0000220
	NT_STATUS_IMAGE_CHECKSUM_MISMATCH                                     NT_STATUS = 0xC0000221
	NT_STATUS_LOST_WRITEBEHIND_DATA                                       NT_STATUS = 0xC0000222
	NT_STATUS_CLIENT_SERVER_PARAMETERS_INVALID                            NT_STATUS = 0xC0000223
	NT_STATUS_PASSWORD_MUST_CHANGE                                        NT_STATUS = 0xC0000224
	NT_STATUS_NOT_FOUND                                                   NT_STATUS = 0xC0000225
	NT_STATUS_NOT_TINY_STREAM                                             NT_STATUS = 0xC0000226
	NT_STATUS_RECOVERY_FAILURE                                            NT_STATUS = 0xC0000227
	NT_STATUS_STACK_OVERFLOW_READ                                         NT_STATUS = 0xC0000228
	NT_STATUS_FAIL_CHECK                                                  NT_STATUS = 0xC0000229
	NT_STATUS_DUPLICATE_OBJECTID                                          NT_STATUS = 0xC000022A
	NT_STATUS_OBJECTID_EXISTS                                             NT_STATUS = 0xC000022B
	NT_STATUS_CONVERT_TO_LARGE                                            NT_STATUS = 0xC000022C
	NT_STATUS_RETRY                                                       NT_STATUS = 0xC000022D
	NT_STATUS_FOUND_OUT_OF_SCOPE                                          NT_STATUS = 0xC000022E
	NT_STATUS_ALLOCATE_BUCKET                                             NT_STATUS = 0xC000022F
	NT_STATUS_PROPSET_NOT_FOUND                                           NT_STATUS = 0xC0000230
	NT_STATUS_MARSHALL_OVERFLOW                                           NT_STATUS = 0xC0000231
	NT_STATUS_INVALID_VARIANT                                             NT_STATUS = 0xC0000232
	NT_STATUS_DOMAIN_CONTROLLER_NOT_FOUND                                 NT_STATUS = 0xC0000233
	NT_STATUS_ACCOUNT_LOCKED_OUT                                          NT_STATUS = 0xC0000234
	NT_STATUS_HANDLE_NOT_CLOSABLE                                         NT_STATUS = 0xC0000235
	NT_STATUS_CONNECTION_REFUSED                                          NT_STATUS = 0xC0000236
	NT_STATUS_GRACEFUL_DISCONNECT                                         NT_STATUS = 0xC0000237
	NT_STATUS_ADDRESS_ALREADY_ASSOCIATED                                  NT_STATUS = 0xC0000238
	NT_STATUS_ADDRESS_NOT_ASSOCIATED                                      NT_STATUS = 0xC0000239
	NT_STATUS_CONNECTION_INVALID                                          NT_STATUS = 0xC000023A
	NT_STATUS_CONNECTION_ACTIVE                                           NT_STATUS = 0xC000023B
	NT_STATUS_NETWORK_UNREACHABLE                                         NT_STATUS = 0xC000023C
	NT_STATUS_HOST_UNREACHABLE                                            NT_STATUS = 0xC000023D
	NT_STATUS_PROTOCOL_UNREACHABLE                                        NT_STATUS = 0xC000023E
	NT_STATUS_PORT_UNREACHABLE                                            NT_STATUS = 0xC000023F
	NT_STATUS_REQUEST_ABORTED                                             NT_STATUS = 0xC0000240
	NT_STATUS_CONNECTION_ABORTED                                          NT_STATUS = 0xC0000241
	NT_STATUS_BAD_COMPRESSION_BUFFER                                      NT_STATUS = 0xC0000242
	NT_STATUS_USER_MAPPED_FILE                                            NT_STATUS = 0xC0000243
	NT_STATUS_AUDIT_FAILED                                                NT_STATUS = 0xC0000244
	NT_STATUS_TIMER_RESOLUTION_NOT_SET                                    NT_STATUS = 0xC0000245
	NT_STATUS_CONNECTION_COUNT_LIMIT                                      NT_STATUS = 0xC0000246
	NT_STATUS_LOGIN_TIME_RESTRICTION                                      NT_STATUS = 0xC0000247
	NT_STATUS_LOGIN_WKSTA_RESTRICTION                                     NT_STATUS = 0xC0000248
	NT_STATUS_IMAGE_MP_UP_MISMATCH                                        NT_STATUS = 0xC0000249
	NT_STATUS_INSUFFICIENT_LOGON_INFO                                     NT_STATUS = 0xC0000250
	NT_STATUS_BAD_DLL_ENTRYPOINT                                          NT_STATUS = 0xC0000251
	NT_STATUS_BAD_SERVICE_ENTRYPOINT                                      NT_STATUS = 0xC0000252
	NT_STATUS_LPC_REPLY_LOST                                              NT_STATUS = 0xC0000253
	NT_STATUS_IP_ADDRESS_CONFLICT1                                        NT_STATUS = 0xC0000254
	NT_STATUS_IP_ADDRESS_CONFLICT2                                        NT_STATUS = 0xC0000255
	NT_STATUS_REGISTRY_QUOTA_LIMIT                                        NT_STATUS = 0xC0000256
	NT_STATUS_PATH_NOT_COVERED                                            NT_STATUS = 0xC0000257
	NT_STATUS_NO_CALLBACK_ACTIVE                                          NT_STATUS = 0xC0000258
	NT_STATUS_LICENSE_QUOTA_EXCEEDED                                      NT_STATUS = 0xC0000259
	NT_STATUS_PWD_TOO_SHORT                                               NT_STATUS = 0xC000025A
	NT_STATUS_PWD_TOO_RECENT                                              NT_STATUS = 0xC000025B
	NT_STATUS_PWD_HISTORY_CONFLICT                                        NT_STATUS = 0xC000025C
	NT_STATUS_PLUGPLAY_NO_DEVICE                                          NT_STATUS = 0xC000025E
	NT_STATUS_UNSUPPORTED_COMPRESSION                                     NT_STATUS = 0xC000025F
	NT_STATUS_INVALID_HW_PROFILE                                          NT_STATUS = 0xC0000260
	NT_STATUS_INVALID_PLUGPLAY_DEVICE_PATH                                NT_STATUS = 0xC0000261
	NT_STATUS_DRIVER_ORDINAL_NOT_FOUND                                    NT_STATUS = 0xC0000262
	NT_STATUS_DRIVER_ENTRYPOINT_NOT_FOUND                                 NT_STATUS = 0xC0000263
	NT_STATUS_RESOURCE_NOT_OWNED                                          NT_STATUS = 0xC0000264
	NT_STATUS_TOO_MANY_LINKS                                              NT_STATUS = 0xC0000265
	NT_STATUS_QUOTA_LIST_INCONSISTENT                                     NT_STATUS = 0xC0000266
	NT_STATUS_FILE_IS_OFFLINE                                             NT_STATUS = 0xC0000267
	NT_STATUS_EVALUATION_EXPIRATION                                       NT_STATUS = 0xC0000268
	NT_STATUS_ILLEGAL_DLL_RELOCATION                                      NT_STATUS = 0xC0000269
	NT_STATUS_LICENSE_VIOLATION                                           NT_STATUS = 0xC000026A
	NT_STATUS_DLL_INIT_FAILED_LOGOFF                                      NT_STATUS = 0xC000026B
	NT_STATUS_DRIVER_UNABLE_TO_LOAD                                       NT_STATUS = 0xC000026C
	NT_STATUS_DFS_UNAVAILABLE                                             NT_STATUS = 0xC000026D
	NT_STATUS_VOLUME_DISMOUNTED                                           NT_STATUS = 0xC000026E
	NT_STATUS_WX86_INTERNAL_ERROR                                         NT_STATUS = 0xC000026F
	NT_STATUS_WX86_FLOAT_STACK_CHECK                                      NT_STATUS = 0xC0000270
	NT_STATUS_VALIDATE_CONTINUE                                           NT_STATUS = 0xC0000271
	NT_STATUS_NO_MATCH                                                    NT_STATUS = 0xC0000272
	NT_STATUS_NO_MORE_MATCHES                                             NT_STATUS = 0xC0000273
	NT_STATUS_NOT_A_REPARSE_POINT                                         NT_STATUS = 0xC0000275
	NT_STATUS_IO_REPARSE_TAG_INVALID                                      NT_STATUS = 0xC0000276
	NT_STATUS_IO_REPARSE_TAG_MISMATCH                                     NT_STATUS = 0xC0000277
	NT_STATUS_IO_REPARSE_DATA_INVALID                                     NT_STATUS = 0xC0000278
	NT_STATUS_IO_REPARSE_TAG_NOT_HANDLED                                  NT_STATUS = 0xC0000279
	NT_STATUS_REPARSE_POINT_NOT_RESOLVED                                  NT_STATUS = 0xC0000280
	NT_STATUS_DIRECTORY_IS_A_REPARSE_POINT                                NT_STATUS = 0xC0000281
	NT_STATUS_RANGE_LIST_CONFLICT                                         NT_STATUS = 0xC0000282
	NT_STATUS_SOURCE_ELEMENT_EMPTY                                        NT_STATUS = 0xC0000283
	NT_STATUS_DESTINATION_ELEMENT_FULL                                    NT_STATUS = 0xC0000284
	NT_STATUS_ILLEGAL_ELEMENT_ADDRESS                                     NT_STATUS = 0xC0000285
	NT_STATUS_MAGAZINE_NOT_PRESENT                                        NT_STATUS = 0xC0000286
	NT_STATUS_REINITIALIZATION_NEEDED                                     NT_STATUS = 0xC0000287
	NT_STATUS_ENCRYPTION_FAILED                                           NT_STATUS = 0xC000028A
	NT_STATUS_DECRYPTION_FAILED                                           NT_STATUS = 0xC000028B
	NT_STATUS_RANGE_NOT_FOUND                                             NT_STATUS = 0xC000028C
	NT_STATUS_NO_RECOVERY_POLICY                                          NT_STATUS = 0xC000028D
	NT_STATUS_NO_EFS                                                      NT_STATUS = 0xC000028E
	NT_STATUS_WRONG_EFS                                                   NT_STATUS = 0xC000028F
	NT_STATUS_NO_USER_KEYS                                                NT_STATUS = 0xC0000290
	NT_STATUS_FILE_NOT_ENCRYPTED                                          NT_STATUS = 0xC0000291
	NT_STATUS_NOT_EXPORT_FORMAT                                           NT_STATUS = 0xC0000292
	NT_STATUS_FILE_ENCRYPTED                                              NT_STATUS = 0xC0000293
	NT_STATUS_WMI_GUID_NOT_FOUND                                          NT_STATUS = 0xC0000295
	NT_STATUS_WMI_INSTANCE_NOT_FOUND                                      NT_STATUS = 0xC0000296
	NT_STATUS_WMI_ITEMID_NOT_FOUND                                        NT_STATUS = 0xC0000297
	NT_STATUS_WMI_TRY_AGAIN                                               NT_STATUS = 0xC0000298
	NT_STATUS_SHARED_POLICY                                               NT_STATUS = 0xC0000299
	NT_STATUS_POLICY_OBJECT_NOT_FOUND                                     NT_STATUS = 0xC000029A
	NT_STATUS_POLICY_ONLY_IN_DS                                           NT_STATUS = 0xC000029B
	NT_STATUS_VOLUME_NOT_UPGRADED                                         NT_STATUS = 0xC000029C
	NT_STATUS_REMOTE_STORAGE_NOT_ACTIVE                                   NT_STATUS = 0xC000029D
	NT_STATUS_REMOTE_STORAGE_MEDIA_ERROR                                  NT_STATUS = 0xC000029E
	NT_STATUS_NO_TRACKING_SERVICE                                         NT_STATUS = 0xC000029F
	NT_STATUS_SERVER_SID_MISMATCH                                         NT_STATUS = 0xC00002A0
	NT_STATUS_DS_NO_ATTRIBUTE_OR_VALUE                                    NT_STATUS = 0xC00002A1
	NT_STATUS_DS_INVALID_ATTRIBUTE_SYNTAX                                 NT_STATUS = 0xC00002A2
	NT_STATUS_DS_ATTRIBUTE_TYPE_UNDEFINED                                 NT_STATUS = 0xC00002A3
	NT_STATUS_DS_ATTRIBUTE_OR_VALUE_EXISTS                                NT_STATUS = 0xC00002A4
	NT_STATUS_DS_BUSY                                                     NT_STATUS = 0xC00002A5
	NT_STATUS_DS_UNAVAILABLE                                              NT_STATUS = 0xC00002A6
	NT_STATUS_DS_NO_RIDS_ALLOCATED                                        NT_STATUS = 0xC00002A7
	NT_STATUS_DS_NO_MORE_RIDS                                             NT_STATUS = 0xC00002A8
	NT_STATUS_DS_INCORRECT_ROLE_OWNER                                     NT_STATUS = 0xC00002A9
	NT_STATUS_DS_RIDMGR_INIT_ERROR                                        NT_STATUS = 0xC00002AA
	NT_STATUS_DS_OBJ_CLASS_VIOLATION                                      NT_STATUS = 0xC00002AB
	NT_STATUS_DS_CANT_ON_NON_LEAF                                         NT_STATUS = 0xC00002AC
	NT_STATUS_DS_CANT_ON_RDN                                              NT_STATUS = 0xC00002AD
	NT_STATUS_DS_CANT_MOD_OBJ_CLASS                                       NT_STATUS = 0xC00002AE
	NT_STATUS_DS_CROSS_DOM_MOVE_FAILED                                    NT_STATUS = 0xC00002AF
	NT_STATUS_DS_GC_NOT_AVAILABLE                                         NT_STATUS = 0xC00002B0
	NT_STATUS_DIRECTORY_SERVICE_REQUIRED                                  NT_STATUS = 0xC00002B1
	NT_STATUS_REPARSE_ATTRIBUTE_CONFLICT                                  NT_STATUS = 0xC00002B2
	NT_STATUS_CANT_ENABLE_DENY_ONLY                                       NT_STATUS = 0xC00002B3
	NT_STATUS_FLOAT_MULTIPLE_FAULTS                                       NT_STATUS = 0xC00002B4
	NT_STATUS_FLOAT_MULTIPLE_TRAPS                                        NT_STATUS = 0xC00002B5
	NT_STATUS_DEVICE_REMOVED                                              NT_STATUS = 0xC00002B6
	NT_STATUS_JOURNAL_DELETE_IN_PROGRESS                                  NT_STATUS = 0xC00002B7
	NT_STATUS_JOURNAL_NOT_ACTIVE                                          NT_STATUS = 0xC00002B8
	NT_STATUS_NOINTERFACE                                                 NT_STATUS = 0xC00002B9
	NT_STATUS_DS_ADMIN_LIMIT_EXCEEDED                                     NT_STATUS = 0xC00002C1
	NT_STATUS_DRIVER_FAILED_SLEEP                                         NT_STATUS = 0xC00002C2
	NT_STATUS_MUTUAL_AUTHENTICATION_FAILED                                NT_STATUS = 0xC00002C3
	NT_STATUS_CORRUPT_SYSTEM_FILE                                         NT_STATUS = 0xC00002C4
	NT_STATUS_DATATYPE_MISALIGNMENT_ERROR                                 NT_STATUS = 0xC00002C5
	NT_STATUS_WMI_READ_ONLY                                               NT_STATUS = 0xC00002C6
	NT_STATUS_WMI_SET_FAILURE                                             NT_STATUS = 0xC00002C7
	NT_STATUS_COMMITMENT_MINIMUM                                          NT_STATUS = 0xC00002C8
	NT_STATUS_REG_NAT_CONSUMPTION                                         NT_STATUS = 0xC00002C9
	NT_STATUS_TRANSPORT_FULL                                              NT_STATUS = 0xC00002CA
	NT_STATUS_DS_SAM_INIT_FAILURE                                         NT_STATUS = 0xC00002CB
	NT_STATUS_ONLY_IF_CONNECTED                                           NT_STATUS = 0xC00002CC
	NT_STATUS_DS_SENSITIVE_GROUP_VIOLATION                                NT_STATUS = 0xC00002CD
	NT_STATUS_PNP_RESTART_ENUMERATION                                     NT_STATUS = 0xC00002CE
	NT_STATUS_JOURNAL_ENTRY_DELETED                                       NT_STATUS = 0xC00002CF
	NT_STATUS_DS_CANT_MOD_PRIMARYGROUPID                                  NT_STATUS = 0xC00002D0
	NT_STATUS_SYSTEM_IMAGE_BAD_SIGNATURE                                  NT_STATUS = 0xC00002D1
	NT_STATUS_PNP_REBOOT_REQUIRED                                         NT_STATUS = 0xC00002D2
	NT_STATUS_POWER_STATE_INVALID                                         NT_STATUS = 0xC00002D3
	NT_STATUS_DS_INVALID_GROUP_TYPE                                       NT_STATUS = 0xC00002D4
	NT_STATUS_DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN                       NT_STATUS = 0xC00002D5
	NT_STATUS_DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN                        NT_STATUS = 0xC00002D6
	NT_STATUS_DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER                            NT_STATUS = 0xC00002D7
	NT_STATUS_DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER                        NT_STATUS = 0xC00002D8
	NT_STATUS_DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER                         NT_STATUS = 0xC00002D9
	NT_STATUS_DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER                      NT_STATUS = 0xC00002DA
	NT_STATUS_DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER                 NT_STATUS = 0xC00002DB
	NT_STATUS_DS_HAVE_PRIMARY_MEMBERS                                     NT_STATUS = 0xC00002DC
	NT_STATUS_WMI_NOT_SUPPORTED                                           NT_STATUS = 0xC00002DD
	NT_STATUS_INSUFFICIENT_POWER                                          NT_STATUS = 0xC00002DE
	NT_STATUS_SAM_NEED_BOOTKEY_PASSWORD                                   NT_STATUS = 0xC00002DF
	NT_STATUS_SAM_NEED_BOOTKEY_FLOPPY                                     NT_STATUS = 0xC00002E0
	NT_STATUS_DS_CANT_START                                               NT_STATUS = 0xC00002E1
	NT_STATUS_DS_INIT_FAILURE                                             NT_STATUS = 0xC00002E2
	NT_STATUS_SAM_INIT_FAILURE                                            NT_STATUS = 0xC00002E3
	NT_STATUS_DS_GC_REQUIRED                                              NT_STATUS = 0xC00002E4
	NT_STATUS_DS_LOCAL_MEMBER_OF_LOCAL_ONLY                               NT_STATUS = 0xC00002E5
	NT_STATUS_DS_NO_FPO_IN_UNIVERSAL_GROUPS                               NT_STATUS = 0xC00002E6
	NT_STATUS_DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED                           NT_STATUS = 0xC00002E7
	NT_STATUS_CURRENT_DOMAIN_NOT_ALLOWED                                  NT_STATUS = 0xC00002E9
	NT_STATUS_CANNOT_MAKE                                                 NT_STATUS = 0xC00002EA
	NT_STATUS_SYSTEM_SHUTDOWN                                             NT_STATUS = 0xC00002EB
	NT_STATUS_DS_INIT_FAILURE_CONSOLE                                     NT_STATUS = 0xC00002EC
	NT_STATUS_DS_SAM_INIT_FAILURE_CONSOLE                                 NT_STATUS = 0xC00002ED
	NT_STATUS_UNFINISHED_CONTEXT_DELETED                                  NT_STATUS = 0xC00002EE
	NT_STATUS_NO_TGT_REPLY                                                NT_STATUS = 0xC00002EF
	NT_STATUS_OBJECTID_NOT_FOUND                                          NT_STATUS = 0xC00002F0
	NT_STATUS_NO_IP_ADDRESSES                                             NT_STATUS = 0xC00002F1
	NT_STATUS_WRONG_CREDENTIAL_HANDLE                                     NT_STATUS = 0xC00002F2
	NT_STATUS_CRYPTO_SYSTEM_INVALID                                       NT_STATUS = 0xC00002F3
	NT_STATUS_MAX_REFERRALS_EXCEEDED                                      NT_STATUS = 0xC00002F4
	NT_STATUS_MUST_BE_KDC                                                 NT_STATUS = 0xC00002F5
	NT_STATUS_STRONG_CRYPTO_NOT_SUPPORTED                                 NT_STATUS = 0xC00002F6
	NT_STATUS_TOO_MANY_PRINCIPALS                                         NT_STATUS = 0xC00002F7
	NT_STATUS_NO_PA_DATA                                                  NT_STATUS = 0xC00002F8
	NT_STATUS_PKINIT_NAME_MISMATCH                                        NT_STATUS = 0xC00002F9
	NT_STATUS_SMARTCARD_LOGON_REQUIRED                                    NT_STATUS = 0xC00002FA
	NT_STATUS_KDC_INVALID_REQUEST                                         NT_STATUS = 0xC00002FB
	NT_STATUS_KDC_UNABLE_TO_REFER                                         NT_STATUS = 0xC00002FC
	NT_STATUS_KDC_UNKNOWN_ETYPE                                           NT_STATUS = 0xC00002FD
	NT_STATUS_SHUTDOWN_IN_PROGRESS                                        NT_STATUS = 0xC00002FE
	NT_STATUS_SERVER_SHUTDOWN_IN_PROGRESS                                 NT_STATUS = 0xC00002FF
	NT_STATUS_NOT_SUPPORTED_ON_SBS                                        NT_STATUS = 0xC0000300
	NT_STATUS_WMI_GUID_DISCONNECTED                                       NT_STATUS = 0xC0000301
	NT_STATUS_WMI_ALREADY_DISABLED                                        NT_STATUS = 0xC0000302
	NT_STATUS_WMI_ALREADY_ENABLED                                         NT_STATUS = 0xC0000303
	NT_STATUS_MFT_TOO_FRAGMENTED                                          NT_STATUS = 0xC0000304
	NT_STATUS_COPY_PROTECTION_FAILURE                                     NT_STATUS = 0xC0000305
	NT_STATUS_CSS_AUTHENTICATION_FAILURE                                  NT_STATUS = 0xC0000306
	NT_STATUS_CSS_KEY_NOT_PRESENT                                         NT_STATUS = 0xC0000307
	NT_STATUS_CSS_KEY_NOT_ESTABLISHED                                     NT_STATUS = 0xC0000308
	NT_STATUS_CSS_SCRAMBLED_SECTOR                                        NT_STATUS = 0xC0000309
	NT_STATUS_CSS_REGION_MISMATCH                                         NT_STATUS = 0xC000030A
	NT_STATUS_CSS_RESETS_EXHAUSTED                                        NT_STATUS = 0xC000030B
	NT_STATUS_PKINIT_FAILURE                                              NT_STATUS = 0xC0000320
	NT_STATUS_SMARTCARD_SUBSYSTEM_FAILURE                                 NT_STATUS = 0xC0000321
	NT_STATUS_NO_KERB_KEY                                                 NT_STATUS = 0xC0000322
	NT_STATUS_HOST_DOWN                                                   NT_STATUS = 0xC0000350
	NT_STATUS_UNSUPPORTED_PREAUTH                                         NT_STATUS = 0xC0000351
	NT_STATUS_EFS_ALG_BLOB_TOO_BIG                                        NT_STATUS = 0xC0000352
	NT_STATUS_PORT_NOT_SET                                                NT_STATUS = 0xC0000353
	NT_STATUS_DEBUGGER_INACTIVE                                           NT_STATUS = 0xC0000354
	NT_STATUS_DS_VERSION_CHECK_FAILURE                                    NT_STATUS = 0xC0000355
	NT_STATUS_AUDITING_DISABLED                                           NT_STATUS = 0xC0000356
	NT_STATUS_PRENT4_MACHINE_ACCOUNT                                      NT_STATUS = 0xC0000357
	NT_STATUS_DS_AG_CANT_HAVE_UNIVERSAL_MEMBER                            NT_STATUS = 0xC0000358
	NT_STATUS_INVALID_IMAGE_WIN_32                                        NT_STATUS = 0xC0000359
	NT_STATUS_INVALID_IMAGE_WIN_64                                        NT_STATUS = 0xC000035A
	NT_STATUS_BAD_BINDINGS                                                NT_STATUS = 0xC000035B
	NT_STATUS_NETWORK_SESSION_EXPIRED                                     NT_STATUS = 0xC000035C
	NT_STATUS_APPHELP_BLOCK                                               NT_STATUS = 0xC000035D
	NT_STATUS_ALL_SIDS_FILTERED                                           NT_STATUS = 0xC000035E
	NT_STATUS_NOT_SAFE_MODE_DRIVER                                        NT_STATUS = 0xC000035F
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_DEFAULT                           NT_STATUS = 0xC0000361
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PATH                              NT_STATUS = 0xC0000362
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PUBLISHER                         NT_STATUS = 0xC0000363
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_OTHER                             NT_STATUS = 0xC0000364
	NT_STATUS_FAILED_DRIVER_ENTRY                                         NT_STATUS = 0xC0000365
	NT_STATUS_DEVICE_ENUMERATION_ERROR                                    NT_STATUS = 0xC0000366
	NT_STATUS_MOUNT_POINT_NOT_RESOLVED                                    NT_STATUS = 0xC0000368
	NT_STATUS_INVALID_DEVICE_OBJECT_PARAMETER                             NT_STATUS = 0xC0000369
	NT_STATUS_MCA_OCCURED                                                 NT_STATUS = 0xC000036A
	NT_STATUS_DRIVER_BLOCKED_CRITICAL                                     NT_STATUS = 0xC000036B
	NT_STATUS_DRIVER_BLOCKED                                              NT_STATUS = 0xC000036C
	NT_STATUS_DRIVER_DATABASE_ERROR                                       NT_STATUS = 0xC000036D
	NT_STATUS_SYSTEM_HIVE_TOO_LARGE                                       NT_STATUS = 0xC000036E
	NT_STATUS_INVALID_IMPORT_OF_NON_DLL                                   NT_STATUS = 0xC000036F
	NT_STATUS_NO_SECRETS                                                  NT_STATUS = 0xC0000371
	NT_STATUS_ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY                       NT_STATUS = 0xC0000372
	NT_STATUS_FAILED_STACK_SWITCH                                         NT_STATUS = 0xC0000373
	NT_STATUS_HEAP_CORRUPTION                                             NT_STATUS = 0xC0000374
	NT_STATUS_SMARTCARD_WRONG_PIN                                         NT_STATUS = 0xC0000380
	NT_STATUS_SMARTCARD_CARD_BLOCKED                                      NT_STATUS = 0xC0000381
	NT_STATUS_SMARTCARD_CARD_NOT_AUTHENTICATED                            NT_STATUS = 0xC0000382
	NT_STATUS_SMARTCARD_NO_CARD                                           NT_STATUS = 0xC0000383
	NT_STATUS_SMARTCARD_NO_KEY_CONTAINER                                  NT_STATUS = 0xC0000384
	NT_STATUS_SMARTCARD_NO_CERTIFICATE                                    NT_STATUS = 0xC0000385
	NT_STATUS_SMARTCARD_NO_KEYSET                                         NT_STATUS = 0xC0000386
	NT_STATUS_SMARTCARD_IO_ERROR                                          NT_STATUS = 0xC0000387
	NT_STATUS_DOWNGRADE_DETECTED                                          NT_STATUS = 0xC0000388
	NT_STATUS_SMARTCARD_CERT_REVOKED                                      NT_STATUS = 0xC0000389
	NT_STATUS_ISSUING_CA_UNTRUSTED                                        NT_STATUS = 0xC000038A
	NT_STATUS_REVOCATION_OFFLINE_C                                        NT_STATUS = 0xC000038B
	NT_STATUS_PKINIT_CLIENT_FAILURE                                       NT_STATUS = 0xC000038C
	NT_STATUS_SMARTCARD_CERT_EXPIRED                                      NT_STATUS = 0xC000038D
	NT_STATUS_DRIVER_FAILED_PRIOR_UNLOAD                                  NT_STATUS = 0xC000038E
	NT_STATUS_SMARTCARD_SILENT_CONTEXT                                    NT_STATUS = 0xC000038F
	NT_STATUS_PER_USER_TRUST_QUOTA_EXCEEDED                               NT_STATUS = 0xC0000401
	NT_STATUS_ALL_USER_TRUST_QUOTA_EXCEEDED                               NT_STATUS = 0xC0000402
	NT_STATUS_USER_DELETE_TRUST_QUOTA_EXCEEDED                            NT_STATUS = 0xC0000403
	NT_STATUS_DS_NAME_NOT_UNIQUE                                          NT_STATUS = 0xC0000404
	NT_STATUS_DS_DUPLICATE_ID_FOUND                                       NT_STATUS = 0xC0000405
	NT_STATUS_DS_GROUP_CONVERSION_ERROR                                   NT_STATUS = 0xC0000406
	NT_STATUS_VOLSNAP_PREPARE_HIBERNATE                                   NT_STATUS = 0xC0000407
	NT_STATUS_USER2USER_REQUIRED                                          NT_STATUS = 0xC0000408
	NT_STATUS_STACK_BUFFER_OVERRUN                                        NT_STATUS = 0xC0000409
	NT_STATUS_NO_S4U_PROT_SUPPORT                                         NT_STATUS = 0xC000040A
	NT_STATUS_CROSSREALM_DELEGATION_FAILURE                               NT_STATUS = 0xC000040B
	NT_STATUS_REVOCATION_OFFLINE_KDC                                      NT_STATUS = 0xC000040C
	NT_STATUS_ISSUING_CA_UNTRUSTED_KDC                                    NT_STATUS = 0xC000040D
	NT_STATUS_KDC_CERT_EXPIRED                                            NT_STATUS = 0xC000040E
	NT_STATUS_KDC_CERT_REVOKED                                            NT_STATUS = 0xC000040F
	NT_STATUS_PARAMETER_QUOTA_EXCEEDED                                    NT_STATUS = 0xC0000410
	NT_STATUS_HIBERNATION_FAILURE                                         NT_STATUS = 0xC0000411
	NT_STATUS_DELAY_LOAD_FAILED                                           NT_STATUS = 0xC0000412
	NT_STATUS_AUTHENTICATION_FIREWALL_FAILED                              NT_STATUS = 0xC0000413
	NT_STATUS_VDM_DISALLOWED                                              NT_STATUS = 0xC0000414
	NT_STATUS_HUNG_DISPLAY_DRIVER_THREAD                                  NT_STATUS = 0xC0000415
	NT_STATUS_INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE     NT_STATUS = 0xC0000416
	NT_STATUS_INVALID_CRUNTIME_PARAMETER                                  NT_STATUS = 0xC0000417
	NT_STATUS_NTLM_BLOCKED                                                NT_STATUS = 0xC0000418
	NT_STATUS_DS_SRC_SID_EXISTS_IN_FOREST                                 NT_STATUS = 0xC0000419
	NT_STATUS_DS_DOMAIN_NAME_EXISTS_IN_FOREST                             NT_STATUS = 0xC000041A
	NT_STATUS_DS_FLAT_NAME_EXISTS_IN_FOREST                               NT_STATUS = 0xC000041B
	NT_STATUS_INVALID_USER_PRINCIPAL_NAME                                 NT_STATUS = 0xC000041C
	NT_STATUS_ASSERTION_FAILURE                                           NT_STATUS = 0xC0000420
	NT_STATUS_VERIFIER_STOP                                               NT_STATUS = 0xC0000421
	NT_STATUS_CALLBACK_POP_STACK                                          NT_STATUS = 0xC0000423
	NT_STATUS_INCOMPATIBLE_DRIVER_BLOCKED                                 NT_STATUS = 0xC0000424
	NT_STATUS_HIVE_UNLOADED                                               NT_STATUS = 0xC0000425
	NT_STATUS_COMPRESSION_DISABLED                                        NT_STATUS = 0xC0000426
	NT_STATUS_FILE_SYSTEM_LIMITATION                                      NT_STATUS = 0xC0000427
	NT_STATUS_INVALID_IMAGE_HASH                                          NT_STATUS = 0xC0000428
	NT_STATUS_NOT_CAPABLE                                                 NT_STATUS = 0xC0000429
	NT_STATUS_REQUEST_OUT_OF_SEQUENCE                                     NT_STATUS = 0xC000042A
	NT_STATUS_IMPLEMENTATION_LIMIT                                        NT_STATUS = 0xC000042B
	NT_STATUS_ELEVATION_REQUIRED                                          NT_STATUS = 0xC000042C
	NT_STATUS_NO_SECURITY_CONTEXT                                         NT_STATUS = 0xC000042D
	NT_STATUS_PKU2U_CERT_FAILURE                                          NT_STATUS = 0xC000042E
	NT_STATUS_BEYOND_VDL                                                  NT_STATUS = 0xC0000432
	NT_STATUS_ENCOUNTERED_WRITE_IN_PROGRESS                               NT_STATUS = 0xC0000433
	NT_STATUS_PTE_CHANGED                                                 NT_STATUS = 0xC0000434
	NT_STATUS_PURGE_FAILED                                                NT_STATUS = 0xC0000435
	NT_STATUS_CRED_REQUIRES_CONFIRMATION                                  NT_STATUS = 0xC0000440
	NT_STATUS_CS_ENCRYPTION_INVALID_SERVER_RESPONSE                       NT_STATUS = 0xC0000441
	NT_STATUS_CS_ENCRYPTION_UNSUPPORTED_SERVER                            NT_STATUS = 0xC0000442
	NT_STATUS_CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE                       NT_STATUS = 0xC0000443
	NT_STATUS_CS_ENCRYPTION_NEW_ENCRYPTED_FILE                            NT_STATUS = 0xC0000444
	NT_STATUS_CS_ENCRYPTION_FILE_NOT_CSE                                  NT_STATUS = 0xC0000445
	NT_STATUS_INVALID_LABEL                                               NT_STATUS = 0xC0000446
	NT_STATUS_DRIVER_PROCESS_TERMINATED                                   NT_STATUS = 0xC0000450
	NT_STATUS_AMBIGUOUS_SYSTEM_DEVICE                                     NT_STATUS = 0xC0000451
	NT_STATUS_SYSTEM_DEVICE_NOT_FOUND                                     NT_STATUS = 0xC0000452
	NT_STATUS_RESTART_BOOT_APPLICATION                                    NT_STATUS = 0xC0000453
	NT_STATUS_INSUFFICIENT_NVRAM_RESOURCES                                NT_STATUS = 0xC0000454
	NT_STATUS_NO_RANGES_PROCESSED                                         NT_STATUS = 0xC0000460
	NT_STATUS_DEVICE_FEATURE_NOT_SUPPORTED                                NT_STATUS = 0xC0000463
	NT_STATUS_DEVICE_UNREACHABLE                                          NT_STATUS = 0xC0000464
	NT_STATUS_INVALID_TOKEN                                               NT_STATUS = 0xC0000465
	NT_STATUS_SERVER_UNAVAILABLE                                          NT_STATUS = 0xC0000466
	NT_STATUS_INVALID_TASK_NAME                                           NT_STATUS = 0xC0000500
	NT_STATUS_INVALID_TASK_INDEX                                          NT_STATUS = 0xC0000501
	NT_STATUS_THREAD_ALREADY_IN_TASK                                      NT_STATUS = 0xC0000502
	NT_STATUS_CALLBACK_BYPASS                                             NT_STATUS = 0xC0000503
	NT_STATUS_FAIL_FAST_EXCEPTION                                         NT_STATUS = 0xC0000602
	NT_STATUS_IMAGE_CERT_REVOKED                                          NT_STATUS = 0xC0000603
	NT_STATUS_PORT_CLOSED                                                 NT_STATUS = 0xC0000700
	NT_STATUS_MESSAGE_LOST                                                NT_STATUS = 0xC0000701
	NT_STATUS_INVALID_MESSAGE                                             NT_STATUS = 0xC0000702
	NT_STATUS_REQUEST_CANCELED                                            NT_STATUS = 0xC0000703
	NT_STATUS_RECURSIVE_DISPATCH                                          NT_STATUS = 0xC0000704
	NT_STATUS_LPC_RECEIVE_BUFFER_EXPECTED                                 NT_STATUS = 0xC0000705
	NT_STATUS_LPC_INVALID_CONNECTION_USAGE                                NT_STATUS = 0xC0000706
	NT_STATUS_LPC_REQUESTS_NOT_ALLOWED                                    NT_STATUS = 0xC0000707
	NT_STATUS_RESOURCE_IN_USE                                             NT_STATUS = 0xC0000708
	NT_STATUS_HARDWARE_MEMORY_ERROR                                       NT_STATUS = 0xC0000709
	NT_STATUS_THREADPOOL_HANDLE_EXCEPTION                                 NT_STATUS = 0xC000070A
	NT_STATUS_THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED                   NT_STATUS = 0xC000070B
	NT_STATUS_THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED           NT_STATUS = 0xC000070C
	NT_STATUS_THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED               NT_STATUS = 0xC000070D
	NT_STATUS_THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED                NT_STATUS = 0xC000070E
	NT_STATUS_THREADPOOL_RELEASED_DURING_OPERATION                        NT_STATUS = 0xC000070F
	NT_STATUS_CALLBACK_RETURNED_WHILE_IMPERSONATING                       NT_STATUS = 0xC0000710
	NT_STATUS_APC_RETURNED_WHILE_IMPERSONATING                            NT_STATUS = 0xC0000711
	NT_STATUS_PROCESS_IS_PROTECTED                                        NT_STATUS = 0xC0000712
	NT_STATUS_MCA_EXCEPTION                                               NT_STATUS = 0xC0000713
	NT_STATUS_CERTIFICATE_MAPPING_NOT_UNIQUE                              NT_STATUS = 0xC0000714
	NT_STATUS_SYMLINK_CLASS_DISABLED                                      NT_STATUS = 0xC0000715
	NT_STATUS_INVALID_IDN_NORMALIZATION                                   NT_STATUS = 0xC0000716
	NT_STATUS_NO_UNICODE_TRANSLATION                                      NT_STATUS = 0xC0000717
	NT_STATUS_ALREADY_REGISTERED                                          NT_STATUS = 0xC0000718
	NT_STATUS_CONTEXT_MISMATCH                                            NT_STATUS = 0xC0000719
	NT_STATUS_PORT_ALREADY_HAS_COMPLETION_LIST                            NT_STATUS = 0xC000071A
	NT_STATUS_CALLBACK_RETURNED_THREAD_PRIORITY                           NT_STATUS = 0xC000071B
	NT_STATUS_INVALID_THREAD                                              NT_STATUS = 0xC000071C
	NT_STATUS_CALLBACK_RETURNED_TRANSACTION                               NT_STATUS = 0xC000071D
	NT_STATUS_CALLBACK_RETURNED_LDR_LOCK                                  NT_STATUS = 0xC000071E
	NT_STATUS_CALLBACK_RETURNED_LANG                                      NT_STATUS = 0xC000071F
	NT_STATUS_CALLBACK_RETURNED_PRI_BACK                                  NT_STATUS = 0xC0000720
	NT_STATUS_DISK_REPAIR_DISABLED                                        NT_STATUS = 0xC0000800
	NT_STATUS_DS_DOMAIN_RENAME_IN_PROGRESS                                NT_STATUS = 0xC0000801
	NT_STATUS_DISK_QUOTA_EXCEEDED                                         NT_STATUS = 0xC0000802
	NT_STATUS_CONTENT_BLOCKED                                             NT_STATUS = 0xC0000804
	NT_STATUS_BAD_CLUSTERS                                                NT_STATUS = 0xC0000805
	NT_STATUS_VOLUME_DIRTY                                                NT_STATUS = 0xC0000806
	NT_STATUS_FILE_CHECKED_OUT                                            NT_STATUS = 0xC0000901
	NT_STATUS_CHECKOUT_REQUIRED                                           NT_STATUS = 0xC0000902
	NT_STATUS_BAD_FILE_TYPE                                               NT_STATUS = 0xC0000903
	NT_STATUS_FILE_TOO_LARGE                                              NT_STATUS = 0xC0000904
	NT_STATUS_FORMS_AUTH_REQUIRED                                         NT_STATUS = 0xC0000905
	NT_STATUS_VIRUS_INFECTED                                              NT_STATUS = 0xC0000906
	NT_STATUS_VIRUS_DELETED                                               NT_STATUS = 0xC0000907
	NT_STATUS_BAD_MCFG_TABLE                                              NT_STATUS = 0xC0000908
	NT_STATUS_BAD_DATA                                                    NT_STATUS = 0xC000090B
	NT_STATUS_CANNOT_BREAK_OPLOCK                                         NT_STATUS = 0xC0000909
	NT_STATUS_WOW_ASSERTION                                               NT_STATUS = 0xC0009898
	NT_STATUS_INVALID_SIGNATURE                                           NT_STATUS = 0xC000A000
	NT_STATUS_HMAC_NOT_SUPPORTED                                          NT_STATUS = 0xC000A001
	NT_STATUS_AUTH_TAG_MISMATCH                                           NT_STATUS = 0xC000A002
	NT_STATUS_IPSEC_QUEUE_OVERFLOW                                        NT_STATUS = 0xC000A010
	NT_STATUS_ND_QUEUE_OVERFLOW                                           NT_STATUS = 0xC000A011
	NT_STATUS_HOPLIMIT_EXCEEDED                                           NT_STATUS = 0xC000A012
	NT_STATUS_PROTOCOL_NOT_SUPPORTED                                      NT_STATUS = 0xC000A013
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED                  NT_STATUS = 0xC000A080
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR                  NT_STATUS = 0xC000A081
	NT_STATUS_LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR                      NT_STATUS = 0xC000A082
	NT_STATUS_XML_PARSE_ERROR                                             NT_STATUS = 0xC000A083
	NT_STATUS_XMLDSIG_ERROR                                               NT_STATUS = 0xC000A084
	NT_STATUS_WRONG_COMPARTMENT                                           NT_STATUS = 0xC000A085
	NT_STATUS_AUTHIP_FAILURE                                              NT_STATUS = 0xC000A086
	NT_STATUS_DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS                       NT_STATUS = 0xC000A087
	NT_STATUS_DS_OID_NOT_FOUND                                            NT_STATUS = 0xC000A088
	NT_STATUS_HASH_NOT_SUPPORTED                                          NT_STATUS = 0xC000A100
	NT_STATUS_HASH_NOT_PRESENT                                            NT_STATUS = 0xC000A101
	NT_STATUS_OFFLOAD_READ_FLT_NOT_SUPPORTED                              NT_STATUS = 0xC000A2A1
	NT_STATUS_OFFLOAD_WRITE_FLT_NOT_SUPPORTED                             NT_STATUS = 0xC000A2A2
	NT_STATUS_OFFLOAD_READ_FILE_NOT_SUPPORTED                             NT_STATUS = 0xC000A2A3
	NT_STATUS_OFFLOAD_WRITE_FILE_NOT_SUPPORTED                            NT_STATUS = 0xC000A2A4
	NT_STATUS_DBG_NO_STATE_CHANGE                                         NT_STATUS = 0xC0010001
	NT_STATUS_DBG_APP_NOT_IDLE                                            NT_STATUS = 0xC0010002
	NT_STATUS_RPC_NT_INVALID_STRING_BINDING                               NT_STATUS = 0xC0020001
	NT_STATUS_RPC_NT_WRONG_KIND_OF_BINDING                                NT_STATUS = 0xC0020002
	NT_STATUS_RPC_NT_INVALID_BINDING                                      NT_STATUS = 0xC0020003
	NT_STATUS_RPC_NT_PROTSEQ_NOT_SUPPORTED                                NT_STATUS = 0xC0020004
	NT_STATUS_RPC_NT_INVALID_RPC_PROTSEQ                                  NT_STATUS = 0xC0020005
	NT_STATUS_RPC_NT_INVALID_STRING_UUID                                  NT_STATUS = 0xC0020006
	NT_STATUS_RPC_NT_INVALID_ENDPOINT_FORMAT                              NT_STATUS = 0xC0020007
	NT_STATUS_RPC_NT_INVALID_NET_ADDR                                     NT_STATUS = 0xC0020008
	NT_STATUS_RPC_NT_NO_ENDPOINT_FOUND                                    NT_STATUS = 0xC0020009
	NT_STATUS_RPC_NT_INVALID_TIMEOUT                                      NT_STATUS = 0xC002000A
	NT_STATUS_RPC_NT_OBJECT_NOT_FOUND                                     NT_STATUS = 0xC002000B
	NT_STATUS_RPC_NT_ALREADY_REGISTERED                                   NT_STATUS = 0xC002000C
	NT_STATUS_RPC_NT_TYPE_ALREADY_REGISTERED                              NT_STATUS = 0xC002000D
	NT_STATUS_RPC_NT_ALREADY_LISTENING                                    NT_STATUS = 0xC002000E
	NT_STATUS_RPC_NT_NO_PROTSEQS_REGISTERED                               NT_STATUS = 0xC002000F
	NT_STATUS_RPC_NT_NOT_LISTENING                                        NT_STATUS = 0xC0020010
	NT_STATUS_RPC_NT_UNKNOWN_MGR_TYPE                                     NT_STATUS = 0xC0020011
	NT_STATUS_RPC_NT_UNKNOWN_IF                                           NT_STATUS = 0xC0020012
	NT_STATUS_RPC_NT_NO_BINDINGS                                          NT_STATUS = 0xC0020013
	NT_STATUS_RPC_NT_NO_PROTSEQS                                          NT_STATUS = 0xC0020014
	NT_STATUS_RPC_NT_CANT_CREATE_ENDPOINT                                 NT_STATUS = 0xC0020015
	NT_STATUS_RPC_NT_OUT_OF_RESOURCES                                     NT_STATUS = 0xC0020016
	NT_STATUS_RPC_NT_SERVER_UNAVAILABLE                                   NT_STATUS = 0xC0020017
	NT_STATUS_RPC_NT_SERVER_TOO_BUSY                                      NT_STATUS = 0xC0020018
	NT_STATUS_RPC_NT_INVALID_NETWORK_OPTIONS                              NT_STATUS = 0xC0020019
	NT_STATUS_RPC_NT_NO_CALL_ACTIVE                                       NT_STATUS = 0xC002001A
	NT_STATUS_RPC_NT_CALL_FAILED                                          NT_STATUS = 0xC002001B
	NT_STATUS_RPC_NT_CALL_FAILED_DNE                                      NT_STATUS = 0xC002001C
	NT_STATUS_RPC_NT_PROTOCOL_ERROR                                       NT_STATUS = 0xC002001D
	NT_STATUS_RPC_NT_UNSUPPORTED_TRANS_SYN                                NT_STATUS = 0xC002001F
	NT_STATUS_RPC_NT_UNSUPPORTED_TYPE                                     NT_STATUS = 0xC0020021
	NT_STATUS_RPC_NT_INVALID_TAG                                          NT_STATUS = 0xC0020022
	NT_STATUS_RPC_NT_INVALID_BOUND                                        NT_STATUS = 0xC0020023
	NT_STATUS_RPC_NT_NO_ENTRY_NAME                                        NT_STATUS = 0xC0020024
	NT_STATUS_RPC_NT_INVALID_NAME_SYNTAX                                  NT_STATUS = 0xC0020025
	NT_STATUS_RPC_NT_UNSUPPORTED_NAME_SYNTAX                              NT_STATUS = 0xC0020026
	NT_STATUS_RPC_NT_UUID_NO_ADDRESS                                      NT_STATUS = 0xC0020028
	NT_STATUS_RPC_NT_DUPLICATE_ENDPOINT                                   NT_STATUS = 0xC0020029
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_TYPE                                   NT_STATUS = 0xC002002A
	NT_STATUS_RPC_NT_MAX_CALLS_TOO_SMALL                                  NT_STATUS = 0xC002002B
	NT_STATUS_RPC_NT_STRING_TOO_LONG                                      NT_STATUS = 0xC002002C
	NT_STATUS_RPC_NT_PROTSEQ_NOT_FOUND                                    NT_STATUS = 0xC002002D
	NT_STATUS_RPC_NT_PROCNUM_OUT_OF_RANGE                                 NT_STATUS = 0xC002002E
	NT_STATUS_RPC_NT_BINDING_HAS_NO_AUTH                                  NT_STATUS = 0xC002002F
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_SERVICE                                NT_STATUS = 0xC0020030
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_LEVEL                                  NT_STATUS = 0xC0020031
	NT_STATUS_RPC_NT_INVALID_AUTH_IDENTITY                                NT_STATUS = 0xC0020032
	NT_STATUS_RPC_NT_UNKNOWN_AUTHZ_SERVICE                                NT_STATUS = 0xC0020033
	NT_STATUS_EPT_NT_INVALID_ENTRY                                        NT_STATUS = 0xC0020034
	NT_STATUS_EPT_NT_CANT_PERFORM_OP                                      NT_STATUS = 0xC0020035
	NT_STATUS_EPT_NT_NOT_REGISTERED                                       NT_STATUS = 0xC0020036
	NT_STATUS_RPC_NT_NOTHING_TO_EXPORT                                    NT_STATUS = 0xC0020037
	NT_STATUS_RPC_NT_INCOMPLETE_NAME                                      NT_STATUS = 0xC0020038
	NT_STATUS_RPC_NT_INVALID_VERS_OPTION                                  NT_STATUS = 0xC0020039
	NT_STATUS_RPC_NT_NO_MORE_MEMBERS                                      NT_STATUS = 0xC002003A
	NT_STATUS_RPC_NT_NOT_ALL_OBJS_UNEXPORTED                              NT_STATUS = 0xC002003B
	NT_STATUS_RPC_NT_INTERFACE_NOT_FOUND                                  NT_STATUS = 0xC002003C
	NT_STATUS_RPC_NT_ENTRY_ALREADY_EXISTS                                 NT_STATUS = 0xC002003D
	NT_STATUS_RPC_NT_ENTRY_NOT_FOUND                                      NT_STATUS = 0xC002003E
	NT_STATUS_RPC_NT_NAME_SERVICE_UNAVAILABLE                             NT_STATUS = 0xC002003F
	NT_STATUS_RPC_NT_INVALID_NAF_ID                                       NT_STATUS = 0xC0020040
	NT_STATUS_RPC_NT_CANNOT_SUPPORT                                       NT_STATUS = 0xC0020041
	NT_STATUS_RPC_NT_NO_CONTEXT_AVAILABLE                                 NT_STATUS = 0xC0020042
	NT_STATUS_RPC_NT_INTERNAL_ERROR                                       NT_STATUS = 0xC0020043
	NT_STATUS_RPC_NT_ZERO_DIVIDE                                          NT_STATUS = 0xC0020044
	NT_STATUS_RPC_NT_ADDRESS_ERROR                                        NT_STATUS = 0xC0020045
	NT_STATUS_RPC_NT_FP_DIV_ZERO                                          NT_STATUS = 0xC0020046
	NT_STATUS_RPC_NT_FP_UNDERFLOW                                         NT_STATUS = 0xC0020047
	NT_STATUS_RPC_NT_FP_OVERFLOW                                          NT_STATUS = 0xC0020048
	NT_STATUS_RPC_NT_CALL_IN_PROGRESS                                     NT_STATUS = 0xC0020049
	NT_STATUS_RPC_NT_NO_MORE_BINDINGS                                     NT_STATUS = 0xC002004A
	NT_STATUS_RPC_NT_GROUP_MEMBER_NOT_FOUND                               NT_STATUS = 0xC002004B
	NT_STATUS_EPT_NT_CANT_CREATE                                          NT_STATUS = 0xC002004C
	NT_STATUS_RPC_NT_INVALID_OBJECT                                       NT_STATUS = 0xC002004D
	NT_STATUS_RPC_NT_NO_INTERFACES                                        NT_STATUS = 0xC002004F
	NT_STATUS_RPC_NT_CALL_CANCELLED                                       NT_STATUS = 0xC0020050
	NT_STATUS_RPC_NT_BINDING_INCOMPLETE                                   NT_STATUS = 0xC0020051
	NT_STATUS_RPC_NT_COMM_FAILURE                                         NT_STATUS = 0xC0020052
	NT_STATUS_RPC_NT_UNSUPPORTED_AUTHN_LEVEL                              NT_STATUS = 0xC0020053
	NT_STATUS_RPC_NT_NO_PRINC_NAME                                        NT_STATUS = 0xC0020054
	NT_STATUS_RPC_NT_NOT_RPC_ERROR                                        NT_STATUS = 0xC0020055
	NT_STATUS_RPC_NT_SEC_PKG_ERROR                                        NT_STATUS = 0xC0020057
	NT_STATUS_RPC_NT_NOT_CANCELLED                                        NT_STATUS = 0xC0020058
	NT_STATUS_RPC_NT_INVALID_ASYNC_HANDLE                                 NT_STATUS = 0xC0020062
	NT_STATUS_RPC_NT_INVALID_ASYNC_CALL                                   NT_STATUS = 0xC0020063
	NT_STATUS_RPC_NT_PROXY_ACCESS_DENIED                                  NT_STATUS = 0xC0020064
	NT_STATUS_RPC_NT_NO_MORE_ENTRIES                                      NT_STATUS = 0xC0030001
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_OPEN_FAIL                              NT_STATUS = 0xC0030002
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_SHORT_FILE                             NT_STATUS = 0xC0030003
	NT_STATUS_RPC_NT_SS_IN_NULL_CONTEXT                                   NT_STATUS = 0xC0030004
	NT_STATUS_RPC_NT_SS_CONTEXT_MISMATCH                                  NT_STATUS = 0xC0030005
	NT_STATUS_RPC_NT_SS_CONTEXT_DAMAGED                                   NT_STATUS = 0xC0030006
	NT_STATUS_RPC_NT_SS_HANDLES_MISMATCH                                  NT_STATUS = 0xC0030007
	NT_STATUS_RPC_NT_SS_CANNOT_GET_CALL_HANDLE                            NT_STATUS = 0xC0030008
	NT_STATUS_RPC_NT_NULL_REF_POINTER                                     NT_STATUS = 0xC0030009
	NT_STATUS_RPC_NT_ENUM_VALUE_OUT_OF_RANGE                              NT_STATUS = 0xC003000A
	NT_STATUS_RPC_NT_BYTE_COUNT_TOO_SMALL                                 NT_STATUS = 0xC003000B
	NT_STATUS_RPC_NT_BAD_STUB_DATA                                        NT_STATUS = 0xC003000C
	NT_STATUS_RPC_NT_INVALID_ES_ACTION                                    NT_STATUS = 0xC0030059
	NT_STATUS_RPC_NT_WRONG_ES_VERSION                                     NT_STATUS = 0xC003005A
	NT_STATUS_RPC_NT_WRONG_STUB_VERSION                                   NT_STATUS = 0xC003005B
	NT_STATUS_RPC_NT_INVALID_PIPE_OBJECT                                  NT_STATUS = 0xC003005C
	NT_STATUS_RPC_NT_INVALID_PIPE_OPERATION                               NT_STATUS = 0xC003005D
	NT_STATUS_RPC_NT_WRONG_PIPE_VERSION                                   NT_STATUS = 0xC003005E
	NT_STATUS_RPC_NT_PIPE_CLOSED                                          NT_STATUS = 0xC003005F
	NT_STATUS_RPC_NT_PIPE_DISCIPLINE_ERROR                                NT_STATUS = 0xC0030060
	NT_STATUS_RPC_NT_PIPE_EMPTY                                           NT_STATUS = 0xC0030061
	NT_STATUS_PNP_BAD_MPS_TABLE                                           NT_STATUS = 0xC0040035
	NT_STATUS_PNP_TRANSLATION_FAILED                                      NT_STATUS = 0xC0040036
	NT_STATUS_PNP_IRQ_TRANSLATION_FAILED                                  NT_STATUS = 0xC0040037
	NT_STATUS_PNP_INVALID_ID                                              NT_STATUS = 0xC0040038
	NT_STATUS_IO_REISSUE_AS_CACHED                                        NT_STATUS = 0xC0040039
	NT_STATUS_CTX_WINSTATION_NAME_INVALID                                 NT_STATUS = 0xC00A0001
	NT_STATUS_CTX_INVALID_PD                                              NT_STATUS = 0xC00A0002
	NT_STATUS_CTX_PD_NOT_FOUND                                            NT_STATUS = 0xC00A0003
	NT_STATUS_CTX_CLOSE_PENDING                                           NT_STATUS = 0xC00A0006
	NT_STATUS_CTX_NO_OUTBUF                                               NT_STATUS = 0xC00A0007
	NT_STATUS_CTX_MODEM_INF_NOT_FOUND                                     NT_STATUS = 0xC00A0008
	NT_STATUS_CTX_INVALID_MODEMNAME                                       NT_STATUS = 0xC00A0009
	NT_STATUS_CTX_RESPONSE_ERROR                                          NT_STATUS = 0xC00A000A
	NT_STATUS_CTX_MODEM_RESPONSE_TIMEOUT                                  NT_STATUS = 0xC00A000B
	NT_STATUS_CTX_MODEM_RESPONSE_NO_CARRIER                               NT_STATUS = 0xC00A000C
	NT_STATUS_CTX_MODEM_RESPONSE_NO_DIALTONE                              NT_STATUS = 0xC00A000D
	NT_STATUS_CTX_MODEM_RESPONSE_BUSY                                     NT_STATUS = 0xC00A000E
	NT_STATUS_CTX_MODEM_RESPONSE_VOICE                                    NT_STATUS = 0xC00A000F
	NT_STATUS_CTX_TD_ERROR                                                NT_STATUS = 0xC00A0010
	NT_STATUS_CTX_LICENSE_CLIENT_INVALID                                  NT_STATUS = 0xC00A0012
	NT_STATUS_CTX_LICENSE_NOT_AVAILABLE                                   NT_STATUS = 0xC00A0013
	NT_STATUS_CTX_LICENSE_EXPIRED                                         NT_STATUS = 0xC00A0014
	NT_STATUS_CTX_WINSTATION_NOT_FOUND                                    NT_STATUS = 0xC00A0015
	NT_STATUS_CTX_WINSTATION_NAME_COLLISION                               NT_STATUS = 0xC00A0016
	NT_STATUS_CTX_WINSTATION_BUSY                                         NT_STATUS = 0xC00A0017
	NT_STATUS_CTX_BAD_VIDEO_MODE                                          NT_STATUS = 0xC00A0018
	NT_STATUS_CTX_GRAPHICS_INVALID                                        NT_STATUS = 0xC00A0022
	NT_STATUS_CTX_NOT_CONSOLE                                             NT_STATUS = 0xC00A0024
	NT_STATUS_CTX_CLIENT_QUERY_TIMEOUT                                    NT_STATUS = 0xC00A0026
	NT_STATUS_CTX_CONSOLE_DISCONNECT                                      NT_STATUS = 0xC00A0027
	NT_STATUS_CTX_CONSOLE_CONNECT                                         NT_STATUS = 0xC00A0028
	NT_STATUS_CTX_SHADOW_DENIED                                           NT_STATUS = 0xC00A002A
	NT_STATUS_CTX_WINSTATION_ACCESS_DENIED                                NT_STATUS = 0xC00A002B
	NT_STATUS_CTX_INVALID_WD                                              NT_STATUS = 0xC00A002E
	NT_STATUS_CTX_WD_NOT_FOUND                                            NT_STATUS = 0xC00A002F
	NT_STATUS_CTX_SHADOW_INVALID                                          NT_STATUS = 0xC00A0030
	NT_STATUS_CTX_SHADOW_DISABLED                                         NT_STATUS = 0xC00A0031
	NT_STATUS_RDP_PROTOCOL_ERROR                                          NT_STATUS = 0xC00A0032
	NT_STATUS_CTX_CLIENT_LICENSE_NOT_SET                                  NT_STATUS = 0xC00A0033
	NT_STATUS_CTX_CLIENT_LICENSE_IN_USE                                   NT_STATUS = 0xC00A0034
	NT_STATUS_CTX_SHADOW_ENDED_BY_MODE_CHANGE                             NT_STATUS = 0xC00A0035
	NT_STATUS_CTX_SHADOW_NOT_RUNNING                                      NT_STATUS = 0xC00A0036
	NT_STATUS_CTX_LOGON_DISABLED                                          NT_STATUS = 0xC00A0037
	NT_STATUS_CTX_SECURITY_LAYER_ERROR                                    NT_STATUS = 0xC00A0038
	NT_STATUS_TS_INCOMPATIBLE_SESSIONS                                    NT_STATUS = 0xC00A0039
	NT_STATUS_MUI_FILE_NOT_FOUND                                          NT_STATUS = 0xC00B0001
	NT_STATUS_MUI_INVALID_FILE                                            NT_STATUS = 0xC00B0002
	NT_STATUS_MUI_INVALID_RC_CONFIG                                       NT_STATUS = 0xC00B0003
	NT_STATUS_MUI_INVALID_LOCALE_NAME                                     NT_STATUS = 0xC00B0004
	NT_STATUS_MUI_INVALID_ULTIMATEFALLBACK_NAME                           NT_STATUS = 0xC00B0005
	NT_STATUS_MUI_FILE_NOT_LOADED                                         NT_STATUS = 0xC00B0006
	NT_STATUS_RESOURCE_ENUM_USER_STOP                                     NT_STATUS = 0xC00B0007
	NT_STATUS_CLUSTER_INVALID_NODE                                        NT_STATUS = 0xC0130001
	NT_STATUS_CLUSTER_NODE_EXISTS                                         NT_STATUS = 0xC0130002
	NT_STATUS_CLUSTER_JOIN_IN_PROGRESS                                    NT_STATUS = 0xC0130003
	NT_STATUS_CLUSTER_NODE_NOT_FOUND                                      NT_STATUS = 0xC0130004
	NT_STATUS_CLUSTER_LOCAL_NODE_NOT_FOUND                                NT_STATUS = 0xC0130005
	NT_STATUS_CLUSTER_NETWORK_EXISTS                                      NT_STATUS = 0xC0130006
	NT_STATUS_CLUSTER_NETWORK_NOT_FOUND                                   NT_STATUS = 0xC0130007
	NT_STATUS_CLUSTER_NETINTERFACE_EXISTS                                 NT_STATUS = 0xC0130008
	NT_STATUS_CLUSTER_NETINTERFACE_NOT_FOUND                              NT_STATUS = 0xC0130009
	NT_STATUS_CLUSTER_INVALID_REQUEST                                     NT_STATUS = 0xC013000A
	NT_STATUS_CLUSTER_INVALID_NETWORK_PROVIDER                            NT_STATUS = 0xC013000B
	NT_STATUS_CLUSTER_NODE_DOWN                                           NT_STATUS = 0xC013000C
	NT_STATUS_CLUSTER_NODE_UNREACHABLE                                    NT_STATUS = 0xC013000D
	NT_STATUS_CLUSTER_NODE_NOT_MEMBER                                     NT_STATUS = 0xC013000E
	NT_STATUS_CLUSTER_JOIN_NOT_IN_PROGRESS                                NT_STATUS = 0xC013000F
	NT_STATUS_CLUSTER_INVALID_NETWORK                                     NT_STATUS = 0xC0130010
	NT_STATUS_CLUSTER_NO_NET_ADAPTERS                                     NT_STATUS = 0xC0130011
	NT_STATUS_CLUSTER_NODE_UP                                             NT_STATUS = 0xC0130012
	NT_STATUS_CLUSTER_NODE_PAUSED                                         NT_STATUS = 0xC0130013
	NT_STATUS_CLUSTER_NODE_NOT_PAUSED                                     NT_STATUS = 0xC0130014
	NT_STATUS_CLUSTER_NO_SECURITY_CONTEXT                                 NT_STATUS = 0xC0130015
	NT_STATUS_CLUSTER_NETWORK_NOT_INTERNAL                                NT_STATUS = 0xC0130016
	NT_STATUS_CLUSTER_POISONED                                            NT_STATUS = 0xC0130017
	NT_STATUS_ACPI_INVALID_OPCODE                                         NT_STATUS = 0xC0140001
	NT_STATUS_ACPI_STACK_OVERFLOW                                         NT_STATUS = 0xC0140002
	NT_STATUS_ACPI_ASSERT_FAILED                                          NT_STATUS = 0xC0140003
	NT_STATUS_ACPI_INVALID_INDEX                                          NT_STATUS = 0xC0140004
	NT_STATUS_ACPI_INVALID_ARGUMENT                                       NT_STATUS = 0xC0140005
	NT_STATUS_ACPI_FATAL                                                  NT_STATUS = 0xC0140006
	NT_STATUS_ACPI_INVALID_SUPERNAME                                      NT_STATUS = 0xC0140007
	NT_STATUS_ACPI_INVALID_ARGTYPE                                        NT_STATUS = 0xC0140008
	NT_STATUS_ACPI_INVALID_OBJTYPE                                        NT_STATUS = 0xC0140009
	NT_STATUS_ACPI_INVALID_TARGETTYPE                                     NT_STATUS = 0xC014000A
	NT_STATUS_ACPI_INCORRECT_ARGUMENT_COUNT                               NT_STATUS = 0xC014000B
	NT_STATUS_ACPI_ADDRESS_NOT_MAPPED                                     NT_STATUS = 0xC014000C
	NT_STATUS_ACPI_INVALID_EVENTTYPE                                      NT_STATUS = 0xC014000D
	NT_STATUS_ACPI_HANDLER_COLLISION                                      NT_STATUS = 0xC014000E
	NT_STATUS_ACPI_INVALID_DATA                                           NT_STATUS = 0xC014000F
	NT_STATUS_ACPI_INVALID_REGION                                         NT_STATUS = 0xC0140010
	NT_STATUS_ACPI_INVALID_ACCESS_SIZE                                    NT_STATUS = 0xC0140011
	NT_STATUS_ACPI_ACQUIRE_GLOBAL_LOCK                                    NT_STATUS = 0xC0140012
	NT_STATUS_ACPI_ALREADY_INITIALIZED                                    NT_STATUS = 0xC0140013
	NT_STATUS_ACPI_NOT_INITIALIZED                                        NT_STATUS = 0xC0140014
	NT_STATUS_ACPI_INVALID_MUTEX_LEVEL                                    NT_STATUS = 0xC0140015
	NT_STATUS_ACPI_MUTEX_NOT_OWNED                                        NT_STATUS = 0xC0140016
	NT_STATUS_ACPI_MUTEX_NOT_OWNER                                        NT_STATUS = 0xC0140017
	NT_STATUS_ACPI_RS_ACCESS                                              NT_STATUS = 0xC0140018
	NT_STATUS_ACPI_INVALID_TABLE                                          NT_STATUS = 0xC0140019
	NT_STATUS_ACPI_REG_HANDLER_FAILED                                     NT_STATUS = 0xC0140020
	NT_STATUS_ACPI_POWER_REQUEST_FAILED                                   NT_STATUS = 0xC0140021
	NT_STATUS_SXS_SECTION_NOT_FOUND                                       NT_STATUS = 0xC0150001
	NT_STATUS_SXS_CANT_GEN_ACTCTX                                         NT_STATUS = 0xC0150002
	NT_STATUS_SXS_INVALID_ACTCTXDATA_FORMAT                               NT_STATUS = 0xC0150003
	NT_STATUS_SXS_ASSEMBLY_NOT_FOUND                                      NT_STATUS = 0xC0150004
	NT_STATUS_SXS_MANIFEST_FORMAT_ERROR                                   NT_STATUS = 0xC0150005
	NT_STATUS_SXS_MANIFEST_PARSE_ERROR                                    NT_STATUS = 0xC0150006
	NT_STATUS_SXS_ACTIVATION_CONTEXT_DISABLED                             NT_STATUS = 0xC0150007
	NT_STATUS_SXS_KEY_NOT_FOUND                                           NT_STATUS = 0xC0150008
	NT_STATUS_SXS_VERSION_CONFLICT                                        NT_STATUS = 0xC0150009
	NT_STATUS_SXS_WRONG_SECTION_TYPE                                      NT_STATUS = 0xC015000A
	NT_STATUS_SXS_THREAD_QUERIES_DISABLED                                 NT_STATUS = 0xC015000B
	NT_STATUS_SXS_ASSEMBLY_MISSING                                        NT_STATUS = 0xC015000C
	NT_STATUS_SXS_PROCESS_DEFAULT_ALREADY_SET                             NT_STATUS = 0xC015000E
	NT_STATUS_SXS_EARLY_DEACTIVATION                                      NT_STATUS = 0xC015000F
	NT_STATUS_SXS_INVALID_DEACTIVATION                                    NT_STATUS = 0xC0150010
	NT_STATUS_SXS_MULTIPLE_DEACTIVATION                                   NT_STATUS = 0xC0150011
	NT_STATUS_SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY                 NT_STATUS = 0xC0150012
	NT_STATUS_SXS_PROCESS_TERMINATION_REQUESTED                           NT_STATUS = 0xC0150013
	NT_STATUS_SXS_CORRUPT_ACTIVATION_STACK                                NT_STATUS = 0xC0150014
	NT_STATUS_SXS_CORRUPTION                                              NT_STATUS = 0xC0150015
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE                        NT_STATUS = 0xC0150016
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_NAME                         NT_STATUS = 0xC0150017
	NT_STATUS_SXS_IDENTITY_DUPLICATE_ATTRIBUTE                            NT_STATUS = 0xC0150018
	NT_STATUS_SXS_IDENTITY_PARSE_ERROR                                    NT_STATUS = 0xC0150019
	NT_STATUS_SXS_COMPONENT_STORE_CORRUPT                                 NT_STATUS = 0xC015001A
	NT_STATUS_SXS_FILE_HASH_MISMATCH                                      NT_STATUS = 0xC015001B
	NT_STATUS_SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT           NT_STATUS = 0xC015001C
	NT_STATUS_SXS_IDENTITIES_DIFFERENT                                    NT_STATUS = 0xC015001D
	NT_STATUS_SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT                            NT_STATUS = 0xC015001E
	NT_STATUS_SXS_FILE_NOT_PART_OF_ASSEMBLY                               NT_STATUS = 0xC015001F
	NT_STATUS_ADVANCED_INSTALLER_FAILED                                   NT_STATUS = 0xC0150020
	NT_STATUS_XML_ENCODING_MISMATCH                                       NT_STATUS = 0xC0150021
	NT_STATUS_SXS_MANIFEST_TOO_BIG                                        NT_STATUS = 0xC0150022
	NT_STATUS_SXS_SETTING_NOT_REGISTERED                                  NT_STATUS = 0xC0150023
	NT_STATUS_SXS_TRANSACTION_CLOSURE_INCOMPLETE                          NT_STATUS = 0xC0150024
	NT_STATUS_SMI_PRIMITIVE_INSTALLER_FAILED                              NT_STATUS = 0xC0150025
	NT_STATUS_GENERIC_COMMAND_FAILED                                      NT_STATUS = 0xC0150026
	NT_STATUS_SXS_FILE_HASH_MISSING                                       NT_STATUS = 0xC0150027
	NT_STATUS_TRANSACTIONAL_CONFLICT                                      NT_STATUS = 0xC0190001
	NT_STATUS_INVALID_TRANSACTION                                         NT_STATUS = 0xC0190002
	NT_STATUS_TRANSACTION_NOT_ACTIVE                                      NT_STATUS = 0xC0190003
	NT_STATUS_TM_INITIALIZATION_FAILED                                    NT_STATUS = 0xC0190004
	NT_STATUS_RM_NOT_ACTIVE                                               NT_STATUS = 0xC0190005
	NT_STATUS_RM_METADATA_CORRUPT                                         NT_STATUS = 0xC0190006
	NT_STATUS_TRANSACTION_NOT_JOINED                                      NT_STATUS = 0xC0190007
	NT_STATUS_DIRECTORY_NOT_RM                                            NT_STATUS = 0xC0190008
	NT_STATUS_TRANSACTIONS_UNSUPPORTED_REMOTE                             NT_STATUS = 0xC019000A
	NT_STATUS_LOG_RESIZE_INVALID_SIZE                                     NT_STATUS = 0xC019000B
	NT_STATUS_REMOTE_FILE_VERSION_MISMATCH                                NT_STATUS = 0xC019000C
	NT_STATUS_CRM_PROTOCOL_ALREADY_EXISTS                                 NT_STATUS = 0xC019000F
	NT_STATUS_TRANSACTION_PROPAGATION_FAILED                              NT_STATUS = 0xC0190010
	NT_STATUS_CRM_PROTOCOL_NOT_FOUND                                      NT_STATUS = 0xC0190011
	NT_STATUS_TRANSACTION_SUPERIOR_EXISTS                                 NT_STATUS = 0xC0190012
	NT_STATUS_TRANSACTION_REQUEST_NOT_VALID                               NT_STATUS = 0xC0190013
	NT_STATUS_TRANSACTION_NOT_REQUESTED                                   NT_STATUS = 0xC0190014
	NT_STATUS_TRANSACTION_ALREADY_ABORTED                                 NT_STATUS = 0xC0190015
	NT_STATUS_TRANSACTION_ALREADY_COMMITTED                               NT_STATUS = 0xC0190016
	NT_STATUS_TRANSACTION_INVALID_MARSHALL_BUFFER                         NT_STATUS = 0xC0190017
	NT_STATUS_CURRENT_TRANSACTION_NOT_VALID                               NT_STATUS = 0xC0190018
	NT_STATUS_LOG_GROWTH_FAILED                                           NT_STATUS = 0xC0190019
	NT_STATUS_OBJECT_NO_LONGER_EXISTS                                     NT_STATUS = 0xC0190021
	NT_STATUS_STREAM_MINIVERSION_NOT_FOUND                                NT_STATUS = 0xC0190022
	NT_STATUS_STREAM_MINIVERSION_NOT_VALID                                NT_STATUS = 0xC0190023
	NT_STATUS_MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION         NT_STATUS = 0xC0190024
	NT_STATUS_CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT                    NT_STATUS = 0xC0190025
	NT_STATUS_CANT_CREATE_MORE_STREAM_MINIVERSIONS                        NT_STATUS = 0xC0190026
	NT_STATUS_HANDLE_NO_LONGER_VALID                                      NT_STATUS = 0xC0190028
	NT_STATUS_LOG_CORRUPTION_DETECTED                                     NT_STATUS = 0xC0190030
	NT_STATUS_RM_DISCONNECTED                                             NT_STATUS = 0xC0190032
	NT_STATUS_ENLISTMENT_NOT_SUPERIOR                                     NT_STATUS = 0xC0190033
	NT_STATUS_FILE_IDENTITY_NOT_PERSISTENT                                NT_STATUS = 0xC0190036
	NT_STATUS_CANT_BREAK_TRANSACTIONAL_DEPENDENCY                         NT_STATUS = 0xC0190037
	NT_STATUS_CANT_CROSS_RM_BOUNDARY                                      NT_STATUS = 0xC0190038
	NT_STATUS_TXF_DIR_NOT_EMPTY                                           NT_STATUS = 0xC0190039
	NT_STATUS_INDOUBT_TRANSACTIONS_EXIST                                  NT_STATUS = 0xC019003A
	NT_STATUS_TM_VOLATILE                                                 NT_STATUS = 0xC019003B
	NT_STATUS_ROLLBACK_TIMER_EXPIRED                                      NT_STATUS = 0xC019003C
	NT_STATUS_TXF_ATTRIBUTE_CORRUPT                                       NT_STATUS = 0xC019003D
	NT_STATUS_EFS_NOT_ALLOWED_IN_TRANSACTION                              NT_STATUS = 0xC019003E
	NT_STATUS_TRANSACTIONAL_OPEN_NOT_ALLOWED                              NT_STATUS = 0xC019003F
	NT_STATUS_TRANSACTED_MAPPING_UNSUPPORTED_REMOTE                       NT_STATUS = 0xC0190040
	NT_STATUS_TRANSACTION_REQUIRED_PROMOTION                              NT_STATUS = 0xC0190043
	NT_STATUS_CANNOT_EXECUTE_FILE_IN_TRANSACTION                          NT_STATUS = 0xC0190044
	NT_STATUS_TRANSACTIONS_NOT_FROZEN                                     NT_STATUS = 0xC0190045
	NT_STATUS_TRANSACTION_FREEZE_IN_PROGRESS                              NT_STATUS = 0xC0190046
	NT_STATUS_NOT_SNAPSHOT_VOLUME                                         NT_STATUS = 0xC0190047
	NT_STATUS_NO_SAVEPOINT_WITH_OPEN_FILES                                NT_STATUS = 0xC0190048
	NT_STATUS_SPARSE_NOT_ALLOWED_IN_TRANSACTION                           NT_STATUS = 0xC0190049
	NT_STATUS_TM_IDENTITY_MISMATCH                                        NT_STATUS = 0xC019004A
	NT_STATUS_FLOATED_SECTION                                             NT_STATUS = 0xC019004B
	NT_STATUS_CANNOT_ACCEPT_TRANSACTED_WORK                               NT_STATUS = 0xC019004C
	NT_STATUS_CANNOT_ABORT_TRANSACTIONS                                   NT_STATUS = 0xC019004D
	NT_STATUS_TRANSACTION_NOT_FOUND                                       NT_STATUS = 0xC019004E
	NT_STATUS_RESOURCEMANAGER_NOT_FOUND                                   NT_STATUS = 0xC019004F
	NT_STATUS_ENLISTMENT_NOT_FOUND                                        NT_STATUS = 0xC0190050
	NT_STATUS_TRANSACTIONMANAGER_NOT_FOUND                                NT_STATUS = 0xC0190051
	NT_STATUS_TRANSACTIONMANAGER_NOT_ONLINE                               NT_STATUS = 0xC0190052
	NT_STATUS_TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION                  NT_STATUS = 0xC0190053
	NT_STATUS_TRANSACTION_NOT_ROOT                                        NT_STATUS = 0xC0190054
	NT_STATUS_TRANSACTION_OBJECT_EXPIRED                                  NT_STATUS = 0xC0190055
	NT_STATUS_COMPRESSION_NOT_ALLOWED_IN_TRANSACTION                      NT_STATUS = 0xC0190056
	NT_STATUS_TRANSACTION_RESPONSE_NOT_ENLISTED                           NT_STATUS = 0xC0190057
	NT_STATUS_TRANSACTION_RECORD_TOO_LONG                                 NT_STATUS = 0xC0190058
	NT_STATUS_NO_LINK_TRACKING_IN_TRANSACTION                             NT_STATUS = 0xC0190059
	NT_STATUS_OPERATION_NOT_SUPPORTED_IN_TRANSACTION                      NT_STATUS = 0xC019005A
	NT_STATUS_TRANSACTION_INTEGRITY_VIOLATED                              NT_STATUS = 0xC019005B
	NT_STATUS_EXPIRED_HANDLE                                              NT_STATUS = 0xC0190060
	NT_STATUS_TRANSACTION_NOT_ENLISTED                                    NT_STATUS = 0xC0190061
	NT_STATUS_LOG_SECTOR_INVALID                                          NT_STATUS = 0xC01A0001
	NT_STATUS_LOG_SECTOR_PARITY_INVALID                                   NT_STATUS = 0xC01A0002
	NT_STATUS_LOG_SECTOR_REMAPPED                                         NT_STATUS = 0xC01A0003
	NT_STATUS_LOG_BLOCK_INCOMPLETE                                        NT_STATUS = 0xC01A0004
	NT_STATUS_LOG_INVALID_RANGE                                           NT_STATUS = 0xC01A0005
	NT_STATUS_LOG_BLOCKS_EXHAUSTED                                        NT_STATUS = 0xC01A0006
	NT_STATUS_LOG_READ_CONTEXT_INVALID                                    NT_STATUS = 0xC01A0007
	NT_STATUS_LOG_RESTART_INVALID                                         NT_STATUS = 0xC01A0008
	NT_STATUS_LOG_BLOCK_VERSION                                           NT_STATUS = 0xC01A0009
	NT_STATUS_LOG_BLOCK_INVALID                                           NT_STATUS = 0xC01A000A
	NT_STATUS_LOG_READ_MODE_INVALID                                       NT_STATUS = 0xC01A000B
	NT_STATUS_LOG_METADATA_CORRUPT                                        NT_STATUS = 0xC01A000D
	NT_STATUS_LOG_METADATA_INVALID                                        NT_STATUS = 0xC01A000E
	NT_STATUS_LOG_METADATA_INCONSISTENT                                   NT_STATUS = 0xC01A000F
	NT_STATUS_LOG_RESERVATION_INVALID                                     NT_STATUS = 0xC01A0010
	NT_STATUS_LOG_CANT_DELETE                                             NT_STATUS = 0xC01A0011
	NT_STATUS_LOG_CONTAINER_LIMIT_EXCEEDED                                NT_STATUS = 0xC01A0012
	NT_STATUS_LOG_START_OF_LOG                                            NT_STATUS = 0xC01A0013
	NT_STATUS_LOG_POLICY_ALREADY_INSTALLED                                NT_STATUS = 0xC01A0014
	NT_STATUS_LOG_POLICY_NOT_INSTALLED                                    NT_STATUS = 0xC01A0015
	NT_STATUS_LOG_POLICY_INVALID                                          NT_STATUS = 0xC01A0016
	NT_STATUS_LOG_POLICY_CONFLICT                                         NT_STATUS = 0xC01A0017
	NT_STATUS_LOG_PINNED_ARCHIVE_TAIL                                     NT_STATUS = 0xC01A0018
	NT_STATUS_LOG_RECORD_NONEXISTENT                                      NT_STATUS = 0xC01A0019
	NT_STATUS_LOG_RECORDS_RESERVED_INVALID                                NT_STATUS = 0xC01A001A
	NT_STATUS_LOG_SPACE_RESERVED_INVALID                                  NT_STATUS = 0xC01A001B
	NT_STATUS_LOG_TAIL_INVALID                                            NT_STATUS = 0xC01A001C
	NT_STATUS_LOG_FULL                                                    NT_STATUS = 0xC01A001D
	NT_STATUS_LOG_MULTIPLEXED                                             NT_STATUS = 0xC01A001E
	NT_STATUS_LOG_DEDICATED                                               NT_STATUS = 0xC01A001F
	NT_STATUS_LOG_ARCHIVE_NOT_IN_PROGRESS                                 NT_STATUS = 0xC01A0020
	NT_STATUS_LOG_ARCHIVE_IN_PROGRESS                                     NT_STATUS = 0xC01A0021
	NT_STATUS_LOG_EPHEMERAL                                               NT_STATUS = 0xC01A0022
	NT_STATUS_LOG_NOT_ENOUGH_CONTAINERS                                   NT_STATUS = 0xC01A0023
	NT_STATUS_LOG_CLIENT_ALREADY_REGISTERED                               NT_STATUS = 0xC01A0024
	NT_STATUS_LOG_CLIENT_NOT_REGISTERED                                   NT_STATUS = 0xC01A0025
	NT_STATUS_LOG_FULL_HANDLER_IN_PROGRESS                                NT_STATUS = 0xC01A0026
	NT_STATUS_LOG_CONTAINER_READ_FAILED                                   NT_STATUS = 0xC01A0027
	NT_STATUS_LOG_CONTAINER_WRITE_FAILED                                  NT_STATUS = 0xC01A0028
	NT_STATUS_LOG_CONTAINER_OPEN_FAILED                                   NT_STATUS = 0xC01A0029
	NT_STATUS_LOG_CONTAINER_STATE_INVALID                                 NT_STATUS = 0xC01A002A
	NT_STATUS_LOG_STATE_INVALID                                           NT_STATUS = 0xC01A002B
	NT_STATUS_LOG_PINNED                                                  NT_STATUS = 0xC01A002C
	NT_STATUS_LOG_METADATA_FLUSH_FAILED                                   NT_STATUS = 0xC01A002D
	NT_STATUS_LOG_INCONSISTENT_SECURITY                                   NT_STATUS = 0xC01A002E
	NT_STATUS_LOG_APPENDED_FLUSH_FAILED                                   NT_STATUS = 0xC01A002F
	NT_STATUS_LOG_PINNED_RESERVATION                                      NT_STATUS = 0xC01A0030
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD                            NT_STATUS = 0xC01B00EA
	NT_STATUS_FLT_NO_HANDLER_DEFINED                                      NT_STATUS = 0xC01C0001
	NT_STATUS_FLT_CONTEXT_ALREADY_DEFINED                                 NT_STATUS = 0xC01C0002
	NT_STATUS_FLT_INVALID_ASYNCHRONOUS_REQUEST                            NT_STATUS = 0xC01C0003
	NT_STATUS_FLT_DISALLOW_FAST_IO                                        NT_STATUS = 0xC01C0004
	NT_STATUS_FLT_INVALID_NAME_REQUEST                                    NT_STATUS = 0xC01C0005
	NT_STATUS_FLT_NOT_SAFE_TO_POST_OPERATION                              NT_STATUS = 0xC01C0006
	NT_STATUS_FLT_NOT_INITIALIZED                                         NT_STATUS = 0xC01C0007
	NT_STATUS_FLT_FILTER_NOT_READY                                        NT_STATUS = 0xC01C0008
	NT_STATUS_FLT_POST_OPERATION_CLEANUP                                  NT_STATUS = 0xC01C0009
	NT_STATUS_FLT_INTERNAL_ERROR                                          NT_STATUS = 0xC01C000A
	NT_STATUS_FLT_DELETING_OBJECT                                         NT_STATUS = 0xC01C000B
	NT_STATUS_FLT_MUST_BE_NONPAGED_POOL                                   NT_STATUS = 0xC01C000C
	NT_STATUS_FLT_DUPLICATE_ENTRY                                         NT_STATUS = 0xC01C000D
	NT_STATUS_FLT_CBDQ_DISABLED                                           NT_STATUS = 0xC01C000E
	NT_STATUS_FLT_DO_NOT_ATTACH                                           NT_STATUS = 0xC01C000F
	NT_STATUS_FLT_DO_NOT_DETACH                                           NT_STATUS = 0xC01C0010
	NT_STATUS_FLT_INSTANCE_ALTITUDE_COLLISION                             NT_STATUS = 0xC01C0011
	NT_STATUS_FLT_INSTANCE_NAME_COLLISION                                 NT_STATUS = 0xC01C0012
	NT_STATUS_FLT_FILTER_NOT_FOUND                                        NT_STATUS = 0xC01C0013
	NT_STATUS_FLT_VOLUME_NOT_FOUND                                        NT_STATUS = 0xC01C0014
	NT_STATUS_FLT_INSTANCE_NOT_FOUND                                      NT_STATUS = 0xC01C0015
	NT_STATUS_FLT_CONTEXT_ALLOCATION_NOT_FOUND                            NT_STATUS = 0xC01C0016
	NT_STATUS_FLT_INVALID_CONTEXT_REGISTRATION                            NT_STATUS = 0xC01C0017
	NT_STATUS_FLT_NAME_CACHE_MISS                                         NT_STATUS = 0xC01C0018
	NT_STATUS_FLT_NO_DEVICE_OBJECT                                        NT_STATUS = 0xC01C0019
	NT_STATUS_FLT_VOLUME_ALREADY_MOUNTED                                  NT_STATUS = 0xC01C001A
	NT_STATUS_FLT_ALREADY_ENLISTED                                        NT_STATUS = 0xC01C001B
	NT_STATUS_FLT_CONTEXT_ALREADY_LINKED                                  NT_STATUS = 0xC01C001C
	NT_STATUS_FLT_NO_WAITER_FOR_REPLY                                     NT_STATUS = 0xC01C0020
	NT_STATUS_MONITOR_NO_DESCRIPTOR                                       NT_STATUS = 0xC01D0001
	NT_STATUS_MONITOR_UNKNOWN_DESCRIPTOR_FORMAT                           NT_STATUS = 0xC01D0002
	NT_STATUS_MONITOR_INVALID_DESCRIPTOR_CHECKSUM                         NT_STATUS = 0xC01D0003
	NT_STATUS_MONITOR_INVALID_STANDARD_TIMING_BLOCK                       NT_STATUS = 0xC01D0004
	NT_STATUS_MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED                   NT_STATUS = 0xC01D0005
	NT_STATUS_MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK                  NT_STATUS = 0xC01D0006
	NT_STATUS_MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK                  NT_STATUS = 0xC01D0007
	NT_STATUS_MONITOR_NO_MORE_DESCRIPTOR_DATA                             NT_STATUS = 0xC01D0008
	NT_STATUS_MONITOR_INVALID_DETAILED_TIMING_BLOCK                       NT_STATUS = 0xC01D0009
	NT_STATUS_MONITOR_INVALID_MANUFACTURE_DATE                            NT_STATUS = 0xC01D000A
	NT_STATUS_GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER                           NT_STATUS = 0xC01E0000
	NT_STATUS_GRAPHICS_INSUFFICIENT_DMA_BUFFER                            NT_STATUS = 0xC01E0001
	NT_STATUS_GRAPHICS_INVALID_DISPLAY_ADAPTER                            NT_STATUS = 0xC01E0002
	NT_STATUS_GRAPHICS_ADAPTER_WAS_RESET                                  NT_STATUS = 0xC01E0003
	NT_STATUS_GRAPHICS_INVALID_DRIVER_MODEL                               NT_STATUS = 0xC01E0004
	NT_STATUS_GRAPHICS_PRESENT_MODE_CHANGED                               NT_STATUS = 0xC01E0005
	NT_STATUS_GRAPHICS_PRESENT_OCCLUDED                                   NT_STATUS = 0xC01E0006
	NT_STATUS_GRAPHICS_PRESENT_DENIED                                     NT_STATUS = 0xC01E0007
	NT_STATUS_GRAPHICS_CANNOTCOLORCONVERT                                 NT_STATUS = 0xC01E0008
	NT_STATUS_GRAPHICS_PRESENT_REDIRECTION_DISABLED                       NT_STATUS = 0xC01E000B
	NT_STATUS_GRAPHICS_PRESENT_UNOCCLUDED                                 NT_STATUS = 0xC01E000C
	NT_STATUS_GRAPHICS_NO_VIDEO_MEMORY                                    NT_STATUS = 0xC01E0100
	NT_STATUS_GRAPHICS_CANT_LOCK_MEMORY                                   NT_STATUS = 0xC01E0101
	NT_STATUS_GRAPHICS_ALLOCATION_BUSY                                    NT_STATUS = 0xC01E0102
	NT_STATUS_GRAPHICS_TOO_MANY_REFERENCES                                NT_STATUS = 0xC01E0103
	NT_STATUS_GRAPHICS_TRY_AGAIN_LATER                                    NT_STATUS = 0xC01E0104
	NT_STATUS_GRAPHICS_TRY_AGAIN_NOW                                      NT_STATUS = 0xC01E0105
	NT_STATUS_GRAPHICS_ALLOCATION_INVALID                                 NT_STATUS = 0xC01E0106
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE                   NT_STATUS = 0xC01E0107
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED                   NT_STATUS = 0xC01E0108
	NT_STATUS_GRAPHICS_CANT_EVICT_PINNED_ALLOCATION                       NT_STATUS = 0xC01E0109
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_USAGE                           NT_STATUS = 0xC01E0110
	NT_STATUS_GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION                      NT_STATUS = 0xC01E0111
	NT_STATUS_GRAPHICS_ALLOCATION_CLOSED                                  NT_STATUS = 0xC01E0112
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_INSTANCE                        NT_STATUS = 0xC01E0113
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_HANDLE                          NT_STATUS = 0xC01E0114
	NT_STATUS_GRAPHICS_WRONG_ALLOCATION_DEVICE                            NT_STATUS = 0xC01E0115
	NT_STATUS_GRAPHICS_ALLOCATION_CONTENT_LOST                            NT_STATUS = 0xC01E0116
	NT_STATUS_GRAPHICS_GPU_EXCEPTION_ON_DEVICE                            NT_STATUS = 0xC01E0200
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY                             NT_STATUS = 0xC01E0300
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED                       NT_STATUS = 0xC01E0301
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED             NT_STATUS = 0xC01E0302
	NT_STATUS_GRAPHICS_INVALID_VIDPN                                      NT_STATUS = 0xC01E0303
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE                       NT_STATUS = 0xC01E0304
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET                       NT_STATUS = 0xC01E0305
	NT_STATUS_GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED                       NT_STATUS = 0xC01E0306
	NT_STATUS_GRAPHICS_INVALID_VIDPN_SOURCEMODESET                        NT_STATUS = 0xC01E0308
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGETMODESET                        NT_STATUS = 0xC01E0309
	NT_STATUS_GRAPHICS_INVALID_FREQUENCY                                  NT_STATUS = 0xC01E030A
	NT_STATUS_GRAPHICS_INVALID_ACTIVE_REGION                              NT_STATUS = 0xC01E030B
	NT_STATUS_GRAPHICS_INVALID_TOTAL_REGION                               NT_STATUS = 0xC01E030C
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE                  NT_STATUS = 0xC01E0310
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE                  NT_STATUS = 0xC01E0311
	NT_STATUS_GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET                     NT_STATUS = 0xC01E0312
	NT_STATUS_GRAPHICS_PATH_ALREADY_IN_TOPOLOGY                           NT_STATUS = 0xC01E0313
	NT_STATUS_GRAPHICS_MODE_ALREADY_IN_MODESET                            NT_STATUS = 0xC01E0314
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTSOURCESET                      NT_STATUS = 0xC01E0315
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTTARGETSET                      NT_STATUS = 0xC01E0316
	NT_STATUS_GRAPHICS_SOURCE_ALREADY_IN_SET                              NT_STATUS = 0xC01E0317
	NT_STATUS_GRAPHICS_TARGET_ALREADY_IN_SET                              NT_STATUS = 0xC01E0318
	NT_STATUS_GRAPHICS_INVALID_VIDPN_PRESENT_PATH                         NT_STATUS = 0xC01E0319
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY                      NT_STATUS = 0xC01E031A
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET                  NT_STATUS = 0xC01E031B
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE                     NT_STATUS = 0xC01E031C
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_NOT_IN_SET                          NT_STATUS = 0xC01E031D
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET                      NT_STATUS = 0xC01E031F
	NT_STATUS_GRAPHICS_STALE_MODESET                                      NT_STATUS = 0xC01E0320
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCEMODESET                      NT_STATUS = 0xC01E0321
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCE_MODE                        NT_STATUS = 0xC01E0322
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN                    NT_STATUS = 0xC01E0323
	NT_STATUS_GRAPHICS_MODE_ID_MUST_BE_UNIQUE                             NT_STATUS = 0xC01E0324
	NT_STATUS_GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION    NT_STATUS = 0xC01E0325
	NT_STATUS_GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES            NT_STATUS = 0xC01E0326
	NT_STATUS_GRAPHICS_PATH_NOT_IN_TOPOLOGY                               NT_STATUS = 0xC01E0327
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE              NT_STATUS = 0xC01E0328
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET              NT_STATUS = 0xC01E0329
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTORSET                       NT_STATUS = 0xC01E032A
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTOR                          NT_STATUS = 0xC01E032B
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET                       NT_STATUS = 0xC01E032C
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET                   NT_STATUS = 0xC01E032D
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE                NT_STATUS = 0xC01E032E
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE                   NT_STATUS = 0xC01E032F
	NT_STATUS_GRAPHICS_RESOURCES_NOT_RELATED                              NT_STATUS = 0xC01E0330
	NT_STATUS_GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE                           NT_STATUS = 0xC01E0331
	NT_STATUS_GRAPHICS_TARGET_ID_MUST_BE_UNIQUE                           NT_STATUS = 0xC01E0332
	NT_STATUS_GRAPHICS_NO_AVAILABLE_VIDPN_TARGET                          NT_STATUS = 0xC01E0333
	NT_STATUS_GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER       NT_STATUS = 0xC01E0334
	NT_STATUS_GRAPHICS_NO_VIDPNMGR                                        NT_STATUS = 0xC01E0335
	NT_STATUS_GRAPHICS_NO_ACTIVE_VIDPN                                    NT_STATUS = 0xC01E0336
	NT_STATUS_GRAPHICS_STALE_VIDPN_TOPOLOGY                               NT_STATUS = 0xC01E0337
	NT_STATUS_GRAPHICS_MONITOR_NOT_CONNECTED                              NT_STATUS = 0xC01E0338
	NT_STATUS_GRAPHICS_SOURCE_NOT_IN_TOPOLOGY                             NT_STATUS = 0xC01E0339
	NT_STATUS_GRAPHICS_INVALID_PRIMARYSURFACE_SIZE                        NT_STATUS = 0xC01E033A
	NT_STATUS_GRAPHICS_INVALID_VISIBLEREGION_SIZE                         NT_STATUS = 0xC01E033B
	NT_STATUS_GRAPHICS_INVALID_STRIDE                                     NT_STATUS = 0xC01E033C
	NT_STATUS_GRAPHICS_INVALID_PIXELFORMAT                                NT_STATUS = 0xC01E033D
	NT_STATUS_GRAPHICS_INVALID_COLORBASIS                                 NT_STATUS = 0xC01E033E
	NT_STATUS_GRAPHICS_INVALID_PIXELVALUEACCESSMODE                       NT_STATUS = 0xC01E033F
	NT_STATUS_GRAPHICS_TARGET_NOT_IN_TOPOLOGY                             NT_STATUS = 0xC01E0340
	NT_STATUS_GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT                 NT_STATUS = 0xC01E0341
	NT_STATUS_GRAPHICS_VIDPN_SOURCE_IN_USE                                NT_STATUS = 0xC01E0342
	NT_STATUS_GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN                           NT_STATUS = 0xC01E0343
	NT_STATUS_GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL                    NT_STATUS = 0xC01E0344
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION       NT_STATUS = 0xC01E0345
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED NT_STATUS = 0xC01E0346
	NT_STATUS_GRAPHICS_INVALID_GAMMA_RAMP                                 NT_STATUS = 0xC01E0347
	NT_STATUS_GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED                           NT_STATUS = 0xC01E0348
	NT_STATUS_GRAPHICS_MULTISAMPLING_NOT_SUPPORTED                        NT_STATUS = 0xC01E0349
	NT_STATUS_GRAPHICS_MODE_NOT_IN_MODESET                                NT_STATUS = 0xC01E034A
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON       NT_STATUS = 0xC01E034D
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_TYPE                          NT_STATUS = 0xC01E034E
	NT_STATUS_GRAPHICS_INVALID_COPYPROTECTION_TYPE                        NT_STATUS = 0xC01E034F
	NT_STATUS_GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS                  NT_STATUS = 0xC01E0350
	NT_STATUS_GRAPHICS_INVALID_SCANLINE_ORDERING                          NT_STATUS = 0xC01E0352
	NT_STATUS_GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED                       NT_STATUS = 0xC01E0353
	NT_STATUS_GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS                   NT_STATUS = 0xC01E0354
	NT_STATUS_GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT                        NT_STATUS = 0xC01E0355
	NT_STATUS_GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM                     NT_STATUS = 0xC01E0356
	NT_STATUS_GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN                  NT_STATUS = 0xC01E0357
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT          NT_STATUS = 0xC01E0358
	NT_STATUS_GRAPHICS_MAX_NUM_PATHS_REACHED                              NT_STATUS = 0xC01E0359
	NT_STATUS_GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION                 NT_STATUS = 0xC01E035A
	NT_STATUS_GRAPHICS_INVALID_CLIENT_TYPE                                NT_STATUS = 0xC01E035B
	NT_STATUS_GRAPHICS_CLIENTVIDPN_NOT_SET                                NT_STATUS = 0xC01E035C
	NT_STATUS_GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED                  NT_STATUS = 0xC01E0400
	NT_STATUS_GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED                     NT_STATUS = 0xC01E0401
	NT_STATUS_GRAPHICS_NOT_A_LINKED_ADAPTER                               NT_STATUS = 0xC01E0430
	NT_STATUS_GRAPHICS_LEADLINK_NOT_ENUMERATED                            NT_STATUS = 0xC01E0431
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_ENUMERATED                          NT_STATUS = 0xC01E0432
	NT_STATUS_GRAPHICS_ADAPTER_CHAIN_NOT_READY                            NT_STATUS = 0xC01E0433
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_STARTED                             NT_STATUS = 0xC01E0434
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_POWERED_ON                          NT_STATUS = 0xC01E0435
	NT_STATUS_GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE                     NT_STATUS = 0xC01E0436
	NT_STATUS_GRAPHICS_NOT_POST_DEVICE_DRIVER                             NT_STATUS = 0xC01E0438
	NT_STATUS_GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED                        NT_STATUS = 0xC01E043B
	NT_STATUS_GRAPHICS_OPM_NOT_SUPPORTED                                  NT_STATUS = 0xC01E0500
	NT_STATUS_GRAPHICS_COPP_NOT_SUPPORTED                                 NT_STATUS = 0xC01E0501
	NT_STATUS_GRAPHICS_UAB_NOT_SUPPORTED                                  NT_STATUS = 0xC01E0502
	NT_STATUS_GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS                   NT_STATUS = 0xC01E0503
	NT_STATUS_GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL                      NT_STATUS = 0xC01E0504
	NT_STATUS_GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST                     NT_STATUS = 0xC01E0505
	NT_STATUS_GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME          NT_STATUS = 0xC01E0506
	NT_STATUS_GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP         NT_STATUS = 0xC01E0507
	NT_STATUS_GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED                NT_STATUS = 0xC01E0508
	NT_STATUS_GRAPHICS_OPM_INVALID_POINTER                                NT_STATUS = 0xC01E050A
	NT_STATUS_GRAPHICS_OPM_INTERNAL_ERROR                                 NT_STATUS = 0xC01E050B
	NT_STATUS_GRAPHICS_OPM_INVALID_HANDLE                                 NT_STATUS = 0xC01E050C
	NT_STATUS_GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE       NT_STATUS = 0xC01E050D
	NT_STATUS_GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH                     NT_STATUS = 0xC01E050E
	NT_STATUS_GRAPHICS_OPM_SPANNING_MODE_ENABLED                          NT_STATUS = 0xC01E050F
	NT_STATUS_GRAPHICS_OPM_THEATER_MODE_ENABLED                           NT_STATUS = 0xC01E0510
	NT_STATUS_GRAPHICS_PVP_HFS_FAILED                                     NT_STATUS = 0xC01E0511
	NT_STATUS_GRAPHICS_OPM_INVALID_SRM                                    NT_STATUS = 0xC01E0512
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP                   NT_STATUS = 0xC01E0513
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP                    NT_STATUS = 0xC01E0514
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA                  NT_STATUS = 0xC01E0515
	NT_STATUS_GRAPHICS_OPM_HDCP_SRM_NEVER_SET                             NT_STATUS = 0xC01E0516
	NT_STATUS_GRAPHICS_OPM_RESOLUTION_TOO_HIGH                            NT_STATUS = 0xC01E0517
	NT_STATUS_GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE               NT_STATUS = 0xC01E0518
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS              NT_STATUS = 0xC01E051A
	NT_STATUS_GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS                NT_STATUS = 0xC01E051B
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS  NT_STATUS = 0xC01E051C
	NT_STATUS_GRAPHICS_OPM_INVALID_INFORMATION_REQUEST                    NT_STATUS = 0xC01E051D
	NT_STATUS_GRAPHICS_OPM_DRIVER_INTERNAL_ERROR                          NT_STATUS = 0xC01E051E
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS   NT_STATUS = 0xC01E051F
	NT_STATUS_GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED                        NT_STATUS = 0xC01E0520
	NT_STATUS_GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST                  NT_STATUS = 0xC01E0521
	NT_STATUS_GRAPHICS_I2C_NOT_SUPPORTED                                  NT_STATUS = 0xC01E0580
	NT_STATUS_GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST                          NT_STATUS = 0xC01E0581
	NT_STATUS_GRAPHICS_I2C_ERROR_TRANSMITTING_DATA                        NT_STATUS = 0xC01E0582
	NT_STATUS_GRAPHICS_I2C_ERROR_RECEIVING_DATA                           NT_STATUS = 0xC01E0583
	NT_STATUS_GRAPHICS_DDCCI_VCP_NOT_SUPPORTED                            NT_STATUS = 0xC01E0584
	NT_STATUS_GRAPHICS_DDCCI_INVALID_DATA                                 NT_STATUS = 0xC01E0585
	NT_STATUS_GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE  NT_STATUS = 0xC01E0586
	NT_STATUS_GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING                  NT_STATUS = 0xC01E0587
	NT_STATUS_GRAPHICS_MCA_INTERNAL_ERROR                                 NT_STATUS = 0xC01E0588
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND                      NT_STATUS = 0xC01E0589
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH                       NT_STATUS = 0xC01E058A
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM                     NT_STATUS = 0xC01E058B
	NT_STATUS_GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE                    NT_STATUS = 0xC01E058C
	NT_STATUS_GRAPHICS_MONITOR_NO_LONGER_EXISTS                           NT_STATUS = 0xC01E058D
	NT_STATUS_GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED                     NT_STATUS = 0xC01E05E0
	NT_STATUS_GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME              NT_STATUS = 0xC01E05E1
	NT_STATUS_GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP             NT_STATUS = 0xC01E05E2
	NT_STATUS_GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED                    NT_STATUS = 0xC01E05E3
	NT_STATUS_GRAPHICS_INVALID_POINTER                                    NT_STATUS = 0xC01E05E4
	NT_STATUS_GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE           NT_STATUS = 0xC01E05E5
	NT_STATUS_GRAPHICS_PARAMETER_ARRAY_TOO_SMALL                          NT_STATUS = 0xC01E05E6
	NT_STATUS_GRAPHICS_INTERNAL_ERROR                                     NT_STATUS = 0xC01E05E7
	NT_STATUS_GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS                    NT_STATUS = 0xC01E05E8
	NT_STATUS_FVE_LOCKED_VOLUME                                           NT_STATUS = 0xC0210000
	NT_STATUS_FVE_NOT_ENCRYPTED                                           NT_STATUS = 0xC0210001
	NT_STATUS_FVE_BAD_INFORMATION                                         NT_STATUS = 0xC0210002
	NT_STATUS_FVE_TOO_SMALL                                               NT_STATUS = 0xC0210003
	NT_STATUS_FVE_FAILED_WRONG_FS                                         NT_STATUS = 0xC0210004
	NT_STATUS_FVE_FAILED_BAD_FS                                           NT_STATUS = 0xC0210005
	NT_STATUS_FVE_FS_NOT_EXTENDED                                         NT_STATUS = 0xC0210006
	NT_STATUS_FVE_FS_MOUNTED                                              NT_STATUS = 0xC0210007
	NT_STATUS_FVE_NO_LICENSE                                              NT_STATUS = 0xC0210008
	NT_STATUS_FVE_ACTION_NOT_ALLOWED                                      NT_STATUS = 0xC0210009
	NT_STATUS_FVE_BAD_DATA                                                NT_STATUS = 0xC021000A
	NT_STATUS_FVE_VOLUME_NOT_BOUND                                        NT_STATUS = 0xC021000B
	NT_STATUS_FVE_NOT_DATA_VOLUME                                         NT_STATUS = 0xC021000C
	NT_STATUS_FVE_CONV_READ_ERROR                                         NT_STATUS = 0xC021000D
	NT_STATUS_FVE_CONV_WRITE_ERROR                                        NT_STATUS = 0xC021000E
	NT_STATUS_FVE_OVERLAPPED_UPDATE                                       NT_STATUS = 0xC021000F
	NT_STATUS_FVE_FAILED_SECTOR_SIZE                                      NT_STATUS = 0xC0210010
	NT_STATUS_FVE_FAILED_AUTHENTICATION                                   NT_STATUS = 0xC0210011
	NT_STATUS_FVE_NOT_OS_VOLUME                                           NT_STATUS = 0xC0210012
	NT_STATUS_FVE_KEYFILE_NOT_FOUND                                       NT_STATUS = 0xC0210013
	NT_STATUS_FVE_KEYFILE_INVALID                                         NT_STATUS = 0xC0210014
	NT_STATUS_FVE_KEYFILE_NO_VMK                                          NT_STATUS = 0xC0210015
	NT_STATUS_FVE_TPM_DISABLED                                            NT_STATUS = 0xC0210016
	NT_STATUS_FVE_TPM_SRK_AUTH_NOT_ZERO                                   NT_STATUS = 0xC0210017
	NT_STATUS_FVE_TPM_INVALID_PCR                                         NT_STATUS = 0xC0210018
	NT_STATUS_FVE_TPM_NO_VMK                                              NT_STATUS = 0xC0210019
	NT_STATUS_FVE_PIN_INVALID                                             NT_STATUS = 0xC021001A
	NT_STATUS_FVE_AUTH_INVALID_APPLICATION                                NT_STATUS = 0xC021001B
	NT_STATUS_FVE_AUTH_INVALID_CONFIG                                     NT_STATUS = 0xC021001C
	NT_STATUS_FVE_DEBUGGER_ENABLED                                        NT_STATUS = 0xC021001D
	NT_STATUS_FVE_DRY_RUN_FAILED                                          NT_STATUS = 0xC021001E
	NT_STATUS_FVE_BAD_METADATA_POINTER                                    NT_STATUS = 0xC021001F
	NT_STATUS_FVE_OLD_METADATA_COPY                                       NT_STATUS = 0xC0210020
	NT_STATUS_FVE_REBOOT_REQUIRED                                         NT_STATUS = 0xC0210021
	NT_STATUS_FVE_RAW_ACCESS                                              NT_STATUS = 0xC0210022
	NT_STATUS_FVE_RAW_BLOCKED                                             NT_STATUS = 0xC0210023
	NT_STATUS_FVE_NO_FEATURE_LICENSE                                      NT_STATUS = 0xC0210026
	NT_STATUS_FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED                     NT_STATUS = 0xC0210027
	NT_STATUS_FVE_CONV_RECOVERY_FAILED                                    NT_STATUS = 0xC0210028
	NT_STATUS_FVE_VIRTUALIZED_SPACE_TOO_BIG                               NT_STATUS = 0xC0210029
	NT_STATUS_FVE_VOLUME_TOO_SMALL                                        NT_STATUS = 0xC0210030
	NT_STATUS_FWP_CALLOUT_NOT_FOUND                                       NT_STATUS = 0xC0220001
	NT_STATUS_FWP_CONDITION_NOT_FOUND                                     NT_STATUS = 0xC0220002
	NT_STATUS_FWP_FILTER_NOT_FOUND                                        NT_STATUS = 0xC0220003
	NT_STATUS_FWP_LAYER_NOT_FOUND                                         NT_STATUS = 0xC0220004
	NT_STATUS_FWP_PROVIDER_NOT_FOUND                                      NT_STATUS = 0xC0220005
	NT_STATUS_FWP_PROVIDER_CONTEXT_NOT_FOUND                              NT_STATUS = 0xC0220006
	NT_STATUS_FWP_SUBLAYER_NOT_FOUND                                      NT_STATUS = 0xC0220007
	NT_STATUS_FWP_NOT_FOUND                                               NT_STATUS = 0xC0220008
	NT_STATUS_FWP_ALREADY_EXISTS                                          NT_STATUS = 0xC0220009
	NT_STATUS_FWP_IN_USE                                                  NT_STATUS = 0xC022000A
	NT_STATUS_FWP_DYNAMIC_SESSION_IN_PROGRESS                             NT_STATUS = 0xC022000B
	NT_STATUS_FWP_WRONG_SESSION                                           NT_STATUS = 0xC022000C
	NT_STATUS_FWP_NO_TXN_IN_PROGRESS                                      NT_STATUS = 0xC022000D
	NT_STATUS_FWP_TXN_IN_PROGRESS                                         NT_STATUS = 0xC022000E
	NT_STATUS_FWP_TXN_ABORTED                                             NT_STATUS = 0xC022000F
	NT_STATUS_FWP_SESSION_ABORTED                                         NT_STATUS = 0xC0220010
	NT_STATUS_FWP_INCOMPATIBLE_TXN                                        NT_STATUS = 0xC0220011
	NT_STATUS_FWP_TIMEOUT                                                 NT_STATUS = 0xC0220012
	NT_STATUS_FWP_NET_EVENTS_DISABLED                                     NT_STATUS = 0xC0220013
	NT_STATUS_FWP_INCOMPATIBLE_LAYER                                      NT_STATUS = 0xC0220014
	NT_STATUS_FWP_KM_CLIENTS_ONLY                                         NT_STATUS = 0xC0220015
	NT_STATUS_FWP_LIFETIME_MISMATCH                                       NT_STATUS = 0xC0220016
	NT_STATUS_FWP_BUILTIN_OBJECT                                          NT_STATUS = 0xC0220017
	NT_STATUS_FWP_TOO_MANY_BOOTTIME_FILTERS                               NT_STATUS = 0xC0220018
	NT_STATUS_FWP_TOO_MANY_CALLOUTS                                       NT_STATUS = 0xC0220018
	NT_STATUS_FWP_NOTIFICATION_DROPPED                                    NT_STATUS = 0xC0220019
	NT_STATUS_FWP_TRAFFIC_MISMATCH                                        NT_STATUS = 0xC022001A
	NT_STATUS_FWP_INCOMPATIBLE_SA_STATE                                   NT_STATUS = 0xC022001B
	NT_STATUS_FWP_NULL_POINTER                                            NT_STATUS = 0xC022001C
	NT_STATUS_FWP_INVALID_ENUMERATOR                                      NT_STATUS = 0xC022001D
	NT_STATUS_FWP_INVALID_FLAGS                                           NT_STATUS = 0xC022001E
	NT_STATUS_FWP_INVALID_NET_MASK                                        NT_STATUS = 0xC022001F
	NT_STATUS_FWP_INVALID_RANGE                                           NT_STATUS = 0xC0220020
	NT_STATUS_FWP_INVALID_INTERVAL                                        NT_STATUS = 0xC0220021
	NT_STATUS_FWP_ZERO_LENGTH_ARRAY                                       NT_STATUS = 0xC0220022
	NT_STATUS_FWP_NULL_DISPLAY_NAME                                       NT_STATUS = 0xC0220023
	NT_STATUS_FWP_INVALID_ACTION_TYPE                                     NT_STATUS = 0xC0220024
	NT_STATUS_FWP_INVALID_WEIGHT                                          NT_STATUS = 0xC0220025
	NT_STATUS_FWP_MATCH_TYPE_MISMATCH                                     NT_STATUS = 0xC0220026
	NT_STATUS_FWP_TYPE_MISMATCH                                           NT_STATUS = 0xC0220027
	NT_STATUS_FWP_OUT_OF_BOUNDS                                           NT_STATUS = 0xC0220028
	NT_STATUS_FWP_RESERVED                                                NT_STATUS = 0xC0220029
	NT_STATUS_FWP_DUPLICATE_CONDITION                                     NT_STATUS = 0xC022002A
	NT_STATUS_FWP_DUPLICATE_KEYMOD                                        NT_STATUS = 0xC022002B
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_LAYER                          NT_STATUS = 0xC022002C
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER                       NT_STATUS = 0xC022002D
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER                         NT_STATUS = 0xC022002E
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT                       NT_STATUS = 0xC022002F
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_METHOD                                NT_STATUS = 0xC0220030
	NT_STATUS_FWP_INCOMPATIBLE_DH_GROUP                                   NT_STATUS = 0xC0220031
	NT_STATUS_FWP_EM_NOT_SUPPORTED                                        NT_STATUS = 0xC0220032
	NT_STATUS_FWP_NEVER_MATCH                                             NT_STATUS = 0xC0220033
	NT_STATUS_FWP_PROVIDER_CONTEXT_MISMATCH                               NT_STATUS = 0xC0220034
	NT_STATUS_FWP_INVALID_PARAMETER                                       NT_STATUS = 0xC0220035
	NT_STATUS_FWP_TOO_MANY_SUBLAYERS                                      NT_STATUS = 0xC0220036
	NT_STATUS_FWP_CALLOUT_NOTIFICATION_FAILED                             NT_STATUS = 0xC0220037
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_CONFIG                                NT_STATUS = 0xC0220038
	NT_STATUS_FWP_INCOMPATIBLE_CIPHER_CONFIG                              NT_STATUS = 0xC0220039
	NT_STATUS_FWP_DUPLICATE_AUTH_METHOD                                   NT_STATUS = 0xC022003C
	NT_STATUS_FWP_TCPIP_NOT_READY                                         NT_STATUS = 0xC0220100
	NT_STATUS_FWP_INJECT_HANDLE_CLOSING                                   NT_STATUS = 0xC0220101
	NT_STATUS_FWP_INJECT_HANDLE_STALE                                     NT_STATUS = 0xC0220102
	NT_STATUS_FWP_CANNOT_PEND                                             NT_STATUS = 0xC0220103
	NT_STATUS_NDIS_CLOSING                                                NT_STATUS = 0xC0230002
	NT_STATUS_NDIS_BAD_VERSION                                            NT_STATUS = 0xC0230004
	NT_STATUS_NDIS_BAD_CHARACTERISTICS                                    NT_STATUS = 0xC0230005
	NT_STATUS_NDIS_ADAPTER_NOT_FOUND                                      NT_STATUS = 0xC0230006
	NT_STATUS_NDIS_OPEN_FAILED                                            NT_STATUS = 0xC0230007
	NT_STATUS_NDIS_DEVICE_FAILED                                          NT_STATUS = 0xC0230008
	NT_STATUS_NDIS_MULTICAST_FULL                                         NT_STATUS = 0xC0230009
	NT_STATUS_NDIS_MULTICAST_EXISTS                                       NT_STATUS = 0xC023000A
	NT_STATUS_NDIS_MULTICAST_NOT_FOUND                                    NT_STATUS = 0xC023000B
	NT_STATUS_NDIS_REQUEST_ABORTED                                        NT_STATUS = 0xC023000C
	NT_STATUS_NDIS_RESET_IN_PROGRESS                                      NT_STATUS = 0xC023000D
	NT_STATUS_NDIS_INVALID_PACKET                                         NT_STATUS = 0xC023000F
	NT_STATUS_NDIS_INVALID_DEVICE_REQUEST                                 NT_STATUS = 0xC0230010
	NT_STATUS_NDIS_ADAPTER_NOT_READY                                      NT_STATUS = 0xC0230011
	NT_STATUS_NDIS_INVALID_LENGTH                                         NT_STATUS = 0xC0230014
	NT_STATUS_NDIS_INVALID_DATA                                           NT_STATUS = 0xC0230015
	NT_STATUS_NDIS_BUFFER_TOO_SHORT                                       NT_STATUS = 0xC0230016
	NT_STATUS_NDIS_INVALID_OID                                            NT_STATUS = 0xC0230017
	NT_STATUS_NDIS_ADAPTER_REMOVED                                        NT_STATUS = 0xC0230018
	NT_STATUS_NDIS_UNSUPPORTED_MEDIA                                      NT_STATUS = 0xC0230019
	NT_STATUS_NDIS_GROUP_ADDRESS_IN_USE                                   NT_STATUS = 0xC023001A
	NT_STATUS_NDIS_FILE_NOT_FOUND                                         NT_STATUS = 0xC023001B
	NT_STATUS_NDIS_ERROR_READING_FILE                                     NT_STATUS = 0xC023001C
	NT_STATUS_NDIS_ALREADY_MAPPED                                         NT_STATUS = 0xC023001D
	NT_STATUS_NDIS_RESOURCE_CONFLICT                                      NT_STATUS = 0xC023001E
	NT_STATUS_NDIS_MEDIA_DISCONNECTED                                     NT_STATUS = 0xC023001F
	NT_STATUS_NDIS_INVALID_ADDRESS                                        NT_STATUS = 0xC0230022
	NT_STATUS_NDIS_PAUSED                                                 NT_STATUS = 0xC023002A
	NT_STATUS_NDIS_INTERFACE_NOT_FOUND                                    NT_STATUS = 0xC023002B
	NT_STATUS_NDIS_UNSUPPORTED_REVISION                                   NT_STATUS = 0xC023002C
	NT_STATUS_NDIS_INVALID_PORT                                           NT_STATUS = 0xC023002D
	NT_STATUS_NDIS_INVALID_PORT_STATE                                     NT_STATUS = 0xC023002E
	NT_STATUS_NDIS_LOW_POWER_STATE                                        NT_STATUS = 0xC023002F
	NT_STATUS_NDIS_NOT_SUPPORTED                                          NT_STATUS = 0xC02300BB
	NT_STATUS_NDIS_OFFLOAD_POLICY                                         NT_STATUS = 0xC023100F
	NT_STATUS_NDIS_OFFLOAD_CONNECTION_REJECTED                            NT_STATUS = 0xC0231012
	NT_STATUS_NDIS_OFFLOAD_PATH_REJECTED                                  NT_STATUS = 0xC0231013
	NT_STATUS_NDIS_DOT11_AUTO_CONFIG_ENABLED                              NT_STATUS = 0xC0232000
	NT_STATUS_NDIS_DOT11_MEDIA_IN_USE                                     NT_STATUS = 0xC0232001
	NT_STATUS_NDIS_DOT11_POWER_STATE_INVALID                              NT_STATUS = 0xC0232002
	NT_STATUS_NDIS_PM_WOL_PATTERN_LIST_FULL                               NT_STATUS = 0xC0232003
	NT_STATUS_NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL                          NT_STATUS = 0xC0232004
	NT_STATUS_IPSEC_BAD_SPI                                               NT_STATUS = 0xC0360001
	NT_STATUS_IPSEC_SA_LIFETIME_EXPIRED                                   NT_STATUS = 0xC0360002
	NT_STATUS_IPSEC_WRONG_SA                                              NT_STATUS = 0xC0360003
	NT_STATUS_IPSEC_REPLAY_CHECK_FAILED                                   NT_STATUS = 0xC0360004
	NT_STATUS_IPSEC_INVALID_PACKET                                        NT_STATUS = 0xC0360005
	NT_STATUS_IPSEC_INTEGRITY_CHECK_FAILED                                NT_STATUS = 0xC0360006
	NT_STATUS_IPSEC_CLEAR_TEXT_DROP                                       NT_STATUS = 0xC0360007
	NT_STATUS_IPSEC_AUTH_FIREWALL_DROP                                    NT_STATUS = 0xC0360008
	NT_STATUS_IPSEC_THROTTLE_DROP                                         NT_STATUS = 0xC0360009
	NT_STATUS_IPSEC_DOSP_BLOCK                                            NT_STATUS = 0xC0368000
	NT_STATUS_IPSEC_DOSP_RECEIVED_MULTICAST                               NT_STATUS = 0xC0368001
	NT_STATUS_IPSEC_DOSP_INVALID_PACKET                                   NT_STATUS = 0xC0368002
	NT_STATUS_IPSEC_DOSP_STATE_LOOKUP_FAILED                              NT_STATUS = 0xC0368003
	NT_STATUS_IPSEC_DOSP_MAX_ENTRIES                                      NT_STATUS = 0xC0368004
	NT_STATUS_IPSEC_DOSP_KEYMOD_NOT_ALLOWED                               NT_STATUS = 0xC0368005
	NT_STATUS_IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES                      NT_STATUS = 0xC0368006
	NT_STATUS_VOLMGR_MIRROR_NOT_SUPPORTED                                 NT_STATUS = 0xC038005B
	NT_STATUS_VOLMGR_RAID5_NOT_SUPPORTED                                  NT_STATUS = 0xC038005C
	NT_STATUS_VIRTDISK_PROVIDER_NOT_FOUND                                 NT_STATUS = 0xC03A0014
	NT_STATUS_VIRTDISK_NOT_VIRTUAL_DISK                                   NT_STATUS = 0xC03A0015
	NT_STATUS_VHD_PARENT_VHD_ACCESS_DENIED                                NT_STATUS = 0xC03A0016
	NT_STATUS_VHD_CHILD_PARENT_SIZE_MISMATCH                              NT_STATUS = 0xC03A0017
	NT_STATUS_VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED                       NT_STATUS = 0xC03A0018
	NT_STATUS_VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT                      NT_STATUS = 0xC03A0019
	NT_STATUS_SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP                       NT_STATUS = 0xC05D0000
	NT_STATUS_SMB_BAD_CLUSTER_DIALECT                                     NT_STATUS = 0xC05D0001
)

var (
	ERROR_SUCCESS = errors.New("the operation completed successfully")
	// NT_STATUS_WAIT_0 = errors.New("the caller specified waitany for waittype and one of the dispatcher objects in the object array has been set to the signaled state")
	ERROR_WAIT_1    = errors.New("the caller specified waitany for waittype and one of the dispatcher objects in the object array has been set to the signaled state")
	ERROR_WAIT_2    = errors.New("the caller specified waitany for waittype and one of the dispatcher objects in the object array has been set to the signaled state")
	ERROR_WAIT_3    = errors.New("the caller specified waitany for waittype and one of the dispatcher objects in the object array has been set to the signaled state")
	ERROR_WAIT_63   = errors.New("the caller specified waitany for waittype and one of the dispatcher objects in the object array has been set to the signaled state")
	ERROR_ABANDONED = errors.New("the caller attempted to wait for a mutex that has been abandoned")
	// NT_STATUS_ABANDONED_WAIT_0 = errors.New("the caller attempted to wait for a mutex that has been abandoned")
	ERROR_ABANDONED_WAIT_63                                           = errors.New("the caller attempted to wait for a mutex that has been abandoned")
	ERROR_USER_APC                                                    = errors.New("a user-mode apc was delivered before the given interval expired")
	ERROR_ALERTED                                                     = errors.New("the delay completed because the thread was alerted")
	ERROR_TIMEOUT                                                     = errors.New("the given timeout interval expired")
	ERROR_PENDING                                                     = errors.New("the operation that was requested is pending completion")
	ERROR_REPARSE                                                     = errors.New("a reparse should be performed by the object manager because the name of the file resulted in a symbolic link")
	ERROR_MORE_ENTRIES                                                = errors.New("returned by enumeration apis to indicate more information is available to successive calls")
	ERROR_NOT_ALL_ASSIGNED                                            = errors.New("indicates not all privileges or groups that are referenced are assigned to the caller. this allows, for example, all privileges to be disabled without having to know exactly which privileges are assigned")
	ERROR_SOME_NOT_MAPPED                                             = errors.New("some of the information to be translated has not been translated")
	ERROR_OPLOCK_BREAK_IN_PROGRESS                                    = errors.New("an open/create operation completed while an opportunistic lock (oplock) break is underway")
	ERROR_VOLUME_MOUNTED                                              = errors.New("a new volume has been mounted by a file system")
	ERROR_RXACT_COMMITTED                                             = errors.New("this success level status indicates that the transaction state already exists for the registry subtree but that a transaction commit was previously aborted. the commit has now been completed")
	ERROR_NOTIFY_CLEANUP                                              = errors.New("indicates that a notify change request has been completed due to closing the handle that made the notify change request")
	ERROR_NOTIFY_ENUM_DIR                                             = errors.New("indicates that a notify change request is being completed and that the information is not being returned in the caller's buffer. the caller now needs to enumerate the files to find the changes")
	ERROR_NO_QUOTAS_FOR_ACCOUNT                                       = errors.New("{no quotas} no system quota limits are specifically set for this account")
	ERROR_PRIMARY_TRANSPORT_CONNECT_FAILED                            = errors.New("{connect failure on primary transport} an attempt was made to connect to the remote server %hs on the primary transport, but the connection failed. the computer was able to connect on a secondary transport")
	ERROR_PAGE_FAULT_TRANSITION                                       = errors.New("the page fault was a transition fault")
	ERROR_PAGE_FAULT_DEMAND_ZERO                                      = errors.New("the page fault was a demand zero fault")
	ERROR_PAGE_FAULT_COPY_ON_WRITE                                    = errors.New("the page fault was a demand zero fault")
	ERROR_PAGE_FAULT_GUARD_PAGE                                       = errors.New("the page fault was a demand zero fault")
	ERROR_PAGE_FAULT_PAGING_FILE                                      = errors.New("the page fault was satisfied by reading from a secondary storage device")
	ERROR_CACHE_PAGE_LOCKED                                           = errors.New("the cached page was locked during operation")
	ERROR_CRASH_DUMP                                                  = errors.New("the crash dump exists in a paging file")
	ERROR_BUFFER_ALL_ZEROS                                            = errors.New("the specified buffer contains all zeros")
	ERROR_REPARSE_OBJECT                                              = errors.New("a reparse should be performed by the object manager because the name of the file resulted in a symbolic link")
	ERROR_RESOURCE_REQUIREMENTS_CHANGED                               = errors.New("the device has succeeded a query-stop and its resource requirements have changed")
	ERROR_TRANSLATION_COMPLETE                                        = errors.New("the translator has translated these resources into the global space and no additional translations should be performed")
	ERROR_DS_MEMBERSHIP_EVALUATED_LOCALLY                             = errors.New("the directory service evaluated group memberships locally, because it was unable to contact a global catalog server")
	ERROR_NOTHING_TO_TERMINATE                                        = errors.New("a process being terminated has no threads to terminate")
	ERROR_PROCESS_NOT_IN_JOB                                          = errors.New("the specified process is not part of a job")
	ERROR_PROCESS_IN_JOB                                              = errors.New("the specified process is part of a job")
	ERROR_VOLSNAP_HIBERNATE_READY                                     = errors.New("{volume shadow copy service} the system is now ready for hibernation")
	ERROR_FSFILTER_OP_COMPLETED_SUCCESSFULLY                          = errors.New("a file system or file system filter driver has successfully completed an fsfilter operation")
	ERROR_INTERRUPT_VECTOR_ALREADY_CONNECTED                          = errors.New("the specified interrupt vector was already connected")
	ERROR_INTERRUPT_STILL_CONNECTED                                   = errors.New("the specified interrupt vector is still connected")
	ERROR_PROCESS_CLONED                                              = errors.New("the current process is a cloned process")
	ERROR_FILE_LOCKED_WITH_ONLY_READERS                               = errors.New("the file was locked and all users of the file can only read")
	ERROR_FILE_LOCKED_WITH_WRITERS                                    = errors.New("the file was locked and at least one user of the file can write")
	ERROR_RESOURCEMANAGER_READ_ONLY                                   = errors.New("the specified resourcemanager made no changes or updates to the resource under this transaction")
	ERROR_WAIT_FOR_OPLOCK                                             = errors.New("an operation is blocked and waiting for an oplock")
	ERROR_DBG_EXCEPTION_HANDLED                                       = errors.New("debugger handled the exception")
	ERROR_DBG_CONTINUE                                                = errors.New("the debugger continued")
	ERROR_FLT_IO_COMPLETE                                             = errors.New("the io was completed by a filter")
	ERROR_FILE_NOT_AVAILABLE                                          = errors.New("the file is temporarily unavailable")
	ERROR_SHARE_UNAVAILABLE                                           = errors.New("the share is temporarily unavailable")
	ERROR_CALLBACK_RETURNED_THREAD_AFFINITY                           = errors.New("a threadpool worker thread entered a callback at thread affinity %p and exited at affinity %p")
	ERROR_OBJECT_NAME_EXISTS                                          = errors.New("{object exists} an attempt was made to create an object but the object name already exists")
	ERROR_THREAD_WAS_SUSPENDED                                        = errors.New("{thread suspended} a thread termination occurred while the thread was suspended. the thread resumed, and termination proceeded")
	ERROR_WORKING_SET_LIMIT_RANGE                                     = errors.New("{working set range error} an attempt was made to set the working set minimum or maximum to values that are outside the allowable range")
	ERROR_IMAGE_NOT_AT_BASE                                           = errors.New("{image relocated} an image file could not be mapped at the address that is specified in the image file. local fixes must be performed on this image")
	ERROR_RXACT_STATE_CREATED                                         = errors.New("this informational level status indicates that a specified registry subtree transaction state did not yet exist and had to be created")
	ERROR_SEGMENT_NOTIFICATION                                        = errors.New("{segment load} a virtual dos machine (vdm) is loading, unloading, or moving an ms-dos or win16 program segment image. an exception is raised so that a debugger can load, unload, or track symbols and breakpoints within these 16-bit segments")
	ERROR_LOCAL_USER_SESSION_KEY                                      = errors.New("{local session key} a user session key was requested for a local remote procedure call (rpc) connection. the session key that is returned is a constant value and not unique to this connection")
	ERROR_BAD_CURRENT_DIRECTORY                                       = errors.New("{invalid current directory} the process cannot switch to the startup current directory %hs. select ok to set the current directory to %hs, or select cancel to exit")
	ERROR_SERIAL_MORE_WRITES                                          = errors.New("{serial ioctl complete} a serial i/o operation was completed by another write to a serial port. (the ioctl_serial_xoff_counter reached zero.)")
	ERROR_REGISTRY_RECOVERED                                          = errors.New("{registry recovery} one of the files that contains the system registry data had to be recovered by using a log or alternate copy. the recovery was successful")
	ERROR_FT_READ_RECOVERY_FROM_BACKUP                                = errors.New("{redundant read} to satisfy a read request, the windows nt operating system fault-tolerant file system successfully read the requested data from a redundant copy. this was done because the file system encountered a failure on a member of the fault-tolerant volume but was unable to reassign the failing area of the device")
	ERROR_FT_WRITE_RECOVERY                                           = errors.New("{redundant write} to satisfy a write request, the windows nt fault-tolerant file system successfully wrote a redundant copy of the information. this was done because the file system encountered a failure on a member of the fault-tolerant volume but was unable to reassign the failing area of the device")
	ERROR_SERIAL_COUNTER_TIMEOUT                                      = errors.New("{serial ioctl timeout} a serial i/o operation completed because the time-out period expired. (the ioctl_serial_xoff_counter had not reached zero.)")
	ERROR_NULL_LM_PASSWORD                                            = errors.New("{password too complex} the windows password is too complex to be converted to a lan manager password. the lan manager password that returned is a null string")
	ERROR_IMAGE_MACHINE_TYPE_MISMATCH                                 = errors.New("{machine type mismatch} the image file %hs is valid but is for a machine type other than the current machine. select ok to continue, or cancel to fail the dll load")
	ERROR_RECEIVE_PARTIAL                                             = errors.New("{partial data received} the network transport returned partial data to its client. the remaining data will be sent later")
	ERROR_RECEIVE_EXPEDITED                                           = errors.New("{expedited data received} the network transport returned data to its client that was marked as expedited by the remote system")
	ERROR_RECEIVE_PARTIAL_EXPEDITED                                   = errors.New("{partial expedited data received} the network transport returned partial data to its client and this data was marked as expedited by the remote system. the remaining data will be sent later")
	ERROR_EVENT_DONE                                                  = errors.New("{tdi event done} the tdi indication has completed successfully")
	ERROR_EVENT_PENDING                                               = errors.New("{tdi event pending} the tdi indication has entered the pending state")
	ERROR_CHECKING_FILE_SYSTEM                                        = errors.New("checking file system on %wz")
	ERROR_FATAL_APP_EXIT                                              = errors.New("{fatal application exit} %hs")
	ERROR_PREDEFINED_HANDLE                                           = errors.New("the specified registry key is referenced by a predefined handle")
	ERROR_WAS_UNLOCKED                                                = errors.New("{page unlocked} the page protection of a locked page was changed to 'no access' and the page was unlocked from memory and from the process")
	ERROR_SERVICE_NOTIFICATION                                        = errors.New("%hs")
	ERROR_WAS_LOCKED                                                  = errors.New("{page locked} one of the pages to lock was already locked")
	ERROR_LOG_HARD_ERROR                                              = errors.New("application popup: %1 : %2")
	ERROR_ALREADY_WIN32                                               = errors.New("a win32 process already exists")
	ERROR_WX86_UNSIMULATE                                             = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_CONTINUE                                               = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_SINGLE_STEP                                            = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_BREAKPOINT                                             = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_EXCEPTION_CONTINUE                                     = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_EXCEPTION_LASTCHANCE                                   = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_WX86_EXCEPTION_CHAIN                                        = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_IMAGE_MACHINE_TYPE_MISMATCH_EXE                             = errors.New("{machine type mismatch} the image file %hs is valid but is for a machine type other than the current machine")
	ERROR_NO_YIELD_PERFORMED                                          = errors.New("a yield execution was performed and no thread was available to run")
	ERROR_TIMER_RESUME_IGNORED                                        = errors.New("the resume flag to a timer api was ignored")
	ERROR_ARBITRATION_UNHANDLED                                       = errors.New("the arbiter has deferred arbitration of these resources to its parent")
	ERROR_CARDBUS_NOT_SUPPORTED                                       = errors.New("the device has detected a cardbus card in its slot")
	ERROR_WX86_CREATEWX86TIB                                          = errors.New("an exception status code that is used by the win32 x86 emulation subsystem")
	ERROR_MP_PROCESSOR_MISMATCH                                       = errors.New("the cpus in this multiprocessor system are not all the same revision level. to use all processors, the operating system restricts itself to the features of the least capable processor in the system. if problems occur with this system, contact the cpu manufacturer to see if this mix of processors is supported")
	ERROR_HIBERNATED                                                  = errors.New("the system was put into hibernation")
	ERROR_RESUME_HIBERNATION                                          = errors.New("the system was resumed from hibernation")
	ERROR_FIRMWARE_UPDATED                                            = errors.New("windows has detected that the system firmware (bios) was updated [previous firmware date = %2, current firmware date %3]")
	ERROR_DRIVERS_LEAKING_LOCKED_PAGES                                = errors.New("a device driver is leaking locked i/o pages and is causing system degradation. the system has automatically enabled the tracking code to try and catch the culprit")
	ERROR_MESSAGE_RETRIEVED                                           = errors.New("the alpc message being canceled has already been retrieved from the queue on the other side")
	ERROR_SYSTEM_POWERSTATE_TRANSITION                                = errors.New("the system power state is transitioning from %2 to %3")
	ERROR_ALPC_CHECK_COMPLETION_LIST                                  = errors.New("the receive operation was successful. check the alpc completion list for the received message")
	ERROR_SYSTEM_POWERSTATE_COMPLEX_TRANSITION                        = errors.New("the system power state is transitioning from %2 to %3 but could enter %4")
	ERROR_ACCESS_AUDIT_BY_POLICY                                      = errors.New("access to %1 is monitored by policy rule %2")
	ERROR_ABANDON_HIBERFILE                                           = errors.New("a valid hibernation file has been invalidated and should be abandoned")
	ERROR_BIZRULES_NOT_ENABLED                                        = errors.New("business rule scripts are disabled for the calling application")
	ERROR_WAKE_SYSTEM                                                 = errors.New("the system has awoken")
	ERROR_DS_SHUTTING_DOWN                                            = errors.New("the directory service is shutting down")
	ERROR_DBG_REPLY_LATER                                             = errors.New("debugger will reply later")
	ERROR_DBG_UNABLE_TO_PROVIDE_HANDLE                                = errors.New("debugger cannot provide a handle")
	ERROR_DBG_TERMINATE_THREAD                                        = errors.New("debugger terminated the thread")
	ERROR_DBG_TERMINATE_PROCESS                                       = errors.New("debugger terminated the process")
	ERROR_DBG_CONTROL_C                                               = errors.New("debugger obtained control of c")
	ERROR_DBG_PRINTEXCEPTION_C                                        = errors.New("debugger printed an exception on control c")
	ERROR_DBG_RIPEXCEPTION                                            = errors.New("debugger received a rip exception")
	ERROR_DBG_CONTROL_BREAK                                           = errors.New("debugger received a control break")
	ERROR_DBG_COMMAND_EXCEPTION                                       = errors.New("debugger command communication exception")
	ERROR_RPC_NT_UUID_LOCAL_ONLY                                      = errors.New("a uuid that is valid only on this computer has been allocated")
	ERROR_RPC_NT_SEND_INCOMPLETE                                      = errors.New("some data remains to be sent in the request buffer")
	ERROR_CTX_CDM_CONNECT                                             = errors.New("the client drive mapping service has connected on terminal connection")
	ERROR_CTX_CDM_DISCONNECT                                          = errors.New("the client drive mapping service has disconnected on terminal connection")
	ERROR_SXS_RELEASE_ACTIVATION_CONTEXT                              = errors.New("a kernel mode component is releasing a reference on an activation context")
	ERROR_RECOVERY_NOT_NEEDED                                         = errors.New("the transactional resource manager is already consistent. recovery is not needed")
	ERROR_RM_ALREADY_STARTED                                          = errors.New("the transactional resource manager has already been started")
	ERROR_LOG_NO_RESTART                                              = errors.New("the log service encountered a log stream with no restart area")
	ERROR_VIDEO_DRIVER_DEBUG_REPORT_REQUEST                           = errors.New("{display driver recovered from failure} the %hs display driver has detected a failure and recovered from it. some graphical operations might have failed. the next time you restart the machine, a dialog box appears, giving you an opportunity to upload data about this failure to microsoft")
	ERROR_GRAPHICS_PARTIAL_DATA_POPULATED                             = errors.New("the specified buffer is not big enough to contain the entire requested dataset. partial data is populated up to the size of the buffer")
	ERROR_GRAPHICS_DRIVER_MISMATCH                                    = errors.New("the kernel driver detected a version mismatch between it and the user mode driver")
	ERROR_GRAPHICS_MODE_NOT_PINNED                                    = errors.New("no mode is pinned on the specified vidpn source/target")
	ERROR_GRAPHICS_NO_PREFERRED_MODE                                  = errors.New("the specified mode set does not specify a preference for one of its modes")
	ERROR_GRAPHICS_DATASET_IS_EMPTY                                   = errors.New("the specified dataset (for example, mode set, frequency range set, descriptor set, or topology) is empty")
	ERROR_GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET                        = errors.New("the specified dataset (for example, mode set, frequency range set, descriptor set, or topology) does not contain any more elements")
	ERROR_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED    = errors.New("the specified content transformation is not pinned on the specified vidpn present path")
	ERROR_GRAPHICS_UNKNOWN_CHILD_STATUS                               = errors.New("the child device presence was not reliably detected")
	ERROR_GRAPHICS_LEADLINK_START_DEFERRED                            = errors.New("starting the lead adapter in a linked configuration has been temporarily deferred")
	ERROR_GRAPHICS_POLLING_TOO_FREQUENTLY                             = errors.New("the display adapter is being polled for children too frequently at the same polling level")
	ERROR_GRAPHICS_START_DEFERRED                                     = errors.New("starting the adapter has been temporarily deferred")
	ERROR_NDIS_INDICATION_REQUIRED                                    = errors.New("the request will be completed later by an ndis status indication")
	ERROR_GUARD_PAGE_VIOLATION                                        = errors.New("{exception} guard page exception a page of memory that marks the end of a data structure, such as a stack or an array, has been accessed")
	ERROR_DATATYPE_MISALIGNMENT                                       = errors.New("{exception} alignment fault a data type misalignment was detected in a load or store instruction")
	ERROR_BREAKPOINT                                                  = errors.New("{exception} breakpoint a breakpoint has been reached")
	ERROR_SINGLE_STEP                                                 = errors.New("{exception} single step a single step or trace operation has just been completed")
	ERROR_BUFFER_OVERFLOW                                             = errors.New("{buffer overflow} the data was too large to fit into the specified buffer")
	ERROR_NO_MORE_FILES                                               = errors.New("{no more files} no more files were found which match the file specification")
	ERROR_WAKE_SYSTEM_DEBUGGER                                        = errors.New("{kernel debugger awakened} the system debugger was awakened by an interrupt")
	ERROR_HANDLES_CLOSED                                              = errors.New("{handles closed} handles to objects have been automatically closed because of the requested operation")
	ERROR_NO_INHERITANCE                                              = errors.New("{non-inheritable acl} an access control list (acl) contains no components that can be inherited")
	ERROR_GUID_SUBSTITUTION_MADE                                      = errors.New("{guid substitution} during the translation of a globally unique identifier (guid) to a windows security id (sid), no administratively defined guid prefix was found. a substitute prefix was used, which will not compromise system security. however, this might provide a more restrictive access than intended")
	ERROR_PARTIAL_COPY                                                = errors.New("because of protection conflicts, not all the requested bytes could be copied")
	ERROR_DEVICE_PAPER_EMPTY                                          = errors.New("{out of paper} the printer is out of paper")
	ERROR_DEVICE_POWERED_OFF                                          = errors.New("{device power is off} the printer power has been turned off")
	ERROR_DEVICE_OFF_LINE                                             = errors.New("{device offline} the printer has been taken offline")
	ERROR_DEVICE_BUSY                                                 = errors.New("{device busy} the device is currently busy")
	ERROR_NO_MORE_EAS                                                 = errors.New("{no more eas} no more extended attributes (eas) were found for the file")
	ERROR_INVALID_EA_NAME                                             = errors.New("{illegal ea} the specified extended attribute (ea) name contains at least one illegal character")
	ERROR_EA_LIST_INCONSISTENT                                        = errors.New("{inconsistent ea list} the extended attribute (ea) list is inconsistent")
	ERROR_INVALID_EA_FLAG                                             = errors.New("{invalid ea flag} an invalid extended attribute (ea) flag was set")
	ERROR_VERIFY_REQUIRED                                             = errors.New("{verifying disk} the media has changed and a verify operation is in progress; therefore, no reads or writes can be performed to the device, except those that are used in the verify operation")
	ERROR_EXTRANEOUS_INFORMATION                                      = errors.New("{too much information} the specified access control list (acl) contained more information than was expected")
	ERROR_RXACT_COMMIT_NECESSARY                                      = errors.New("this warning level status indicates that the transaction state already exists for the registry subtree, but that a transaction commit was previously aborted. the commit has not been completed but has not been rolled back either; therefore, it can still be committed, if needed")
	ERROR_NO_MORE_ENTRIES                                             = errors.New("{no more entries} no more entries are available from an enumeration operation")
	ERROR_FILEMARK_DETECTED                                           = errors.New("{filemark found} a filemark was detected")
	ERROR_MEDIA_CHANGED                                               = errors.New("{media changed} the media has changed")
	ERROR_BUS_RESET                                                   = errors.New("{i/o bus reset} an i/o bus reset was detected")
	ERROR_END_OF_MEDIA                                                = errors.New("{end of media} the end of the media was encountered")
	ERROR_BEGINNING_OF_MEDIA                                          = errors.New("the beginning of a tape or partition has been detected")
	ERROR_MEDIA_CHECK                                                 = errors.New("{media changed} the media might have changed")
	ERROR_SETMARK_DETECTED                                            = errors.New("a tape access reached a set mark")
	ERROR_NO_DATA_DETECTED                                            = errors.New("during a tape access, the end of the data written is reached")
	ERROR_REDIRECTOR_HAS_OPEN_HANDLES                                 = errors.New("the redirector is in use and cannot be unloaded")
	ERROR_SERVER_HAS_OPEN_HANDLES                                     = errors.New("the server is in use and cannot be unloaded")
	ERROR_ALREADY_DISCONNECTED                                        = errors.New("the specified connection has already been disconnected")
	ERROR_LONGJUMP                                                    = errors.New("a long jump has been executed")
	ERROR_CLEANER_CARTRIDGE_INSTALLED                                 = errors.New("a cleaner cartridge is present in the tape library")
	ERROR_PLUGPLAY_QUERY_VETOED                                       = errors.New("the plug and play query operation was not successful")
	ERROR_UNWIND_CONSOLIDATE                                          = errors.New("a frame consolidation has been executed")
	ERROR_REGISTRY_HIVE_RECOVERED                                     = errors.New("{registry hive recovered} the registry hive (file): %hs was corrupted and it has been recovered. some data might have been lost")
	ERROR_DLL_MIGHT_BE_INSECURE                                       = errors.New("the application is attempting to run executable code from the module %hs. this might be insecure. an alternative, %hs, is available. should the application use the secure module %hs?")
	ERROR_DLL_MIGHT_BE_INCOMPATIBLE                                   = errors.New("the application is loading executable code from the module %hs. this is secure but might be incompatible with previous releases of the operating system. an alternative, %hs, is available. should the application use the secure module %hs?")
	ERROR_STOPPED_ON_SYMLINK                                          = errors.New("the create operation stopped after reaching a symbolic link")
	ERROR_DEVICE_REQUIRES_CLEANING                                    = errors.New("the device has indicated that cleaning is necessary")
	ERROR_DEVICE_DOOR_OPEN                                            = errors.New("the device has indicated that its door is open. further operations require it closed and secured")
	ERROR_DATA_LOST_REPAIR                                            = errors.New("windows discovered a corruption in the file %hs. this file has now been repaired. check if any data in the file was lost because of the corruption")
	ERROR_DBG_EXCEPTION_NOT_HANDLED                                   = errors.New("debugger did not handle the exception")
	ERROR_CLUSTER_NODE_ALREADY_UP                                     = errors.New("the cluster node is already up")
	ERROR_CLUSTER_NODE_ALREADY_DOWN                                   = errors.New("the cluster node is already down")
	ERROR_CLUSTER_NETWORK_ALREADY_ONLINE                              = errors.New("the cluster network is already online")
	ERROR_CLUSTER_NETWORK_ALREADY_OFFLINE                             = errors.New("the cluster network is already offline")
	ERROR_CLUSTER_NODE_ALREADY_MEMBER                                 = errors.New("the cluster node is already a member of the cluster")
	ERROR_COULD_NOT_RESIZE_LOG                                        = errors.New("the log could not be set to the requested size")
	ERROR_NO_TXF_METADATA                                             = errors.New("there is no transaction metadata on the file")
	ERROR_CANT_RECOVER_WITH_HANDLE_OPEN                               = errors.New("the file cannot be recovered because there is a handle still open on it")
	ERROR_TXF_METADATA_ALREADY_PRESENT                                = errors.New("transaction metadata is already present on this file and cannot be superseded")
	ERROR_TRANSACTION_SCOPE_CALLBACKS_NOT_SET                         = errors.New("a transaction scope could not be entered because the scope handler has not been initialized")
	ERROR_VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED                  = errors.New("{display driver stopped responding and recovered} the %hs display driver has stopped working normally. the recovery had been performed")
	ERROR_FLT_BUFFER_TOO_SMALL                                        = errors.New("{buffer too small} the buffer is too small to contain the entry. no information has been written to the buffer")
	ERROR_FVE_PARTIAL_METADATA                                        = errors.New("volume metadata read or write is incomplete")
	ERROR_FVE_TRANSIENT_STATE                                         = errors.New("bitlocker encryption keys were ignored because the volume was in a transient state")
	ERROR_UNSUCCESSFUL                                                = errors.New("{operation failed} the requested operation was unsuccessful")
	ERROR_NOT_IMPLEMENTED                                             = errors.New("{not implemented} the requested operation is not implemented")
	ERROR_INVALID_INFO_CLASS                                          = errors.New("{invalid parameter} the specified information class is not a valid information class for the specified object")
	ERROR_INFO_LENGTH_MISMATCH                                        = errors.New("the specified information record length does not match the length that is required for the specified information class")
	ERROR_ACCESS_VIOLATION                                            = errors.New("the instruction at 0x%08lx referenced memory at 0x%08lx. the memory could not be %s")
	ERROR_IN_PAGE_ERROR                                               = errors.New("the instruction at 0x%08lx referenced memory at 0x%08lx. the required data was not placed into memory because of an i/o error status of 0x%08lx")
	ERROR_PAGEFILE_QUOTA                                              = errors.New("the page file quota for the process has been exhausted")
	ERROR_INVALID_HANDLE                                              = errors.New("an invalid handle was specified")
	ERROR_BAD_INITIAL_STACK                                           = errors.New("an invalid initial stack was specified in a call to ntcreatethread")
	ERROR_BAD_INITIAL_PC                                              = errors.New("an invalid initial start address was specified in a call to ntcreatethread")
	ERROR_INVALID_CID                                                 = errors.New("an invalid client id was specified")
	ERROR_TIMER_NOT_CANCELED                                          = errors.New("an attempt was made to cancel or set a timer that has an associated apc and the specified thread is not the thread that originally set the timer with an associated apc routine")
	ERROR_INVALID_PARAMETER                                           = errors.New("an invalid parameter was passed to a service or function")
	ERROR_NO_SUCH_DEVICE                                              = errors.New("a device that does not exist was specified")
	ERROR_NO_SUCH_FILE                                                = errors.New("{file not found} the file %hs does not exist")
	ERROR_INVALID_DEVICE_REQUEST                                      = errors.New("the specified request is not a valid operation for the target device")
	ERROR_END_OF_FILE                                                 = errors.New("the end-of-file marker has been reached. there is no valid data in the file beyond this marker")
	ERROR_WRONG_VOLUME                                                = errors.New("{wrong volume} the wrong volume is in the drive. insert volume %hs into drive %hs")
	ERROR_NO_MEDIA_IN_DEVICE                                          = errors.New("{no disk} there is no disk in the drive. insert a disk into drive %hs")
	ERROR_UNRECOGNIZED_MEDIA                                          = errors.New("{unknown disk format} the disk in drive %hs is not formatted properly. check the disk, and reformat it, if needed")
	ERROR_NONEXISTENT_SECTOR                                          = errors.New("{sector not found} the specified sector does not exist")
	ERROR_MORE_PROCESSING_REQUIRED                                    = errors.New("{still busy} the specified i/o request packet (irp) cannot be disposed of because the i/o operation is not complete")
	ERROR_NO_MEMORY                                                   = errors.New("{not enough quota} not enough virtual memory or paging file quota is available to complete the specified operation")
	ERROR_CONFLICTING_ADDRESSES                                       = errors.New("{conflicting address range} the specified address range conflicts with the address space")
	ERROR_NOT_MAPPED_VIEW                                             = errors.New("the address range to unmap is not a mapped view")
	ERROR_UNABLE_TO_FREE_VM                                           = errors.New("the virtual memory cannot be freed")
	ERROR_UNABLE_TO_DELETE_SECTION                                    = errors.New("the specified section cannot be deleted")
	ERROR_INVALID_SYSTEM_SERVICE                                      = errors.New("an invalid system service was specified in a system service call")
	ERROR_ILLEGAL_INSTRUCTION                                         = errors.New("{exception} illegal instruction an attempt was made to execute an illegal instruction")
	ERROR_INVALID_LOCK_SEQUENCE                                       = errors.New("{invalid lock sequence} an attempt was made to execute an invalid lock sequence")
	ERROR_INVALID_VIEW_SIZE                                           = errors.New("{invalid mapping} an attempt was made to create a view for a section that is bigger than the section")
	ERROR_INVALID_FILE_FOR_SECTION                                    = errors.New("{bad file} the attributes of the specified mapping file for a section of memory cannot be read")
	ERROR_ALREADY_COMMITTED                                           = errors.New("{already committed} the specified address range is already committed")
	ERROR_ACCESS_DENIED                                               = errors.New("{access denied} a process has requested access to an object but has not been granted those access rights")
	ERROR_BUFFER_TOO_SMALL                                            = errors.New("{buffer too small} the buffer is too small to contain the entry. no information has been written to the buffer")
	ERROR_OBJECT_TYPE_MISMATCH                                        = errors.New("{wrong type} there is a mismatch between the type of object that is required by the requested operation and the type of object that is specified in the request")
	ERROR_NONCONTINUABLE_EXCEPTION                                    = errors.New("{exception} cannot continue windows cannot continue from this exception")
	ERROR_INVALID_DISPOSITION                                         = errors.New("an invalid exception disposition was returned by an exception handler")
	ERROR_UNWIND                                                      = errors.New("unwind exception code")
	ERROR_BAD_STACK                                                   = errors.New("an invalid or unaligned stack was encountered during an unwind operation")
	ERROR_INVALID_UNWIND_TARGET                                       = errors.New("an invalid unwind target was encountered during an unwind operation")
	ERROR_NOT_LOCKED                                                  = errors.New("an attempt was made to unlock a page of memory that was not locked")
	ERROR_PARITY_ERROR                                                = errors.New("a device parity error on an i/o operation")
	ERROR_UNABLE_TO_DECOMMIT_VM                                       = errors.New("an attempt was made to decommit uncommitted virtual memory")
	ERROR_NOT_COMMITTED                                               = errors.New("an attempt was made to change the attributes on memory that has not been committed")
	ERROR_INVALID_PORT_ATTRIBUTES                                     = errors.New("invalid object attributes specified to ntcreateport or invalid port attributes specified to ntconnectport")
	ERROR_PORT_MESSAGE_TOO_LONG                                       = errors.New("the length of the message that was passed to ntrequestport or ntrequestwaitreplyport is longer than the maximum message that is allowed by the port")
	ERROR_INVALID_PARAMETER_MIX                                       = errors.New("an invalid combination of parameters was specified")
	ERROR_INVALID_QUOTA_LOWER                                         = errors.New("an attempt was made to lower a quota limit below the current usage")
	ERROR_DISK_CORRUPT_ERROR                                          = errors.New("{corrupt disk} the file system structure on the disk is corrupt and unusable. run the chkdsk utility on the volume %hs")
	ERROR_OBJECT_NAME_INVALID                                         = errors.New("the object name is invalid")
	ERROR_OBJECT_NAME_NOT_FOUND                                       = errors.New("the object name is not found")
	ERROR_OBJECT_NAME_COLLISION                                       = errors.New("the object name already exists")
	ERROR_PORT_DISCONNECTED                                           = errors.New("an attempt was made to send a message to a disconnected communication port")
	ERROR_DEVICE_ALREADY_ATTACHED                                     = errors.New("an attempt was made to attach to a device that was already attached to another device")
	ERROR_OBJECT_PATH_INVALID                                         = errors.New("the object path component was not a directory object")
	ERROR_OBJECT_PATH_NOT_FOUND                                       = errors.New("{path not found} the path %hs does not exist")
	ERROR_OBJECT_PATH_SYNTAX_BAD                                      = errors.New("the object path component was not a directory object")
	ERROR_DATA_OVERRUN                                                = errors.New("{data overrun} a data overrun error occurred")
	ERROR_DATA_LATE_ERROR                                             = errors.New("{data late} a data late error occurred")
	ERROR_DATA_ERROR                                                  = errors.New("{data error} an error occurred in reading or writing data")
	ERROR_CRC_ERROR                                                   = errors.New("{bad crc} a cyclic redundancy check (crc) checksum error occurred")
	ERROR_SECTION_TOO_BIG                                             = errors.New("{section too large} the specified section is too big to map the file")
	ERROR_PORT_CONNECTION_REFUSED                                     = errors.New("the ntconnectport request is refused")
	ERROR_INVALID_PORT_HANDLE                                         = errors.New("the type of port handle is invalid for the operation that is requested")
	ERROR_SHARING_VIOLATION                                           = errors.New("a file cannot be opened because the share access flags are incompatible")
	ERROR_QUOTA_EXCEEDED                                              = errors.New("insufficient quota exists to complete the operation")
	ERROR_INVALID_PAGE_PROTECTION                                     = errors.New("the specified page protection was not valid")
	ERROR_MUTANT_NOT_OWNED                                            = errors.New("an attempt to release a mutant object was made by a thread that was not the owner of the mutant object")
	ERROR_SEMAPHORE_LIMIT_EXCEEDED                                    = errors.New("an attempt was made to release a semaphore such that its maximum count would have been exceeded")
	ERROR_PORT_ALREADY_SET                                            = errors.New("an attempt was made to set the debugport or exceptionport of a process, but a port already exists in the process, or an attempt was made to set the completionport of a file but a port was already set in the file, or an attempt was made to set the associated completion port of an alpc port but it is already set")
	ERROR_SECTION_NOT_IMAGE                                           = errors.New("an attempt was made to query image information on a section that does not map an image")
	ERROR_SUSPEND_COUNT_EXCEEDED                                      = errors.New("an attempt was made to suspend a thread whose suspend count was at its maximum")
	ERROR_THREAD_IS_TERMINATING                                       = errors.New("an attempt was made to suspend a thread that has begun termination")
	ERROR_BAD_WORKING_SET_LIMIT                                       = errors.New("an attempt was made to set the working set limit to an invalid value (for example, the minimum greater than maximum)")
	ERROR_INCOMPATIBLE_FILE_MAP                                       = errors.New("a section was created to map a file that is not compatible with an already existing section that maps the same file")
	ERROR_SECTION_PROTECTION                                          = errors.New("a view to a section specifies a protection that is incompatible with the protection of the initial view")
	ERROR_EAS_NOT_SUPPORTED                                           = errors.New("an operation involving eas failed because the file system does not support eas")
	ERROR_EA_TOO_LARGE                                                = errors.New("an ea operation failed because the ea set is too large")
	ERROR_NONEXISTENT_EA_ENTRY                                        = errors.New("an ea operation failed because the name or ea index is invalid")
	ERROR_NO_EAS_ON_FILE                                              = errors.New("the file for which eas were requested has no eas")
	ERROR_EA_CORRUPT_ERROR                                            = errors.New("the ea is corrupt and cannot be read")
	ERROR_FILE_LOCK_CONFLICT                                          = errors.New("a requested read/write cannot be granted due to a conflicting file lock")
	ERROR_LOCK_NOT_GRANTED                                            = errors.New("a requested file lock cannot be granted due to other existing locks")
	ERROR_DELETE_PENDING                                              = errors.New("a non-close operation has been requested of a file object that has a delete pending")
	ERROR_CTL_FILE_NOT_SUPPORTED                                      = errors.New("an attempt was made to set the control attribute on a file. this attribute is not supported in the destination file system")
	ERROR_UNKNOWN_REVISION                                            = errors.New("indicates a revision number that was encountered or specified is not one that is known by the service. it might be a more recent revision than the service is aware of")
	ERROR_REVISION_MISMATCH                                           = errors.New("indicates that two revision levels are incompatible")
	ERROR_INVALID_OWNER                                               = errors.New("indicates a particular security id cannot be assigned as the owner of an object")
	ERROR_INVALID_PRIMARY_GROUP                                       = errors.New("indicates a particular security id cannot be assigned as the primary group of an object")
	ERROR_NO_IMPERSONATION_TOKEN                                      = errors.New("an attempt has been made to operate on an impersonation token by a thread that is not currently impersonating a client")
	ERROR_CANT_DISABLE_MANDATORY                                      = errors.New("a mandatory group cannot be disabled")
	ERROR_NO_LOGON_SERVERS                                            = errors.New("no logon servers are currently available to service the logon request")
	ERROR_NO_SUCH_LOGON_SESSION                                       = errors.New("a specified logon session does not exist. it might already have been terminated")
	ERROR_NO_SUCH_PRIVILEGE                                           = errors.New("a specified privilege does not exist")
	ERROR_PRIVILEGE_NOT_HELD                                          = errors.New("a required privilege is not held by the client")
	ERROR_INVALID_ACCOUNT_NAME                                        = errors.New("the name provided is not a properly formed account name")
	ERROR_USER_EXISTS                                                 = errors.New("the specified account already exists")
	ERROR_NO_SUCH_USER                                                = errors.New("the specified account does not exist")
	ERROR_GROUP_EXISTS                                                = errors.New("the specified group already exists")
	ERROR_NO_SUCH_GROUP                                               = errors.New("the specified group does not exist")
	ERROR_MEMBER_IN_GROUP                                             = errors.New("the specified user account is already in the specified group account. also used to indicate a group cannot be deleted because it contains a member")
	ERROR_MEMBER_NOT_IN_GROUP                                         = errors.New("the specified user account is not a member of the specified group account")
	ERROR_LAST_ADMIN                                                  = errors.New("indicates the requested operation would disable or delete the last remaining administration account. this is not allowed to prevent creating a situation in which the system cannot be administrated")
	ERROR_WRONG_PASSWORD                                              = errors.New("when trying to update a password, this return status indicates that the value provided as the current password is not correct")
	ERROR_ILL_FORMED_PASSWORD                                         = errors.New("when trying to update a password, this return status indicates that the value provided for the new password contains values that are not allowed in passwords")
	ERROR_PASSWORD_RESTRICTION                                        = errors.New("when trying to update a password, this status indicates that some password update rule has been violated. for example, the password might not meet length criteria")
	ERROR_LOGON_FAILURE                                               = errors.New("the attempted logon is invalid. this is either due to a bad username or authentication information")
	ERROR_ACCOUNT_RESTRICTION                                         = errors.New("indicates a referenced user name and authentication information are valid, but some user account restriction has prevented successful authentication (such as time-of-day restrictions)")
	ERROR_INVALID_LOGON_HOURS                                         = errors.New("the user account has time restrictions and cannot be logged onto at this time")
	ERROR_INVALID_WORKSTATION                                         = errors.New("the user account is restricted so that it cannot be used to log on from the source workstation")
	ERROR_PASSWORD_EXPIRED                                            = errors.New("the user account password has expired")
	ERROR_ACCOUNT_DISABLED                                            = errors.New("the referenced account is currently disabled and cannot be logged on to")
	ERROR_NONE_MAPPED                                                 = errors.New("none of the information to be translated has been translated")
	ERROR_TOO_MANY_LUIDS_REQUESTED                                    = errors.New("the number of luids requested cannot be allocated with a single allocation")
	ERROR_LUIDS_EXHAUSTED                                             = errors.New("indicates there are no more luids to allocate")
	ERROR_INVALID_SUB_AUTHORITY                                       = errors.New("indicates the sub-authority value is invalid for the particular use")
	ERROR_INVALID_ACL                                                 = errors.New("indicates the acl structure is not valid")
	ERROR_INVALID_SID                                                 = errors.New("indicates the sid structure is not valid")
	ERROR_INVALID_SECURITY_DESCR                                      = errors.New("indicates the security_descriptor structure is not valid")
	ERROR_PROCEDURE_NOT_FOUND                                         = errors.New("indicates the specified procedure address cannot be found in the dll")
	ERROR_INVALID_IMAGE_FORMAT                                        = errors.New("{bad image} %hs is either not designed to run on windows or it contains an error. try installing the program again using the original installation media or contact your system administrator or the software vendor for support")
	ERROR_NO_TOKEN                                                    = errors.New("an attempt was made to reference a token that does not exist. this is typically done by referencing the token that is associated with a thread when the thread is not impersonating a client")
	ERROR_BAD_INHERITANCE_ACL                                         = errors.New("indicates that an attempt to build either an inherited acl or ace was not successful. this can be caused by a number of things. one of the more probable causes is the replacement of a creatorid with a sid that did not fit into the ace or acl")
	ERROR_RANGE_NOT_LOCKED                                            = errors.New("the range specified in ntunlockfile was not locked")
	ERROR_DISK_FULL                                                   = errors.New("an operation failed because the disk was full")
	ERROR_SERVER_DISABLED                                             = errors.New("the guid allocation server is disabled at the moment")
	ERROR_SERVER_NOT_DISABLED                                         = errors.New("the guid allocation server is enabled at the moment")
	ERROR_TOO_MANY_GUIDS_REQUESTED                                    = errors.New("too many guids were requested from the allocation server at once")
	ERROR_GUIDS_EXHAUSTED                                             = errors.New("the guids could not be allocated because the authority agent was exhausted")
	ERROR_INVALID_ID_AUTHORITY                                        = errors.New("the value provided was an invalid value for an identifier authority")
	ERROR_AGENTS_EXHAUSTED                                            = errors.New("no more authority agent values are available for the particular identifier authority value")
	ERROR_INVALID_VOLUME_LABEL                                        = errors.New("an invalid volume label has been specified")
	ERROR_SECTION_NOT_EXTENDED                                        = errors.New("a mapped section could not be extended")
	ERROR_NOT_MAPPED_DATA                                             = errors.New("specified section to flush does not map a data file")
	ERROR_RESOURCE_DATA_NOT_FOUND                                     = errors.New("indicates the specified image file did not contain a resource section")
	ERROR_RESOURCE_TYPE_NOT_FOUND                                     = errors.New("indicates the specified resource type cannot be found in the image file")
	ERROR_RESOURCE_NAME_NOT_FOUND                                     = errors.New("indicates the specified resource name cannot be found in the image file")
	ERROR_ARRAY_BOUNDS_EXCEEDED                                       = errors.New("{exception} array bounds exceeded")
	ERROR_FLOAT_DENORMAL_OPERAND                                      = errors.New("{exception} floating-point denormal operand")
	ERROR_FLOAT_DIVIDE_BY_ZERO                                        = errors.New("{exception} floating-point division by zero")
	ERROR_FLOAT_INEXACT_RESULT                                        = errors.New("{exception} floating-point inexact result")
	ERROR_FLOAT_INVALID_OPERATION                                     = errors.New("{exception} floating-point invalid operation")
	ERROR_FLOAT_OVERFLOW                                              = errors.New("{exception} floating-point overflow")
	ERROR_FLOAT_STACK_CHECK                                           = errors.New("{exception} floating-point stack check")
	ERROR_FLOAT_UNDERFLOW                                             = errors.New("{exception} floating-point underflow")
	ERROR_INTEGER_DIVIDE_BY_ZERO                                      = errors.New("{exception} integer division by zero")
	ERROR_INTEGER_OVERFLOW                                            = errors.New("{exception} integer overflow")
	ERROR_PRIVILEGED_INSTRUCTION                                      = errors.New("{exception} privileged instruction")
	ERROR_TOO_MANY_PAGING_FILES                                       = errors.New("an attempt was made to install more paging files than the system supports")
	ERROR_FILE_INVALID                                                = errors.New("the volume for a file has been externally altered such that the opened file is no longer valid")
	ERROR_ALLOTTED_SPACE_EXCEEDED                                     = errors.New("when a block of memory is allotted for future updates, such as the memory allocated to hold discretionary access control and primary group information, successive updates might exceed the amount of memory originally allotted. because a quota might already have been charged to several processes that have handles to the object, it is not reasonable to alter the size of the allocated memory. instead, a request that requires more memory than has been allotted must fail and the status_allotted_space_exceeded error returned")
	ERROR_INSUFFICIENT_RESOURCES                                      = errors.New("insufficient system resources exist to complete the api")
	ERROR_DFS_EXIT_PATH_FOUND                                         = errors.New("an attempt has been made to open a dfs exit path control file")
	ERROR_DEVICE_DATA_ERROR                                           = errors.New("there are bad blocks (sectors) on the hard disk")
	ERROR_DEVICE_NOT_CONNECTED                                        = errors.New("there is bad cabling, non-termination, or the controller is not able to obtain access to the hard disk")
	ERROR_FREE_VM_NOT_AT_BASE                                         = errors.New("virtual memory cannot be freed because the base address is not the base of the region and a region size of zero was specified")
	ERROR_MEMORY_NOT_ALLOCATED                                        = errors.New("an attempt was made to free virtual memory that is not allocated")
	ERROR_WORKING_SET_QUOTA                                           = errors.New("the working set is not big enough to allow the requested pages to be locked")
	ERROR_MEDIA_WRITE_PROTECTED                                       = errors.New("{write protect error} the disk cannot be written to because it is write-protected. remove the write protection from the volume %hs in drive %hs")
	ERROR_DEVICE_NOT_READY                                            = errors.New("{drive not ready} the drive is not ready for use; its door might be open. check drive %hs and make sure that a disk is inserted and that the drive door is closed")
	ERROR_INVALID_GROUP_ATTRIBUTES                                    = errors.New("the specified attributes are invalid or are incompatible with the attributes for the group as a whole")
	ERROR_BAD_IMPERSONATION_LEVEL                                     = errors.New("a specified impersonation level is invalid. also used to indicate that a required impersonation level was not provided")
	ERROR_CANT_OPEN_ANONYMOUS                                         = errors.New("an attempt was made to open an anonymous-level token. anonymous tokens cannot be opened")
	ERROR_BAD_VALIDATION_CLASS                                        = errors.New("the validation information class requested was invalid")
	ERROR_BAD_TOKEN_TYPE                                              = errors.New("the type of a token object is inappropriate for its attempted use")
	ERROR_BAD_MASTER_BOOT_RECORD                                      = errors.New("the type of a token object is inappropriate for its attempted use")
	ERROR_INSTRUCTION_MISALIGNMENT                                    = errors.New("an attempt was made to execute an instruction at an unaligned address and the host system does not support unaligned instruction references")
	ERROR_INSTANCE_NOT_AVAILABLE                                      = errors.New("the maximum named pipe instance count has been reached")
	ERROR_PIPE_NOT_AVAILABLE                                          = errors.New("an instance of a named pipe cannot be found in the listening state")
	ERROR_INVALID_PIPE_STATE                                          = errors.New("the named pipe is not in the connected or closing state")
	ERROR_PIPE_BUSY                                                   = errors.New("the specified pipe is set to complete operations and there are current i/o operations queued so that it cannot be changed to queue operations")
	ERROR_ILLEGAL_FUNCTION                                            = errors.New("the specified handle is not open to the server end of the named pipe")
	ERROR_PIPE_DISCONNECTED                                           = errors.New("the specified named pipe is in the disconnected state")
	ERROR_PIPE_CLOSING                                                = errors.New("the specified named pipe is in the closing state")
	ERROR_PIPE_CONNECTED                                              = errors.New("the specified named pipe is in the connected state")
	ERROR_PIPE_LISTENING                                              = errors.New("the specified named pipe is in the listening state")
	ERROR_INVALID_READ_MODE                                           = errors.New("the specified named pipe is not in message mode")
	ERROR_IO_TIMEOUT                                                  = errors.New("{device timeout} the specified i/o operation on %hs was not completed before the time-out period expired")
	ERROR_FILE_FORCED_CLOSED                                          = errors.New("the specified file has been closed by another process")
	ERROR_PROFILING_NOT_STARTED                                       = errors.New("profiling is not started")
	ERROR_PROFILING_NOT_STOPPED                                       = errors.New("profiling is not stopped")
	ERROR_COULD_NOT_INTERPRET                                         = errors.New("the passed acl did not contain the minimum required information")
	ERROR_FILE_IS_A_DIRECTORY                                         = errors.New("the file that was specified as a target is a directory, and the caller specified that it could be anything but a directory")
	ERROR_NOT_SUPPORTED                                               = errors.New("the request is not supported")
	ERROR_REMOTE_NOT_LISTENING                                        = errors.New("this remote computer is not listening")
	ERROR_DUPLICATE_NAME                                              = errors.New("a duplicate name exists on the network")
	ERROR_BAD_NETWORK_PATH                                            = errors.New("the network path cannot be located")
	ERROR_NETWORK_BUSY                                                = errors.New("the network is busy")
	ERROR_DEVICE_DOES_NOT_EXIST                                       = errors.New("this device does not exist")
	ERROR_TOO_MANY_COMMANDS                                           = errors.New("the network bios command limit has been reached")
	ERROR_ADAPTER_HARDWARE_ERROR                                      = errors.New("an i/o adapter hardware error has occurred")
	ERROR_INVALID_NETWORK_RESPONSE                                    = errors.New("the network responded incorrectly")
	ERROR_UNEXPECTED_NETWORK_ERROR                                    = errors.New("an unexpected network error occurred")
	ERROR_BAD_REMOTE_ADAPTER                                          = errors.New("the remote adapter is not compatible")
	ERROR_PRINT_QUEUE_FULL                                            = errors.New("the print queue is full")
	ERROR_NO_SPOOL_SPACE                                              = errors.New("space to store the file that is waiting to be printed is not available on the server")
	ERROR_PRINT_CANCELLED                                             = errors.New("the requested print file has been canceled")
	ERROR_NETWORK_NAME_DELETED                                        = errors.New("the network name was deleted")
	ERROR_NETWORK_ACCESS_DENIED                                       = errors.New("network access is denied")
	ERROR_BAD_DEVICE_TYPE                                             = errors.New("{incorrect network resource type} the specified device type (lpt, for example) conflicts with the actual device type on the remote resource")
	ERROR_BAD_NETWORK_NAME                                            = errors.New("{network name not found} the specified share name cannot be found on the remote server")
	ERROR_TOO_MANY_NAMES                                              = errors.New("the name limit for the network adapter card of the local computer was exceeded")
	ERROR_TOO_MANY_SESSIONS                                           = errors.New("the network bios session limit was exceeded")
	ERROR_SHARING_PAUSED                                              = errors.New("file sharing has been temporarily paused")
	ERROR_REQUEST_NOT_ACCEPTED                                        = errors.New("no more connections can be made to this remote computer at this time because the computer has already accepted the maximum number of connections")
	ERROR_REDIRECTOR_PAUSED                                           = errors.New("print or disk redirection is temporarily paused")
	ERROR_NET_WRITE_FAULT                                             = errors.New("a network data fault occurred")
	ERROR_PROFILING_AT_LIMIT                                          = errors.New("the number of active profiling objects is at the maximum and no more can be started")
	ERROR_NOT_SAME_DEVICE                                             = errors.New("{incorrect volume} the destination file of a rename request is located on a different device than the source of the rename request")
	ERROR_FILE_RENAMED                                                = errors.New("the specified file has been renamed and thus cannot be modified")
	ERROR_VIRTUAL_CIRCUIT_CLOSED                                      = errors.New("{network request timeout} the session with a remote server has been disconnected because the time-out interval for a request has expired")
	ERROR_NO_SECURITY_ON_OBJECT                                       = errors.New("indicates an attempt was made to operate on the security of an object that does not have security associated with it")
	ERROR_CANT_WAIT                                                   = errors.New("used to indicate that an operation cannot continue without blocking for i/o")
	ERROR_PIPE_EMPTY                                                  = errors.New("used to indicate that a read operation was done on an empty pipe")
	ERROR_CANT_ACCESS_DOMAIN_INFO                                     = errors.New("configuration information could not be read from the domain controller, either because the machine is unavailable or access has been denied")
	ERROR_CANT_TERMINATE_SELF                                         = errors.New("indicates that a thread attempted to terminate itself by default (called ntterminatethread with null) and it was the last thread in the current process")
	ERROR_INVALID_SERVER_STATE                                        = errors.New("indicates the sam server was in the wrong state to perform the desired operation")
	ERROR_INVALID_DOMAIN_STATE                                        = errors.New("indicates the domain was in the wrong state to perform the desired operation")
	ERROR_INVALID_DOMAIN_ROLE                                         = errors.New("this operation is only allowed for the primary domain controller of the domain")
	ERROR_NO_SUCH_DOMAIN                                              = errors.New("the specified domain did not exist")
	ERROR_DOMAIN_EXISTS                                               = errors.New("the specified domain already exists")
	ERROR_DOMAIN_LIMIT_EXCEEDED                                       = errors.New("an attempt was made to exceed the limit on the number of domains per server for this release")
	ERROR_OPLOCK_NOT_GRANTED                                          = errors.New("an error status returned when the opportunistic lock (oplock) request is denied")
	ERROR_INVALID_OPLOCK_PROTOCOL                                     = errors.New("an error status returned when an invalid opportunistic lock (oplock) acknowledgment is received by a file system")
	ERROR_INTERNAL_DB_CORRUPTION                                      = errors.New("this error indicates that the requested operation cannot be completed due to a catastrophic media failure or an on-disk data structure corruption")
	ERROR_INTERNAL_ERROR                                              = errors.New("an internal error occurred")
	ERROR_GENERIC_NOT_MAPPED                                          = errors.New("indicates generic access types were contained in an access mask which should already be mapped to non-generic access types")
	ERROR_BAD_DESCRIPTOR_FORMAT                                       = errors.New("indicates a security descriptor is not in the necessary format (absolute or self-relative)")
	ERROR_INVALID_USER_BUFFER                                         = errors.New("an access to a user buffer failed at an expected point in time. this code is defined because the caller does not want to accept status_access_violation in its filter")
	ERROR_UNEXPECTED_IO_ERROR                                         = errors.New("if an i/o error that is not defined in the standard fsrtl filter is returned, it is converted to the following error, which is guaranteed to be in the filter. in this case, information is lost; however, the filter correctly handles the exception")
	ERROR_UNEXPECTED_MM_CREATE_ERR                                    = errors.New("if an mm error that is not defined in the standard fsrtl filter is returned, it is converted to one of the following errors, which are guaranteed to be in the filter. in this case, information is lost; however, the filter correctly handles the exception")
	ERROR_UNEXPECTED_MM_MAP_ERROR                                     = errors.New("if an mm error that is not defined in the standard fsrtl filter is returned, it is converted to one of the following errors, which are guaranteed to be in the filter. in this case, information is lost; however, the filter correctly handles the exception")
	ERROR_UNEXPECTED_MM_EXTEND_ERR                                    = errors.New("if an mm error that is not defined in the standard fsrtl filter is returned, it is converted to one of the following errors, which are guaranteed to be in the filter. in this case, information is lost; however, the filter correctly handles the exception")
	ERROR_NOT_LOGON_PROCESS                                           = errors.New("the requested action is restricted for use by logon processes only. the calling process has not registered as a logon process")
	ERROR_LOGON_SESSION_EXISTS                                        = errors.New("an attempt has been made to start a new session manager or lsa logon session by using an id that is already in use")
	ERROR_INVALID_PARAMETER_1                                         = errors.New("an invalid parameter was passed to a service or function as the first argument")
	ERROR_INVALID_PARAMETER_2                                         = errors.New("an invalid parameter was passed to a service or function as the second argument")
	ERROR_INVALID_PARAMETER_3                                         = errors.New("an invalid parameter was passed to a service or function as the third argument")
	ERROR_INVALID_PARAMETER_4                                         = errors.New("an invalid parameter was passed to a service or function as the fourth argument")
	ERROR_INVALID_PARAMETER_5                                         = errors.New("an invalid parameter was passed to a service or function as the fifth argument")
	ERROR_INVALID_PARAMETER_6                                         = errors.New("an invalid parameter was passed to a service or function as the sixth argument")
	ERROR_INVALID_PARAMETER_7                                         = errors.New("an invalid parameter was passed to a service or function as the seventh argument")
	ERROR_INVALID_PARAMETER_8                                         = errors.New("an invalid parameter was passed to a service or function as the eighth argument")
	ERROR_INVALID_PARAMETER_9                                         = errors.New("an invalid parameter was passed to a service or function as the ninth argument")
	ERROR_INVALID_PARAMETER_10                                        = errors.New("an invalid parameter was passed to a service or function as the tenth argument")
	ERROR_INVALID_PARAMETER_11                                        = errors.New("an invalid parameter was passed to a service or function as the eleventh argument")
	ERROR_INVALID_PARAMETER_12                                        = errors.New("an invalid parameter was passed to a service or function as the twelfth argument")
	ERROR_REDIRECTOR_NOT_STARTED                                      = errors.New("an attempt was made to access a network file, but the network software was not yet started")
	ERROR_REDIRECTOR_STARTED                                          = errors.New("an attempt was made to start the redirector, but the redirector has already been started")
	ERROR_STACK_OVERFLOW                                              = errors.New("a new guard page for the stack cannot be created")
	ERROR_NO_SUCH_PACKAGE                                             = errors.New("a specified authentication package is unknown")
	ERROR_BAD_FUNCTION_TABLE                                          = errors.New("a malformed function table was encountered during an unwind operation")
	ERROR_VARIABLE_NOT_FOUND                                          = errors.New("indicates the specified environment variable name was not found in the specified environment block")
	ERROR_DIRECTORY_NOT_EMPTY                                         = errors.New("indicates that the directory trying to be deleted is not empty")
	ERROR_FILE_CORRUPT_ERROR                                          = errors.New("{corrupt file} the file or directory %hs is corrupt and unreadable. run the chkdsk utility")
	ERROR_NOT_A_DIRECTORY                                             = errors.New("a requested opened file is not a directory")
	ERROR_BAD_LOGON_SESSION_STATE                                     = errors.New("the logon session is not in a state that is consistent with the requested operation")
	ERROR_LOGON_SESSION_COLLISION                                     = errors.New("an internal lsa error has occurred. an authentication package has requested the creation of a logon session but the id of an already existing logon session has been specified")
	ERROR_NAME_TOO_LONG                                               = errors.New("a specified name string is too long for its intended use")
	ERROR_FILES_OPEN                                                  = errors.New("the user attempted to force close the files on a redirected drive, but there were opened files on the drive, and the user did not specify a sufficient level of force")
	ERROR_CONNECTION_IN_USE                                           = errors.New("the user attempted to force close the files on a redirected drive, but there were opened directories on the drive, and the user did not specify a sufficient level of force")
	ERROR_MESSAGE_NOT_FOUND                                           = errors.New("rtlfindmessage could not locate the requested message id in the message table resource")
	ERROR_PROCESS_IS_TERMINATING                                      = errors.New("an attempt was made to duplicate an object handle into or out of an exiting process")
	ERROR_INVALID_LOGON_TYPE                                          = errors.New("indicates an invalid value has been provided for the logontype requested")
	ERROR_NO_GUID_TRANSLATION                                         = errors.New("indicates that an attempt was made to assign protection to a file system file or directory and one of the sids in the security descriptor could not be translated into a guid that could be stored by the file system. this causes the protection attempt to fail, which might cause a file creation attempt to fail")
	ERROR_CANNOT_IMPERSONATE                                          = errors.New("indicates that an attempt has been made to impersonate via a named pipe that has not yet been read from")
	ERROR_IMAGE_ALREADY_LOADED                                        = errors.New("indicates that the specified image is already loaded")
	ERROR_NO_LDT                                                      = errors.New("indicates that an attempt was made to change the size of the ldt for a process that has no ldt")
	ERROR_INVALID_LDT_SIZE                                            = errors.New("indicates that an attempt was made to grow an ldt by setting its size, or that the size was not an even number of selectors")
	ERROR_INVALID_LDT_OFFSET                                          = errors.New("indicates that the starting value for the ldt information was not an integral multiple of the selector size")
	ERROR_INVALID_LDT_DESCRIPTOR                                      = errors.New("indicates that the user supplied an invalid descriptor when trying to set up ldt descriptors")
	ERROR_INVALID_IMAGE_NE_FORMAT                                     = errors.New("the specified image file did not have the correct format. it appears to be ne format")
	ERROR_RXACT_INVALID_STATE                                         = errors.New("indicates that the transaction state of a registry subtree is incompatible with the requested operation. for example, a request has been made to start a new transaction with one already in progress, or a request has been made to apply a transaction when one is not currently in progress")
	ERROR_RXACT_COMMIT_FAILURE                                        = errors.New("indicates an error has occurred during a registry transaction commit. the database has been left in an unknown, but probably inconsistent, state. the state of the registry transaction is left as committing")
	ERROR_MAPPED_FILE_SIZE_ZERO                                       = errors.New("an attempt was made to map a file of size zero with the maximum size specified as zero")
	ERROR_TOO_MANY_OPENED_FILES                                       = errors.New("too many files are opened on a remote server. this error should only be returned by the windows redirector on a remote drive")
	ERROR_CANCELLED                                                   = errors.New("the i/o request was canceled")
	ERROR_CANNOT_DELETE                                               = errors.New("an attempt has been made to remove a file or directory that cannot be deleted")
	ERROR_INVALID_COMPUTER_NAME                                       = errors.New("indicates a name that was specified as a remote computer name is syntactically invalid")
	ERROR_FILE_DELETED                                                = errors.New("an i/o request other than close was performed on a file after it was deleted, which can only happen to a request that did not complete before the last handle was closed via ntclose")
	ERROR_SPECIAL_ACCOUNT                                             = errors.New("indicates an operation that is incompatible with built-in accounts has been attempted on a built-in (special) sam account. for example, built-in accounts cannot be deleted")
	ERROR_SPECIAL_GROUP                                               = errors.New("the operation requested cannot be performed on the specified group because it is a built-in special group")
	ERROR_SPECIAL_USER                                                = errors.New("the operation requested cannot be performed on the specified user because it is a built-in special user")
	ERROR_MEMBERS_PRIMARY_GROUP                                       = errors.New("indicates a member cannot be removed from a group because the group is currently the member's primary group")
	ERROR_FILE_CLOSED                                                 = errors.New("an i/o request other than close and several other special case operations was attempted using a file object that had already been closed")
	ERROR_TOO_MANY_THREADS                                            = errors.New("indicates a process has too many threads to perform the requested action. for example, assignment of a primary token can be performed only when a process has zero or one threads")
	ERROR_THREAD_NOT_IN_PROCESS                                       = errors.New("an attempt was made to operate on a thread within a specific process, but the specified thread is not in the specified process")
	ERROR_TOKEN_ALREADY_IN_USE                                        = errors.New("an attempt was made to establish a token for use as a primary token but the token is already in use. a token can only be the primary token of one process at a time")
	ERROR_PAGEFILE_QUOTA_EXCEEDED                                     = errors.New("the page file quota was exceeded")
	ERROR_COMMITMENT_LIMIT                                            = errors.New("{out of virtual memory} your system is low on virtual memory. to ensure that windows runs correctly, increase the size of your virtual memory paging file. for more information, see help")
	ERROR_INVALID_IMAGE_LE_FORMAT                                     = errors.New("the specified image file did not have the correct format: it appears to be le format")
	ERROR_INVALID_IMAGE_NOT_MZ                                        = errors.New("the specified image file did not have the correct format: it did not have an initial mz")
	ERROR_INVALID_IMAGE_PROTECT                                       = errors.New("the specified image file did not have the correct format: it did not have a proper e_lfarlc in the mz header")
	ERROR_INVALID_IMAGE_WIN_16                                        = errors.New("the specified image file did not have the correct format: it appears to be a 16-bit windows image")
	ERROR_LOGON_SERVER_CONFLICT                                       = errors.New("the netlogon service cannot start because another netlogon service running in the domain conflicts with the specified role")
	ERROR_TIME_DIFFERENCE_AT_DC                                       = errors.New("the time at the primary domain controller is different from the time at the backup domain controller or member server by too large an amount")
	ERROR_SYNCHRONIZATION_REQUIRED                                    = errors.New("on applicable windows server releases, the sam database is significantly out of synchronization with the copy on the domain controller. a complete synchronization is required")
	ERROR_DLL_NOT_FOUND                                               = errors.New("{unable to locate component} this application has failed to start because %hs was not found. reinstalling the application might fix this problem")
	ERROR_OPEN_FAILED                                                 = errors.New("the ntcreatefile api failed. this error should never be returned to an application; it is a place holder for the windows lan manager redirector to use in its internal error-mapping routines")
	ERROR_IO_PRIVILEGE_FAILED                                         = errors.New("{privilege failed} the i/o permissions for the process could not be changed")
	ERROR_ORDINAL_NOT_FOUND                                           = errors.New("{ordinal not found} the ordinal %ld could not be located in the dynamic link library %hs")
	ERROR_ENTRYPOINT_NOT_FOUND                                        = errors.New("{entry point not found} the procedure entry point %hs could not be located in the dynamic link library %hs")
	ERROR_CONTROL_C_EXIT                                              = errors.New("{application exit by ctrl+c} the application terminated as a result of a ctrl+c")
	ERROR_LOCAL_DISCONNECT                                            = errors.New("{virtual circuit closed} the network transport on your computer has closed a network connection. there might or might not be i/o requests outstanding")
	ERROR_REMOTE_DISCONNECT                                           = errors.New("{virtual circuit closed} the network transport on a remote computer has closed a network connection. there might or might not be i/o requests outstanding")
	ERROR_REMOTE_RESOURCES                                            = errors.New("{insufficient resources on remote computer} the remote computer has insufficient resources to complete the network request. for example, the remote computer might not have enough available memory to carry out the request at this time")
	ERROR_LINK_FAILED                                                 = errors.New("{virtual circuit closed} an existing connection (virtual circuit) has been broken at the remote computer. there is probably something wrong with the network software protocol or the network hardware on the remote computer")
	ERROR_LINK_TIMEOUT                                                = errors.New("{virtual circuit closed} the network transport on your computer has closed a network connection because it had to wait too long for a response from the remote computer")
	ERROR_INVALID_CONNECTION                                          = errors.New("the connection handle that was given to the transport was invalid")
	ERROR_INVALID_ADDRESS                                             = errors.New("the address handle that was given to the transport was invalid")
	ERROR_DLL_INIT_FAILED                                             = errors.New("{dll initialization failed} initialization of the dynamic link library %hs failed. the process is terminating abnormally")
	ERROR_MISSING_SYSTEMFILE                                          = errors.New("{missing system file} the required system file %hs is bad or missing")
	ERROR_UNHANDLED_EXCEPTION                                         = errors.New("{application error} the exception %s (0x%08lx) occurred in the application at location 0x%08lx")
	ERROR_APP_INIT_FAILURE                                            = errors.New("{application error} the application failed to initialize properly (0x%lx). click ok to terminate the application")
	ERROR_PAGEFILE_CREATE_FAILED                                      = errors.New("{unable to create paging file} the creation of the paging file %hs failed (%lx). the requested size was %ld")
	ERROR_NO_PAGEFILE                                                 = errors.New("{no paging file specified} no paging file was specified in the system configuration")
	ERROR_INVALID_LEVEL                                               = errors.New("{incorrect system call level} an invalid level was passed into the specified system call")
	ERROR_WRONG_PASSWORD_CORE                                         = errors.New("{incorrect password to lan manager server} you specified an incorrect password to a lan manager 2.x or ms-net server")
	ERROR_ILLEGAL_FLOAT_CONTEXT                                       = errors.New("{exception} a real-mode application issued a floating-point instruction and floating-point hardware is not present")
	ERROR_PIPE_BROKEN                                                 = errors.New("the pipe operation has failed because the other end of the pipe has been closed")
	ERROR_REGISTRY_CORRUPT                                            = errors.New("{the registry is corrupt} the structure of one of the files that contains registry data is corrupt; the image of the file in memory is corrupt; or the file could not be recovered because the alternate copy or log was absent or corrupt")
	ERROR_REGISTRY_IO_FAILED                                          = errors.New("an i/o operation initiated by the registry failed and cannot be recovered. the registry could not read in, write out, or flush one of the files that contain the system's image of the registry")
	ERROR_NO_EVENT_PAIR                                               = errors.New("an event pair synchronization operation was performed using the thread-specific client/server event pair object, but no event pair object was associated with the thread")
	ERROR_UNRECOGNIZED_VOLUME                                         = errors.New("the volume does not contain a recognized file system. be sure that all required file system drivers are loaded and that the volume is not corrupt")
	ERROR_SERIAL_NO_DEVICE_INITED                                     = errors.New("no serial device was successfully initialized. the serial driver will unload")
	ERROR_NO_SUCH_ALIAS                                               = errors.New("the specified local group does not exist")
	ERROR_MEMBER_NOT_IN_ALIAS                                         = errors.New("the specified account name is not a member of the group")
	ERROR_MEMBER_IN_ALIAS                                             = errors.New("the specified account name is already a member of the group")
	ERROR_ALIAS_EXISTS                                                = errors.New("the specified local group already exists")
	ERROR_LOGON_NOT_GRANTED                                           = errors.New("a requested type of logon (for example, interactive, network, and service) is not granted by the local security policy of the target system. ask the system administrator to grant the necessary form of logon")
	ERROR_TOO_MANY_SECRETS                                            = errors.New("the maximum number of secrets that can be stored in a single system was exceeded. the length and number of secrets is limited to satisfy u.s. state department export restrictions")
	ERROR_SECRET_TOO_LONG                                             = errors.New("the length of a secret exceeds the maximum allowable length. the length and number of secrets is limited to satisfy u.s. state department export restrictions")
	ERROR_INTERNAL_DB_ERROR                                           = errors.New("the local security authority (lsa) database contains an internal inconsistency")
	ERROR_FULLSCREEN_MODE                                             = errors.New("the requested operation cannot be performed in full-screen mode")
	ERROR_TOO_MANY_CONTEXT_IDS                                        = errors.New("during a logon attempt, the user's security context accumulated too many security ids. this is a very unusual situation. remove the user from some global or local groups to reduce the number of security ids to incorporate into the security context")
	ERROR_LOGON_TYPE_NOT_GRANTED                                      = errors.New("a user has requested a type of logon (for example, interactive or network) that has not been granted. an administrator has control over who can logon interactively and through the network")
	ERROR_NOT_REGISTRY_FILE                                           = errors.New("the system has attempted to load or restore a file into the registry, and the specified file is not in the format of a registry file")
	ERROR_NT_CROSS_ENCRYPTION_REQUIRED                                = errors.New("an attempt was made to change a user password in the security account manager without providing the necessary windows cross-encrypted password")
	ERROR_DOMAIN_CTRLR_CONFIG_ERROR                                   = errors.New("a domain server has an incorrect configuration")
	ERROR_FT_MISSING_MEMBER                                           = errors.New("an attempt was made to explicitly access the secondary copy of information via a device control to the fault tolerance driver and the secondary copy is not present in the system")
	ERROR_ILL_FORMED_SERVICE_ENTRY                                    = errors.New("a configuration registry node that represents a driver service entry was ill-formed and did not contain the required value entries")
	ERROR_ILLEGAL_CHARACTER                                           = errors.New("an illegal character was encountered. for a multibyte character set, this includes a lead byte without a succeeding trail byte. for the unicode character set this includes the characters 0xffff and 0xfffe")
	ERROR_UNMAPPABLE_CHARACTER                                        = errors.New("no mapping for the unicode character exists in the target multibyte code page")
	ERROR_UNDEFINED_CHARACTER                                         = errors.New("the unicode character is not defined in the unicode character set that is installed on the system")
	ERROR_FLOPPY_VOLUME                                               = errors.New("the paging file cannot be created on a floppy disk")
	ERROR_FLOPPY_ID_MARK_NOT_FOUND                                    = errors.New("{floppy disk error} while accessing a floppy disk, an id address mark was not found")
	ERROR_FLOPPY_WRONG_CYLINDER                                       = errors.New("{floppy disk error} while accessing a floppy disk, the track address from the sector id field was found to be different from the track address that is maintained by the controller")
	ERROR_FLOPPY_UNKNOWN_ERROR                                        = errors.New("{floppy disk error} the floppy disk controller reported an error that is not recognized by the floppy disk driver")
	ERROR_FLOPPY_BAD_REGISTERS                                        = errors.New("{floppy disk error} while accessing a floppy-disk, the controller returned inconsistent results via its registers")
	ERROR_DISK_RECALIBRATE_FAILED                                     = errors.New("{hard disk error} while accessing the hard disk, a recalibrate operation failed, even after retries")
	ERROR_DISK_OPERATION_FAILED                                       = errors.New("{hard disk error} while accessing the hard disk, a disk operation failed even after retries")
	ERROR_DISK_RESET_FAILED                                           = errors.New("{hard disk error} while accessing the hard disk, a disk controller reset was needed, but even that failed")
	ERROR_SHARED_IRQ_BUSY                                             = errors.New("an attempt was made to open a device that was sharing an interrupt request (irq) with other devices. at least one other device that uses that irq was already opened. two concurrent opens of devices that share an irq and only work via interrupts is not supported for the particular bus type that the devices use")
	ERROR_FT_ORPHANING                                                = errors.New("{ft orphaning} a disk that is part of a fault-tolerant volume can no longer be accessed")
	ERROR_BIOS_FAILED_TO_CONNECT_INTERRUPT                            = errors.New("the basic input/output system (bios) failed to connect a system interrupt to the device or bus for which the device is connected")
	ERROR_PARTITION_FAILURE                                           = errors.New("the tape could not be partitioned")
	ERROR_INVALID_BLOCK_LENGTH                                        = errors.New("when accessing a new tape of a multi-volume partition, the current blocksize is incorrect")
	ERROR_DEVICE_NOT_PARTITIONED                                      = errors.New("the tape partition information could not be found when loading a tape")
	ERROR_UNABLE_TO_LOCK_MEDIA                                        = errors.New("an attempt to lock the eject media mechanism failed")
	ERROR_UNABLE_TO_UNLOAD_MEDIA                                      = errors.New("an attempt to unload media failed")
	ERROR_EOM_OVERFLOW                                                = errors.New("the physical end of tape was detected")
	ERROR_NO_MEDIA                                                    = errors.New("{no media} there is no media in the drive. insert media into drive %hs")
	ERROR_NO_SUCH_MEMBER                                              = errors.New("a member could not be added to or removed from the local group because the member does not exist")
	ERROR_INVALID_MEMBER                                              = errors.New("a new member could not be added to a local group because the member has the wrong account type")
	ERROR_KEY_DELETED                                                 = errors.New("an illegal operation was attempted on a registry key that has been marked for deletion")
	ERROR_NO_LOG_SPACE                                                = errors.New("the system could not allocate the required space in a registry log")
	ERROR_TOO_MANY_SIDS                                               = errors.New("too many sids have been specified")
	ERROR_LM_CROSS_ENCRYPTION_REQUIRED                                = errors.New("an attempt was made to change a user password in the security account manager without providing the necessary lm cross-encrypted password")
	ERROR_KEY_HAS_CHILDREN                                            = errors.New("an attempt was made to create a symbolic link in a registry key that already has subkeys or values")
	ERROR_CHILD_MUST_BE_VOLATILE                                      = errors.New("an attempt was made to create a stable subkey under a volatile parent key")
	ERROR_DEVICE_CONFIGURATION_ERROR                                  = errors.New("the i/o device is configured incorrectly or the configuration parameters to the driver are incorrect")
	ERROR_DRIVER_INTERNAL_ERROR                                       = errors.New("an error was detected between two drivers or within an i/o driver")
	ERROR_INVALID_DEVICE_STATE                                        = errors.New("the device is not in a valid state to perform this request")
	ERROR_IO_DEVICE_ERROR                                             = errors.New("the i/o device reported an i/o error")
	ERROR_DEVICE_PROTOCOL_ERROR                                       = errors.New("a protocol error was detected between the driver and the device")
	ERROR_BACKUP_CONTROLLER                                           = errors.New("this operation is only allowed for the primary domain controller of the domain")
	ERROR_LOG_FILE_FULL                                               = errors.New("the log file space is insufficient to support this operation")
	ERROR_TOO_LATE                                                    = errors.New("a write operation was attempted to a volume after it was dismounted")
	ERROR_NO_TRUST_LSA_SECRET                                         = errors.New("the workstation does not have a trust secret for the primary domain in the local lsa database")
	ERROR_NO_TRUST_SAM_ACCOUNT                                        = errors.New("on applicable windows server releases, the sam database does not have a computer account for this workstation trust relationship")
	ERROR_TRUSTED_DOMAIN_FAILURE                                      = errors.New("the logon request failed because the trust relationship between the primary domain and the trusted domain failed")
	ERROR_TRUSTED_RELATIONSHIP_FAILURE                                = errors.New("the logon request failed because the trust relationship between this workstation and the primary domain failed")
	ERROR_EVENTLOG_FILE_CORRUPT                                       = errors.New("the eventlog log file is corrupt")
	ERROR_EVENTLOG_CANT_START                                         = errors.New("no eventlog log file could be opened. the eventlog service did not start")
	ERROR_TRUST_FAILURE                                               = errors.New("the network logon failed. this might be because the validation authority cannot be reached")
	ERROR_MUTANT_LIMIT_EXCEEDED                                       = errors.New("an attempt was made to acquire a mutant such that its maximum count would have been exceeded")
	ERROR_NETLOGON_NOT_STARTED                                        = errors.New("an attempt was made to logon, but the netlogon service was not started")
	ERROR_ACCOUNT_EXPIRED                                             = errors.New("the user account has expired")
	ERROR_POSSIBLE_DEADLOCK                                           = errors.New("{exception} possible deadlock condition")
	ERROR_NETWORK_CREDENTIAL_CONFLICT                                 = errors.New("multiple connections to a server or shared resource by the same user, using more than one user name, are not allowed. disconnect all previous connections to the server or shared resource and try again")
	ERROR_REMOTE_SESSION_LIMIT                                        = errors.New("an attempt was made to establish a session to a network server, but there are already too many sessions established to that server")
	ERROR_EVENTLOG_FILE_CHANGED                                       = errors.New("the log file has changed between reads")
	ERROR_NOLOGON_INTERDOMAIN_TRUST_ACCOUNT                           = errors.New("the account used is an interdomain trust account. use your global user account or local user account to access this server")
	ERROR_NOLOGON_WORKSTATION_TRUST_ACCOUNT                           = errors.New("the account used is a computer account. use your global user account or local user account to access this server")
	ERROR_NOLOGON_SERVER_TRUST_ACCOUNT                                = errors.New("the account used is a server trust account. use your global user account or local user account to access this server")
	ERROR_DOMAIN_TRUST_INCONSISTENT                                   = errors.New("the name or sid of the specified domain is inconsistent with the trust information for that domain")
	ERROR_FS_DRIVER_REQUIRED                                          = errors.New("a volume has been accessed for which a file system driver is required that has not yet been loaded")
	ERROR_IMAGE_ALREADY_LOADED_AS_DLL                                 = errors.New("indicates that the specified image is already loaded as a dll")
	ERROR_INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING        = errors.New("short name settings cannot be changed on this volume due to the global registry setting")
	ERROR_SHORT_NAMES_NOT_ENABLED_ON_VOLUME                           = errors.New("short names are not enabled on this volume")
	ERROR_SECURITY_STREAM_IS_INCONSISTENT                             = errors.New("the security stream for the given volume is in an inconsistent state. please run chkdsk on the volume")
	ERROR_INVALID_LOCK_RANGE                                          = errors.New("a requested file lock operation cannot be processed due to an invalid byte range")
	ERROR_INVALID_ACE_CONDITION                                       = errors.New("the specified access control entry (ace) contains an invalid condition")
	ERROR_IMAGE_SUBSYSTEM_NOT_PRESENT                                 = errors.New("the subsystem needed to support the image type is not present")
	ERROR_NOTIFICATION_GUID_ALREADY_DEFINED                           = errors.New("the specified file already has a notification guid associated with it")
	ERROR_NETWORK_OPEN_RESTRICTION                                    = errors.New("a remote open failed because the network open restrictions were not satisfied")
	ERROR_NO_USER_SESSION_KEY                                         = errors.New("there is no user session key for the specified logon session")
	ERROR_USER_SESSION_DELETED                                        = errors.New("the remote user session has been deleted")
	ERROR_RESOURCE_LANG_NOT_FOUND                                     = errors.New("indicates the specified resource language id cannot be found in the image file")
	ERROR_INSUFF_SERVER_RESOURCES                                     = errors.New("insufficient server resources exist to complete the request")
	ERROR_INVALID_BUFFER_SIZE                                         = errors.New("the size of the buffer is invalid for the specified operation")
	ERROR_INVALID_ADDRESS_COMPONENT                                   = errors.New("the transport rejected the specified network address as invalid")
	ERROR_INVALID_ADDRESS_WILDCARD                                    = errors.New("the transport rejected the specified network address due to invalid use of a wildcard")
	ERROR_TOO_MANY_ADDRESSES                                          = errors.New("the transport address could not be opened because all the available addresses are in use")
	ERROR_ADDRESS_ALREADY_EXISTS                                      = errors.New("the transport address could not be opened because it already exists")
	ERROR_ADDRESS_CLOSED                                              = errors.New("the transport address is now closed")
	ERROR_CONNECTION_DISCONNECTED                                     = errors.New("the transport connection is now disconnected")
	ERROR_CONNECTION_RESET                                            = errors.New("the transport connection has been reset")
	ERROR_TOO_MANY_NODES                                              = errors.New("the transport cannot dynamically acquire any more nodes")
	ERROR_TRANSACTION_ABORTED                                         = errors.New("the transport aborted a pending transaction")
	ERROR_TRANSACTION_TIMED_OUT                                       = errors.New("the transport timed out a request that is waiting for a response")
	ERROR_TRANSACTION_NO_RELEASE                                      = errors.New("the transport did not receive a release for a pending response")
	ERROR_TRANSACTION_NO_MATCH                                        = errors.New("the transport did not find a transaction that matches the specific token")
	ERROR_TRANSACTION_RESPONDED                                       = errors.New("the transport had previously responded to a transaction request")
	ERROR_TRANSACTION_INVALID_ID                                      = errors.New("the transport does not recognize the specified transaction request id")
	ERROR_TRANSACTION_INVALID_TYPE                                    = errors.New("the transport does not recognize the specified transaction request type")
	ERROR_NOT_SERVER_SESSION                                          = errors.New("the transport can only process the specified request on the server side of a session")
	ERROR_NOT_CLIENT_SESSION                                          = errors.New("the transport can only process the specified request on the client side of a session")
	ERROR_CANNOT_LOAD_REGISTRY_FILE                                   = errors.New("{registry file failure} the registry cannot load the hive (file): %hs or its log or alternate. it is corrupt, absent, or not writable")
	ERROR_DEBUG_ATTACH_FAILED                                         = errors.New("{unexpected failure in debugactiveprocess} an unexpected failure occurred while processing a debugactiveprocess api request. choosing ok will terminate the process, and choosing cancel will ignore the error")
	ERROR_SYSTEM_PROCESS_TERMINATED                                   = errors.New("{fatal system error} the %hs system process terminated unexpectedly with a status of 0x%08x (0x%08x 0x%08x). the system has been shut down")
	ERROR_DATA_NOT_ACCEPTED                                           = errors.New("{data not accepted} the tdi client could not handle the data received during an indication")
	ERROR_NO_BROWSER_SERVERS_FOUND                                    = errors.New("{unable to retrieve browser server list} the list of servers for this workgroup is not currently available")
	ERROR_VDM_HARD_ERROR                                              = errors.New("ntvdm encountered a hard error")
	ERROR_DRIVER_CANCEL_TIMEOUT                                       = errors.New("{cancel timeout} the driver %hs failed to complete a canceled i/o request in the allotted time")
	ERROR_REPLY_MESSAGE_MISMATCH                                      = errors.New("{reply message mismatch} an attempt was made to reply to an lpc message, but the thread specified by the client id in the message was not waiting on that message")
	ERROR_MAPPED_ALIGNMENT                                            = errors.New("{mapped view alignment incorrect} an attempt was made to map a view of a file, but either the specified base address or the offset into the file were not aligned on the proper allocation granularity")
	ERROR_IMAGE_CHECKSUM_MISMATCH                                     = errors.New("{bad image checksum} the image %hs is possibly corrupt. the header checksum does not match the computed checksum")
	ERROR_LOST_WRITEBEHIND_DATA                                       = errors.New("{delayed write failed} windows was unable to save all the data for the file %hs. the data has been lost. this error might be caused by a failure of your computer hardware or network connection. try to save this file elsewhere")
	ERROR_CLIENT_SERVER_PARAMETERS_INVALID                            = errors.New("the parameters passed to the server in the client/server shared memory window were invalid. too much data might have been put in the shared memory window")
	ERROR_PASSWORD_MUST_CHANGE                                        = errors.New("the user password must be changed before logging on the first time")
	ERROR_NOT_FOUND                                                   = errors.New("the object was not found")
	ERROR_NOT_TINY_STREAM                                             = errors.New("the stream is not a tiny stream")
	ERROR_RECOVERY_FAILURE                                            = errors.New("a transaction recovery failed")
	ERROR_STACK_OVERFLOW_READ                                         = errors.New("the request must be handled by the stack overflow code")
	ERROR_FAIL_CHECK                                                  = errors.New("a consistency check failed")
	ERROR_DUPLICATE_OBJECTID                                          = errors.New("the attempt to insert the id in the index failed because the id is already in the index")
	ERROR_OBJECTID_EXISTS                                             = errors.New("the attempt to set the object id failed because the object already has an id")
	ERROR_CONVERT_TO_LARGE                                            = errors.New("internal ofs status codes indicating how an allocation operation is handled. either it is retried after the containing onode is moved or the extent stream is converted to a large stream")
	ERROR_RETRY                                                       = errors.New("the request needs to be retried")
	ERROR_FOUND_OUT_OF_SCOPE                                          = errors.New("the attempt to find the object found an object on the volume that matches by id; however, it is out of the scope of the handle that is used for the operation")
	ERROR_ALLOCATE_BUCKET                                             = errors.New("the bucket array must be grown. retry the transaction after doing so")
	ERROR_PROPSET_NOT_FOUND                                           = errors.New("the specified property set does not exist on the object")
	ERROR_MARSHALL_OVERFLOW                                           = errors.New("the user/kernel marshaling buffer has overflowed")
	ERROR_INVALID_VARIANT                                             = errors.New("the supplied variant structure contains invalid data")
	ERROR_DOMAIN_CONTROLLER_NOT_FOUND                                 = errors.New("a domain controller for this domain was not found")
	ERROR_ACCOUNT_LOCKED_OUT                                          = errors.New("the user account has been automatically locked because too many invalid logon attempts or password change attempts have been requested")
	ERROR_HANDLE_NOT_CLOSABLE                                         = errors.New("ntclose was called on a handle that was protected from close via ntsetinformationobject")
	ERROR_CONNECTION_REFUSED                                          = errors.New("the transport-connection attempt was refused by the remote system")
	ERROR_GRACEFUL_DISCONNECT                                         = errors.New("the transport connection was gracefully closed")
	ERROR_ADDRESS_ALREADY_ASSOCIATED                                  = errors.New("the transport endpoint already has an address associated with it")
	ERROR_ADDRESS_NOT_ASSOCIATED                                      = errors.New("an address has not yet been associated with the transport endpoint")
	ERROR_CONNECTION_INVALID                                          = errors.New("an operation was attempted on a nonexistent transport connection")
	ERROR_CONNECTION_ACTIVE                                           = errors.New("an invalid operation was attempted on an active transport connection")
	ERROR_NETWORK_UNREACHABLE                                         = errors.New("the remote network is not reachable by the transport")
	ERROR_HOST_UNREACHABLE                                            = errors.New("the remote system is not reachable by the transport")
	ERROR_PROTOCOL_UNREACHABLE                                        = errors.New("the remote system does not support the transport protocol")
	ERROR_PORT_UNREACHABLE                                            = errors.New("no service is operating at the destination port of the transport on the remote system")
	ERROR_REQUEST_ABORTED                                             = errors.New("the request was aborted")
	ERROR_CONNECTION_ABORTED                                          = errors.New("the transport connection was aborted by the local system")
	ERROR_BAD_COMPRESSION_BUFFER                                      = errors.New("the specified buffer contains ill-formed data")
	ERROR_USER_MAPPED_FILE                                            = errors.New("the requested operation cannot be performed on a file with a user mapped section open")
	ERROR_AUDIT_FAILED                                                = errors.New("{audit failed} an attempt to generate a security audit failed")
	ERROR_TIMER_RESOLUTION_NOT_SET                                    = errors.New("the timer resolution was not previously set by the current process")
	ERROR_CONNECTION_COUNT_LIMIT                                      = errors.New("a connection to the server could not be made because the limit on the number of concurrent connections for this account has been reached")
	ERROR_LOGIN_TIME_RESTRICTION                                      = errors.New("attempting to log on during an unauthorized time of day for this account")
	ERROR_LOGIN_WKSTA_RESTRICTION                                     = errors.New("the account is not authorized to log on from this station")
	ERROR_IMAGE_MP_UP_MISMATCH                                        = errors.New("{up/mp image mismatch} the image %hs has been modified for use on a uniprocessor system, but you are running it on a multiprocessor machine. reinstall the image file")
	ERROR_INSUFFICIENT_LOGON_INFO                                     = errors.New("there is insufficient account information to log you on")
	ERROR_BAD_DLL_ENTRYPOINT                                          = errors.New("{invalid dll entrypoint} the dynamic link library %hs is not written correctly. the stack pointer has been left in an inconsistent state. the entry point should be declared as winapi or stdcall. select yes to fail the dll load. select no to continue execution. selecting no might cause the application to operate incorrectly")
	ERROR_BAD_SERVICE_ENTRYPOINT                                      = errors.New("{invalid service callback entrypoint} the %hs service is not written correctly. the stack pointer has been left in an inconsistent state. the callback entry point should be declared as winapi or stdcall. selecting ok will cause the service to continue operation. however, the service process might operate incorrectly")
	ERROR_LPC_REPLY_LOST                                              = errors.New("the server received the messages but did not send a reply")
	ERROR_IP_ADDRESS_CONFLICT1                                        = errors.New("there is an ip address conflict with another system on the network")
	ERROR_IP_ADDRESS_CONFLICT2                                        = errors.New("there is an ip address conflict with another system on the network")
	ERROR_REGISTRY_QUOTA_LIMIT                                        = errors.New("{low on registry space} the system has reached the maximum size that is allowed for the system part of the registry. additional storage requests will be ignored")
	ERROR_PATH_NOT_COVERED                                            = errors.New("the contacted server does not support the indicated part of the dfs namespace")
	ERROR_NO_CALLBACK_ACTIVE                                          = errors.New("a callback return system service cannot be executed when no callback is active")
	ERROR_LICENSE_QUOTA_EXCEEDED                                      = errors.New("the service being accessed is licensed for a particular number of connections. no more connections can be made to the service at this time because the service has already accepted the maximum number of connections")
	ERROR_PWD_TOO_SHORT                                               = errors.New("the password provided is too short to meet the policy of your user account. choose a longer password")
	ERROR_PWD_TOO_RECENT                                              = errors.New("the policy of your user account does not allow you to change passwords too frequently. this is done to prevent users from changing back to a familiar, but potentially discovered, password. if you feel your password has been compromised, contact your administrator immediately to have a new one assigned")
	ERROR_PWD_HISTORY_CONFLICT                                        = errors.New("you have attempted to change your password to one that you have used in the past. the policy of your user account does not allow this. select a password that you have not previously used")
	ERROR_PLUGPLAY_NO_DEVICE                                          = errors.New("you have attempted to load a legacy device driver while its device instance had been disabled")
	ERROR_UNSUPPORTED_COMPRESSION                                     = errors.New("the specified compression format is unsupported")
	ERROR_INVALID_HW_PROFILE                                          = errors.New("the specified hardware profile configuration is invalid")
	ERROR_INVALID_PLUGPLAY_DEVICE_PATH                                = errors.New("the specified plug and play registry device path is invalid")
	ERROR_DRIVER_ORDINAL_NOT_FOUND                                    = errors.New("{driver entry point not found} the %hs device driver could not locate the ordinal %ld in driver %hs")
	ERROR_DRIVER_ENTRYPOINT_NOT_FOUND                                 = errors.New("{driver entry point not found} the %hs device driver could not locate the entry point %hs in driver %hs")
	ERROR_RESOURCE_NOT_OWNED                                          = errors.New("{application error} the application attempted to release a resource it did not own. click ok to terminate the application")
	ERROR_TOO_MANY_LINKS                                              = errors.New("an attempt was made to create more links on a file than the file system supports")
	ERROR_QUOTA_LIST_INCONSISTENT                                     = errors.New("the specified quota list is internally inconsistent with its descriptor")
	ERROR_FILE_IS_OFFLINE                                             = errors.New("the specified file has been relocated to offline storage")
	ERROR_EVALUATION_EXPIRATION                                       = errors.New("{windows evaluation notification} the evaluation period for this installation of windows has expired. this system will shutdown in 1 hour. to restore access to this installation of windows, upgrade this installation by using a licensed distribution of this product")
	ERROR_ILLEGAL_DLL_RELOCATION                                      = errors.New("{illegal system dll relocation} the system dll %hs was relocated in memory. the application will not run properly. the relocation occurred because the dll %hs occupied an address range that is reserved for windows system dlls. the vendor supplying the dll should be contacted for a new dll")
	ERROR_LICENSE_VIOLATION                                           = errors.New("{license violation} the system has detected tampering with your registered product type. this is a violation of your software license. tampering with the product type is not permitted")
	ERROR_DLL_INIT_FAILED_LOGOFF                                      = errors.New("{dll initialization failed} the application failed to initialize because the window station is shutting down")
	ERROR_DRIVER_UNABLE_TO_LOAD                                       = errors.New("{unable to load device driver} %hs device driver could not be loaded. error status was 0x%x")
	ERROR_DFS_UNAVAILABLE                                             = errors.New("dfs is unavailable on the contacted server")
	ERROR_VOLUME_DISMOUNTED                                           = errors.New("an operation was attempted to a volume after it was dismounted")
	ERROR_WX86_INTERNAL_ERROR                                         = errors.New("an internal error occurred in the win32 x86 emulation subsystem")
	ERROR_WX86_FLOAT_STACK_CHECK                                      = errors.New("win32 x86 emulation subsystem floating-point stack check")
	ERROR_VALIDATE_CONTINUE                                           = errors.New("the validation process needs to continue on to the next step")
	ERROR_NO_MATCH                                                    = errors.New("there was no match for the specified key in the index")
	ERROR_NO_MORE_MATCHES                                             = errors.New("there are no more matches for the current index enumeration")
	ERROR_NOT_A_REPARSE_POINT                                         = errors.New("the ntfs file or directory is not a reparse point")
	ERROR_IO_REPARSE_TAG_INVALID                                      = errors.New("the windows i/o reparse tag passed for the ntfs reparse point is invalid")
	ERROR_IO_REPARSE_TAG_MISMATCH                                     = errors.New("the windows i/o reparse tag does not match the one that is in the ntfs reparse point")
	ERROR_IO_REPARSE_DATA_INVALID                                     = errors.New("the user data passed for the ntfs reparse point is invalid")
	ERROR_IO_REPARSE_TAG_NOT_HANDLED                                  = errors.New("the layered file system driver for this i/o tag did not handle it when needed")
	ERROR_REPARSE_POINT_NOT_RESOLVED                                  = errors.New("the ntfs symbolic link could not be resolved even though the initial file name is valid")
	ERROR_DIRECTORY_IS_A_REPARSE_POINT                                = errors.New("the ntfs directory is a reparse point")
	ERROR_RANGE_LIST_CONFLICT                                         = errors.New("the range could not be added to the range list because of a conflict")
	ERROR_SOURCE_ELEMENT_EMPTY                                        = errors.New("the specified medium changer source element contains no media")
	ERROR_DESTINATION_ELEMENT_FULL                                    = errors.New("the specified medium changer destination element already contains media")
	ERROR_ILLEGAL_ELEMENT_ADDRESS                                     = errors.New("the specified medium changer element does not exist")
	ERROR_MAGAZINE_NOT_PRESENT                                        = errors.New("the specified element is contained in a magazine that is no longer present")
	ERROR_REINITIALIZATION_NEEDED                                     = errors.New("the device requires re-initialization due to hardware errors")
	ERROR_ENCRYPTION_FAILED                                           = errors.New("the file encryption attempt failed")
	ERROR_DECRYPTION_FAILED                                           = errors.New("the file decryption attempt failed")
	ERROR_RANGE_NOT_FOUND                                             = errors.New("the specified range could not be found in the range list")
	ERROR_NO_RECOVERY_POLICY                                          = errors.New("there is no encryption recovery policy configured for this system")
	ERROR_NO_EFS                                                      = errors.New("the required encryption driver is not loaded for this system")
	ERROR_WRONG_EFS                                                   = errors.New("the file was encrypted with a different encryption driver than is currently loaded")
	ERROR_NO_USER_KEYS                                                = errors.New("there are no efs keys defined for the user")
	ERROR_FILE_NOT_ENCRYPTED                                          = errors.New("the specified file is not encrypted")
	ERROR_NOT_EXPORT_FORMAT                                           = errors.New("the specified file is not in the defined efs export format")
	ERROR_FILE_ENCRYPTED                                              = errors.New("the specified file is encrypted and the user does not have the ability to decrypt it")
	ERROR_WMI_GUID_NOT_FOUND                                          = errors.New("the guid passed was not recognized as valid by a wmi data provider")
	ERROR_WMI_INSTANCE_NOT_FOUND                                      = errors.New("the instance name passed was not recognized as valid by a wmi data provider")
	ERROR_WMI_ITEMID_NOT_FOUND                                        = errors.New("the data item id passed was not recognized as valid by a wmi data provider")
	ERROR_WMI_TRY_AGAIN                                               = errors.New("the wmi request could not be completed and should be retried")
	ERROR_SHARED_POLICY                                               = errors.New("the policy object is shared and can only be modified at the root")
	ERROR_POLICY_OBJECT_NOT_FOUND                                     = errors.New("the policy object does not exist when it should")
	ERROR_POLICY_ONLY_IN_DS                                           = errors.New("the requested policy information only lives in the ds")
	ERROR_VOLUME_NOT_UPGRADED                                         = errors.New("the volume must be upgraded to enable this feature")
	ERROR_REMOTE_STORAGE_NOT_ACTIVE                                   = errors.New("the remote storage service is not operational at this time")
	ERROR_REMOTE_STORAGE_MEDIA_ERROR                                  = errors.New("the remote storage service encountered a media error")
	ERROR_NO_TRACKING_SERVICE                                         = errors.New("the tracking (workstation) service is not running")
	ERROR_SERVER_SID_MISMATCH                                         = errors.New("the server process is running under a sid that is different from the sid that is required by client")
	ERROR_DS_NO_ATTRIBUTE_OR_VALUE                                    = errors.New("the specified directory service attribute or value does not exist")
	ERROR_DS_INVALID_ATTRIBUTE_SYNTAX                                 = errors.New("the attribute syntax specified to the directory service is invalid")
	ERROR_DS_ATTRIBUTE_TYPE_UNDEFINED                                 = errors.New("the attribute type specified to the directory service is not defined")
	ERROR_DS_ATTRIBUTE_OR_VALUE_EXISTS                                = errors.New("the specified directory service attribute or value already exists")
	ERROR_DS_BUSY                                                     = errors.New("the directory service is busy")
	ERROR_DS_UNAVAILABLE                                              = errors.New("the directory service is unavailable")
	ERROR_DS_NO_RIDS_ALLOCATED                                        = errors.New("the directory service was unable to allocate a relative identifier")
	ERROR_DS_NO_MORE_RIDS                                             = errors.New("the directory service has exhausted the pool of relative identifiers")
	ERROR_DS_INCORRECT_ROLE_OWNER                                     = errors.New("the requested operation could not be performed because the directory service is not the master for that type of operation")
	ERROR_DS_RIDMGR_INIT_ERROR                                        = errors.New("the directory service was unable to initialize the subsystem that allocates relative identifiers")
	ERROR_DS_OBJ_CLASS_VIOLATION                                      = errors.New("the requested operation did not satisfy one or more constraints that are associated with the class of the object")
	ERROR_DS_CANT_ON_NON_LEAF                                         = errors.New("the directory service can perform the requested operation only on a leaf object")
	ERROR_DS_CANT_ON_RDN                                              = errors.New("the directory service cannot perform the requested operation on the relatively defined name (rdn) attribute of an object")
	ERROR_DS_CANT_MOD_OBJ_CLASS                                       = errors.New("the directory service detected an attempt to modify the object class of an object")
	ERROR_DS_CROSS_DOM_MOVE_FAILED                                    = errors.New("an error occurred while performing a cross domain move operation")
	ERROR_DS_GC_NOT_AVAILABLE                                         = errors.New("unable to contact the global catalog server")
	ERROR_DIRECTORY_SERVICE_REQUIRED                                  = errors.New("the requested operation requires a directory service, and none was available")
	ERROR_REPARSE_ATTRIBUTE_CONFLICT                                  = errors.New("the reparse attribute cannot be set because it is incompatible with an existing attribute")
	ERROR_CANT_ENABLE_DENY_ONLY                                       = errors.New("a group marked \"use for deny only\" cannot be enabled")
	ERROR_FLOAT_MULTIPLE_FAULTS                                       = errors.New("{exception} multiple floating-point faults")
	ERROR_FLOAT_MULTIPLE_TRAPS                                        = errors.New("{exception} multiple floating-point traps")
	ERROR_DEVICE_REMOVED                                              = errors.New("the device has been removed")
	ERROR_JOURNAL_DELETE_IN_PROGRESS                                  = errors.New("the volume change journal is being deleted")
	ERROR_JOURNAL_NOT_ACTIVE                                          = errors.New("the volume change journal is not active")
	ERROR_NOINTERFACE                                                 = errors.New("the requested interface is not supported")
	ERROR_DS_ADMIN_LIMIT_EXCEEDED                                     = errors.New("a directory service resource limit has been exceeded")
	ERROR_DRIVER_FAILED_SLEEP                                         = errors.New("{system standby failed} the driver %hs does not support standby mode. updating this driver allows the system to go to standby mode")
	ERROR_MUTUAL_AUTHENTICATION_FAILED                                = errors.New("mutual authentication failed. the server password is out of date at the domain controller")
	ERROR_CORRUPT_SYSTEM_FILE                                         = errors.New("the system file %1 has become corrupt and has been replaced")
	ERROR_DATATYPE_MISALIGNMENT_ERROR                                 = errors.New("{exception} alignment error a data type misalignment error was detected in a load or store instruction")
	ERROR_WMI_READ_ONLY                                               = errors.New("the wmi data item or data block is read-only")
	ERROR_WMI_SET_FAILURE                                             = errors.New("the wmi data item or data block could not be changed")
	ERROR_COMMITMENT_MINIMUM                                          = errors.New("{virtual memory minimum too low} your system is low on virtual memory. windows is increasing the size of your virtual memory paging file. during this process, memory requests for some applications might be denied. for more information, see help")
	ERROR_REG_NAT_CONSUMPTION                                         = errors.New("{exception} register nat consumption faults. a nat value is consumed on a non-speculative instruction")
	ERROR_TRANSPORT_FULL                                              = errors.New("the transport element of the medium changer contains media, which is causing the operation to fail")
	ERROR_DS_SAM_INIT_FAILURE                                         = errors.New("security accounts manager initialization failed because of the following error: %hs error status: 0x%x. click ok to shut down this system and restart in directory services restore mode. check the event log for more detailed information")
	ERROR_ONLY_IF_CONNECTED                                           = errors.New("this operation is supported only when you are connected to the server")
	ERROR_DS_SENSITIVE_GROUP_VIOLATION                                = errors.New("only an administrator can modify the membership list of an administrative group")
	ERROR_PNP_RESTART_ENUMERATION                                     = errors.New("a device was removed so enumeration must be restarted")
	ERROR_JOURNAL_ENTRY_DELETED                                       = errors.New("the journal entry has been deleted from the journal")
	ERROR_DS_CANT_MOD_PRIMARYGROUPID                                  = errors.New("cannot change the primary group id of a domain controller account")
	ERROR_SYSTEM_IMAGE_BAD_SIGNATURE                                  = errors.New("{fatal system error} the system image %s is not properly signed. the file has been replaced with the signed file. the system has been shut down")
	ERROR_PNP_REBOOT_REQUIRED                                         = errors.New("the device will not start without a reboot")
	ERROR_POWER_STATE_INVALID                                         = errors.New("the power state of the current device cannot support this request")
	ERROR_DS_INVALID_GROUP_TYPE                                       = errors.New("the specified group type is invalid")
	ERROR_DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN                       = errors.New("in a mixed domain, no nesting of a global group if the group is security enabled")
	ERROR_DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN                        = errors.New("in a mixed domain, cannot nest local groups with other local groups, if the group is security enabled")
	ERROR_DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER                            = errors.New("a global group cannot have a local group as a member")
	ERROR_DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER                        = errors.New("a global group cannot have a universal group as a member")
	ERROR_DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER                         = errors.New("a universal group cannot have a local group as a member")
	ERROR_DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER                      = errors.New("a global group cannot have a cross-domain member")
	ERROR_DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER                 = errors.New("a local group cannot have another cross-domain local group as a member")
	ERROR_DS_HAVE_PRIMARY_MEMBERS                                     = errors.New("cannot change to a security-disabled group because primary members are in this group")
	ERROR_WMI_NOT_SUPPORTED                                           = errors.New("the wmi operation is not supported by the data block or method")
	ERROR_INSUFFICIENT_POWER                                          = errors.New("there is not enough power to complete the requested operation")
	ERROR_SAM_NEED_BOOTKEY_PASSWORD                                   = errors.New("the security accounts manager needs to get the boot password")
	ERROR_SAM_NEED_BOOTKEY_FLOPPY                                     = errors.New("the security accounts manager needs to get the boot key from the floppy disk")
	ERROR_DS_CANT_START                                               = errors.New("the directory service cannot start")
	ERROR_DS_INIT_FAILURE                                             = errors.New("the directory service could not start because of the following error: %hs error status: 0x%x. click ok to shut down this system and restart in directory services restore mode. check the event log for more detailed information")
	ERROR_SAM_INIT_FAILURE                                            = errors.New("the security accounts manager initialization failed because of the following error: %hs error status: 0x%x. click ok to shut down this system and restart in safe mode. check the event log for more detailed information")
	ERROR_DS_GC_REQUIRED                                              = errors.New("the requested operation can be performed only on a global catalog server")
	ERROR_DS_LOCAL_MEMBER_OF_LOCAL_ONLY                               = errors.New("a local group can only be a member of other local groups in the same domain")
	ERROR_DS_NO_FPO_IN_UNIVERSAL_GROUPS                               = errors.New("foreign security principals cannot be members of universal groups")
	ERROR_DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED                           = errors.New("your computer could not be joined to the domain. you have exceeded the maximum number of computer accounts you are allowed to create in this domain. contact your system administrator to have this limit reset or increased")
	ERROR_CURRENT_DOMAIN_NOT_ALLOWED                                  = errors.New("this operation cannot be performed on the current domain")
	ERROR_CANNOT_MAKE                                                 = errors.New("the directory or file cannot be created")
	ERROR_SYSTEM_SHUTDOWN                                             = errors.New("the system is in the process of shutting down")
	ERROR_DS_INIT_FAILURE_CONSOLE                                     = errors.New("directory services could not start because of the following error: %hs error status: 0x%x. click ok to shut down the system. you can use the recovery console to diagnose the system further")
	ERROR_DS_SAM_INIT_FAILURE_CONSOLE                                 = errors.New("security accounts manager initialization failed because of the following error: %hs error status: 0x%x. click ok to shut down the system. you can use the recovery console to diagnose the system further")
	ERROR_UNFINISHED_CONTEXT_DELETED                                  = errors.New("a security context was deleted before the context was completed. this is considered a logon failure")
	ERROR_NO_TGT_REPLY                                                = errors.New("the client is trying to negotiate a context and the server requires user-to-user but did not send a tgt reply")
	ERROR_OBJECTID_NOT_FOUND                                          = errors.New("an object id was not found in the file")
	ERROR_NO_IP_ADDRESSES                                             = errors.New("unable to accomplish the requested task because the local machine does not have any ip addresses")
	ERROR_WRONG_CREDENTIAL_HANDLE                                     = errors.New("the supplied credential handle does not match the credential that is associated with the security context")
	ERROR_CRYPTO_SYSTEM_INVALID                                       = errors.New("the crypto system or checksum function is invalid because a required function is unavailable")
	ERROR_MAX_REFERRALS_EXCEEDED                                      = errors.New("the number of maximum ticket referrals has been exceeded")
	ERROR_MUST_BE_KDC                                                 = errors.New("the local machine must be a kerberos kdc (domain controller) and it is not")
	ERROR_STRONG_CRYPTO_NOT_SUPPORTED                                 = errors.New("the other end of the security negotiation requires strong crypto but it is not supported on the local machine")
	ERROR_TOO_MANY_PRINCIPALS                                         = errors.New("the kdc reply contained more than one principal name")
	ERROR_NO_PA_DATA                                                  = errors.New("expected to find pa data for a hint of what etype to use, but it was not found")
	ERROR_PKINIT_NAME_MISMATCH                                        = errors.New("the client certificate does not contain a valid upn, or does not match the client name in the logon request. contact your administrator")
	ERROR_SMARTCARD_LOGON_REQUIRED                                    = errors.New("smart card logon is required and was not used")
	ERROR_KDC_INVALID_REQUEST                                         = errors.New("an invalid request was sent to the kdc")
	ERROR_KDC_UNABLE_TO_REFER                                         = errors.New("the kdc was unable to generate a referral for the service requested")
	ERROR_KDC_UNKNOWN_ETYPE                                           = errors.New("the encryption type requested is not supported by the kdc")
	ERROR_SHUTDOWN_IN_PROGRESS                                        = errors.New("a system shutdown is in progress")
	ERROR_SERVER_SHUTDOWN_IN_PROGRESS                                 = errors.New("the server machine is shutting down")
	ERROR_NOT_SUPPORTED_ON_SBS                                        = errors.New("this operation is not supported on a computer running windows server 2003 operating system for small business server")
	ERROR_WMI_GUID_DISCONNECTED                                       = errors.New("the wmi guid is no longer available")
	ERROR_WMI_ALREADY_DISABLED                                        = errors.New("collection or events for the wmi guid is already disabled")
	ERROR_WMI_ALREADY_ENABLED                                         = errors.New("collection or events for the wmi guid is already enabled")
	ERROR_MFT_TOO_FRAGMENTED                                          = errors.New("the master file table on the volume is too fragmented to complete this operation")
	ERROR_COPY_PROTECTION_FAILURE                                     = errors.New("copy protection failure")
	ERROR_CSS_AUTHENTICATION_FAILURE                                  = errors.New("copy protection error—dvd css authentication failed")
	ERROR_CSS_KEY_NOT_PRESENT                                         = errors.New("copy protection error—the specified sector does not contain a valid key")
	ERROR_CSS_KEY_NOT_ESTABLISHED                                     = errors.New("copy protection error—dvd session key not established")
	ERROR_CSS_SCRAMBLED_SECTOR                                        = errors.New("copy protection error—the read failed because the sector is encrypted")
	ERROR_CSS_REGION_MISMATCH                                         = errors.New("copy protection error—the region of the specified dvd does not correspond to the region setting of the drive")
	ERROR_CSS_RESETS_EXHAUSTED                                        = errors.New("copy protection error—the region setting of the drive might be permanent")
	ERROR_PKINIT_FAILURE                                              = errors.New("the kerberos protocol encountered an error while validating the kdc certificate during smart card logon. there is more information in the system event log")
	ERROR_SMARTCARD_SUBSYSTEM_FAILURE                                 = errors.New("the kerberos protocol encountered an error while attempting to use the smart card subsystem")
	ERROR_NO_KERB_KEY                                                 = errors.New("the target server does not have acceptable kerberos credentials")
	ERROR_HOST_DOWN                                                   = errors.New("the transport determined that the remote system is down")
	ERROR_UNSUPPORTED_PREAUTH                                         = errors.New("an unsupported pre-authentication mechanism was presented to the kerberos package")
	ERROR_EFS_ALG_BLOB_TOO_BIG                                        = errors.New("the encryption algorithm that is used on the source file needs a bigger key buffer than the one that is used on the destination file")
	ERROR_PORT_NOT_SET                                                = errors.New("an attempt to remove a processes debugport was made, but a port was not already associated with the process")
	ERROR_DEBUGGER_INACTIVE                                           = errors.New("an attempt to do an operation on a debug port failed because the port is in the process of being deleted")
	ERROR_DS_VERSION_CHECK_FAILURE                                    = errors.New("this version of windows is not compatible with the behavior version of the directory forest, domain, or domain controller")
	ERROR_AUDITING_DISABLED                                           = errors.New("the specified event is currently not being audited")
	ERROR_PRENT4_MACHINE_ACCOUNT                                      = errors.New("the machine account was created prior to windows nt 4.0 operating system. the account needs to be recreated")
	ERROR_DS_AG_CANT_HAVE_UNIVERSAL_MEMBER                            = errors.New("an account group cannot have a universal group as a member")
	ERROR_INVALID_IMAGE_WIN_32                                        = errors.New("the specified image file did not have the correct format; it appears to be a 32-bit windows image")
	ERROR_INVALID_IMAGE_WIN_64                                        = errors.New("the specified image file did not have the correct format; it appears to be a 64-bit windows image")
	ERROR_BAD_BINDINGS                                                = errors.New("the client's supplied sspi channel bindings were incorrect")
	ERROR_NETWORK_SESSION_EXPIRED                                     = errors.New("the client session has expired; so the client must re-authenticate to continue accessing the remote resources")
	ERROR_APPHELP_BLOCK                                               = errors.New("the apphelp dialog box canceled; thus preventing the application from starting")
	ERROR_ALL_SIDS_FILTERED                                           = errors.New("the sid filtering operation removed all sids")
	ERROR_NOT_SAFE_MODE_DRIVER                                        = errors.New("the driver was not loaded because the system is starting in safe mode")
	ERROR_ACCESS_DISABLED_BY_POLICY_DEFAULT                           = errors.New("access to %1 has been restricted by your administrator by the default software restriction policy level")
	ERROR_ACCESS_DISABLED_BY_POLICY_PATH                              = errors.New("access to %1 has been restricted by your administrator by location with policy rule %2 placed on path %3")
	ERROR_ACCESS_DISABLED_BY_POLICY_PUBLISHER                         = errors.New("access to %1 has been restricted by your administrator by software publisher policy")
	ERROR_ACCESS_DISABLED_BY_POLICY_OTHER                             = errors.New("access to %1 has been restricted by your administrator by policy rule %2")
	ERROR_FAILED_DRIVER_ENTRY                                         = errors.New("the driver was not loaded because it failed its initialization call")
	ERROR_DEVICE_ENUMERATION_ERROR                                    = errors.New("the device encountered an error while applying power or reading the device configuration. this might be caused by a failure of your hardware or by a poor connection")
	ERROR_MOUNT_POINT_NOT_RESOLVED                                    = errors.New("the create operation failed because the name contained at least one mount point that resolves to a volume to which the specified device object is not attached")
	ERROR_INVALID_DEVICE_OBJECT_PARAMETER                             = errors.New("the device object parameter is either not a valid device object or is not attached to the volume that is specified by the file name")
	ERROR_MCA_OCCURED                                                 = errors.New("a machine check error has occurred. check the system event log for additional information")
	ERROR_DRIVER_BLOCKED_CRITICAL                                     = errors.New("driver %2 has been blocked from loading")
	ERROR_DRIVER_BLOCKED                                              = errors.New("driver %2 has been blocked from loading")
	ERROR_DRIVER_DATABASE_ERROR                                       = errors.New("there was error [%2] processing the driver database")
	ERROR_SYSTEM_HIVE_TOO_LARGE                                       = errors.New("system hive size has exceeded its limit")
	ERROR_INVALID_IMPORT_OF_NON_DLL                                   = errors.New("a dynamic link library (dll) referenced a module that was neither a dll nor the process's executable image")
	ERROR_NO_SECRETS                                                  = errors.New("the local account store does not contain secret material for the specified account")
	ERROR_ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY                       = errors.New("access to %1 has been restricted by your administrator by policy rule %2")
	ERROR_FAILED_STACK_SWITCH                                         = errors.New("the system was not able to allocate enough memory to perform a stack switch")
	ERROR_HEAP_CORRUPTION                                             = errors.New("a heap has been corrupted")
	ERROR_SMARTCARD_WRONG_PIN                                         = errors.New("an incorrect pin was presented to the smart card")
	ERROR_SMARTCARD_CARD_BLOCKED                                      = errors.New("the smart card is blocked")
	ERROR_SMARTCARD_CARD_NOT_AUTHENTICATED                            = errors.New("no pin was presented to the smart card")
	ERROR_SMARTCARD_NO_CARD                                           = errors.New("no smart card is available")
	ERROR_SMARTCARD_NO_KEY_CONTAINER                                  = errors.New("the requested key container does not exist on the smart card")
	ERROR_SMARTCARD_NO_CERTIFICATE                                    = errors.New("the requested certificate does not exist on the smart card")
	ERROR_SMARTCARD_NO_KEYSET                                         = errors.New("the requested keyset does not exist")
	ERROR_SMARTCARD_IO_ERROR                                          = errors.New("a communication error with the smart card has been detected")
	ERROR_DOWNGRADE_DETECTED                                          = errors.New("the system detected a possible attempt to compromise security. ensure that you can contact the server that authenticated you")
	ERROR_SMARTCARD_CERT_REVOKED                                      = errors.New("the smart card certificate used for authentication has been revoked. contact your system administrator. there might be additional information in the event log")
	ERROR_ISSUING_CA_UNTRUSTED                                        = errors.New("an untrusted certificate authority was detected while processing the smart card certificate that is used for authentication. contact your system administrator")
	ERROR_REVOCATION_OFFLINE_C                                        = errors.New("the revocation status of the smart card certificate that is used for authentication could not be determined. contact your system administrator")
	ERROR_PKINIT_CLIENT_FAILURE                                       = errors.New("the smart card certificate used for authentication was not trusted. contact your system administrator")
	ERROR_SMARTCARD_CERT_EXPIRED                                      = errors.New("the smart card certificate used for authentication has expired. contact your system administrator")
	ERROR_DRIVER_FAILED_PRIOR_UNLOAD                                  = errors.New("the driver could not be loaded because a previous version of the driver is still in memory")
	ERROR_SMARTCARD_SILENT_CONTEXT                                    = errors.New("the smart card provider could not perform the action because the context was acquired as silent")
	ERROR_PER_USER_TRUST_QUOTA_EXCEEDED                               = errors.New("the delegated trust creation quota of the current user has been exceeded")
	ERROR_ALL_USER_TRUST_QUOTA_EXCEEDED                               = errors.New("the total delegated trust creation quota has been exceeded")
	ERROR_USER_DELETE_TRUST_QUOTA_EXCEEDED                            = errors.New("the delegated trust deletion quota of the current user has been exceeded")
	ERROR_DS_NAME_NOT_UNIQUE                                          = errors.New("the requested name already exists as a unique identifier")
	ERROR_DS_DUPLICATE_ID_FOUND                                       = errors.New("the requested object has a non-unique identifier and cannot be retrieved")
	ERROR_DS_GROUP_CONVERSION_ERROR                                   = errors.New("the group cannot be converted due to attribute restrictions on the requested group type")
	ERROR_VOLSNAP_PREPARE_HIBERNATE                                   = errors.New("{volume shadow copy service} wait while the volume shadow copy service prepares volume %hs for hibernation")
	ERROR_USER2USER_REQUIRED                                          = errors.New("kerberos sub-protocol user2user is required")
	ERROR_STACK_BUFFER_OVERRUN                                        = errors.New("the system detected an overrun of a stack-based buffer in this application. this overrun could potentially allow a malicious user to gain control of this application")
	ERROR_NO_S4U_PROT_SUPPORT                                         = errors.New("the kerberos subsystem encountered an error. a service for user protocol request was made against a domain controller which does not support service for user")
	ERROR_CROSSREALM_DELEGATION_FAILURE                               = errors.New("an attempt was made by this server to make a kerberos constrained delegation request for a target that is outside the server realm. this action is not supported and the resulting error indicates a misconfiguration on the allowed-to-delegate-to list for this server. contact your administrator")
	ERROR_REVOCATION_OFFLINE_KDC                                      = errors.New("the revocation status of the domain controller certificate used for smart card authentication could not be determined. there is additional information in the system event log. contact your system administrator")
	ERROR_ISSUING_CA_UNTRUSTED_KDC                                    = errors.New("an untrusted certificate authority was detected while processing the domain controller certificate used for authentication. there is additional information in the system event log. contact your system administrator")
	ERROR_KDC_CERT_EXPIRED                                            = errors.New("the domain controller certificate used for smart card logon has expired. contact your system administrator with the contents of your system event log")
	ERROR_KDC_CERT_REVOKED                                            = errors.New("the domain controller certificate used for smart card logon has been revoked. contact your system administrator with the contents of your system event log")
	ERROR_PARAMETER_QUOTA_EXCEEDED                                    = errors.New("data present in one of the parameters is more than the function can operate on")
	ERROR_HIBERNATION_FAILURE                                         = errors.New("the system has failed to hibernate (the error code is %hs). hibernation will be disabled until the system is restarted")
	ERROR_DELAY_LOAD_FAILED                                           = errors.New("an attempt to delay-load a .dll or get a function address in a delay-loaded .dll failed")
	ERROR_AUTHENTICATION_FIREWALL_FAILED                              = errors.New("logon failure: the machine you are logging onto is protected by an authentication firewall. the specified account is not allowed to authenticate to the machine")
	ERROR_VDM_DISALLOWED                                              = errors.New("%hs is a 16-bit application. you do not have permissions to execute 16-bit applications. check your permissions with your system administrator")
	ERROR_HUNG_DISPLAY_DRIVER_THREAD                                  = errors.New("{display driver stopped responding} the %hs display driver has stopped working normally. save your work and reboot the system to restore full display functionality. the next time you reboot the machine a dialog will be displayed giving you a chance to report this failure to microsoft")
	ERROR_INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE     = errors.New("the desktop heap encountered an error while allocating session memory. there is more information in the system event log")
	ERROR_INVALID_CRUNTIME_PARAMETER                                  = errors.New("an invalid parameter was passed to a c runtime function")
	ERROR_NTLM_BLOCKED                                                = errors.New("the authentication failed because ntlm was blocked")
	ERROR_DS_SRC_SID_EXISTS_IN_FOREST                                 = errors.New("the source object's sid already exists in destination forest")
	ERROR_DS_DOMAIN_NAME_EXISTS_IN_FOREST                             = errors.New("the domain name of the trusted domain already exists in the forest")
	ERROR_DS_FLAT_NAME_EXISTS_IN_FOREST                               = errors.New("the flat name of the trusted domain already exists in the forest")
	ERROR_INVALID_USER_PRINCIPAL_NAME                                 = errors.New("the user principal name (upn) is invalid")
	ERROR_ASSERTION_FAILURE                                           = errors.New("there has been an assertion failure")
	ERROR_VERIFIER_STOP                                               = errors.New("application verifier has found an error in the current process")
	ERROR_CALLBACK_POP_STACK                                          = errors.New("a user mode unwind is in progress")
	ERROR_INCOMPATIBLE_DRIVER_BLOCKED                                 = errors.New("%2 has been blocked from loading due to incompatibility with this system. contact your software vendor for a compatible version of the driver")
	ERROR_HIVE_UNLOADED                                               = errors.New("illegal operation attempted on a registry key which has already been unloaded")
	ERROR_COMPRESSION_DISABLED                                        = errors.New("compression is disabled for this volume")
	ERROR_FILE_SYSTEM_LIMITATION                                      = errors.New("the requested operation could not be completed due to a file system limitation")
	ERROR_INVALID_IMAGE_HASH                                          = errors.New("the hash for image %hs cannot be found in the system catalogs. the image is likely corrupt or the victim of tampering")
	ERROR_NOT_CAPABLE                                                 = errors.New("the implementation is not capable of performing the request")
	ERROR_REQUEST_OUT_OF_SEQUENCE                                     = errors.New("the requested operation is out of order with respect to other operations")
	ERROR_IMPLEMENTATION_LIMIT                                        = errors.New("an operation attempted to exceed an implementation-defined limit")
	ERROR_ELEVATION_REQUIRED                                          = errors.New("the requested operation requires elevation")
	ERROR_NO_SECURITY_CONTEXT                                         = errors.New("the required security context does not exist")
	ERROR_PKU2U_CERT_FAILURE                                          = errors.New("the pku2u protocol encountered an error while attempting to utilize the associated certificates")
	ERROR_BEYOND_VDL                                                  = errors.New("the operation was attempted beyond the valid data length of the file")
	ERROR_ENCOUNTERED_WRITE_IN_PROGRESS                               = errors.New("the attempted write operation encountered a write already in progress for some portion of the range")
	ERROR_PTE_CHANGED                                                 = errors.New("the page fault mappings changed in the middle of processing a fault so the operation must be retried")
	ERROR_PURGE_FAILED                                                = errors.New("the attempt to purge this file from memory failed to purge some or all the data from memory")
	ERROR_CRED_REQUIRES_CONFIRMATION                                  = errors.New("the requested credential requires confirmation")
	ERROR_CS_ENCRYPTION_INVALID_SERVER_RESPONSE                       = errors.New("the remote server sent an invalid response for a file being opened with client side encryption")
	ERROR_CS_ENCRYPTION_UNSUPPORTED_SERVER                            = errors.New("client side encryption is not supported by the remote server even though it claims to support it")
	ERROR_CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE                       = errors.New("file is encrypted and should be opened in client side encryption mode")
	ERROR_CS_ENCRYPTION_NEW_ENCRYPTED_FILE                            = errors.New("a new encrypted file is being created and a $efs needs to be provided")
	ERROR_CS_ENCRYPTION_FILE_NOT_CSE                                  = errors.New("the smb client requested a cse fsctl on a non-cse file")
	ERROR_INVALID_LABEL                                               = errors.New("indicates a particular security id cannot be assigned as the label of an object")
	ERROR_DRIVER_PROCESS_TERMINATED                                   = errors.New("the process hosting the driver for this device has terminated")
	ERROR_AMBIGUOUS_SYSTEM_DEVICE                                     = errors.New("the requested system device cannot be identified due to multiple indistinguishable devices potentially matching the identification criteria")
	ERROR_SYSTEM_DEVICE_NOT_FOUND                                     = errors.New("the requested system device cannot be found")
	ERROR_RESTART_BOOT_APPLICATION                                    = errors.New("this boot application must be restarted")
	ERROR_INSUFFICIENT_NVRAM_RESOURCES                                = errors.New("insufficient nvram resources exist to complete the api. a reboot might be required")
	ERROR_NO_RANGES_PROCESSED                                         = errors.New("no ranges for the specified operation were able to be processed")
	ERROR_DEVICE_FEATURE_NOT_SUPPORTED                                = errors.New("the storage device does not support offload write")
	ERROR_DEVICE_UNREACHABLE                                          = errors.New("data cannot be moved because the source device cannot communicate with the destination device")
	ERROR_INVALID_TOKEN                                               = errors.New("the token representing the data is invalid or expired")
	ERROR_SERVER_UNAVAILABLE                                          = errors.New("the file server is temporarily unavailable")
	ERROR_INVALID_TASK_NAME                                           = errors.New("the specified task name is invalid")
	ERROR_INVALID_TASK_INDEX                                          = errors.New("the specified task index is invalid")
	ERROR_THREAD_ALREADY_IN_TASK                                      = errors.New("the specified thread is already joining a task")
	ERROR_CALLBACK_BYPASS                                             = errors.New("a callback has requested to bypass native code")
	ERROR_FAIL_FAST_EXCEPTION                                         = errors.New("a fail fast exception occurred. exception handlers will not be invoked and the process will be terminated immediately")
	ERROR_IMAGE_CERT_REVOKED                                          = errors.New("windows cannot verify the digital signature for this file. the signing certificate for this file has been revoked")
	ERROR_PORT_CLOSED                                                 = errors.New("the alpc port is closed")
	ERROR_MESSAGE_LOST                                                = errors.New("the alpc message requested is no longer available")
	ERROR_INVALID_MESSAGE                                             = errors.New("the alpc message supplied is invalid")
	ERROR_REQUEST_CANCELED                                            = errors.New("the alpc message has been canceled")
	ERROR_RECURSIVE_DISPATCH                                          = errors.New("invalid recursive dispatch attempt")
	ERROR_LPC_RECEIVE_BUFFER_EXPECTED                                 = errors.New("no receive buffer has been supplied in a synchronous request")
	ERROR_LPC_INVALID_CONNECTION_USAGE                                = errors.New("the connection port is used in an invalid context")
	ERROR_LPC_REQUESTS_NOT_ALLOWED                                    = errors.New("the alpc port does not accept new request messages")
	ERROR_RESOURCE_IN_USE                                             = errors.New("the resource requested is already in use")
	ERROR_HARDWARE_MEMORY_ERROR                                       = errors.New("the hardware has reported an uncorrectable memory error")
	ERROR_THREADPOOL_HANDLE_EXCEPTION                                 = errors.New("status 0x%08x was returned, waiting on handle 0x%x for wait 0x%p, in waiter 0x%p")
	ERROR_THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED                   = errors.New("after a callback to 0x%p(0x%p), a completion call to set event(0x%p) failed with status 0x%08x")
	ERROR_THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED           = errors.New("after a callback to 0x%p(0x%p), a completion call to releasesemaphore(0x%p, %d) failed with status 0x%08x")
	ERROR_THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED               = errors.New("after a callback to 0x%p(0x%p), a completion call to releasemutex(%p) failed with status 0x%08x")
	ERROR_THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED                = errors.New("after a callback to 0x%p(0x%p), a completion call to freelibrary(%p) failed with status 0x%08x")
	ERROR_THREADPOOL_RELEASED_DURING_OPERATION                        = errors.New("the thread pool 0x%p was released while a thread was posting a callback to 0x%p(0x%p) to it")
	ERROR_CALLBACK_RETURNED_WHILE_IMPERSONATING                       = errors.New("a thread pool worker thread is impersonating a client, after a callback to 0x%p(0x%p). this is unexpected, indicating that the callback is missing a call to revert the impersonation")
	ERROR_APC_RETURNED_WHILE_IMPERSONATING                            = errors.New("a thread pool worker thread is impersonating a client, after executing an apc. this is unexpected, indicating that the apc is missing a call to revert the impersonation")
	ERROR_PROCESS_IS_PROTECTED                                        = errors.New("either the target process, or the target thread's containing process, is a protected process")
	ERROR_MCA_EXCEPTION                                               = errors.New("a thread is getting dispatched with mca exception because of mca")
	ERROR_CERTIFICATE_MAPPING_NOT_UNIQUE                              = errors.New("the client certificate account mapping is not unique")
	ERROR_SYMLINK_CLASS_DISABLED                                      = errors.New("the symbolic link cannot be followed because its type is disabled")
	ERROR_INVALID_IDN_NORMALIZATION                                   = errors.New("indicates that the specified string is not valid for idn normalization")
	ERROR_NO_UNICODE_TRANSLATION                                      = errors.New("no mapping for the unicode character exists in the target multi-byte code page")
	ERROR_ALREADY_REGISTERED                                          = errors.New("the provided callback is already registered")
	ERROR_CONTEXT_MISMATCH                                            = errors.New("the provided context did not match the target")
	ERROR_PORT_ALREADY_HAS_COMPLETION_LIST                            = errors.New("the specified port already has a completion list")
	ERROR_CALLBACK_RETURNED_THREAD_PRIORITY                           = errors.New("a threadpool worker thread entered a callback at thread base priority 0x%x and exited at priority 0x%x")
	ERROR_INVALID_THREAD                                              = errors.New("an invalid thread, handle %p, is specified for this operation. possibly, a threadpool worker thread was specified")
	ERROR_CALLBACK_RETURNED_TRANSACTION                               = errors.New("a threadpool worker thread entered a callback, which left transaction state")
	ERROR_CALLBACK_RETURNED_LDR_LOCK                                  = errors.New("a threadpool worker thread entered a callback, which left the loader lock held")
	ERROR_CALLBACK_RETURNED_LANG                                      = errors.New("a threadpool worker thread entered a callback, which left with preferred languages set")
	ERROR_CALLBACK_RETURNED_PRI_BACK                                  = errors.New("a threadpool worker thread entered a callback, which left with background priorities set")
	ERROR_DISK_REPAIR_DISABLED                                        = errors.New("the attempted operation required self healing to be enabled")
	ERROR_DS_DOMAIN_RENAME_IN_PROGRESS                                = errors.New("the directory service cannot perform the requested operation because a domain rename operation is in progress")
	ERROR_DISK_QUOTA_EXCEEDED                                         = errors.New("an operation failed because the storage quota was exceeded")
	ERROR_CONTENT_BLOCKED                                             = errors.New("an operation failed because the content was blocked")
	ERROR_BAD_CLUSTERS                                                = errors.New("the operation could not be completed due to bad clusters on disk")
	ERROR_VOLUME_DIRTY                                                = errors.New("the operation could not be completed because the volume is dirty. please run the chkdsk utility and try again")
	ERROR_FILE_CHECKED_OUT                                            = errors.New("this file is checked out or locked for editing by another user")
	ERROR_CHECKOUT_REQUIRED                                           = errors.New("the file must be checked out before saving changes")
	ERROR_BAD_FILE_TYPE                                               = errors.New("the file type being saved or retrieved has been blocked")
	ERROR_FILE_TOO_LARGE                                              = errors.New("the file size exceeds the limit allowed and cannot be saved")
	ERROR_FORMS_AUTH_REQUIRED                                         = errors.New("access denied. before opening files in this location, you must first browse to the e.g. site and select the option to log on automatically")
	ERROR_VIRUS_INFECTED                                              = errors.New("the operation did not complete successfully because the file contains a virus")
	ERROR_VIRUS_DELETED                                               = errors.New("this file contains a virus and cannot be opened. due to the nature of this virus, the file has been removed from this location")
	ERROR_BAD_MCFG_TABLE                                              = errors.New("the resources required for this device conflict with the mcfg table")
	ERROR_BAD_DATA                                                    = errors.New("bad data")
	ERROR_CANNOT_BREAK_OPLOCK                                         = errors.New("the operation did not complete successfully because it would cause an oplock to be broken. the caller has requested that existing oplocks not be broken")
	ERROR_WOW_ASSERTION                                               = errors.New("wow assertion error")
	ERROR_INVALID_SIGNATURE                                           = errors.New("the cryptographic signature is invalid")
	ERROR_HMAC_NOT_SUPPORTED                                          = errors.New("the cryptographic provider does not support hmac")
	ERROR_AUTH_TAG_MISMATCH                                           = errors.New("the computed authentication tag did not match the input authentication tag")
	ERROR_IPSEC_QUEUE_OVERFLOW                                        = errors.New("the ipsec queue overflowed")
	ERROR_ND_QUEUE_OVERFLOW                                           = errors.New("the neighbor discovery queue overflowed")
	ERROR_HOPLIMIT_EXCEEDED                                           = errors.New("an internet control message protocol (icmp) hop limit exceeded error was received")
	ERROR_PROTOCOL_NOT_SUPPORTED                                      = errors.New("the protocol is not installed on the local machine")
	ERROR_LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED                  = errors.New("{delayed write failed} windows was unable to save all the data for the file %hs; the data has been lost. this error might be caused by network connectivity issues. try to save this file elsewhere")
	ERROR_LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR                  = errors.New("{delayed write failed} windows was unable to save all the data for the file %hs; the data has been lost. this error was returned by the server on which the file exists. try to save this file elsewhere")
	ERROR_LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR                      = errors.New("{delayed write failed} windows was unable to save all the data for the file %hs; the data has been lost. this error might be caused if the device has been removed or the media is write-protected")
	ERROR_XML_PARSE_ERROR                                             = errors.New("windows was unable to parse the requested xml data")
	ERROR_XMLDSIG_ERROR                                               = errors.New("an error was encountered while processing an xml digital signature")
	ERROR_WRONG_COMPARTMENT                                           = errors.New("this indicates that the caller made the connection request in the wrong routing compartment")
	ERROR_AUTHIP_FAILURE                                              = errors.New("this indicates that there was an authip failure when attempting to connect to the remote host")
	ERROR_DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS                       = errors.New("oid mapped groups cannot have members")
	ERROR_DS_OID_NOT_FOUND                                            = errors.New("the specified oid cannot be found")
	ERROR_HASH_NOT_SUPPORTED                                          = errors.New("hash generation for the specified version and hash type is not enabled on server")
	ERROR_HASH_NOT_PRESENT                                            = errors.New("the hash requests is not present or not up to date with the current file contents")
	ERROR_OFFLOAD_READ_FLT_NOT_SUPPORTED                              = errors.New("a file system filter on the server has not opted in for offload read support")
	ERROR_OFFLOAD_WRITE_FLT_NOT_SUPPORTED                             = errors.New("a file system filter on the server has not opted in for offload write support")
	ERROR_OFFLOAD_READ_FILE_NOT_SUPPORTED                             = errors.New("offload read operations cannot be performed on: compressed files, sparse files, encrypted files, file system metadata files")
	ERROR_OFFLOAD_WRITE_FILE_NOT_SUPPORTED                            = errors.New("offload write operations cannot be performed on: compressed files, sparse files, encrypted files, file system metadata files")
	ERROR_DBG_NO_STATE_CHANGE                                         = errors.New("the debugger did not perform a state change")
	ERROR_DBG_APP_NOT_IDLE                                            = errors.New("the debugger found that the application is not idle")
	ERROR_RPC_NT_INVALID_STRING_BINDING                               = errors.New("the string binding is invalid")
	ERROR_RPC_NT_WRONG_KIND_OF_BINDING                                = errors.New("the binding handle is not the correct type")
	ERROR_RPC_NT_INVALID_BINDING                                      = errors.New("the binding handle is invalid")
	ERROR_RPC_NT_PROTSEQ_NOT_SUPPORTED                                = errors.New("the rpc protocol sequence is not supported")
	ERROR_RPC_NT_INVALID_RPC_PROTSEQ                                  = errors.New("the rpc protocol sequence is invalid")
	ERROR_RPC_NT_INVALID_STRING_UUID                                  = errors.New("the string uuid is invalid")
	ERROR_RPC_NT_INVALID_ENDPOINT_FORMAT                              = errors.New("the endpoint format is invalid")
	ERROR_RPC_NT_INVALID_NET_ADDR                                     = errors.New("the network address is invalid")
	ERROR_RPC_NT_NO_ENDPOINT_FOUND                                    = errors.New("no endpoint was found")
	ERROR_RPC_NT_INVALID_TIMEOUT                                      = errors.New("the time-out value is invalid")
	ERROR_RPC_NT_OBJECT_NOT_FOUND                                     = errors.New("the object uuid was not found")
	ERROR_RPC_NT_ALREADY_REGISTERED                                   = errors.New("the object uuid has already been registered")
	ERROR_RPC_NT_TYPE_ALREADY_REGISTERED                              = errors.New("the type uuid has already been registered")
	ERROR_RPC_NT_ALREADY_LISTENING                                    = errors.New("the rpc server is already listening")
	ERROR_RPC_NT_NO_PROTSEQS_REGISTERED                               = errors.New("no protocol sequences have been registered")
	ERROR_RPC_NT_NOT_LISTENING                                        = errors.New("the rpc server is not listening")
	ERROR_RPC_NT_UNKNOWN_MGR_TYPE                                     = errors.New("the manager type is unknown")
	ERROR_RPC_NT_UNKNOWN_IF                                           = errors.New("the interface is unknown")
	ERROR_RPC_NT_NO_BINDINGS                                          = errors.New("there are no bindings")
	ERROR_RPC_NT_NO_PROTSEQS                                          = errors.New("there are no protocol sequences")
	ERROR_RPC_NT_CANT_CREATE_ENDPOINT                                 = errors.New("the endpoint cannot be created")
	ERROR_RPC_NT_OUT_OF_RESOURCES                                     = errors.New("insufficient resources are available to complete this operation")
	ERROR_RPC_NT_SERVER_UNAVAILABLE                                   = errors.New("the rpc server is unavailable")
	ERROR_RPC_NT_SERVER_TOO_BUSY                                      = errors.New("the rpc server is too busy to complete this operation")
	ERROR_RPC_NT_INVALID_NETWORK_OPTIONS                              = errors.New("the network options are invalid")
	ERROR_RPC_NT_NO_CALL_ACTIVE                                       = errors.New("no rpcs are active on this thread")
	ERROR_RPC_NT_CALL_FAILED                                          = errors.New("the rpc failed")
	ERROR_RPC_NT_CALL_FAILED_DNE                                      = errors.New("the rpc failed and did not execute")
	ERROR_RPC_NT_PROTOCOL_ERROR                                       = errors.New("an rpc protocol error occurred")
	ERROR_RPC_NT_UNSUPPORTED_TRANS_SYN                                = errors.New("the rpc server does not support the transfer syntax")
	ERROR_RPC_NT_UNSUPPORTED_TYPE                                     = errors.New("the type uuid is not supported")
	ERROR_RPC_NT_INVALID_TAG                                          = errors.New("the tag is invalid")
	ERROR_RPC_NT_INVALID_BOUND                                        = errors.New("the array bounds are invalid")
	ERROR_RPC_NT_NO_ENTRY_NAME                                        = errors.New("the binding does not contain an entry name")
	ERROR_RPC_NT_INVALID_NAME_SYNTAX                                  = errors.New("the name syntax is invalid")
	ERROR_RPC_NT_UNSUPPORTED_NAME_SYNTAX                              = errors.New("the name syntax is not supported")
	ERROR_RPC_NT_UUID_NO_ADDRESS                                      = errors.New("no network address is available to construct a uuid")
	ERROR_RPC_NT_DUPLICATE_ENDPOINT                                   = errors.New("the endpoint is a duplicate")
	ERROR_RPC_NT_UNKNOWN_AUTHN_TYPE                                   = errors.New("the authentication type is unknown")
	ERROR_RPC_NT_MAX_CALLS_TOO_SMALL                                  = errors.New("the maximum number of calls is too small")
	ERROR_RPC_NT_STRING_TOO_LONG                                      = errors.New("the string is too long")
	ERROR_RPC_NT_PROTSEQ_NOT_FOUND                                    = errors.New("the rpc protocol sequence was not found")
	ERROR_RPC_NT_PROCNUM_OUT_OF_RANGE                                 = errors.New("the procedure number is out of range")
	ERROR_RPC_NT_BINDING_HAS_NO_AUTH                                  = errors.New("the binding does not contain any authentication information")
	ERROR_RPC_NT_UNKNOWN_AUTHN_SERVICE                                = errors.New("the authentication service is unknown")
	ERROR_RPC_NT_UNKNOWN_AUTHN_LEVEL                                  = errors.New("the authentication level is unknown")
	ERROR_RPC_NT_INVALID_AUTH_IDENTITY                                = errors.New("the security context is invalid")
	ERROR_RPC_NT_UNKNOWN_AUTHZ_SERVICE                                = errors.New("the authorization service is unknown")
	ERROR_EPT_NT_INVALID_ENTRY                                        = errors.New("the entry is invalid")
	ERROR_EPT_NT_CANT_PERFORM_OP                                      = errors.New("the operation cannot be performed")
	ERROR_EPT_NT_NOT_REGISTERED                                       = errors.New("no more endpoints are available from the endpoint mapper")
	ERROR_RPC_NT_NOTHING_TO_EXPORT                                    = errors.New("no interfaces have been exported")
	ERROR_RPC_NT_INCOMPLETE_NAME                                      = errors.New("the entry name is incomplete")
	ERROR_RPC_NT_INVALID_VERS_OPTION                                  = errors.New("the version option is invalid")
	ERROR_RPC_NT_NO_MORE_MEMBERS                                      = errors.New("there are no more members")
	ERROR_RPC_NT_NOT_ALL_OBJS_UNEXPORTED                              = errors.New("there is nothing to unexport")
	ERROR_RPC_NT_INTERFACE_NOT_FOUND                                  = errors.New("the interface was not found")
	ERROR_RPC_NT_ENTRY_ALREADY_EXISTS                                 = errors.New("the entry already exists")
	ERROR_RPC_NT_ENTRY_NOT_FOUND                                      = errors.New("the entry was not found")
	ERROR_RPC_NT_NAME_SERVICE_UNAVAILABLE                             = errors.New("the name service is unavailable")
	ERROR_RPC_NT_INVALID_NAF_ID                                       = errors.New("the network address family is invalid")
	ERROR_RPC_NT_CANNOT_SUPPORT                                       = errors.New("the requested operation is not supported")
	ERROR_RPC_NT_NO_CONTEXT_AVAILABLE                                 = errors.New("no security context is available to allow impersonation")
	ERROR_RPC_NT_INTERNAL_ERROR                                       = errors.New("an internal error occurred in the rpc")
	ERROR_RPC_NT_ZERO_DIVIDE                                          = errors.New("the rpc server attempted to divide an integer by zero")
	ERROR_RPC_NT_ADDRESS_ERROR                                        = errors.New("an addressing error occurred in the rpc server")
	ERROR_RPC_NT_FP_DIV_ZERO                                          = errors.New("a floating point operation at the rpc server caused a divide by zero")
	ERROR_RPC_NT_FP_UNDERFLOW                                         = errors.New("a floating point underflow occurred at the rpc server")
	ERROR_RPC_NT_FP_OVERFLOW                                          = errors.New("a floating point overflow occurred at the rpc server")
	ERROR_RPC_NT_CALL_IN_PROGRESS                                     = errors.New("an rpc is already in progress for this thread")
	ERROR_RPC_NT_NO_MORE_BINDINGS                                     = errors.New("there are no more bindings")
	ERROR_RPC_NT_GROUP_MEMBER_NOT_FOUND                               = errors.New("the group member was not found")
	ERROR_EPT_NT_CANT_CREATE                                          = errors.New("the endpoint mapper database entry could not be created")
	ERROR_RPC_NT_INVALID_OBJECT                                       = errors.New("the object uuid is the nil uuid")
	ERROR_RPC_NT_NO_INTERFACES                                        = errors.New("no interfaces have been registered")
	ERROR_RPC_NT_CALL_CANCELLED                                       = errors.New("the rpc was canceled")
	ERROR_RPC_NT_BINDING_INCOMPLETE                                   = errors.New("the binding handle does not contain all the required information")
	ERROR_RPC_NT_COMM_FAILURE                                         = errors.New("a communications failure occurred during an rpc")
	ERROR_RPC_NT_UNSUPPORTED_AUTHN_LEVEL                              = errors.New("the requested authentication level is not supported")
	ERROR_RPC_NT_NO_PRINC_NAME                                        = errors.New("no principal name was registered")
	ERROR_RPC_NT_NOT_RPC_ERROR                                        = errors.New("the error specified is not a valid windows rpc error code")
	ERROR_RPC_NT_SEC_PKG_ERROR                                        = errors.New("a security package-specific error occurred")
	ERROR_RPC_NT_NOT_CANCELLED                                        = errors.New("the thread was not canceled")
	ERROR_RPC_NT_INVALID_ASYNC_HANDLE                                 = errors.New("invalid asynchronous rpc handle")
	ERROR_RPC_NT_INVALID_ASYNC_CALL                                   = errors.New("invalid asynchronous rpc call handle for this operation")
	ERROR_RPC_NT_PROXY_ACCESS_DENIED                                  = errors.New("access to the http proxy is denied")
	ERROR_RPC_NT_NO_MORE_ENTRIES                                      = errors.New("the list of rpc servers available for auto-handle binding has been exhausted")
	ERROR_RPC_NT_SS_CHAR_TRANS_OPEN_FAIL                              = errors.New("the file designated by dcerpcchartrans cannot be opened")
	ERROR_RPC_NT_SS_CHAR_TRANS_SHORT_FILE                             = errors.New("the file containing the character translation table has fewer than 512 bytes")
	ERROR_RPC_NT_SS_IN_NULL_CONTEXT                                   = errors.New("a null context handle is passed as an [in] parameter")
	ERROR_RPC_NT_SS_CONTEXT_MISMATCH                                  = errors.New("the context handle does not match any known context handles")
	ERROR_RPC_NT_SS_CONTEXT_DAMAGED                                   = errors.New("the context handle changed during a call")
	ERROR_RPC_NT_SS_HANDLES_MISMATCH                                  = errors.New("the binding handles passed to an rpc do not match")
	ERROR_RPC_NT_SS_CANNOT_GET_CALL_HANDLE                            = errors.New("the stub is unable to get the call handle")
	ERROR_RPC_NT_NULL_REF_POINTER                                     = errors.New("a null reference pointer was passed to the stub")
	ERROR_RPC_NT_ENUM_VALUE_OUT_OF_RANGE                              = errors.New("the enumeration value is out of range")
	ERROR_RPC_NT_BYTE_COUNT_TOO_SMALL                                 = errors.New("the byte count is too small")
	ERROR_RPC_NT_BAD_STUB_DATA                                        = errors.New("the stub received bad data")
	ERROR_RPC_NT_INVALID_ES_ACTION                                    = errors.New("invalid operation on the encoding/decoding handle")
	ERROR_RPC_NT_WRONG_ES_VERSION                                     = errors.New("incompatible version of the serializing package")
	ERROR_RPC_NT_WRONG_STUB_VERSION                                   = errors.New("incompatible version of the rpc stub")
	ERROR_RPC_NT_INVALID_PIPE_OBJECT                                  = errors.New("the rpc pipe object is invalid or corrupt")
	ERROR_RPC_NT_INVALID_PIPE_OPERATION                               = errors.New("an invalid operation was attempted on an rpc pipe object")
	ERROR_RPC_NT_WRONG_PIPE_VERSION                                   = errors.New("unsupported rpc pipe version")
	ERROR_RPC_NT_PIPE_CLOSED                                          = errors.New("the rpc pipe object has already been closed")
	ERROR_RPC_NT_PIPE_DISCIPLINE_ERROR                                = errors.New("the rpc call completed before all pipes were processed")
	ERROR_RPC_NT_PIPE_EMPTY                                           = errors.New("no more data is available from the rpc pipe")
	ERROR_PNP_BAD_MPS_TABLE                                           = errors.New("a device is missing in the system bios mps table. this device will not be used. contact your system vendor for a system bios update")
	ERROR_PNP_TRANSLATION_FAILED                                      = errors.New("a translator failed to translate resources")
	ERROR_PNP_IRQ_TRANSLATION_FAILED                                  = errors.New("an irq translator failed to translate resources")
	ERROR_PNP_INVALID_ID                                              = errors.New("driver %2 returned an invalid id for a child device (%3)")
	ERROR_IO_REISSUE_AS_CACHED                                        = errors.New("reissue the given operation as a cached i/o operation")
	ERROR_CTX_WINSTATION_NAME_INVALID                                 = errors.New("session name %1 is invalid")
	ERROR_CTX_INVALID_PD                                              = errors.New("the protocol driver %1 is invalid")
	ERROR_CTX_PD_NOT_FOUND                                            = errors.New("the protocol driver %1 was not found in the system path")
	ERROR_CTX_CLOSE_PENDING                                           = errors.New("a close operation is pending on the terminal connection")
	ERROR_CTX_NO_OUTBUF                                               = errors.New("no free output buffers are available")
	ERROR_CTX_MODEM_INF_NOT_FOUND                                     = errors.New("the modem.inf file was not found")
	ERROR_CTX_INVALID_MODEMNAME                                       = errors.New("the modem (%1) was not found in the modem.inf file")
	ERROR_CTX_RESPONSE_ERROR                                          = errors.New("the modem did not accept the command sent to it. verify that the configured modem name matches the attached modem")
	ERROR_CTX_MODEM_RESPONSE_TIMEOUT                                  = errors.New("the modem did not respond to the command sent to it. verify that the modem cable is properly attached and the modem is turned on")
	ERROR_CTX_MODEM_RESPONSE_NO_CARRIER                               = errors.New("carrier detection has failed or the carrier has been dropped due to disconnection")
	ERROR_CTX_MODEM_RESPONSE_NO_DIALTONE                              = errors.New("a dial tone was not detected within the required time. verify that the phone cable is properly attached and functional")
	ERROR_CTX_MODEM_RESPONSE_BUSY                                     = errors.New("a busy signal was detected at a remote site on callback")
	ERROR_CTX_MODEM_RESPONSE_VOICE                                    = errors.New("a voice was detected at a remote site on callback")
	ERROR_CTX_TD_ERROR                                                = errors.New("transport driver error")
	ERROR_CTX_LICENSE_CLIENT_INVALID                                  = errors.New("the client you are using is not licensed to use this system. your logon request is denied")
	ERROR_CTX_LICENSE_NOT_AVAILABLE                                   = errors.New("the system has reached its licensed logon limit. try again later")
	ERROR_CTX_LICENSE_EXPIRED                                         = errors.New("the system license has expired. your logon request is denied")
	ERROR_CTX_WINSTATION_NOT_FOUND                                    = errors.New("the specified session cannot be found")
	ERROR_CTX_WINSTATION_NAME_COLLISION                               = errors.New("the specified session name is already in use")
	ERROR_CTX_WINSTATION_BUSY                                         = errors.New("the requested operation cannot be completed because the terminal connection is currently processing a connect, disconnect, reset, or delete operation")
	ERROR_CTX_BAD_VIDEO_MODE                                          = errors.New("an attempt has been made to connect to a session whose video mode is not supported by the current client")
	ERROR_CTX_GRAPHICS_INVALID                                        = errors.New("the application attempted to enable dos graphics mode. dos graphics mode is not supported")
	ERROR_CTX_NOT_CONSOLE                                             = errors.New("the requested operation can be performed only on the system console. this is most often the result of a driver or system dll requiring direct console access")
	ERROR_CTX_CLIENT_QUERY_TIMEOUT                                    = errors.New("the client failed to respond to the server connect message")
	ERROR_CTX_CONSOLE_DISCONNECT                                      = errors.New("disconnecting the console session is not supported")
	ERROR_CTX_CONSOLE_CONNECT                                         = errors.New("reconnecting a disconnected session to the console is not supported")
	ERROR_CTX_SHADOW_DENIED                                           = errors.New("the request to control another session remotely was denied")
	ERROR_CTX_WINSTATION_ACCESS_DENIED                                = errors.New("a process has requested access to a session, but has not been granted those access rights")
	ERROR_CTX_INVALID_WD                                              = errors.New("the terminal connection driver %1 is invalid")
	ERROR_CTX_WD_NOT_FOUND                                            = errors.New("the terminal connection driver %1 was not found in the system path")
	ERROR_CTX_SHADOW_INVALID                                          = errors.New("the requested session cannot be controlled remotely. you cannot control your own session, a session that is trying to control your session, a session that has no user logged on, or other sessions from the console")
	ERROR_CTX_SHADOW_DISABLED                                         = errors.New("the requested session is not configured to allow remote control")
	ERROR_RDP_PROTOCOL_ERROR                                          = errors.New("the rdp protocol component %2 detected an error in the protocol stream and has disconnected the client")
	ERROR_CTX_CLIENT_LICENSE_NOT_SET                                  = errors.New("your request to connect to this terminal server has been rejected. your terminal server client license number has not been entered for this copy of the terminal client. contact your system administrator for help in entering a valid, unique license number for this terminal server client. click ok to continue")
	ERROR_CTX_CLIENT_LICENSE_IN_USE                                   = errors.New("your request to connect to this terminal server has been rejected. your terminal server client license number is currently being used by another user. contact your system administrator to obtain a new copy of the terminal server client with a valid, unique license number. click ok to continue")
	ERROR_CTX_SHADOW_ENDED_BY_MODE_CHANGE                             = errors.New("the remote control of the console was terminated because the display mode was changed. changing the display mode in a remote control session is not supported")
	ERROR_CTX_SHADOW_NOT_RUNNING                                      = errors.New("remote control could not be terminated because the specified session is not currently being remotely controlled")
	ERROR_CTX_LOGON_DISABLED                                          = errors.New("your interactive logon privilege has been disabled. contact your system administrator")
	ERROR_CTX_SECURITY_LAYER_ERROR                                    = errors.New("the terminal server security layer detected an error in the protocol stream and has disconnected the client")
	ERROR_TS_INCOMPATIBLE_SESSIONS                                    = errors.New("the target session is incompatible with the current session")
	ERROR_MUI_FILE_NOT_FOUND                                          = errors.New("the resource loader failed to find an mui file")
	ERROR_MUI_INVALID_FILE                                            = errors.New("the resource loader failed to load an mui file because the file failed to pass validation")
	ERROR_MUI_INVALID_RC_CONFIG                                       = errors.New("the rc manifest is corrupted with garbage data, is an unsupported version, or is missing a required item")
	ERROR_MUI_INVALID_LOCALE_NAME                                     = errors.New("the rc manifest has an invalid culture name")
	ERROR_MUI_INVALID_ULTIMATEFALLBACK_NAME                           = errors.New("the rc manifest has and invalid ultimate fallback name")
	ERROR_MUI_FILE_NOT_LOADED                                         = errors.New("the resource loader cache does not have a loaded mui entry")
	ERROR_RESOURCE_ENUM_USER_STOP                                     = errors.New("the user stopped resource enumeration")
	ERROR_CLUSTER_INVALID_NODE                                        = errors.New("the cluster node is not valid")
	ERROR_CLUSTER_NODE_EXISTS                                         = errors.New("the cluster node already exists")
	ERROR_CLUSTER_JOIN_IN_PROGRESS                                    = errors.New("a node is in the process of joining the cluster")
	ERROR_CLUSTER_NODE_NOT_FOUND                                      = errors.New("the cluster node was not found")
	ERROR_CLUSTER_LOCAL_NODE_NOT_FOUND                                = errors.New("the cluster local node information was not found")
	ERROR_CLUSTER_NETWORK_EXISTS                                      = errors.New("the cluster network already exists")
	ERROR_CLUSTER_NETWORK_NOT_FOUND                                   = errors.New("the cluster network was not found")
	ERROR_CLUSTER_NETINTERFACE_EXISTS                                 = errors.New("the cluster network interface already exists")
	ERROR_CLUSTER_NETINTERFACE_NOT_FOUND                              = errors.New("the cluster network interface was not found")
	ERROR_CLUSTER_INVALID_REQUEST                                     = errors.New("the cluster request is not valid for this object")
	ERROR_CLUSTER_INVALID_NETWORK_PROVIDER                            = errors.New("the cluster network provider is not valid")
	ERROR_CLUSTER_NODE_DOWN                                           = errors.New("the cluster node is down")
	ERROR_CLUSTER_NODE_UNREACHABLE                                    = errors.New("the cluster node is not reachable")
	ERROR_CLUSTER_NODE_NOT_MEMBER                                     = errors.New("the cluster node is not a member of the cluster")
	ERROR_CLUSTER_JOIN_NOT_IN_PROGRESS                                = errors.New("a cluster join operation is not in progress")
	ERROR_CLUSTER_INVALID_NETWORK                                     = errors.New("the cluster network is not valid")
	ERROR_CLUSTER_NO_NET_ADAPTERS                                     = errors.New("no network adapters are available")
	ERROR_CLUSTER_NODE_UP                                             = errors.New("the cluster node is up")
	ERROR_CLUSTER_NODE_PAUSED                                         = errors.New("the cluster node is paused")
	ERROR_CLUSTER_NODE_NOT_PAUSED                                     = errors.New("the cluster node is not paused")
	ERROR_CLUSTER_NO_SECURITY_CONTEXT                                 = errors.New("no cluster security context is available")
	ERROR_CLUSTER_NETWORK_NOT_INTERNAL                                = errors.New("the cluster network is not configured for internal cluster communication")
	ERROR_CLUSTER_POISONED                                            = errors.New("the cluster node has been poisoned")
	ERROR_ACPI_INVALID_OPCODE                                         = errors.New("an attempt was made to run an invalid aml opcode")
	ERROR_ACPI_STACK_OVERFLOW                                         = errors.New("the aml interpreter stack has overflowed")
	ERROR_ACPI_ASSERT_FAILED                                          = errors.New("an inconsistent state has occurred")
	ERROR_ACPI_INVALID_INDEX                                          = errors.New("an attempt was made to access an array outside its bounds")
	ERROR_ACPI_INVALID_ARGUMENT                                       = errors.New("a required argument was not specified")
	ERROR_ACPI_FATAL                                                  = errors.New("a fatal error has occurred")
	ERROR_ACPI_INVALID_SUPERNAME                                      = errors.New("an invalid supername was specified")
	ERROR_ACPI_INVALID_ARGTYPE                                        = errors.New("an argument with an incorrect type was specified")
	ERROR_ACPI_INVALID_OBJTYPE                                        = errors.New("an object with an incorrect type was specified")
	ERROR_ACPI_INVALID_TARGETTYPE                                     = errors.New("a target with an incorrect type was specified")
	ERROR_ACPI_INCORRECT_ARGUMENT_COUNT                               = errors.New("an incorrect number of arguments was specified")
	ERROR_ACPI_ADDRESS_NOT_MAPPED                                     = errors.New("an address failed to translate")
	ERROR_ACPI_INVALID_EVENTTYPE                                      = errors.New("an incorrect event type was specified")
	ERROR_ACPI_HANDLER_COLLISION                                      = errors.New("a handler for the target already exists")
	ERROR_ACPI_INVALID_DATA                                           = errors.New("invalid data for the target was specified")
	ERROR_ACPI_INVALID_REGION                                         = errors.New("an invalid region for the target was specified")
	ERROR_ACPI_INVALID_ACCESS_SIZE                                    = errors.New("an attempt was made to access a field outside the defined range")
	ERROR_ACPI_ACQUIRE_GLOBAL_LOCK                                    = errors.New("the global system lock could not be acquired")
	ERROR_ACPI_ALREADY_INITIALIZED                                    = errors.New("an attempt was made to reinitialize the acpi subsystem")
	ERROR_ACPI_NOT_INITIALIZED                                        = errors.New("the acpi subsystem has not been initialized")
	ERROR_ACPI_INVALID_MUTEX_LEVEL                                    = errors.New("an incorrect mutex was specified")
	ERROR_ACPI_MUTEX_NOT_OWNED                                        = errors.New("the mutex is not currently owned")
	ERROR_ACPI_MUTEX_NOT_OWNER                                        = errors.New("an attempt was made to access the mutex by a process that was not the owner")
	ERROR_ACPI_RS_ACCESS                                              = errors.New("an error occurred during an access to region space")
	ERROR_ACPI_INVALID_TABLE                                          = errors.New("an attempt was made to use an incorrect table")
	ERROR_ACPI_REG_HANDLER_FAILED                                     = errors.New("the registration of an acpi event failed")
	ERROR_ACPI_POWER_REQUEST_FAILED                                   = errors.New("an acpi power object failed to transition state")
	ERROR_SXS_SECTION_NOT_FOUND                                       = errors.New("the requested section is not present in the activation context")
	ERROR_SXS_CANT_GEN_ACTCTX                                         = errors.New("windows was unble to process the application binding information. refer to the system event log for further information")
	ERROR_SXS_INVALID_ACTCTXDATA_FORMAT                               = errors.New("the application binding data format is invalid")
	ERROR_SXS_ASSEMBLY_NOT_FOUND                                      = errors.New("the referenced assembly is not installed on the system")
	ERROR_SXS_MANIFEST_FORMAT_ERROR                                   = errors.New("the manifest file does not begin with the required tag and format information")
	ERROR_SXS_MANIFEST_PARSE_ERROR                                    = errors.New("the manifest file contains one or more syntax errors")
	ERROR_SXS_ACTIVATION_CONTEXT_DISABLED                             = errors.New("the application attempted to activate a disabled activation context")
	ERROR_SXS_KEY_NOT_FOUND                                           = errors.New("the requested lookup key was not found in any active activation context")
	ERROR_SXS_VERSION_CONFLICT                                        = errors.New("a component version required by the application conflicts with another component version that is already active")
	ERROR_SXS_WRONG_SECTION_TYPE                                      = errors.New("the type requested activation context section does not match the query api used")
	ERROR_SXS_THREAD_QUERIES_DISABLED                                 = errors.New("lack of system resources has required isolated activation to be disabled for the current thread of execution")
	ERROR_SXS_ASSEMBLY_MISSING                                        = errors.New("the referenced assembly could not be found")
	ERROR_SXS_PROCESS_DEFAULT_ALREADY_SET                             = errors.New("an attempt to set the process default activation context failed because the process default activation context was already set")
	ERROR_SXS_EARLY_DEACTIVATION                                      = errors.New("the activation context being deactivated is not the most recently activated one")
	ERROR_SXS_INVALID_DEACTIVATION                                    = errors.New("the activation context being deactivated is not active for the current thread of execution")
	ERROR_SXS_MULTIPLE_DEACTIVATION                                   = errors.New("the activation context being deactivated has already been deactivated")
	ERROR_SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY                 = errors.New("the activation context of the system default assembly could not be generated")
	ERROR_SXS_PROCESS_TERMINATION_REQUESTED                           = errors.New("a component used by the isolation facility has requested that the process be terminated")
	ERROR_SXS_CORRUPT_ACTIVATION_STACK                                = errors.New("the activation context activation stack for the running thread of execution is corrupt")
	ERROR_SXS_CORRUPTION                                              = errors.New("the application isolation metadata for this process or thread has become corrupt")
	ERROR_SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE                        = errors.New("the value of an attribute in an identity is not within the legal range")
	ERROR_SXS_INVALID_IDENTITY_ATTRIBUTE_NAME                         = errors.New("the name of an attribute in an identity is not within the legal range")
	ERROR_SXS_IDENTITY_DUPLICATE_ATTRIBUTE                            = errors.New("an identity contains two definitions for the same attribute")
	ERROR_SXS_IDENTITY_PARSE_ERROR                                    = errors.New("the identity string is malformed. this might be due to a trailing comma, more than two unnamed attributes, a missing attribute name, or a missing attribute value")
	ERROR_SXS_COMPONENT_STORE_CORRUPT                                 = errors.New("the component store has become corrupted")
	ERROR_SXS_FILE_HASH_MISMATCH                                      = errors.New("a component's file does not match the verification information present in the component manifest")
	ERROR_SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT           = errors.New("the identities of the manifests are identical, but their contents are different")
	ERROR_SXS_IDENTITIES_DIFFERENT                                    = errors.New("the component identities are different")
	ERROR_SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT                            = errors.New("the assembly is not a deployment")
	ERROR_SXS_FILE_NOT_PART_OF_ASSEMBLY                               = errors.New("the file is not a part of the assembly")
	ERROR_ADVANCED_INSTALLER_FAILED                                   = errors.New("an advanced installer failed during setup or servicing")
	ERROR_XML_ENCODING_MISMATCH                                       = errors.New("the character encoding in the xml declaration did not match the encoding used in the document")
	ERROR_SXS_MANIFEST_TOO_BIG                                        = errors.New("the size of the manifest exceeds the maximum allowed")
	ERROR_SXS_SETTING_NOT_REGISTERED                                  = errors.New("the setting is not registered")
	ERROR_SXS_TRANSACTION_CLOSURE_INCOMPLETE                          = errors.New("one or more required transaction members are not present")
	ERROR_SMI_PRIMITIVE_INSTALLER_FAILED                              = errors.New("the smi primitive installer failed during setup or servicing")
	ERROR_GENERIC_COMMAND_FAILED                                      = errors.New("a generic command executable returned a result that indicates failure")
	ERROR_SXS_FILE_HASH_MISSING                                       = errors.New("a component is missing file verification information in its manifest")
	ERROR_TRANSACTIONAL_CONFLICT                                      = errors.New("the function attempted to use a name that is reserved for use by another transaction")
	ERROR_INVALID_TRANSACTION                                         = errors.New("the transaction handle associated with this operation is invalid")
	ERROR_TRANSACTION_NOT_ACTIVE                                      = errors.New("the requested operation was made in the context of a transaction that is no longer active")
	ERROR_TM_INITIALIZATION_FAILED                                    = errors.New("the transaction manager was unable to be successfully initialized. transacted operations are not supported")
	ERROR_RM_NOT_ACTIVE                                               = errors.New("transaction support within the specified file system resource manager was not started or was shut down due to an error")
	ERROR_RM_METADATA_CORRUPT                                         = errors.New("the metadata of the resource manager has been corrupted. the resource manager will not function")
	ERROR_TRANSACTION_NOT_JOINED                                      = errors.New("the resource manager attempted to prepare a transaction that it has not successfully joined")
	ERROR_DIRECTORY_NOT_RM                                            = errors.New("the specified directory does not contain a file system resource manager")
	ERROR_TRANSACTIONS_UNSUPPORTED_REMOTE                             = errors.New("the remote server or share does not support transacted file operations")
	ERROR_LOG_RESIZE_INVALID_SIZE                                     = errors.New("the requested log size for the file system resource manager is invalid")
	ERROR_REMOTE_FILE_VERSION_MISMATCH                                = errors.New("the remote server sent mismatching version number or fid for a file opened with transactions")
	ERROR_CRM_PROTOCOL_ALREADY_EXISTS                                 = errors.New("the resource manager tried to register a protocol that already exists")
	ERROR_TRANSACTION_PROPAGATION_FAILED                              = errors.New("the attempt to propagate the transaction failed")
	ERROR_CRM_PROTOCOL_NOT_FOUND                                      = errors.New("the requested propagation protocol was not registered as a crm")
	ERROR_TRANSACTION_SUPERIOR_EXISTS                                 = errors.New("the transaction object already has a superior enlistment, and the caller attempted an operation that would have created a new superior. only a single superior enlistment is allowed")
	ERROR_TRANSACTION_REQUEST_NOT_VALID                               = errors.New("the requested operation is not valid on the transaction object in its current state")
	ERROR_TRANSACTION_NOT_REQUESTED                                   = errors.New("the caller has called a response api, but the response is not expected because the transaction manager did not issue the corresponding request to the caller")
	ERROR_TRANSACTION_ALREADY_ABORTED                                 = errors.New("it is too late to perform the requested operation, because the transaction has already been aborted")
	ERROR_TRANSACTION_ALREADY_COMMITTED                               = errors.New("it is too late to perform the requested operation, because the transaction has already been committed")
	ERROR_TRANSACTION_INVALID_MARSHALL_BUFFER                         = errors.New("the buffer passed in to ntpushtransaction or ntpulltransaction is not in a valid format")
	ERROR_CURRENT_TRANSACTION_NOT_VALID                               = errors.New("the current transaction context associated with the thread is not a valid handle to a transaction object")
	ERROR_LOG_GROWTH_FAILED                                           = errors.New("an attempt to create space in the transactional resource manager's log failed. the failure status has been recorded in the event log")
	ERROR_OBJECT_NO_LONGER_EXISTS                                     = errors.New("the object (file, stream, or link) that corresponds to the handle has been deleted by a transaction savepoint rollback")
	ERROR_STREAM_MINIVERSION_NOT_FOUND                                = errors.New("the specified file miniversion was not found for this transacted file open")
	ERROR_STREAM_MINIVERSION_NOT_VALID                                = errors.New("the specified file miniversion was found but has been invalidated. the most likely cause is a transaction savepoint rollback")
	ERROR_MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION         = errors.New("a miniversion can be opened only in the context of the transaction that created it")
	ERROR_CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT                    = errors.New("it is not possible to open a miniversion with modify access")
	ERROR_CANT_CREATE_MORE_STREAM_MINIVERSIONS                        = errors.New("it is not possible to create any more miniversions for this stream")
	ERROR_HANDLE_NO_LONGER_VALID                                      = errors.New("the handle has been invalidated by a transaction. the most likely cause is the presence of memory mapping on a file or an open handle when the transaction ended or rolled back to savepoint")
	ERROR_LOG_CORRUPTION_DETECTED                                     = errors.New("the log data is corrupt")
	ERROR_RM_DISCONNECTED                                             = errors.New("the transaction outcome is unavailable because the resource manager responsible for it is disconnected")
	ERROR_ENLISTMENT_NOT_SUPERIOR                                     = errors.New("the request was rejected because the enlistment in question is not a superior enlistment")
	ERROR_FILE_IDENTITY_NOT_PERSISTENT                                = errors.New("the file cannot be opened in a transaction because its identity depends on the outcome of an unresolved transaction")
	ERROR_CANT_BREAK_TRANSACTIONAL_DEPENDENCY                         = errors.New("the operation cannot be performed because another transaction is depending on this property not changing")
	ERROR_CANT_CROSS_RM_BOUNDARY                                      = errors.New("the operation would involve a single file with two transactional resource managers and is, therefore, not allowed")
	ERROR_TXF_DIR_NOT_EMPTY                                           = errors.New("the $txf directory must be empty for this operation to succeed")
	ERROR_INDOUBT_TRANSACTIONS_EXIST                                  = errors.New("the operation would leave a transactional resource manager in an inconsistent state and is therefore not allowed")
	ERROR_TM_VOLATILE                                                 = errors.New("the operation could not be completed because the transaction manager does not have a log")
	ERROR_ROLLBACK_TIMER_EXPIRED                                      = errors.New("a rollback could not be scheduled because a previously scheduled rollback has already executed or been queued for execution")
	ERROR_TXF_ATTRIBUTE_CORRUPT                                       = errors.New("the transactional metadata attribute on the file or directory %hs is corrupt and unreadable")
	ERROR_EFS_NOT_ALLOWED_IN_TRANSACTION                              = errors.New("the encryption operation could not be completed because a transaction is active")
	ERROR_TRANSACTIONAL_OPEN_NOT_ALLOWED                              = errors.New("this object is not allowed to be opened in a transaction")
	ERROR_TRANSACTED_MAPPING_UNSUPPORTED_REMOTE                       = errors.New("memory mapping (creating a mapped section) a remote file under a transaction is not supported")
	ERROR_TRANSACTION_REQUIRED_PROMOTION                              = errors.New("promotion was required to allow the resource manager to enlist, but the transaction was set to disallow it")
	ERROR_CANNOT_EXECUTE_FILE_IN_TRANSACTION                          = errors.New("this file is open for modification in an unresolved transaction and can be opened for execute only by a transacted reader")
	ERROR_TRANSACTIONS_NOT_FROZEN                                     = errors.New("the request to thaw frozen transactions was ignored because transactions were not previously frozen")
	ERROR_TRANSACTION_FREEZE_IN_PROGRESS                              = errors.New("transactions cannot be frozen because a freeze is already in progress")
	ERROR_NOT_SNAPSHOT_VOLUME                                         = errors.New("the target volume is not a snapshot volume. this operation is valid only on a volume mounted as a snapshot")
	ERROR_NO_SAVEPOINT_WITH_OPEN_FILES                                = errors.New("the savepoint operation failed because files are open on the transaction, which is not permitted")
	ERROR_SPARSE_NOT_ALLOWED_IN_TRANSACTION                           = errors.New("the sparse operation could not be completed because a transaction is active on the file")
	ERROR_TM_IDENTITY_MISMATCH                                        = errors.New("the call to create a transaction manager object failed because the tm identity that is stored in the log file does not match the tm identity that was passed in as an argument")
	ERROR_FLOATED_SECTION                                             = errors.New("i/o was attempted on a section object that has been floated as a result of a transaction ending. there is no valid data")
	ERROR_CANNOT_ACCEPT_TRANSACTED_WORK                               = errors.New("the transactional resource manager cannot currently accept transacted work due to a transient condition, such as low resources")
	ERROR_CANNOT_ABORT_TRANSACTIONS                                   = errors.New("the transactional resource manager had too many transactions outstanding that could not be aborted. the transactional resource manager has been shut down")
	ERROR_TRANSACTION_NOT_FOUND                                       = errors.New("the specified transaction was unable to be opened because it was not found")
	ERROR_RESOURCEMANAGER_NOT_FOUND                                   = errors.New("the specified resource manager was unable to be opened because it was not found")
	ERROR_ENLISTMENT_NOT_FOUND                                        = errors.New("the specified enlistment was unable to be opened because it was not found")
	ERROR_TRANSACTIONMANAGER_NOT_FOUND                                = errors.New("the specified transaction manager was unable to be opened because it was not found")
	ERROR_TRANSACTIONMANAGER_NOT_ONLINE                               = errors.New("the specified resource manager was unable to create an enlistment because its associated transaction manager is not online")
	ERROR_TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION                  = errors.New("the specified transaction manager was unable to create the objects contained in its log file in the ob namespace. therefore, the transaction manager was unable to recover")
	ERROR_TRANSACTION_NOT_ROOT                                        = errors.New("the call to create a superior enlistment on this transaction object could not be completed because the transaction object specified for the enlistment is a subordinate branch of the transaction. only the root of the transaction can be enlisted as a superior")
	ERROR_TRANSACTION_OBJECT_EXPIRED                                  = errors.New("because the associated transaction manager or resource manager has been closed, the handle is no longer valid")
	ERROR_COMPRESSION_NOT_ALLOWED_IN_TRANSACTION                      = errors.New("the compression operation could not be completed because a transaction is active on the file")
	ERROR_TRANSACTION_RESPONSE_NOT_ENLISTED                           = errors.New("the specified operation could not be performed on this superior enlistment because the enlistment was not created with the corresponding completion response in the notificationmask")
	ERROR_TRANSACTION_RECORD_TOO_LONG                                 = errors.New("the specified operation could not be performed because the record to be logged was too long. this can occur because either there are too many enlistments on this transaction or the combined recoveryinformation being logged on behalf of those enlistments is too long")
	ERROR_NO_LINK_TRACKING_IN_TRANSACTION                             = errors.New("the link-tracking operation could not be completed because a transaction is active")
	ERROR_OPERATION_NOT_SUPPORTED_IN_TRANSACTION                      = errors.New("this operation cannot be performed in a transaction")
	ERROR_TRANSACTION_INTEGRITY_VIOLATED                              = errors.New("the kernel transaction manager had to abort or forget the transaction because it blocked forward progress")
	ERROR_EXPIRED_HANDLE                                              = errors.New("the handle is no longer properly associated with its transaction. it might have been opened in a transactional resource manager that was subsequently forced to restart. please close the handle and open a new one")
	ERROR_TRANSACTION_NOT_ENLISTED                                    = errors.New("the specified operation could not be performed because the resource manager is not enlisted in the transaction")
	ERROR_LOG_SECTOR_INVALID                                          = errors.New("the log service found an invalid log sector")
	ERROR_LOG_SECTOR_PARITY_INVALID                                   = errors.New("the log service encountered a log sector with invalid block parity")
	ERROR_LOG_SECTOR_REMAPPED                                         = errors.New("the log service encountered a remapped log sector")
	ERROR_LOG_BLOCK_INCOMPLETE                                        = errors.New("the log service encountered a partial or incomplete log block")
	ERROR_LOG_INVALID_RANGE                                           = errors.New("the log service encountered an attempt to access data outside the active log range")
	ERROR_LOG_BLOCKS_EXHAUSTED                                        = errors.New("the log service user-log marshaling buffers are exhausted")
	ERROR_LOG_READ_CONTEXT_INVALID                                    = errors.New("the log service encountered an attempt to read from a marshaling area with an invalid read context")
	ERROR_LOG_RESTART_INVALID                                         = errors.New("the log service encountered an invalid log restart area")
	ERROR_LOG_BLOCK_VERSION                                           = errors.New("the log service encountered an invalid log block version")
	ERROR_LOG_BLOCK_INVALID                                           = errors.New("the log service encountered an invalid log block")
	ERROR_LOG_READ_MODE_INVALID                                       = errors.New("the log service encountered an attempt to read the log with an invalid read mode")
	ERROR_LOG_METADATA_CORRUPT                                        = errors.New("the log service encountered a corrupted metadata file")
	ERROR_LOG_METADATA_INVALID                                        = errors.New("the log service encountered a metadata file that could not be created by the log file system")
	ERROR_LOG_METADATA_INCONSISTENT                                   = errors.New("the log service encountered a metadata file with inconsistent data")
	ERROR_LOG_RESERVATION_INVALID                                     = errors.New("the log service encountered an attempt to erroneously allocate or dispose reservation space")
	ERROR_LOG_CANT_DELETE                                             = errors.New("the log service cannot delete the log file or the file system container")
	ERROR_LOG_CONTAINER_LIMIT_EXCEEDED                                = errors.New("the log service has reached the maximum allowable containers allocated to a log file")
	ERROR_LOG_START_OF_LOG                                            = errors.New("the log service has attempted to read or write backward past the start of the log")
	ERROR_LOG_POLICY_ALREADY_INSTALLED                                = errors.New("the log policy could not be installed because a policy of the same type is already present")
	ERROR_LOG_POLICY_NOT_INSTALLED                                    = errors.New("the log policy in question was not installed at the time of the request")
	ERROR_LOG_POLICY_INVALID                                          = errors.New("the installed set of policies on the log is invalid")
	ERROR_LOG_POLICY_CONFLICT                                         = errors.New("a policy on the log in question prevented the operation from completing")
	ERROR_LOG_PINNED_ARCHIVE_TAIL                                     = errors.New("the log space cannot be reclaimed because the log is pinned by the archive tail")
	ERROR_LOG_RECORD_NONEXISTENT                                      = errors.New("the log record is not a record in the log file")
	ERROR_LOG_RECORDS_RESERVED_INVALID                                = errors.New("the number of reserved log records or the adjustment of the number of reserved log records is invalid")
	ERROR_LOG_SPACE_RESERVED_INVALID                                  = errors.New("the reserved log space or the adjustment of the log space is invalid")
	ERROR_LOG_TAIL_INVALID                                            = errors.New("a new or existing archive tail or the base of the active log is invalid")
	ERROR_LOG_FULL                                                    = errors.New("the log space is exhausted")
	ERROR_LOG_MULTIPLEXED                                             = errors.New("the log is multiplexed; no direct writes to the physical log are allowed")
	ERROR_LOG_DEDICATED                                               = errors.New("the operation failed because the log is dedicated")
	ERROR_LOG_ARCHIVE_NOT_IN_PROGRESS                                 = errors.New("the operation requires an archive context")
	ERROR_LOG_ARCHIVE_IN_PROGRESS                                     = errors.New("log archival is in progress")
	ERROR_LOG_EPHEMERAL                                               = errors.New("the operation requires a nonephemeral log, but the log is ephemeral")
	ERROR_LOG_NOT_ENOUGH_CONTAINERS                                   = errors.New("the log must have at least two containers before it can be read from or written to")
	ERROR_LOG_CLIENT_ALREADY_REGISTERED                               = errors.New("a log client has already registered on the stream")
	ERROR_LOG_CLIENT_NOT_REGISTERED                                   = errors.New("a log client has not been registered on the stream")
	ERROR_LOG_FULL_HANDLER_IN_PROGRESS                                = errors.New("a request has already been made to handle the log full condition")
	ERROR_LOG_CONTAINER_READ_FAILED                                   = errors.New("the log service encountered an error when attempting to read from a log container")
	ERROR_LOG_CONTAINER_WRITE_FAILED                                  = errors.New("the log service encountered an error when attempting to write to a log container")
	ERROR_LOG_CONTAINER_OPEN_FAILED                                   = errors.New("the log service encountered an error when attempting to open a log container")
	ERROR_LOG_CONTAINER_STATE_INVALID                                 = errors.New("the log service encountered an invalid container state when attempting a requested action")
	ERROR_LOG_STATE_INVALID                                           = errors.New("the log service is not in the correct state to perform a requested action")
	ERROR_LOG_PINNED                                                  = errors.New("the log space cannot be reclaimed because the log is pinned")
	ERROR_LOG_METADATA_FLUSH_FAILED                                   = errors.New("the log metadata flush failed")
	ERROR_LOG_INCONSISTENT_SECURITY                                   = errors.New("security on the log and its containers is inconsistent")
	ERROR_LOG_APPENDED_FLUSH_FAILED                                   = errors.New("records were appended to the log or reservation changes were made, but the log could not be flushed")
	ERROR_LOG_PINNED_RESERVATION                                      = errors.New("the log is pinned due to reservation consuming most of the log space. free some reserved records to make space available")
	ERROR_VIDEO_HUNG_DISPLAY_DRIVER_THREAD                            = errors.New("{display driver stopped responding} the %hs display driver has stopped working normally. save your work and reboot the system to restore full display functionality. the next time you reboot the computer, a dialog box will allow you to upload data about this failure to microsoft")
	ERROR_FLT_NO_HANDLER_DEFINED                                      = errors.New("a handler was not defined by the filter for this operation")
	ERROR_FLT_CONTEXT_ALREADY_DEFINED                                 = errors.New("a context is already defined for this object")
	ERROR_FLT_INVALID_ASYNCHRONOUS_REQUEST                            = errors.New("asynchronous requests are not valid for this operation")
	ERROR_FLT_DISALLOW_FAST_IO                                        = errors.New("this is an internal error code used by the filter manager to determine if a fast i/o operation should be forced down the input/output request packet (irp) path. minifilters should never return this value")
	ERROR_FLT_INVALID_NAME_REQUEST                                    = errors.New("an invalid name request was made. the name requested cannot be retrieved at this time")
	ERROR_FLT_NOT_SAFE_TO_POST_OPERATION                              = errors.New("posting this operation to a worker thread for further processing is not safe at this time because it could lead to a system deadlock")
	ERROR_FLT_NOT_INITIALIZED                                         = errors.New("the filter manager was not initialized when a filter tried to register. make sure that the filter manager is loaded as a driver")
	ERROR_FLT_FILTER_NOT_READY                                        = errors.New("the filter is not ready for attachment to volumes because it has not finished initializing (fltstartfiltering has not been called)")
	ERROR_FLT_POST_OPERATION_CLEANUP                                  = errors.New("the filter must clean up any operation-specific context at this time because it is being removed from the system before the operation is completed by the lower drivers")
	ERROR_FLT_INTERNAL_ERROR                                          = errors.New("the filter manager had an internal error from which it cannot recover; therefore, the operation has failed. this is usually the result of a filter returning an invalid value from a pre-operation callback")
	ERROR_FLT_DELETING_OBJECT                                         = errors.New("the object specified for this action is in the process of being deleted; therefore, the action requested cannot be completed at this time")
	ERROR_FLT_MUST_BE_NONPAGED_POOL                                   = errors.New("a nonpaged pool must be used for this type of context")
	ERROR_FLT_DUPLICATE_ENTRY                                         = errors.New("a duplicate handler definition has been provided for an operation")
	ERROR_FLT_CBDQ_DISABLED                                           = errors.New("the callback data queue has been disabled")
	ERROR_FLT_DO_NOT_ATTACH                                           = errors.New("do not attach the filter to the volume at this time")
	ERROR_FLT_DO_NOT_DETACH                                           = errors.New("do not detach the filter from the volume at this time")
	ERROR_FLT_INSTANCE_ALTITUDE_COLLISION                             = errors.New("an instance already exists at this altitude on the volume specified")
	ERROR_FLT_INSTANCE_NAME_COLLISION                                 = errors.New("an instance already exists with this name on the volume specified")
	ERROR_FLT_FILTER_NOT_FOUND                                        = errors.New("the system could not find the filter specified")
	ERROR_FLT_VOLUME_NOT_FOUND                                        = errors.New("the system could not find the volume specified")
	ERROR_FLT_INSTANCE_NOT_FOUND                                      = errors.New("the system could not find the instance specified")
	ERROR_FLT_CONTEXT_ALLOCATION_NOT_FOUND                            = errors.New("no registered context allocation definition was found for the given request")
	ERROR_FLT_INVALID_CONTEXT_REGISTRATION                            = errors.New("an invalid parameter was specified during context registration")
	ERROR_FLT_NAME_CACHE_MISS                                         = errors.New("the name requested was not found in the filter manager name cache and could not be retrieved from the file system")
	ERROR_FLT_NO_DEVICE_OBJECT                                        = errors.New("the requested device object does not exist for the given volume")
	ERROR_FLT_VOLUME_ALREADY_MOUNTED                                  = errors.New("the specified volume is already mounted")
	ERROR_FLT_ALREADY_ENLISTED                                        = errors.New("the specified transaction context is already enlisted in a transaction")
	ERROR_FLT_CONTEXT_ALREADY_LINKED                                  = errors.New("the specified context is already attached to another object")
	ERROR_FLT_NO_WAITER_FOR_REPLY                                     = errors.New("no waiter is present for the filter's reply to this message")
	ERROR_MONITOR_NO_DESCRIPTOR                                       = errors.New("a monitor descriptor could not be obtained")
	ERROR_MONITOR_UNKNOWN_DESCRIPTOR_FORMAT                           = errors.New("this release does not support the format of the obtained monitor descriptor")
	ERROR_MONITOR_INVALID_DESCRIPTOR_CHECKSUM                         = errors.New("the checksum of the obtained monitor descriptor is invalid")
	ERROR_MONITOR_INVALID_STANDARD_TIMING_BLOCK                       = errors.New("the monitor descriptor contains an invalid standard timing block")
	ERROR_MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED                   = errors.New("wmi data-block registration failed for one of the msmonitorclass wmi subclasses")
	ERROR_MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK                  = errors.New("the provided monitor descriptor block is either corrupted or does not contain the monitor's detailed serial number")
	ERROR_MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK                  = errors.New("the provided monitor descriptor block is either corrupted or does not contain the monitor's user-friendly name")
	ERROR_MONITOR_NO_MORE_DESCRIPTOR_DATA                             = errors.New("there is no monitor descriptor data at the specified (offset or size) region")
	ERROR_MONITOR_INVALID_DETAILED_TIMING_BLOCK                       = errors.New("the monitor descriptor contains an invalid detailed timing block")
	ERROR_MONITOR_INVALID_MANUFACTURE_DATE                            = errors.New("monitor descriptor contains invalid manufacture date")
	ERROR_GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER                           = errors.New("exclusive mode ownership is needed to create an unmanaged primary allocation")
	ERROR_GRAPHICS_INSUFFICIENT_DMA_BUFFER                            = errors.New("the driver needs more dma buffer space to complete the requested operation")
	ERROR_GRAPHICS_INVALID_DISPLAY_ADAPTER                            = errors.New("the specified display adapter handle is invalid")
	ERROR_GRAPHICS_ADAPTER_WAS_RESET                                  = errors.New("the specified display adapter and all of its state have been reset")
	ERROR_GRAPHICS_INVALID_DRIVER_MODEL                               = errors.New("the driver stack does not match the expected driver model")
	ERROR_GRAPHICS_PRESENT_MODE_CHANGED                               = errors.New("present happened but ended up into the changed desktop mode")
	ERROR_GRAPHICS_PRESENT_OCCLUDED                                   = errors.New("nothing to present due to desktop occlusion")
	ERROR_GRAPHICS_PRESENT_DENIED                                     = errors.New("not able to present due to denial of desktop access")
	ERROR_GRAPHICS_CANNOTCOLORCONVERT                                 = errors.New("not able to present with color conversion")
	ERROR_GRAPHICS_PRESENT_REDIRECTION_DISABLED                       = errors.New("present redirection is disabled (desktop windowing management subsystem is off)")
	ERROR_GRAPHICS_PRESENT_UNOCCLUDED                                 = errors.New("previous exclusive vidpn source owner has released its ownership")
	ERROR_GRAPHICS_NO_VIDEO_MEMORY                                    = errors.New("not enough video memory is available to complete the operation")
	ERROR_GRAPHICS_CANT_LOCK_MEMORY                                   = errors.New("could not probe and lock the underlying memory of an allocation")
	ERROR_GRAPHICS_ALLOCATION_BUSY                                    = errors.New("the allocation is currently busy")
	ERROR_GRAPHICS_TOO_MANY_REFERENCES                                = errors.New("an object being referenced has already reached the maximum reference count and cannot be referenced further")
	ERROR_GRAPHICS_TRY_AGAIN_LATER                                    = errors.New("a problem could not be solved due to an existing condition. try again later")
	ERROR_GRAPHICS_TRY_AGAIN_NOW                                      = errors.New("a problem could not be solved due to an existing condition. try again now")
	ERROR_GRAPHICS_ALLOCATION_INVALID                                 = errors.New("the allocation is invalid")
	ERROR_GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE                   = errors.New("no more unswizzling apertures are currently available")
	ERROR_GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED                   = errors.New("the current allocation cannot be unswizzled by an aperture")
	ERROR_GRAPHICS_CANT_EVICT_PINNED_ALLOCATION                       = errors.New("the request failed because a pinned allocation cannot be evicted")
	ERROR_GRAPHICS_INVALID_ALLOCATION_USAGE                           = errors.New("the allocation cannot be used from its current segment location for the specified operation")
	ERROR_GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION                      = errors.New("a locked allocation cannot be used in the current command buffer")
	ERROR_GRAPHICS_ALLOCATION_CLOSED                                  = errors.New("the allocation being referenced has been closed permanently")
	ERROR_GRAPHICS_INVALID_ALLOCATION_INSTANCE                        = errors.New("an invalid allocation instance is being referenced")
	ERROR_GRAPHICS_INVALID_ALLOCATION_HANDLE                          = errors.New("an invalid allocation handle is being referenced")
	ERROR_GRAPHICS_WRONG_ALLOCATION_DEVICE                            = errors.New("the allocation being referenced does not belong to the current device")
	ERROR_GRAPHICS_ALLOCATION_CONTENT_LOST                            = errors.New("the specified allocation lost its content")
	ERROR_GRAPHICS_GPU_EXCEPTION_ON_DEVICE                            = errors.New("a gpu exception was detected on the given device. the device cannot be scheduled")
	ERROR_GRAPHICS_INVALID_VIDPN_TOPOLOGY                             = errors.New("the specified vidpn topology is invalid")
	ERROR_GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED                       = errors.New("the specified vidpn topology is valid but is not supported by this model of the display adapter")
	ERROR_GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED             = errors.New("the specified vidpn topology is valid but is not currently supported by the display adapter due to allocation of its resources")
	ERROR_GRAPHICS_INVALID_VIDPN                                      = errors.New("the specified vidpn handle is invalid")
	ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE                       = errors.New("the specified video present source is invalid")
	ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET                       = errors.New("the specified video present target is invalid")
	ERROR_GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED                       = errors.New("the specified vidpn modality is not supported (for example, at least two of the pinned modes are not co-functional)")
	ERROR_GRAPHICS_INVALID_VIDPN_SOURCEMODESET                        = errors.New("the specified vidpn source mode set is invalid")
	ERROR_GRAPHICS_INVALID_VIDPN_TARGETMODESET                        = errors.New("the specified vidpn target mode set is invalid")
	ERROR_GRAPHICS_INVALID_FREQUENCY                                  = errors.New("the specified video signal frequency is invalid")
	ERROR_GRAPHICS_INVALID_ACTIVE_REGION                              = errors.New("the specified video signal active region is invalid")
	ERROR_GRAPHICS_INVALID_TOTAL_REGION                               = errors.New("the specified video signal total region is invalid")
	ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE                  = errors.New("the specified video present source mode is invalid")
	ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE                  = errors.New("the specified video present target mode is invalid")
	ERROR_GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET                     = errors.New("the pinned mode must remain in the set on the vidpn's co-functional modality enumeration")
	ERROR_GRAPHICS_PATH_ALREADY_IN_TOPOLOGY                           = errors.New("the specified video present path is already in the vidpn's topology")
	ERROR_GRAPHICS_MODE_ALREADY_IN_MODESET                            = errors.New("the specified mode is already in the mode set")
	ERROR_GRAPHICS_INVALID_VIDEOPRESENTSOURCESET                      = errors.New("the specified video present source set is invalid")
	ERROR_GRAPHICS_INVALID_VIDEOPRESENTTARGETSET                      = errors.New("the specified video present target set is invalid")
	ERROR_GRAPHICS_SOURCE_ALREADY_IN_SET                              = errors.New("the specified video present source is already in the video present source set")
	ERROR_GRAPHICS_TARGET_ALREADY_IN_SET                              = errors.New("the specified video present target is already in the video present target set")
	ERROR_GRAPHICS_INVALID_VIDPN_PRESENT_PATH                         = errors.New("the specified vidpn present path is invalid")
	ERROR_GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY                      = errors.New("the miniport has no recommendation for augmenting the specified vidpn's topology")
	ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET                  = errors.New("the specified monitor frequency range set is invalid")
	ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE                     = errors.New("the specified monitor frequency range is invalid")
	ERROR_GRAPHICS_FREQUENCYRANGE_NOT_IN_SET                          = errors.New("the specified frequency range is not in the specified monitor frequency range set")
	ERROR_GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET                      = errors.New("the specified frequency range is already in the specified monitor frequency range set")
	ERROR_GRAPHICS_STALE_MODESET                                      = errors.New("the specified mode set is stale. reacquire the new mode set")
	ERROR_GRAPHICS_INVALID_MONITOR_SOURCEMODESET                      = errors.New("the specified monitor source mode set is invalid")
	ERROR_GRAPHICS_INVALID_MONITOR_SOURCE_MODE                        = errors.New("the specified monitor source mode is invalid")
	ERROR_GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN                    = errors.New("the miniport does not have a recommendation regarding the request to provide a functional vidpn given the current display adapter configuration")
	ERROR_GRAPHICS_MODE_ID_MUST_BE_UNIQUE                             = errors.New("the id of the specified mode is being used by another mode in the set")
	ERROR_GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION    = errors.New("the system failed to determine a mode that is supported by both the display adapter and the monitor connected to it")
	ERROR_GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES            = errors.New("the number of video present targets must be greater than or equal to the number of video present sources")
	ERROR_GRAPHICS_PATH_NOT_IN_TOPOLOGY                               = errors.New("the specified present path is not in the vidpn's topology")
	ERROR_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE              = errors.New("the display adapter must have at least one video present source")
	ERROR_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET              = errors.New("the display adapter must have at least one video present target")
	ERROR_GRAPHICS_INVALID_MONITORDESCRIPTORSET                       = errors.New("the specified monitor descriptor set is invalid")
	ERROR_GRAPHICS_INVALID_MONITORDESCRIPTOR                          = errors.New("the specified monitor descriptor is invalid")
	ERROR_GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET                       = errors.New("the specified descriptor is not in the specified monitor descriptor set")
	ERROR_GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET                   = errors.New("the specified descriptor is already in the specified monitor descriptor set")
	ERROR_GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE                = errors.New("the id of the specified monitor descriptor is being used by another descriptor in the set")
	ERROR_GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE                   = errors.New("the specified video present target subset type is invalid")
	ERROR_GRAPHICS_RESOURCES_NOT_RELATED                              = errors.New("two or more of the specified resources are not related to each other, as defined by the interface semantics")
	ERROR_GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE                           = errors.New("the id of the specified video present source is being used by another source in the set")
	ERROR_GRAPHICS_TARGET_ID_MUST_BE_UNIQUE                           = errors.New("the id of the specified video present target is being used by another target in the set")
	ERROR_GRAPHICS_NO_AVAILABLE_VIDPN_TARGET                          = errors.New("the specified vidpn source cannot be used because there is no available vidpn target to connect it to")
	ERROR_GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER       = errors.New("the newly arrived monitor could not be associated with a display adapter")
	ERROR_GRAPHICS_NO_VIDPNMGR                                        = errors.New("the particular display adapter does not have an associated vidpn manager")
	ERROR_GRAPHICS_NO_ACTIVE_VIDPN                                    = errors.New("the vidpn manager of the particular display adapter does not have an active vidpn")
	ERROR_GRAPHICS_STALE_VIDPN_TOPOLOGY                               = errors.New("the specified vidpn topology is stale; obtain the new topology")
	ERROR_GRAPHICS_MONITOR_NOT_CONNECTED                              = errors.New("no monitor is connected on the specified video present target")
	ERROR_GRAPHICS_SOURCE_NOT_IN_TOPOLOGY                             = errors.New("the specified source is not part of the specified vidpn's topology")
	ERROR_GRAPHICS_INVALID_PRIMARYSURFACE_SIZE                        = errors.New("the specified primary surface size is invalid")
	ERROR_GRAPHICS_INVALID_VISIBLEREGION_SIZE                         = errors.New("the specified visible region size is invalid")
	ERROR_GRAPHICS_INVALID_STRIDE                                     = errors.New("the specified stride is invalid")
	ERROR_GRAPHICS_INVALID_PIXELFORMAT                                = errors.New("the specified pixel format is invalid")
	ERROR_GRAPHICS_INVALID_COLORBASIS                                 = errors.New("the specified color basis is invalid")
	ERROR_GRAPHICS_INVALID_PIXELVALUEACCESSMODE                       = errors.New("the specified pixel value access mode is invalid")
	ERROR_GRAPHICS_TARGET_NOT_IN_TOPOLOGY                             = errors.New("the specified target is not part of the specified vidpn's topology")
	ERROR_GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT                 = errors.New("failed to acquire the display mode management interface")
	ERROR_GRAPHICS_VIDPN_SOURCE_IN_USE                                = errors.New("the specified vidpn source is already owned by a dmm client and cannot be used until that client releases it")
	ERROR_GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN                           = errors.New("the specified vidpn is active and cannot be accessed")
	ERROR_GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL                    = errors.New("the specified vidpn's present path importance ordinal is invalid")
	ERROR_GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION       = errors.New("the specified vidpn's present path content geometry transformation is invalid")
	ERROR_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED = errors.New("the specified content geometry transformation is not supported on the respective vidpn present path")
	ERROR_GRAPHICS_INVALID_GAMMA_RAMP                                 = errors.New("the specified gamma ramp is invalid")
	ERROR_GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED                           = errors.New("the specified gamma ramp is not supported on the respective vidpn present path")
	ERROR_GRAPHICS_MULTISAMPLING_NOT_SUPPORTED                        = errors.New("multisampling is not supported on the respective vidpn present path")
	ERROR_GRAPHICS_MODE_NOT_IN_MODESET                                = errors.New("the specified mode is not in the specified mode set")
	ERROR_GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON       = errors.New("the specified vidpn topology recommendation reason is invalid")
	ERROR_GRAPHICS_INVALID_PATH_CONTENT_TYPE                          = errors.New("the specified vidpn present path content type is invalid")
	ERROR_GRAPHICS_INVALID_COPYPROTECTION_TYPE                        = errors.New("the specified vidpn present path copy protection type is invalid")
	ERROR_GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS                  = errors.New("only one unassigned mode set can exist at any one time for a particular vidpn source or target")
	ERROR_GRAPHICS_INVALID_SCANLINE_ORDERING                          = errors.New("the specified scan line ordering type is invalid")
	ERROR_GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED                       = errors.New("the topology changes are not allowed for the specified vidpn")
	ERROR_GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS                   = errors.New("all available importance ordinals are being used in the specified topology")
	ERROR_GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT                        = errors.New("the specified primary surface has a different private-format attribute than the current primary surface")
	ERROR_GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM                     = errors.New("the specified mode-pruning algorithm is invalid")
	ERROR_GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN                  = errors.New("the specified monitor-capability origin is invalid")
	ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT          = errors.New("the specified monitor-frequency range constraint is invalid")
	ERROR_GRAPHICS_MAX_NUM_PATHS_REACHED                              = errors.New("the maximum supported number of present paths has been reached")
	ERROR_GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION                 = errors.New("the miniport requested that augmentation be canceled for the specified source of the specified vidpn's topology")
	ERROR_GRAPHICS_INVALID_CLIENT_TYPE                                = errors.New("the specified client type was not recognized")
	ERROR_GRAPHICS_CLIENTVIDPN_NOT_SET                                = errors.New("the client vidpn is not set on this adapter (for example, no user mode-initiated mode changes have taken place on this adapter)")
	ERROR_GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED                  = errors.New("the specified display adapter child device already has an external device connected to it")
	ERROR_GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED                     = errors.New("the display adapter child device does not support reporting a descriptor")
	ERROR_GRAPHICS_NOT_A_LINKED_ADAPTER                               = errors.New("the display adapter is not linked to any other adapters")
	ERROR_GRAPHICS_LEADLINK_NOT_ENUMERATED                            = errors.New("the lead adapter in a linked configuration was not enumerated yet")
	ERROR_GRAPHICS_CHAINLINKS_NOT_ENUMERATED                          = errors.New("some chain adapters in a linked configuration have not yet been enumerated")
	ERROR_GRAPHICS_ADAPTER_CHAIN_NOT_READY                            = errors.New("the chain of linked adapters is not ready to start because of an unknown failure")
	ERROR_GRAPHICS_CHAINLINKS_NOT_STARTED                             = errors.New("an attempt was made to start a lead link display adapter when the chain links had not yet started")
	ERROR_GRAPHICS_CHAINLINKS_NOT_POWERED_ON                          = errors.New("an attempt was made to turn on a lead link display adapter when the chain links were turned off")
	ERROR_GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE                     = errors.New("the adapter link was found in an inconsistent state. not all adapters are in an expected pnp/power state")
	ERROR_GRAPHICS_NOT_POST_DEVICE_DRIVER                             = errors.New("the driver trying to start is not the same as the driver for the posted display adapter")
	ERROR_GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED                        = errors.New("an operation is being attempted that requires the display adapter to be in a quiescent state")
	ERROR_GRAPHICS_OPM_NOT_SUPPORTED                                  = errors.New("the driver does not support opm")
	ERROR_GRAPHICS_COPP_NOT_SUPPORTED                                 = errors.New("the driver does not support copp")
	ERROR_GRAPHICS_UAB_NOT_SUPPORTED                                  = errors.New("the driver does not support uab")
	ERROR_GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS                   = errors.New("the specified encrypted parameters are invalid")
	ERROR_GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL                      = errors.New("an array passed to a function cannot hold all of the data that the function wants to put in it")
	ERROR_GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST                     = errors.New("the gdi display device passed to this function does not have any active protected outputs")
	ERROR_GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME          = errors.New("the pvp cannot find an actual gdi display device that corresponds to the passed-in gdi display device name")
	ERROR_GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP         = errors.New("this function failed because the gdi display device passed to it was not attached to the windows desktop")
	ERROR_GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED                = errors.New("the pvp does not support mirroring display devices because they do not have any protected outputs")
	ERROR_GRAPHICS_OPM_INVALID_POINTER                                = errors.New("the function failed because an invalid pointer parameter was passed to it. a pointer parameter is invalid if it is null, is not correctly aligned, or it points to an invalid address or a kernel mode address")
	ERROR_GRAPHICS_OPM_INTERNAL_ERROR                                 = errors.New("an internal error caused an operation to fail")
	ERROR_GRAPHICS_OPM_INVALID_HANDLE                                 = errors.New("the function failed because the caller passed in an invalid opm user-mode handle")
	ERROR_GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE       = errors.New("this function failed because the gdi device passed to it did not have any monitors associated with it")
	ERROR_GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH                     = errors.New("a certificate could not be returned because the certificate buffer passed to the function was too small")
	ERROR_GRAPHICS_OPM_SPANNING_MODE_ENABLED                          = errors.New("dxgkddiopmcreateprotectedoutput() could not create a protected output because the video present yarget is in spanning mode")
	ERROR_GRAPHICS_OPM_THEATER_MODE_ENABLED                           = errors.New("dxgkddiopmcreateprotectedoutput() could not create a protected output because the video present target is in theater mode")
	ERROR_GRAPHICS_PVP_HFS_FAILED                                     = errors.New("the function call failed because the display adapter's hardware functionality scan (hfs) failed to validate the graphics hardware")
	ERROR_GRAPHICS_OPM_INVALID_SRM                                    = errors.New("the hdcp srm passed to this function did not comply with section 5 of the hdcp 1.1 specification")
	ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP                   = errors.New("the protected output cannot enable the hdcp system because it does not support it")
	ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP                    = errors.New("the protected output cannot enable analog copy protection because it does not support it")
	ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA                  = errors.New("the protected output cannot enable the cgms-a protection technology because it does not support it")
	ERROR_GRAPHICS_OPM_HDCP_SRM_NEVER_SET                             = errors.New("dxgkddiopmgetinformation() cannot return the version of the srm being used because the application never successfully passed an srm to the protected output")
	ERROR_GRAPHICS_OPM_RESOLUTION_TOO_HIGH                            = errors.New("dxgkddiopmconfigureprotectedoutput() cannot enable the specified output protection technology because the output's screen resolution is too high")
	ERROR_GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE               = errors.New("dxgkddiopmconfigureprotectedoutput() cannot enable hdcp because other physical outputs are using the display adapter's hdcp hardware")
	ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS              = errors.New("the operating system asynchronously destroyed this opm-protected output because the operating system state changed. this error typically occurs because the monitor pdo associated with this protected output was removed or stopped, the protected output's session became a nonconsole session, or the protected output's desktop became inactive")
	ERROR_GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS                = errors.New("opm functions cannot be called when a session is changing its type. three types of sessions currently exist: console, disconnected, and remote (rdp or ica)")
	ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS  = errors.New("the dxgkddiopmgetcoppcompatibleinformation, dxgkddiopmgetinformation, or dxgkddiopmconfigureprotectedoutput function failed. this error is returned only if a protected output has opm semantics")
	ERROR_GRAPHICS_OPM_INVALID_INFORMATION_REQUEST                    = errors.New("the dxgkddiopmgetinformation and dxgkddiopmgetcoppcompatibleinformation functions return this error code if the passed-in sequence number is not the expected sequence number or the passed-in omac value is invalid")
	ERROR_GRAPHICS_OPM_DRIVER_INTERNAL_ERROR                          = errors.New("the function failed because an unexpected error occurred inside a display driver")
	ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS   = errors.New("the dxgkddiopmgetcoppcompatibleinformation, dxgkddiopmgetinformation, or dxgkddiopmconfigureprotectedoutput function failed. this error is returned only if a protected output has copp semantics")
	ERROR_GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED                        = errors.New("the dxgkddiopmgetcoppcompatibleinformation and dxgkddiopmconfigureprotectedoutput functions return this error if the display driver does not support the dxgkmdt_opm_get_acp_and_cgmsa_signaling and dxgkmdt_opm_set_acp_and_cgmsa_signaling guids")
	ERROR_GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST                  = errors.New("the dxgkddiopmconfigureprotectedoutput function returns this error code if the passed-in sequence number is not the expected sequence number or the passed-in omac value is invalid")
	ERROR_GRAPHICS_I2C_NOT_SUPPORTED                                  = errors.New("the monitor connected to the specified video output does not have an i2c bus")
	ERROR_GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST                          = errors.New("no device on the i2c bus has the specified address")
	ERROR_GRAPHICS_I2C_ERROR_TRANSMITTING_DATA                        = errors.New("an error occurred while transmitting data to the device on the i2c bus")
	ERROR_GRAPHICS_I2C_ERROR_RECEIVING_DATA                           = errors.New("an error occurred while receiving data from the device on the i2c bus")
	ERROR_GRAPHICS_DDCCI_VCP_NOT_SUPPORTED                            = errors.New("the monitor does not support the specified vcp code")
	ERROR_GRAPHICS_DDCCI_INVALID_DATA                                 = errors.New("the data received from the monitor is invalid")
	ERROR_GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE  = errors.New("a function call failed because a monitor returned an invalid timing status byte when the operating system used the ddc/ci get timing report and timing message command to get a timing report from a monitor")
	ERROR_GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING                  = errors.New("a monitor returned a ddc/ci capabilities string that did not comply with the access.bus 3.0, ddc/ci 1.1, or mccs 2 revision 1 specification")
	ERROR_GRAPHICS_MCA_INTERNAL_ERROR                                 = errors.New("an internal error caused an operation to fail")
	ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND                      = errors.New("an operation failed because a ddc/ci message had an invalid value in its command field")
	ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH                       = errors.New("this error occurred because a ddc/ci message had an invalid value in its length field")
	ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM                     = errors.New("this error occurred because the value in a ddc/ci message's checksum field did not match the message's computed checksum value. this error implies that the data was corrupted while it was being transmitted from a monitor to a computer")
	ERROR_GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE                    = errors.New("this function failed because an invalid monitor handle was passed to it")
	ERROR_GRAPHICS_MONITOR_NO_LONGER_EXISTS                           = errors.New("the operating system asynchronously destroyed the monitor that corresponds to this handle because the operating system's state changed. this error typically occurs because the monitor pdo associated with this handle was removed or stopped, or a display mode change occurred. a display mode change occurs when windows sends a wm_displaychange message to applications")
	ERROR_GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED                     = errors.New("this function can be used only if a program is running in the local console session. it cannot be used if a program is running on a remote desktop session or on a terminal server session")
	ERROR_GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME              = errors.New("this function cannot find an actual gdi display device that corresponds to the specified gdi display device name")
	ERROR_GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP             = errors.New("the function failed because the specified gdi display device was not attached to the windows desktop")
	ERROR_GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED                    = errors.New("this function does not support gdi mirroring display devices because gdi mirroring display devices do not have any physical monitors associated with them")
	ERROR_GRAPHICS_INVALID_POINTER                                    = errors.New("the function failed because an invalid pointer parameter was passed to it. a pointer parameter is invalid if it is null, is not correctly aligned, or points to an invalid address or to a kernel mode address")
	ERROR_GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE           = errors.New("this function failed because the gdi device passed to it did not have a monitor associated with it")
	ERROR_GRAPHICS_PARAMETER_ARRAY_TOO_SMALL                          = errors.New("an array passed to the function cannot hold all of the data that the function must copy into the array")
	ERROR_GRAPHICS_INTERNAL_ERROR                                     = errors.New("an internal error caused an operation to fail")
	ERROR_GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS                    = errors.New("the function failed because the current session is changing its type. this function cannot be called when the current session is changing its type. three types of sessions currently exist: console, disconnected, and remote (rdp or ica)")
	ERROR_FVE_LOCKED_VOLUME                                           = errors.New("the volume must be unlocked before it can be used")
	ERROR_FVE_NOT_ENCRYPTED                                           = errors.New("the volume is fully decrypted and no key is available")
	ERROR_FVE_BAD_INFORMATION                                         = errors.New("the control block for the encrypted volume is not valid")
	ERROR_FVE_TOO_SMALL                                               = errors.New("not enough free space remains on the volume to allow encryption")
	ERROR_FVE_FAILED_WRONG_FS                                         = errors.New("the partition cannot be encrypted because the file system is not supported")
	ERROR_FVE_FAILED_BAD_FS                                           = errors.New("the file system is inconsistent. run the check disk utility")
	ERROR_FVE_FS_NOT_EXTENDED                                         = errors.New("the file system does not extend to the end of the volume")
	ERROR_FVE_FS_MOUNTED                                              = errors.New("this operation cannot be performed while a file system is mounted on the volume")
	ERROR_FVE_NO_LICENSE                                              = errors.New("bitlocker drive encryption is not included with this version of windows")
	ERROR_FVE_ACTION_NOT_ALLOWED                                      = errors.New("the requested action was denied by the fve control engine")
	ERROR_FVE_BAD_DATA                                                = errors.New("the data supplied is malformed")
	ERROR_FVE_VOLUME_NOT_BOUND                                        = errors.New("the volume is not bound to the system")
	ERROR_FVE_NOT_DATA_VOLUME                                         = errors.New("the volume specified is not a data volume")
	ERROR_FVE_CONV_READ_ERROR                                         = errors.New("a read operation failed while converting the volume")
	ERROR_FVE_CONV_WRITE_ERROR                                        = errors.New("a write operation failed while converting the volume")
	ERROR_FVE_OVERLAPPED_UPDATE                                       = errors.New("the control block for the encrypted volume was updated by another thread. try again")
	ERROR_FVE_FAILED_SECTOR_SIZE                                      = errors.New("the volume encryption algorithm cannot be used on this sector size")
	ERROR_FVE_FAILED_AUTHENTICATION                                   = errors.New("bitlocker recovery authentication failed")
	ERROR_FVE_NOT_OS_VOLUME                                           = errors.New("the volume specified is not the boot operating system volume")
	ERROR_FVE_KEYFILE_NOT_FOUND                                       = errors.New("the bitlocker startup key or recovery password could not be read from external media")
	ERROR_FVE_KEYFILE_INVALID                                         = errors.New("the bitlocker startup key or recovery password file is corrupt or invalid")
	ERROR_FVE_KEYFILE_NO_VMK                                          = errors.New("the bitlocker encryption key could not be obtained from the startup key or the recovery password")
	ERROR_FVE_TPM_DISABLED                                            = errors.New("the tpm is disabled")
	ERROR_FVE_TPM_SRK_AUTH_NOT_ZERO                                   = errors.New("the authorization data for the srk of the tpm is not zero")
	ERROR_FVE_TPM_INVALID_PCR                                         = errors.New("the system boot information changed or the tpm locked out access to bitlocker encryption keys until the computer is restarted")
	ERROR_FVE_TPM_NO_VMK                                              = errors.New("the bitlocker encryption key could not be obtained from the tpm")
	ERROR_FVE_PIN_INVALID                                             = errors.New("the bitlocker encryption key could not be obtained from the tpm and pin")
	ERROR_FVE_AUTH_INVALID_APPLICATION                                = errors.New("a boot application hash does not match the hash computed when bitlocker was turned on")
	ERROR_FVE_AUTH_INVALID_CONFIG                                     = errors.New("the boot configuration data (bcd) settings are not supported or have changed because bitlocker was enabled")
	ERROR_FVE_DEBUGGER_ENABLED                                        = errors.New("boot debugging is enabled. run windows boot configuration data store editor (bcdedit.exe) to turn it off")
	ERROR_FVE_DRY_RUN_FAILED                                          = errors.New("the bitlocker encryption key could not be obtained")
	ERROR_FVE_BAD_METADATA_POINTER                                    = errors.New("the metadata disk region pointer is incorrect")
	ERROR_FVE_OLD_METADATA_COPY                                       = errors.New("the backup copy of the metadata is out of date")
	ERROR_FVE_REBOOT_REQUIRED                                         = errors.New("no action was taken because a system restart is required")
	ERROR_FVE_RAW_ACCESS                                              = errors.New("no action was taken because bitlocker drive encryption is in raw access mode")
	ERROR_FVE_RAW_BLOCKED                                             = errors.New("bitlocker drive encryption cannot enter raw access mode for this volume")
	ERROR_FVE_NO_FEATURE_LICENSE                                      = errors.New("this feature of bitlocker drive encryption is not included with this version of windows")
	ERROR_FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED                     = errors.New("group policy does not permit turning off bitlocker drive encryption on roaming data volumes")
	ERROR_FVE_CONV_RECOVERY_FAILED                                    = errors.New("bitlocker drive encryption failed to recover from aborted conversion. this could be due to either all conversion logs being corrupted or the media being write-protected")
	ERROR_FVE_VIRTUALIZED_SPACE_TOO_BIG                               = errors.New("the requested virtualization size is too big")
	ERROR_FVE_VOLUME_TOO_SMALL                                        = errors.New("the drive is too small to be protected using bitlocker drive encryption")
	ERROR_FWP_CALLOUT_NOT_FOUND                                       = errors.New("the callout does not exist")
	ERROR_FWP_CONDITION_NOT_FOUND                                     = errors.New("the filter condition does not exist")
	ERROR_FWP_FILTER_NOT_FOUND                                        = errors.New("the filter does not exist")
	ERROR_FWP_LAYER_NOT_FOUND                                         = errors.New("the layer does not exist")
	ERROR_FWP_PROVIDER_NOT_FOUND                                      = errors.New("the provider does not exist")
	ERROR_FWP_PROVIDER_CONTEXT_NOT_FOUND                              = errors.New("the provider context does not exist")
	ERROR_FWP_SUBLAYER_NOT_FOUND                                      = errors.New("the sublayer does not exist")
	ERROR_FWP_NOT_FOUND                                               = errors.New("the object does not exist")
	ERROR_FWP_ALREADY_EXISTS                                          = errors.New("an object with that guid or luid already exists")
	ERROR_FWP_IN_USE                                                  = errors.New("the object is referenced by other objects and cannot be deleted")
	ERROR_FWP_DYNAMIC_SESSION_IN_PROGRESS                             = errors.New("the call is not allowed from within a dynamic session")
	ERROR_FWP_WRONG_SESSION                                           = errors.New("the call was made from the wrong session and cannot be completed")
	ERROR_FWP_NO_TXN_IN_PROGRESS                                      = errors.New("the call must be made from within an explicit transaction")
	ERROR_FWP_TXN_IN_PROGRESS                                         = errors.New("the call is not allowed from within an explicit transaction")
	ERROR_FWP_TXN_ABORTED                                             = errors.New("the explicit transaction has been forcibly canceled")
	ERROR_FWP_SESSION_ABORTED                                         = errors.New("the session has been canceled")
	ERROR_FWP_INCOMPATIBLE_TXN                                        = errors.New("the call is not allowed from within a read-only transaction")
	ERROR_FWP_TIMEOUT                                                 = errors.New("the call timed out while waiting to acquire the transaction lock")
	ERROR_FWP_NET_EVENTS_DISABLED                                     = errors.New("the collection of network diagnostic events is disabled")
	ERROR_FWP_INCOMPATIBLE_LAYER                                      = errors.New("the operation is not supported by the specified layer")
	ERROR_FWP_KM_CLIENTS_ONLY                                         = errors.New("the call is allowed for kernel-mode callers only")
	ERROR_FWP_LIFETIME_MISMATCH                                       = errors.New("the call tried to associate two objects with incompatible lifetimes")
	ERROR_FWP_BUILTIN_OBJECT                                          = errors.New("the object is built-in and cannot be deleted")
	ERROR_FWP_TOO_MANY_BOOTTIME_FILTERS                               = errors.New("the maximum number of boot-time filters has been reached")
	// NT_STATUS_FWP_TOO_MANY_CALLOUTS = errors.New("the maximum number of callouts has been reached")
	ERROR_FWP_NOTIFICATION_DROPPED               = errors.New("a notification could not be delivered because a message queue has reached maximum capacity")
	ERROR_FWP_TRAFFIC_MISMATCH                   = errors.New("the traffic parameters do not match those for the security association context")
	ERROR_FWP_INCOMPATIBLE_SA_STATE              = errors.New("the call is not allowed for the current security association state")
	ERROR_FWP_NULL_POINTER                       = errors.New("a required pointer is null")
	ERROR_FWP_INVALID_ENUMERATOR                 = errors.New("an enumerator is not valid")
	ERROR_FWP_INVALID_FLAGS                      = errors.New("the flags field contains an invalid value")
	ERROR_FWP_INVALID_NET_MASK                   = errors.New("a network mask is not valid")
	ERROR_FWP_INVALID_RANGE                      = errors.New("an fwp_range is not valid")
	ERROR_FWP_INVALID_INTERVAL                   = errors.New("the time interval is not valid")
	ERROR_FWP_ZERO_LENGTH_ARRAY                  = errors.New("an array that must contain at least one element has a zero length")
	ERROR_FWP_NULL_DISPLAY_NAME                  = errors.New("the displaydata.name field cannot be null")
	ERROR_FWP_INVALID_ACTION_TYPE                = errors.New("the action type is not one of the allowed action types for a filter")
	ERROR_FWP_INVALID_WEIGHT                     = errors.New("the filter weight is not valid")
	ERROR_FWP_MATCH_TYPE_MISMATCH                = errors.New("a filter condition contains a match type that is not compatible with the operands")
	ERROR_FWP_TYPE_MISMATCH                      = errors.New("an fwp_value or fwpm_condition_value is of the wrong type")
	ERROR_FWP_OUT_OF_BOUNDS                      = errors.New("an integer value is outside the allowed range")
	ERROR_FWP_RESERVED                           = errors.New("a reserved field is nonzero")
	ERROR_FWP_DUPLICATE_CONDITION                = errors.New("a filter cannot contain multiple conditions operating on a single field")
	ERROR_FWP_DUPLICATE_KEYMOD                   = errors.New("a policy cannot contain the same keying module more than once")
	ERROR_FWP_ACTION_INCOMPATIBLE_WITH_LAYER     = errors.New("the action type is not compatible with the layer")
	ERROR_FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER  = errors.New("the action type is not compatible with the sublayer")
	ERROR_FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER    = errors.New("the raw context or the provider context is not compatible with the layer")
	ERROR_FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT  = errors.New("the raw context or the provider context is not compatible with the callout")
	ERROR_FWP_INCOMPATIBLE_AUTH_METHOD           = errors.New("the authentication method is not compatible with the policy type")
	ERROR_FWP_INCOMPATIBLE_DH_GROUP              = errors.New("the diffie-hellman group is not compatible with the policy type")
	ERROR_FWP_EM_NOT_SUPPORTED                   = errors.New("an ike policy cannot contain an extended mode policy")
	ERROR_FWP_NEVER_MATCH                        = errors.New("the enumeration template or subscription will never match any objects")
	ERROR_FWP_PROVIDER_CONTEXT_MISMATCH          = errors.New("the provider context is of the wrong type")
	ERROR_FWP_INVALID_PARAMETER                  = errors.New("the parameter is incorrect")
	ERROR_FWP_TOO_MANY_SUBLAYERS                 = errors.New("the maximum number of sublayers has been reached")
	ERROR_FWP_CALLOUT_NOTIFICATION_FAILED        = errors.New("the notification function for a callout returned an error")
	ERROR_FWP_INCOMPATIBLE_AUTH_CONFIG           = errors.New("the ipsec authentication configuration is not compatible with the authentication type")
	ERROR_FWP_INCOMPATIBLE_CIPHER_CONFIG         = errors.New("the ipsec cipher configuration is not compatible with the cipher type")
	ERROR_FWP_DUPLICATE_AUTH_METHOD              = errors.New("a policy cannot contain the same auth method more than once")
	ERROR_FWP_TCPIP_NOT_READY                    = errors.New("the tcp/ip stack is not ready")
	ERROR_FWP_INJECT_HANDLE_CLOSING              = errors.New("the injection handle is being closed by another thread")
	ERROR_FWP_INJECT_HANDLE_STALE                = errors.New("the injection handle is stale")
	ERROR_FWP_CANNOT_PEND                        = errors.New("the classify cannot be pended")
	ERROR_NDIS_CLOSING                           = errors.New("the binding to the network interface is being closed")
	ERROR_NDIS_BAD_VERSION                       = errors.New("an invalid version was specified")
	ERROR_NDIS_BAD_CHARACTERISTICS               = errors.New("an invalid characteristics table was used")
	ERROR_NDIS_ADAPTER_NOT_FOUND                 = errors.New("failed to find the network interface or the network interface is not ready")
	ERROR_NDIS_OPEN_FAILED                       = errors.New("failed to open the network interface")
	ERROR_NDIS_DEVICE_FAILED                     = errors.New("the network interface has encountered an internal unrecoverable failure")
	ERROR_NDIS_MULTICAST_FULL                    = errors.New("the multicast list on the network interface is full")
	ERROR_NDIS_MULTICAST_EXISTS                  = errors.New("an attempt was made to add a duplicate multicast address to the list")
	ERROR_NDIS_MULTICAST_NOT_FOUND               = errors.New("at attempt was made to remove a multicast address that was never added")
	ERROR_NDIS_REQUEST_ABORTED                   = errors.New("the network interface aborted the request")
	ERROR_NDIS_RESET_IN_PROGRESS                 = errors.New("the network interface cannot process the request because it is being reset")
	ERROR_NDIS_INVALID_PACKET                    = errors.New("an attempt was made to send an invalid packet on a network interface")
	ERROR_NDIS_INVALID_DEVICE_REQUEST            = errors.New("the specified request is not a valid operation for the target device")
	ERROR_NDIS_ADAPTER_NOT_READY                 = errors.New("the network interface is not ready to complete this operation")
	ERROR_NDIS_INVALID_LENGTH                    = errors.New("the length of the buffer submitted for this operation is not valid")
	ERROR_NDIS_INVALID_DATA                      = errors.New("the data used for this operation is not valid")
	ERROR_NDIS_BUFFER_TOO_SHORT                  = errors.New("the length of the submitted buffer for this operation is too small")
	ERROR_NDIS_INVALID_OID                       = errors.New("the network interface does not support this object identifier")
	ERROR_NDIS_ADAPTER_REMOVED                   = errors.New("the network interface has been removed")
	ERROR_NDIS_UNSUPPORTED_MEDIA                 = errors.New("the network interface does not support this media type")
	ERROR_NDIS_GROUP_ADDRESS_IN_USE              = errors.New("an attempt was made to remove a token ring group address that is in use by other components")
	ERROR_NDIS_FILE_NOT_FOUND                    = errors.New("an attempt was made to map a file that cannot be found")
	ERROR_NDIS_ERROR_READING_FILE                = errors.New("an error occurred while ndis tried to map the file")
	ERROR_NDIS_ALREADY_MAPPED                    = errors.New("an attempt was made to map a file that is already mapped")
	ERROR_NDIS_RESOURCE_CONFLICT                 = errors.New("an attempt to allocate a hardware resource failed because the resource is used by another component")
	ERROR_NDIS_MEDIA_DISCONNECTED                = errors.New("the i/o operation failed because the network media is disconnected or the wireless access point is out of range")
	ERROR_NDIS_INVALID_ADDRESS                   = errors.New("the network address used in the request is invalid")
	ERROR_NDIS_PAUSED                            = errors.New("the offload operation on the network interface has been paused")
	ERROR_NDIS_INTERFACE_NOT_FOUND               = errors.New("the network interface was not found")
	ERROR_NDIS_UNSUPPORTED_REVISION              = errors.New("the revision number specified in the structure is not supported")
	ERROR_NDIS_INVALID_PORT                      = errors.New("the specified port does not exist on this network interface")
	ERROR_NDIS_INVALID_PORT_STATE                = errors.New("the current state of the specified port on this network interface does not support the requested operation")
	ERROR_NDIS_LOW_POWER_STATE                   = errors.New("the miniport adapter is in a lower power state")
	ERROR_NDIS_NOT_SUPPORTED                     = errors.New("the network interface does not support this request")
	ERROR_NDIS_OFFLOAD_POLICY                    = errors.New("the tcp connection is not offloadable because of a local policy setting")
	ERROR_NDIS_OFFLOAD_CONNECTION_REJECTED       = errors.New("the tcp connection is not offloadable by the chimney offload target")
	ERROR_NDIS_OFFLOAD_PATH_REJECTED             = errors.New("the ip path object is not in an offloadable state")
	ERROR_NDIS_DOT11_AUTO_CONFIG_ENABLED         = errors.New("the wireless lan interface is in auto-configuration mode and does not support the requested parameter change operation")
	ERROR_NDIS_DOT11_MEDIA_IN_USE                = errors.New("the wireless lan interface is busy and cannot perform the requested operation")
	ERROR_NDIS_DOT11_POWER_STATE_INVALID         = errors.New("the wireless lan interface is power down and does not support the requested operation")
	ERROR_NDIS_PM_WOL_PATTERN_LIST_FULL          = errors.New("the list of wake on lan patterns is full")
	ERROR_NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL     = errors.New("the list of low power protocol offloads is full")
	ERROR_IPSEC_BAD_SPI                          = errors.New("the spi in the packet does not match a valid ipsec sa")
	ERROR_IPSEC_SA_LIFETIME_EXPIRED              = errors.New("the packet was received on an ipsec sa whose lifetime has expired")
	ERROR_IPSEC_WRONG_SA                         = errors.New("the packet was received on an ipsec sa that does not match the packet characteristics")
	ERROR_IPSEC_REPLAY_CHECK_FAILED              = errors.New("the packet sequence number replay check failed")
	ERROR_IPSEC_INVALID_PACKET                   = errors.New("the ipsec header and/or trailer in the packet is invalid")
	ERROR_IPSEC_INTEGRITY_CHECK_FAILED           = errors.New("the ipsec integrity check failed")
	ERROR_IPSEC_CLEAR_TEXT_DROP                  = errors.New("ipsec dropped a clear text packet")
	ERROR_IPSEC_AUTH_FIREWALL_DROP               = errors.New("ipsec dropped an incoming esp packet in authenticated firewall mode. this drop is benign")
	ERROR_IPSEC_THROTTLE_DROP                    = errors.New("ipsec dropped a packet due to dos throttle")
	ERROR_IPSEC_DOSP_BLOCK                       = errors.New("ipsec dos protection matched an explicit block rule")
	ERROR_IPSEC_DOSP_RECEIVED_MULTICAST          = errors.New("ipsec dos protection received an ipsec specific multicast packet which is not allowed")
	ERROR_IPSEC_DOSP_INVALID_PACKET              = errors.New("ipsec dos protection received an incorrectly formatted packet")
	ERROR_IPSEC_DOSP_STATE_LOOKUP_FAILED         = errors.New("ipsec dos protection failed to lookup state")
	ERROR_IPSEC_DOSP_MAX_ENTRIES                 = errors.New("ipsec dos protection failed to create state because there are already maximum number of entries allowed by policy")
	ERROR_IPSEC_DOSP_KEYMOD_NOT_ALLOWED          = errors.New("ipsec dos protection received an ipsec negotiation packet for a keying module which is not allowed by policy")
	ERROR_IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES = errors.New("ipsec dos protection failed to create per internal ip ratelimit queue because there is already maximum number of queues allowed by policy")
	ERROR_VOLMGR_MIRROR_NOT_SUPPORTED            = errors.New("the system does not support mirrored volumes")
	ERROR_VOLMGR_RAID5_NOT_SUPPORTED             = errors.New("the system does not support raid-5 volumes")
	ERROR_VIRTDISK_PROVIDER_NOT_FOUND            = errors.New("a virtual disk support provider for the specified file was not found")
	ERROR_VIRTDISK_NOT_VIRTUAL_DISK              = errors.New("the specified disk is not a virtual disk")
	ERROR_VHD_PARENT_VHD_ACCESS_DENIED           = errors.New("the chain of virtual hard disks is inaccessible. the process has not been granted access rights to the parent virtual hard disk for the differencing disk")
	ERROR_VHD_CHILD_PARENT_SIZE_MISMATCH         = errors.New("the chain of virtual hard disks is corrupted. there is a mismatch in the virtual sizes of the parent virtual hard disk and differencing disk")
	ERROR_VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED  = errors.New("the chain of virtual hard disks is corrupted. a differencing disk is indicated in its own parent chain")
	ERROR_VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT = errors.New("the chain of virtual hard disks is inaccessible. there was an error opening a virtual hard disk further up the chain")
	ERROR_SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP  = errors.New("returned in response to a client negotiate request when the server does not support any of the hash algorithms in the request")
	ERROR_SMB_BAD_CLUSTER_DIALECT                = errors.New("the current cluster functional level does not support this smb dialect")
)

var NTStatusToGoErrorMap = map[NT_STATUS]error{
	NT_STATUS_SUCCESS: ERROR_SUCCESS,
	// NT_STATUS_WAIT_0: // NT_STATUS_WAIT_0,
	NT_STATUS_WAIT_1:    ERROR_WAIT_1,
	NT_STATUS_WAIT_2:    ERROR_WAIT_2,
	NT_STATUS_WAIT_3:    ERROR_WAIT_3,
	NT_STATUS_WAIT_63:   ERROR_WAIT_63,
	NT_STATUS_ABANDONED: ERROR_ABANDONED,
	// NT_STATUS_ABANDONED_WAIT_0: // NT_STATUS_ABANDONED_WAIT_0,
	NT_STATUS_ABANDONED_WAIT_63:                                           ERROR_ABANDONED_WAIT_63,
	NT_STATUS_USER_APC:                                                    ERROR_USER_APC,
	NT_STATUS_ALERTED:                                                     ERROR_ALERTED,
	NT_STATUS_TIMEOUT:                                                     ERROR_TIMEOUT,
	NT_STATUS_PENDING:                                                     ERROR_PENDING,
	NT_STATUS_REPARSE:                                                     ERROR_REPARSE,
	NT_STATUS_MORE_ENTRIES:                                                ERROR_MORE_ENTRIES,
	NT_STATUS_NOT_ALL_ASSIGNED:                                            ERROR_NOT_ALL_ASSIGNED,
	NT_STATUS_SOME_NOT_MAPPED:                                             ERROR_SOME_NOT_MAPPED,
	NT_STATUS_OPLOCK_BREAK_IN_PROGRESS:                                    ERROR_OPLOCK_BREAK_IN_PROGRESS,
	NT_STATUS_VOLUME_MOUNTED:                                              ERROR_VOLUME_MOUNTED,
	NT_STATUS_RXACT_COMMITTED:                                             ERROR_RXACT_COMMITTED,
	NT_STATUS_NOTIFY_CLEANUP:                                              ERROR_NOTIFY_CLEANUP,
	NT_STATUS_NOTIFY_ENUM_DIR:                                             ERROR_NOTIFY_ENUM_DIR,
	NT_STATUS_NO_QUOTAS_FOR_ACCOUNT:                                       ERROR_NO_QUOTAS_FOR_ACCOUNT,
	NT_STATUS_PRIMARY_TRANSPORT_CONNECT_FAILED:                            ERROR_PRIMARY_TRANSPORT_CONNECT_FAILED,
	NT_STATUS_PAGE_FAULT_TRANSITION:                                       ERROR_PAGE_FAULT_TRANSITION,
	NT_STATUS_PAGE_FAULT_DEMAND_ZERO:                                      ERROR_PAGE_FAULT_DEMAND_ZERO,
	NT_STATUS_PAGE_FAULT_COPY_ON_WRITE:                                    ERROR_PAGE_FAULT_COPY_ON_WRITE,
	NT_STATUS_PAGE_FAULT_GUARD_PAGE:                                       ERROR_PAGE_FAULT_GUARD_PAGE,
	NT_STATUS_PAGE_FAULT_PAGING_FILE:                                      ERROR_PAGE_FAULT_PAGING_FILE,
	NT_STATUS_CACHE_PAGE_LOCKED:                                           ERROR_CACHE_PAGE_LOCKED,
	NT_STATUS_CRASH_DUMP:                                                  ERROR_CRASH_DUMP,
	NT_STATUS_BUFFER_ALL_ZEROS:                                            ERROR_BUFFER_ALL_ZEROS,
	NT_STATUS_REPARSE_OBJECT:                                              ERROR_REPARSE_OBJECT,
	NT_STATUS_RESOURCE_REQUIREMENTS_CHANGED:                               ERROR_RESOURCE_REQUIREMENTS_CHANGED,
	NT_STATUS_TRANSLATION_COMPLETE:                                        ERROR_TRANSLATION_COMPLETE,
	NT_STATUS_DS_MEMBERSHIP_EVALUATED_LOCALLY:                             ERROR_DS_MEMBERSHIP_EVALUATED_LOCALLY,
	NT_STATUS_NOTHING_TO_TERMINATE:                                        ERROR_NOTHING_TO_TERMINATE,
	NT_STATUS_PROCESS_NOT_IN_JOB:                                          ERROR_PROCESS_NOT_IN_JOB,
	NT_STATUS_PROCESS_IN_JOB:                                              ERROR_PROCESS_IN_JOB,
	NT_STATUS_VOLSNAP_HIBERNATE_READY:                                     ERROR_VOLSNAP_HIBERNATE_READY,
	NT_STATUS_FSFILTER_OP_COMPLETED_SUCCESSFULLY:                          ERROR_FSFILTER_OP_COMPLETED_SUCCESSFULLY,
	NT_STATUS_INTERRUPT_VECTOR_ALREADY_CONNECTED:                          ERROR_INTERRUPT_VECTOR_ALREADY_CONNECTED,
	NT_STATUS_INTERRUPT_STILL_CONNECTED:                                   ERROR_INTERRUPT_STILL_CONNECTED,
	NT_STATUS_PROCESS_CLONED:                                              ERROR_PROCESS_CLONED,
	NT_STATUS_FILE_LOCKED_WITH_ONLY_READERS:                               ERROR_FILE_LOCKED_WITH_ONLY_READERS,
	NT_STATUS_FILE_LOCKED_WITH_WRITERS:                                    ERROR_FILE_LOCKED_WITH_WRITERS,
	NT_STATUS_RESOURCEMANAGER_READ_ONLY:                                   ERROR_RESOURCEMANAGER_READ_ONLY,
	NT_STATUS_WAIT_FOR_OPLOCK:                                             ERROR_WAIT_FOR_OPLOCK,
	NT_STATUS_DBG_EXCEPTION_HANDLED:                                       ERROR_DBG_EXCEPTION_HANDLED,
	NT_STATUS_DBG_CONTINUE:                                                ERROR_DBG_CONTINUE,
	NT_STATUS_FLT_IO_COMPLETE:                                             ERROR_FLT_IO_COMPLETE,
	NT_STATUS_FILE_NOT_AVAILABLE:                                          ERROR_FILE_NOT_AVAILABLE,
	NT_STATUS_SHARE_UNAVAILABLE:                                           ERROR_SHARE_UNAVAILABLE,
	NT_STATUS_CALLBACK_RETURNED_THREAD_AFFINITY:                           ERROR_CALLBACK_RETURNED_THREAD_AFFINITY,
	NT_STATUS_OBJECT_NAME_EXISTS:                                          ERROR_OBJECT_NAME_EXISTS,
	NT_STATUS_THREAD_WAS_SUSPENDED:                                        ERROR_THREAD_WAS_SUSPENDED,
	NT_STATUS_WORKING_SET_LIMIT_RANGE:                                     ERROR_WORKING_SET_LIMIT_RANGE,
	NT_STATUS_IMAGE_NOT_AT_BASE:                                           ERROR_IMAGE_NOT_AT_BASE,
	NT_STATUS_RXACT_STATE_CREATED:                                         ERROR_RXACT_STATE_CREATED,
	NT_STATUS_SEGMENT_NOTIFICATION:                                        ERROR_SEGMENT_NOTIFICATION,
	NT_STATUS_LOCAL_USER_SESSION_KEY:                                      ERROR_LOCAL_USER_SESSION_KEY,
	NT_STATUS_BAD_CURRENT_DIRECTORY:                                       ERROR_BAD_CURRENT_DIRECTORY,
	NT_STATUS_SERIAL_MORE_WRITES:                                          ERROR_SERIAL_MORE_WRITES,
	NT_STATUS_REGISTRY_RECOVERED:                                          ERROR_REGISTRY_RECOVERED,
	NT_STATUS_FT_READ_RECOVERY_FROM_BACKUP:                                ERROR_FT_READ_RECOVERY_FROM_BACKUP,
	NT_STATUS_FT_WRITE_RECOVERY:                                           ERROR_FT_WRITE_RECOVERY,
	NT_STATUS_SERIAL_COUNTER_TIMEOUT:                                      ERROR_SERIAL_COUNTER_TIMEOUT,
	NT_STATUS_NULL_LM_PASSWORD:                                            ERROR_NULL_LM_PASSWORD,
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH:                                 ERROR_IMAGE_MACHINE_TYPE_MISMATCH,
	NT_STATUS_RECEIVE_PARTIAL:                                             ERROR_RECEIVE_PARTIAL,
	NT_STATUS_RECEIVE_EXPEDITED:                                           ERROR_RECEIVE_EXPEDITED,
	NT_STATUS_RECEIVE_PARTIAL_EXPEDITED:                                   ERROR_RECEIVE_PARTIAL_EXPEDITED,
	NT_STATUS_EVENT_DONE:                                                  ERROR_EVENT_DONE,
	NT_STATUS_EVENT_PENDING:                                               ERROR_EVENT_PENDING,
	NT_STATUS_CHECKING_FILE_SYSTEM:                                        ERROR_CHECKING_FILE_SYSTEM,
	NT_STATUS_FATAL_APP_EXIT:                                              ERROR_FATAL_APP_EXIT,
	NT_STATUS_PREDEFINED_HANDLE:                                           ERROR_PREDEFINED_HANDLE,
	NT_STATUS_WAS_UNLOCKED:                                                ERROR_WAS_UNLOCKED,
	NT_STATUS_SERVICE_NOTIFICATION:                                        ERROR_SERVICE_NOTIFICATION,
	NT_STATUS_WAS_LOCKED:                                                  ERROR_WAS_LOCKED,
	NT_STATUS_LOG_HARD_ERROR:                                              ERROR_LOG_HARD_ERROR,
	NT_STATUS_ALREADY_WIN32:                                               ERROR_ALREADY_WIN32,
	NT_STATUS_WX86_UNSIMULATE:                                             ERROR_WX86_UNSIMULATE,
	NT_STATUS_WX86_CONTINUE:                                               ERROR_WX86_CONTINUE,
	NT_STATUS_WX86_SINGLE_STEP:                                            ERROR_WX86_SINGLE_STEP,
	NT_STATUS_WX86_BREAKPOINT:                                             ERROR_WX86_BREAKPOINT,
	NT_STATUS_WX86_EXCEPTION_CONTINUE:                                     ERROR_WX86_EXCEPTION_CONTINUE,
	NT_STATUS_WX86_EXCEPTION_LASTCHANCE:                                   ERROR_WX86_EXCEPTION_LASTCHANCE,
	NT_STATUS_WX86_EXCEPTION_CHAIN:                                        ERROR_WX86_EXCEPTION_CHAIN,
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH_EXE:                             ERROR_IMAGE_MACHINE_TYPE_MISMATCH_EXE,
	NT_STATUS_NO_YIELD_PERFORMED:                                          ERROR_NO_YIELD_PERFORMED,
	NT_STATUS_TIMER_RESUME_IGNORED:                                        ERROR_TIMER_RESUME_IGNORED,
	NT_STATUS_ARBITRATION_UNHANDLED:                                       ERROR_ARBITRATION_UNHANDLED,
	NT_STATUS_CARDBUS_NOT_SUPPORTED:                                       ERROR_CARDBUS_NOT_SUPPORTED,
	NT_STATUS_WX86_CREATEWX86TIB:                                          ERROR_WX86_CREATEWX86TIB,
	NT_STATUS_MP_PROCESSOR_MISMATCH:                                       ERROR_MP_PROCESSOR_MISMATCH,
	NT_STATUS_HIBERNATED:                                                  ERROR_HIBERNATED,
	NT_STATUS_RESUME_HIBERNATION:                                          ERROR_RESUME_HIBERNATION,
	NT_STATUS_FIRMWARE_UPDATED:                                            ERROR_FIRMWARE_UPDATED,
	NT_STATUS_DRIVERS_LEAKING_LOCKED_PAGES:                                ERROR_DRIVERS_LEAKING_LOCKED_PAGES,
	NT_STATUS_MESSAGE_RETRIEVED:                                           ERROR_MESSAGE_RETRIEVED,
	NT_STATUS_SYSTEM_POWERSTATE_TRANSITION:                                ERROR_SYSTEM_POWERSTATE_TRANSITION,
	NT_STATUS_ALPC_CHECK_COMPLETION_LIST:                                  ERROR_ALPC_CHECK_COMPLETION_LIST,
	NT_STATUS_SYSTEM_POWERSTATE_COMPLEX_TRANSITION:                        ERROR_SYSTEM_POWERSTATE_COMPLEX_TRANSITION,
	NT_STATUS_ACCESS_AUDIT_BY_POLICY:                                      ERROR_ACCESS_AUDIT_BY_POLICY,
	NT_STATUS_ABANDON_HIBERFILE:                                           ERROR_ABANDON_HIBERFILE,
	NT_STATUS_BIZRULES_NOT_ENABLED:                                        ERROR_BIZRULES_NOT_ENABLED,
	NT_STATUS_WAKE_SYSTEM:                                                 ERROR_WAKE_SYSTEM,
	NT_STATUS_DS_SHUTTING_DOWN:                                            ERROR_DS_SHUTTING_DOWN,
	NT_STATUS_DBG_REPLY_LATER:                                             ERROR_DBG_REPLY_LATER,
	NT_STATUS_DBG_UNABLE_TO_PROVIDE_HANDLE:                                ERROR_DBG_UNABLE_TO_PROVIDE_HANDLE,
	NT_STATUS_DBG_TERMINATE_THREAD:                                        ERROR_DBG_TERMINATE_THREAD,
	NT_STATUS_DBG_TERMINATE_PROCESS:                                       ERROR_DBG_TERMINATE_PROCESS,
	NT_STATUS_DBG_CONTROL_C:                                               ERROR_DBG_CONTROL_C,
	NT_STATUS_DBG_PRINTEXCEPTION_C:                                        ERROR_DBG_PRINTEXCEPTION_C,
	NT_STATUS_DBG_RIPEXCEPTION:                                            ERROR_DBG_RIPEXCEPTION,
	NT_STATUS_DBG_CONTROL_BREAK:                                           ERROR_DBG_CONTROL_BREAK,
	NT_STATUS_DBG_COMMAND_EXCEPTION:                                       ERROR_DBG_COMMAND_EXCEPTION,
	NT_STATUS_RPC_NT_UUID_LOCAL_ONLY:                                      ERROR_RPC_NT_UUID_LOCAL_ONLY,
	NT_STATUS_RPC_NT_SEND_INCOMPLETE:                                      ERROR_RPC_NT_SEND_INCOMPLETE,
	NT_STATUS_CTX_CDM_CONNECT:                                             ERROR_CTX_CDM_CONNECT,
	NT_STATUS_CTX_CDM_DISCONNECT:                                          ERROR_CTX_CDM_DISCONNECT,
	NT_STATUS_SXS_RELEASE_ACTIVATION_CONTEXT:                              ERROR_SXS_RELEASE_ACTIVATION_CONTEXT,
	NT_STATUS_RECOVERY_NOT_NEEDED:                                         ERROR_RECOVERY_NOT_NEEDED,
	NT_STATUS_RM_ALREADY_STARTED:                                          ERROR_RM_ALREADY_STARTED,
	NT_STATUS_LOG_NO_RESTART:                                              ERROR_LOG_NO_RESTART,
	NT_STATUS_VIDEO_DRIVER_DEBUG_REPORT_REQUEST:                           ERROR_VIDEO_DRIVER_DEBUG_REPORT_REQUEST,
	NT_STATUS_GRAPHICS_PARTIAL_DATA_POPULATED:                             ERROR_GRAPHICS_PARTIAL_DATA_POPULATED,
	NT_STATUS_GRAPHICS_DRIVER_MISMATCH:                                    ERROR_GRAPHICS_DRIVER_MISMATCH,
	NT_STATUS_GRAPHICS_MODE_NOT_PINNED:                                    ERROR_GRAPHICS_MODE_NOT_PINNED,
	NT_STATUS_GRAPHICS_NO_PREFERRED_MODE:                                  ERROR_GRAPHICS_NO_PREFERRED_MODE,
	NT_STATUS_GRAPHICS_DATASET_IS_EMPTY:                                   ERROR_GRAPHICS_DATASET_IS_EMPTY,
	NT_STATUS_GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET:                        ERROR_GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET,
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED:    ERROR_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED,
	NT_STATUS_GRAPHICS_UNKNOWN_CHILD_STATUS:                               ERROR_GRAPHICS_UNKNOWN_CHILD_STATUS,
	NT_STATUS_GRAPHICS_LEADLINK_START_DEFERRED:                            ERROR_GRAPHICS_LEADLINK_START_DEFERRED,
	NT_STATUS_GRAPHICS_POLLING_TOO_FREQUENTLY:                             ERROR_GRAPHICS_POLLING_TOO_FREQUENTLY,
	NT_STATUS_GRAPHICS_START_DEFERRED:                                     ERROR_GRAPHICS_START_DEFERRED,
	NT_STATUS_NDIS_INDICATION_REQUIRED:                                    ERROR_NDIS_INDICATION_REQUIRED,
	NT_STATUS_GUARD_PAGE_VIOLATION:                                        ERROR_GUARD_PAGE_VIOLATION,
	NT_STATUS_DATATYPE_MISALIGNMENT:                                       ERROR_DATATYPE_MISALIGNMENT,
	NT_STATUS_BREAKPOINT:                                                  ERROR_BREAKPOINT,
	NT_STATUS_SINGLE_STEP:                                                 ERROR_SINGLE_STEP,
	NT_STATUS_BUFFER_OVERFLOW:                                             ERROR_BUFFER_OVERFLOW,
	NT_STATUS_NO_MORE_FILES:                                               ERROR_NO_MORE_FILES,
	NT_STATUS_WAKE_SYSTEM_DEBUGGER:                                        ERROR_WAKE_SYSTEM_DEBUGGER,
	NT_STATUS_HANDLES_CLOSED:                                              ERROR_HANDLES_CLOSED,
	NT_STATUS_NO_INHERITANCE:                                              ERROR_NO_INHERITANCE,
	NT_STATUS_GUID_SUBSTITUTION_MADE:                                      ERROR_GUID_SUBSTITUTION_MADE,
	NT_STATUS_PARTIAL_COPY:                                                ERROR_PARTIAL_COPY,
	NT_STATUS_DEVICE_PAPER_EMPTY:                                          ERROR_DEVICE_PAPER_EMPTY,
	NT_STATUS_DEVICE_POWERED_OFF:                                          ERROR_DEVICE_POWERED_OFF,
	NT_STATUS_DEVICE_OFF_LINE:                                             ERROR_DEVICE_OFF_LINE,
	NT_STATUS_DEVICE_BUSY:                                                 ERROR_DEVICE_BUSY,
	NT_STATUS_NO_MORE_EAS:                                                 ERROR_NO_MORE_EAS,
	NT_STATUS_INVALID_EA_NAME:                                             ERROR_INVALID_EA_NAME,
	NT_STATUS_EA_LIST_INCONSISTENT:                                        ERROR_EA_LIST_INCONSISTENT,
	NT_STATUS_INVALID_EA_FLAG:                                             ERROR_INVALID_EA_FLAG,
	NT_STATUS_VERIFY_REQUIRED:                                             ERROR_VERIFY_REQUIRED,
	NT_STATUS_EXTRANEOUS_INFORMATION:                                      ERROR_EXTRANEOUS_INFORMATION,
	NT_STATUS_RXACT_COMMIT_NECESSARY:                                      ERROR_RXACT_COMMIT_NECESSARY,
	NT_STATUS_NO_MORE_ENTRIES:                                             ERROR_NO_MORE_ENTRIES,
	NT_STATUS_FILEMARK_DETECTED:                                           ERROR_FILEMARK_DETECTED,
	NT_STATUS_MEDIA_CHANGED:                                               ERROR_MEDIA_CHANGED,
	NT_STATUS_BUS_RESET:                                                   ERROR_BUS_RESET,
	NT_STATUS_END_OF_MEDIA:                                                ERROR_END_OF_MEDIA,
	NT_STATUS_BEGINNING_OF_MEDIA:                                          ERROR_BEGINNING_OF_MEDIA,
	NT_STATUS_MEDIA_CHECK:                                                 ERROR_MEDIA_CHECK,
	NT_STATUS_SETMARK_DETECTED:                                            ERROR_SETMARK_DETECTED,
	NT_STATUS_NO_DATA_DETECTED:                                            ERROR_NO_DATA_DETECTED,
	NT_STATUS_REDIRECTOR_HAS_OPEN_HANDLES:                                 ERROR_REDIRECTOR_HAS_OPEN_HANDLES,
	NT_STATUS_SERVER_HAS_OPEN_HANDLES:                                     ERROR_SERVER_HAS_OPEN_HANDLES,
	NT_STATUS_ALREADY_DISCONNECTED:                                        ERROR_ALREADY_DISCONNECTED,
	NT_STATUS_LONGJUMP:                                                    ERROR_LONGJUMP,
	NT_STATUS_CLEANER_CARTRIDGE_INSTALLED:                                 ERROR_CLEANER_CARTRIDGE_INSTALLED,
	NT_STATUS_PLUGPLAY_QUERY_VETOED:                                       ERROR_PLUGPLAY_QUERY_VETOED,
	NT_STATUS_UNWIND_CONSOLIDATE:                                          ERROR_UNWIND_CONSOLIDATE,
	NT_STATUS_REGISTRY_HIVE_RECOVERED:                                     ERROR_REGISTRY_HIVE_RECOVERED,
	NT_STATUS_DLL_MIGHT_BE_INSECURE:                                       ERROR_DLL_MIGHT_BE_INSECURE,
	NT_STATUS_DLL_MIGHT_BE_INCOMPATIBLE:                                   ERROR_DLL_MIGHT_BE_INCOMPATIBLE,
	NT_STATUS_STOPPED_ON_SYMLINK:                                          ERROR_STOPPED_ON_SYMLINK,
	NT_STATUS_DEVICE_REQUIRES_CLEANING:                                    ERROR_DEVICE_REQUIRES_CLEANING,
	NT_STATUS_DEVICE_DOOR_OPEN:                                            ERROR_DEVICE_DOOR_OPEN,
	NT_STATUS_DATA_LOST_REPAIR:                                            ERROR_DATA_LOST_REPAIR,
	NT_STATUS_DBG_EXCEPTION_NOT_HANDLED:                                   ERROR_DBG_EXCEPTION_NOT_HANDLED,
	NT_STATUS_CLUSTER_NODE_ALREADY_UP:                                     ERROR_CLUSTER_NODE_ALREADY_UP,
	NT_STATUS_CLUSTER_NODE_ALREADY_DOWN:                                   ERROR_CLUSTER_NODE_ALREADY_DOWN,
	NT_STATUS_CLUSTER_NETWORK_ALREADY_ONLINE:                              ERROR_CLUSTER_NETWORK_ALREADY_ONLINE,
	NT_STATUS_CLUSTER_NETWORK_ALREADY_OFFLINE:                             ERROR_CLUSTER_NETWORK_ALREADY_OFFLINE,
	NT_STATUS_CLUSTER_NODE_ALREADY_MEMBER:                                 ERROR_CLUSTER_NODE_ALREADY_MEMBER,
	NT_STATUS_COULD_NOT_RESIZE_LOG:                                        ERROR_COULD_NOT_RESIZE_LOG,
	NT_STATUS_NO_TXF_METADATA:                                             ERROR_NO_TXF_METADATA,
	NT_STATUS_CANT_RECOVER_WITH_HANDLE_OPEN:                               ERROR_CANT_RECOVER_WITH_HANDLE_OPEN,
	NT_STATUS_TXF_METADATA_ALREADY_PRESENT:                                ERROR_TXF_METADATA_ALREADY_PRESENT,
	NT_STATUS_TRANSACTION_SCOPE_CALLBACKS_NOT_SET:                         ERROR_TRANSACTION_SCOPE_CALLBACKS_NOT_SET,
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED:                  ERROR_VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED,
	NT_STATUS_FLT_BUFFER_TOO_SMALL:                                        ERROR_FLT_BUFFER_TOO_SMALL,
	NT_STATUS_FVE_PARTIAL_METADATA:                                        ERROR_FVE_PARTIAL_METADATA,
	NT_STATUS_FVE_TRANSIENT_STATE:                                         ERROR_FVE_TRANSIENT_STATE,
	NT_STATUS_UNSUCCESSFUL:                                                ERROR_UNSUCCESSFUL,
	NT_STATUS_NOT_IMPLEMENTED:                                             ERROR_NOT_IMPLEMENTED,
	NT_STATUS_INVALID_INFO_CLASS:                                          ERROR_INVALID_INFO_CLASS,
	NT_STATUS_INFO_LENGTH_MISMATCH:                                        ERROR_INFO_LENGTH_MISMATCH,
	NT_STATUS_ACCESS_VIOLATION:                                            ERROR_ACCESS_VIOLATION,
	NT_STATUS_IN_PAGE_ERROR:                                               ERROR_IN_PAGE_ERROR,
	NT_STATUS_PAGEFILE_QUOTA:                                              ERROR_PAGEFILE_QUOTA,
	NT_STATUS_INVALID_HANDLE:                                              ERROR_INVALID_HANDLE,
	NT_STATUS_BAD_INITIAL_STACK:                                           ERROR_BAD_INITIAL_STACK,
	NT_STATUS_BAD_INITIAL_PC:                                              ERROR_BAD_INITIAL_PC,
	NT_STATUS_INVALID_CID:                                                 ERROR_INVALID_CID,
	NT_STATUS_TIMER_NOT_CANCELED:                                          ERROR_TIMER_NOT_CANCELED,
	NT_STATUS_INVALID_PARAMETER:                                           ERROR_INVALID_PARAMETER,
	NT_STATUS_NO_SUCH_DEVICE:                                              ERROR_NO_SUCH_DEVICE,
	NT_STATUS_NO_SUCH_FILE:                                                ERROR_NO_SUCH_FILE,
	NT_STATUS_INVALID_DEVICE_REQUEST:                                      ERROR_INVALID_DEVICE_REQUEST,
	NT_STATUS_END_OF_FILE:                                                 ERROR_END_OF_FILE,
	NT_STATUS_WRONG_VOLUME:                                                ERROR_WRONG_VOLUME,
	NT_STATUS_NO_MEDIA_IN_DEVICE:                                          ERROR_NO_MEDIA_IN_DEVICE,
	NT_STATUS_UNRECOGNIZED_MEDIA:                                          ERROR_UNRECOGNIZED_MEDIA,
	NT_STATUS_NONEXISTENT_SECTOR:                                          ERROR_NONEXISTENT_SECTOR,
	NT_STATUS_MORE_PROCESSING_REQUIRED:                                    ERROR_MORE_PROCESSING_REQUIRED,
	NT_STATUS_NO_MEMORY:                                                   ERROR_NO_MEMORY,
	NT_STATUS_CONFLICTING_ADDRESSES:                                       ERROR_CONFLICTING_ADDRESSES,
	NT_STATUS_NOT_MAPPED_VIEW:                                             ERROR_NOT_MAPPED_VIEW,
	NT_STATUS_UNABLE_TO_FREE_VM:                                           ERROR_UNABLE_TO_FREE_VM,
	NT_STATUS_UNABLE_TO_DELETE_SECTION:                                    ERROR_UNABLE_TO_DELETE_SECTION,
	NT_STATUS_INVALID_SYSTEM_SERVICE:                                      ERROR_INVALID_SYSTEM_SERVICE,
	NT_STATUS_ILLEGAL_INSTRUCTION:                                         ERROR_ILLEGAL_INSTRUCTION,
	NT_STATUS_INVALID_LOCK_SEQUENCE:                                       ERROR_INVALID_LOCK_SEQUENCE,
	NT_STATUS_INVALID_VIEW_SIZE:                                           ERROR_INVALID_VIEW_SIZE,
	NT_STATUS_INVALID_FILE_FOR_SECTION:                                    ERROR_INVALID_FILE_FOR_SECTION,
	NT_STATUS_ALREADY_COMMITTED:                                           ERROR_ALREADY_COMMITTED,
	NT_STATUS_ACCESS_DENIED:                                               ERROR_ACCESS_DENIED,
	NT_STATUS_BUFFER_TOO_SMALL:                                            ERROR_BUFFER_TOO_SMALL,
	NT_STATUS_OBJECT_TYPE_MISMATCH:                                        ERROR_OBJECT_TYPE_MISMATCH,
	NT_STATUS_NONCONTINUABLE_EXCEPTION:                                    ERROR_NONCONTINUABLE_EXCEPTION,
	NT_STATUS_INVALID_DISPOSITION:                                         ERROR_INVALID_DISPOSITION,
	NT_STATUS_UNWIND:                                                      ERROR_UNWIND,
	NT_STATUS_BAD_STACK:                                                   ERROR_BAD_STACK,
	NT_STATUS_INVALID_UNWIND_TARGET:                                       ERROR_INVALID_UNWIND_TARGET,
	NT_STATUS_NOT_LOCKED:                                                  ERROR_NOT_LOCKED,
	NT_STATUS_PARITY_ERROR:                                                ERROR_PARITY_ERROR,
	NT_STATUS_UNABLE_TO_DECOMMIT_VM:                                       ERROR_UNABLE_TO_DECOMMIT_VM,
	NT_STATUS_NOT_COMMITTED:                                               ERROR_NOT_COMMITTED,
	NT_STATUS_INVALID_PORT_ATTRIBUTES:                                     ERROR_INVALID_PORT_ATTRIBUTES,
	NT_STATUS_PORT_MESSAGE_TOO_LONG:                                       ERROR_PORT_MESSAGE_TOO_LONG,
	NT_STATUS_INVALID_PARAMETER_MIX:                                       ERROR_INVALID_PARAMETER_MIX,
	NT_STATUS_INVALID_QUOTA_LOWER:                                         ERROR_INVALID_QUOTA_LOWER,
	NT_STATUS_DISK_CORRUPT_ERROR:                                          ERROR_DISK_CORRUPT_ERROR,
	NT_STATUS_OBJECT_NAME_INVALID:                                         ERROR_OBJECT_NAME_INVALID,
	NT_STATUS_OBJECT_NAME_NOT_FOUND:                                       ERROR_OBJECT_NAME_NOT_FOUND,
	NT_STATUS_OBJECT_NAME_COLLISION:                                       ERROR_OBJECT_NAME_COLLISION,
	NT_STATUS_PORT_DISCONNECTED:                                           ERROR_PORT_DISCONNECTED,
	NT_STATUS_DEVICE_ALREADY_ATTACHED:                                     ERROR_DEVICE_ALREADY_ATTACHED,
	NT_STATUS_OBJECT_PATH_INVALID:                                         ERROR_OBJECT_PATH_INVALID,
	NT_STATUS_OBJECT_PATH_NOT_FOUND:                                       ERROR_OBJECT_PATH_NOT_FOUND,
	NT_STATUS_OBJECT_PATH_SYNTAX_BAD:                                      ERROR_OBJECT_PATH_SYNTAX_BAD,
	NT_STATUS_DATA_OVERRUN:                                                ERROR_DATA_OVERRUN,
	NT_STATUS_DATA_LATE_ERROR:                                             ERROR_DATA_LATE_ERROR,
	NT_STATUS_DATA_ERROR:                                                  ERROR_DATA_ERROR,
	NT_STATUS_CRC_ERROR:                                                   ERROR_CRC_ERROR,
	NT_STATUS_SECTION_TOO_BIG:                                             ERROR_SECTION_TOO_BIG,
	NT_STATUS_PORT_CONNECTION_REFUSED:                                     ERROR_PORT_CONNECTION_REFUSED,
	NT_STATUS_INVALID_PORT_HANDLE:                                         ERROR_INVALID_PORT_HANDLE,
	NT_STATUS_SHARING_VIOLATION:                                           ERROR_SHARING_VIOLATION,
	NT_STATUS_QUOTA_EXCEEDED:                                              ERROR_QUOTA_EXCEEDED,
	NT_STATUS_INVALID_PAGE_PROTECTION:                                     ERROR_INVALID_PAGE_PROTECTION,
	NT_STATUS_MUTANT_NOT_OWNED:                                            ERROR_MUTANT_NOT_OWNED,
	NT_STATUS_SEMAPHORE_LIMIT_EXCEEDED:                                    ERROR_SEMAPHORE_LIMIT_EXCEEDED,
	NT_STATUS_PORT_ALREADY_SET:                                            ERROR_PORT_ALREADY_SET,
	NT_STATUS_SECTION_NOT_IMAGE:                                           ERROR_SECTION_NOT_IMAGE,
	NT_STATUS_SUSPEND_COUNT_EXCEEDED:                                      ERROR_SUSPEND_COUNT_EXCEEDED,
	NT_STATUS_THREAD_IS_TERMINATING:                                       ERROR_THREAD_IS_TERMINATING,
	NT_STATUS_BAD_WORKING_SET_LIMIT:                                       ERROR_BAD_WORKING_SET_LIMIT,
	NT_STATUS_INCOMPATIBLE_FILE_MAP:                                       ERROR_INCOMPATIBLE_FILE_MAP,
	NT_STATUS_SECTION_PROTECTION:                                          ERROR_SECTION_PROTECTION,
	NT_STATUS_EAS_NOT_SUPPORTED:                                           ERROR_EAS_NOT_SUPPORTED,
	NT_STATUS_EA_TOO_LARGE:                                                ERROR_EA_TOO_LARGE,
	NT_STATUS_NONEXISTENT_EA_ENTRY:                                        ERROR_NONEXISTENT_EA_ENTRY,
	NT_STATUS_NO_EAS_ON_FILE:                                              ERROR_NO_EAS_ON_FILE,
	NT_STATUS_EA_CORRUPT_ERROR:                                            ERROR_EA_CORRUPT_ERROR,
	NT_STATUS_FILE_LOCK_CONFLICT:                                          ERROR_FILE_LOCK_CONFLICT,
	NT_STATUS_LOCK_NOT_GRANTED:                                            ERROR_LOCK_NOT_GRANTED,
	NT_STATUS_DELETE_PENDING:                                              ERROR_DELETE_PENDING,
	NT_STATUS_CTL_FILE_NOT_SUPPORTED:                                      ERROR_CTL_FILE_NOT_SUPPORTED,
	NT_STATUS_UNKNOWN_REVISION:                                            ERROR_UNKNOWN_REVISION,
	NT_STATUS_REVISION_MISMATCH:                                           ERROR_REVISION_MISMATCH,
	NT_STATUS_INVALID_OWNER:                                               ERROR_INVALID_OWNER,
	NT_STATUS_INVALID_PRIMARY_GROUP:                                       ERROR_INVALID_PRIMARY_GROUP,
	NT_STATUS_NO_IMPERSONATION_TOKEN:                                      ERROR_NO_IMPERSONATION_TOKEN,
	NT_STATUS_CANT_DISABLE_MANDATORY:                                      ERROR_CANT_DISABLE_MANDATORY,
	NT_STATUS_NO_LOGON_SERVERS:                                            ERROR_NO_LOGON_SERVERS,
	NT_STATUS_NO_SUCH_LOGON_SESSION:                                       ERROR_NO_SUCH_LOGON_SESSION,
	NT_STATUS_NO_SUCH_PRIVILEGE:                                           ERROR_NO_SUCH_PRIVILEGE,
	NT_STATUS_PRIVILEGE_NOT_HELD:                                          ERROR_PRIVILEGE_NOT_HELD,
	NT_STATUS_INVALID_ACCOUNT_NAME:                                        ERROR_INVALID_ACCOUNT_NAME,
	NT_STATUS_USER_EXISTS:                                                 ERROR_USER_EXISTS,
	NT_STATUS_NO_SUCH_USER:                                                ERROR_NO_SUCH_USER,
	NT_STATUS_GROUP_EXISTS:                                                ERROR_GROUP_EXISTS,
	NT_STATUS_NO_SUCH_GROUP:                                               ERROR_NO_SUCH_GROUP,
	NT_STATUS_MEMBER_IN_GROUP:                                             ERROR_MEMBER_IN_GROUP,
	NT_STATUS_MEMBER_NOT_IN_GROUP:                                         ERROR_MEMBER_NOT_IN_GROUP,
	NT_STATUS_LAST_ADMIN:                                                  ERROR_LAST_ADMIN,
	NT_STATUS_WRONG_PASSWORD:                                              ERROR_WRONG_PASSWORD,
	NT_STATUS_ILL_FORMED_PASSWORD:                                         ERROR_ILL_FORMED_PASSWORD,
	NT_STATUS_PASSWORD_RESTRICTION:                                        ERROR_PASSWORD_RESTRICTION,
	NT_STATUS_LOGON_FAILURE:                                               ERROR_LOGON_FAILURE,
	NT_STATUS_ACCOUNT_RESTRICTION:                                         ERROR_ACCOUNT_RESTRICTION,
	NT_STATUS_INVALID_LOGON_HOURS:                                         ERROR_INVALID_LOGON_HOURS,
	NT_STATUS_INVALID_WORKSTATION:                                         ERROR_INVALID_WORKSTATION,
	NT_STATUS_PASSWORD_EXPIRED:                                            ERROR_PASSWORD_EXPIRED,
	NT_STATUS_ACCOUNT_DISABLED:                                            ERROR_ACCOUNT_DISABLED,
	NT_STATUS_NONE_MAPPED:                                                 ERROR_NONE_MAPPED,
	NT_STATUS_TOO_MANY_LUIDS_REQUESTED:                                    ERROR_TOO_MANY_LUIDS_REQUESTED,
	NT_STATUS_LUIDS_EXHAUSTED:                                             ERROR_LUIDS_EXHAUSTED,
	NT_STATUS_INVALID_SUB_AUTHORITY:                                       ERROR_INVALID_SUB_AUTHORITY,
	NT_STATUS_INVALID_ACL:                                                 ERROR_INVALID_ACL,
	NT_STATUS_INVALID_SID:                                                 ERROR_INVALID_SID,
	NT_STATUS_INVALID_SECURITY_DESCR:                                      ERROR_INVALID_SECURITY_DESCR,
	NT_STATUS_PROCEDURE_NOT_FOUND:                                         ERROR_PROCEDURE_NOT_FOUND,
	NT_STATUS_INVALID_IMAGE_FORMAT:                                        ERROR_INVALID_IMAGE_FORMAT,
	NT_STATUS_NO_TOKEN:                                                    ERROR_NO_TOKEN,
	NT_STATUS_BAD_INHERITANCE_ACL:                                         ERROR_BAD_INHERITANCE_ACL,
	NT_STATUS_RANGE_NOT_LOCKED:                                            ERROR_RANGE_NOT_LOCKED,
	NT_STATUS_DISK_FULL:                                                   ERROR_DISK_FULL,
	NT_STATUS_SERVER_DISABLED:                                             ERROR_SERVER_DISABLED,
	NT_STATUS_SERVER_NOT_DISABLED:                                         ERROR_SERVER_NOT_DISABLED,
	NT_STATUS_TOO_MANY_GUIDS_REQUESTED:                                    ERROR_TOO_MANY_GUIDS_REQUESTED,
	NT_STATUS_GUIDS_EXHAUSTED:                                             ERROR_GUIDS_EXHAUSTED,
	NT_STATUS_INVALID_ID_AUTHORITY:                                        ERROR_INVALID_ID_AUTHORITY,
	NT_STATUS_AGENTS_EXHAUSTED:                                            ERROR_AGENTS_EXHAUSTED,
	NT_STATUS_INVALID_VOLUME_LABEL:                                        ERROR_INVALID_VOLUME_LABEL,
	NT_STATUS_SECTION_NOT_EXTENDED:                                        ERROR_SECTION_NOT_EXTENDED,
	NT_STATUS_NOT_MAPPED_DATA:                                             ERROR_NOT_MAPPED_DATA,
	NT_STATUS_RESOURCE_DATA_NOT_FOUND:                                     ERROR_RESOURCE_DATA_NOT_FOUND,
	NT_STATUS_RESOURCE_TYPE_NOT_FOUND:                                     ERROR_RESOURCE_TYPE_NOT_FOUND,
	NT_STATUS_RESOURCE_NAME_NOT_FOUND:                                     ERROR_RESOURCE_NAME_NOT_FOUND,
	NT_STATUS_ARRAY_BOUNDS_EXCEEDED:                                       ERROR_ARRAY_BOUNDS_EXCEEDED,
	NT_STATUS_FLOAT_DENORMAL_OPERAND:                                      ERROR_FLOAT_DENORMAL_OPERAND,
	NT_STATUS_FLOAT_DIVIDE_BY_ZERO:                                        ERROR_FLOAT_DIVIDE_BY_ZERO,
	NT_STATUS_FLOAT_INEXACT_RESULT:                                        ERROR_FLOAT_INEXACT_RESULT,
	NT_STATUS_FLOAT_INVALID_OPERATION:                                     ERROR_FLOAT_INVALID_OPERATION,
	NT_STATUS_FLOAT_OVERFLOW:                                              ERROR_FLOAT_OVERFLOW,
	NT_STATUS_FLOAT_STACK_CHECK:                                           ERROR_FLOAT_STACK_CHECK,
	NT_STATUS_FLOAT_UNDERFLOW:                                             ERROR_FLOAT_UNDERFLOW,
	NT_STATUS_INTEGER_DIVIDE_BY_ZERO:                                      ERROR_INTEGER_DIVIDE_BY_ZERO,
	NT_STATUS_INTEGER_OVERFLOW:                                            ERROR_INTEGER_OVERFLOW,
	NT_STATUS_PRIVILEGED_INSTRUCTION:                                      ERROR_PRIVILEGED_INSTRUCTION,
	NT_STATUS_TOO_MANY_PAGING_FILES:                                       ERROR_TOO_MANY_PAGING_FILES,
	NT_STATUS_FILE_INVALID:                                                ERROR_FILE_INVALID,
	NT_STATUS_ALLOTTED_SPACE_EXCEEDED:                                     ERROR_ALLOTTED_SPACE_EXCEEDED,
	NT_STATUS_INSUFFICIENT_RESOURCES:                                      ERROR_INSUFFICIENT_RESOURCES,
	NT_STATUS_DFS_EXIT_PATH_FOUND:                                         ERROR_DFS_EXIT_PATH_FOUND,
	NT_STATUS_DEVICE_DATA_ERROR:                                           ERROR_DEVICE_DATA_ERROR,
	NT_STATUS_DEVICE_NOT_CONNECTED:                                        ERROR_DEVICE_NOT_CONNECTED,
	NT_STATUS_FREE_VM_NOT_AT_BASE:                                         ERROR_FREE_VM_NOT_AT_BASE,
	NT_STATUS_MEMORY_NOT_ALLOCATED:                                        ERROR_MEMORY_NOT_ALLOCATED,
	NT_STATUS_WORKING_SET_QUOTA:                                           ERROR_WORKING_SET_QUOTA,
	NT_STATUS_MEDIA_WRITE_PROTECTED:                                       ERROR_MEDIA_WRITE_PROTECTED,
	NT_STATUS_DEVICE_NOT_READY:                                            ERROR_DEVICE_NOT_READY,
	NT_STATUS_INVALID_GROUP_ATTRIBUTES:                                    ERROR_INVALID_GROUP_ATTRIBUTES,
	NT_STATUS_BAD_IMPERSONATION_LEVEL:                                     ERROR_BAD_IMPERSONATION_LEVEL,
	NT_STATUS_CANT_OPEN_ANONYMOUS:                                         ERROR_CANT_OPEN_ANONYMOUS,
	NT_STATUS_BAD_VALIDATION_CLASS:                                        ERROR_BAD_VALIDATION_CLASS,
	NT_STATUS_BAD_TOKEN_TYPE:                                              ERROR_BAD_TOKEN_TYPE,
	NT_STATUS_BAD_MASTER_BOOT_RECORD:                                      ERROR_BAD_MASTER_BOOT_RECORD,
	NT_STATUS_INSTRUCTION_MISALIGNMENT:                                    ERROR_INSTRUCTION_MISALIGNMENT,
	NT_STATUS_INSTANCE_NOT_AVAILABLE:                                      ERROR_INSTANCE_NOT_AVAILABLE,
	NT_STATUS_PIPE_NOT_AVAILABLE:                                          ERROR_PIPE_NOT_AVAILABLE,
	NT_STATUS_INVALID_PIPE_STATE:                                          ERROR_INVALID_PIPE_STATE,
	NT_STATUS_PIPE_BUSY:                                                   ERROR_PIPE_BUSY,
	NT_STATUS_ILLEGAL_FUNCTION:                                            ERROR_ILLEGAL_FUNCTION,
	NT_STATUS_PIPE_DISCONNECTED:                                           ERROR_PIPE_DISCONNECTED,
	NT_STATUS_PIPE_CLOSING:                                                ERROR_PIPE_CLOSING,
	NT_STATUS_PIPE_CONNECTED:                                              ERROR_PIPE_CONNECTED,
	NT_STATUS_PIPE_LISTENING:                                              ERROR_PIPE_LISTENING,
	NT_STATUS_INVALID_READ_MODE:                                           ERROR_INVALID_READ_MODE,
	NT_STATUS_IO_TIMEOUT:                                                  ERROR_IO_TIMEOUT,
	NT_STATUS_FILE_FORCED_CLOSED:                                          ERROR_FILE_FORCED_CLOSED,
	NT_STATUS_PROFILING_NOT_STARTED:                                       ERROR_PROFILING_NOT_STARTED,
	NT_STATUS_PROFILING_NOT_STOPPED:                                       ERROR_PROFILING_NOT_STOPPED,
	NT_STATUS_COULD_NOT_INTERPRET:                                         ERROR_COULD_NOT_INTERPRET,
	NT_STATUS_FILE_IS_A_DIRECTORY:                                         ERROR_FILE_IS_A_DIRECTORY,
	NT_STATUS_NOT_SUPPORTED:                                               ERROR_NOT_SUPPORTED,
	NT_STATUS_REMOTE_NOT_LISTENING:                                        ERROR_REMOTE_NOT_LISTENING,
	NT_STATUS_DUPLICATE_NAME:                                              ERROR_DUPLICATE_NAME,
	NT_STATUS_BAD_NETWORK_PATH:                                            ERROR_BAD_NETWORK_PATH,
	NT_STATUS_NETWORK_BUSY:                                                ERROR_NETWORK_BUSY,
	NT_STATUS_DEVICE_DOES_NOT_EXIST:                                       ERROR_DEVICE_DOES_NOT_EXIST,
	NT_STATUS_TOO_MANY_COMMANDS:                                           ERROR_TOO_MANY_COMMANDS,
	NT_STATUS_ADAPTER_HARDWARE_ERROR:                                      ERROR_ADAPTER_HARDWARE_ERROR,
	NT_STATUS_INVALID_NETWORK_RESPONSE:                                    ERROR_INVALID_NETWORK_RESPONSE,
	NT_STATUS_UNEXPECTED_NETWORK_ERROR:                                    ERROR_UNEXPECTED_NETWORK_ERROR,
	NT_STATUS_BAD_REMOTE_ADAPTER:                                          ERROR_BAD_REMOTE_ADAPTER,
	NT_STATUS_PRINT_QUEUE_FULL:                                            ERROR_PRINT_QUEUE_FULL,
	NT_STATUS_NO_SPOOL_SPACE:                                              ERROR_NO_SPOOL_SPACE,
	NT_STATUS_PRINT_CANCELLED:                                             ERROR_PRINT_CANCELLED,
	NT_STATUS_NETWORK_NAME_DELETED:                                        ERROR_NETWORK_NAME_DELETED,
	NT_STATUS_NETWORK_ACCESS_DENIED:                                       ERROR_NETWORK_ACCESS_DENIED,
	NT_STATUS_BAD_DEVICE_TYPE:                                             ERROR_BAD_DEVICE_TYPE,
	NT_STATUS_BAD_NETWORK_NAME:                                            ERROR_BAD_NETWORK_NAME,
	NT_STATUS_TOO_MANY_NAMES:                                              ERROR_TOO_MANY_NAMES,
	NT_STATUS_TOO_MANY_SESSIONS:                                           ERROR_TOO_MANY_SESSIONS,
	NT_STATUS_SHARING_PAUSED:                                              ERROR_SHARING_PAUSED,
	NT_STATUS_REQUEST_NOT_ACCEPTED:                                        ERROR_REQUEST_NOT_ACCEPTED,
	NT_STATUS_REDIRECTOR_PAUSED:                                           ERROR_REDIRECTOR_PAUSED,
	NT_STATUS_NET_WRITE_FAULT:                                             ERROR_NET_WRITE_FAULT,
	NT_STATUS_PROFILING_AT_LIMIT:                                          ERROR_PROFILING_AT_LIMIT,
	NT_STATUS_NOT_SAME_DEVICE:                                             ERROR_NOT_SAME_DEVICE,
	NT_STATUS_FILE_RENAMED:                                                ERROR_FILE_RENAMED,
	NT_STATUS_VIRTUAL_CIRCUIT_CLOSED:                                      ERROR_VIRTUAL_CIRCUIT_CLOSED,
	NT_STATUS_NO_SECURITY_ON_OBJECT:                                       ERROR_NO_SECURITY_ON_OBJECT,
	NT_STATUS_CANT_WAIT:                                                   ERROR_CANT_WAIT,
	NT_STATUS_PIPE_EMPTY:                                                  ERROR_PIPE_EMPTY,
	NT_STATUS_CANT_ACCESS_DOMAIN_INFO:                                     ERROR_CANT_ACCESS_DOMAIN_INFO,
	NT_STATUS_CANT_TERMINATE_SELF:                                         ERROR_CANT_TERMINATE_SELF,
	NT_STATUS_INVALID_SERVER_STATE:                                        ERROR_INVALID_SERVER_STATE,
	NT_STATUS_INVALID_DOMAIN_STATE:                                        ERROR_INVALID_DOMAIN_STATE,
	NT_STATUS_INVALID_DOMAIN_ROLE:                                         ERROR_INVALID_DOMAIN_ROLE,
	NT_STATUS_NO_SUCH_DOMAIN:                                              ERROR_NO_SUCH_DOMAIN,
	NT_STATUS_DOMAIN_EXISTS:                                               ERROR_DOMAIN_EXISTS,
	NT_STATUS_DOMAIN_LIMIT_EXCEEDED:                                       ERROR_DOMAIN_LIMIT_EXCEEDED,
	NT_STATUS_OPLOCK_NOT_GRANTED:                                          ERROR_OPLOCK_NOT_GRANTED,
	NT_STATUS_INVALID_OPLOCK_PROTOCOL:                                     ERROR_INVALID_OPLOCK_PROTOCOL,
	NT_STATUS_INTERNAL_DB_CORRUPTION:                                      ERROR_INTERNAL_DB_CORRUPTION,
	NT_STATUS_INTERNAL_ERROR:                                              ERROR_INTERNAL_ERROR,
	NT_STATUS_GENERIC_NOT_MAPPED:                                          ERROR_GENERIC_NOT_MAPPED,
	NT_STATUS_BAD_DESCRIPTOR_FORMAT:                                       ERROR_BAD_DESCRIPTOR_FORMAT,
	NT_STATUS_INVALID_USER_BUFFER:                                         ERROR_INVALID_USER_BUFFER,
	NT_STATUS_UNEXPECTED_IO_ERROR:                                         ERROR_UNEXPECTED_IO_ERROR,
	NT_STATUS_UNEXPECTED_MM_CREATE_ERR:                                    ERROR_UNEXPECTED_MM_CREATE_ERR,
	NT_STATUS_UNEXPECTED_MM_MAP_ERROR:                                     ERROR_UNEXPECTED_MM_MAP_ERROR,
	NT_STATUS_UNEXPECTED_MM_EXTEND_ERR:                                    ERROR_UNEXPECTED_MM_EXTEND_ERR,
	NT_STATUS_NOT_LOGON_PROCESS:                                           ERROR_NOT_LOGON_PROCESS,
	NT_STATUS_LOGON_SESSION_EXISTS:                                        ERROR_LOGON_SESSION_EXISTS,
	NT_STATUS_INVALID_PARAMETER_1:                                         ERROR_INVALID_PARAMETER_1,
	NT_STATUS_INVALID_PARAMETER_2:                                         ERROR_INVALID_PARAMETER_2,
	NT_STATUS_INVALID_PARAMETER_3:                                         ERROR_INVALID_PARAMETER_3,
	NT_STATUS_INVALID_PARAMETER_4:                                         ERROR_INVALID_PARAMETER_4,
	NT_STATUS_INVALID_PARAMETER_5:                                         ERROR_INVALID_PARAMETER_5,
	NT_STATUS_INVALID_PARAMETER_6:                                         ERROR_INVALID_PARAMETER_6,
	NT_STATUS_INVALID_PARAMETER_7:                                         ERROR_INVALID_PARAMETER_7,
	NT_STATUS_INVALID_PARAMETER_8:                                         ERROR_INVALID_PARAMETER_8,
	NT_STATUS_INVALID_PARAMETER_9:                                         ERROR_INVALID_PARAMETER_9,
	NT_STATUS_INVALID_PARAMETER_10:                                        ERROR_INVALID_PARAMETER_10,
	NT_STATUS_INVALID_PARAMETER_11:                                        ERROR_INVALID_PARAMETER_11,
	NT_STATUS_INVALID_PARAMETER_12:                                        ERROR_INVALID_PARAMETER_12,
	NT_STATUS_REDIRECTOR_NOT_STARTED:                                      ERROR_REDIRECTOR_NOT_STARTED,
	NT_STATUS_REDIRECTOR_STARTED:                                          ERROR_REDIRECTOR_STARTED,
	NT_STATUS_STACK_OVERFLOW:                                              ERROR_STACK_OVERFLOW,
	NT_STATUS_NO_SUCH_PACKAGE:                                             ERROR_NO_SUCH_PACKAGE,
	NT_STATUS_BAD_FUNCTION_TABLE:                                          ERROR_BAD_FUNCTION_TABLE,
	NT_STATUS_VARIABLE_NOT_FOUND:                                          ERROR_VARIABLE_NOT_FOUND,
	NT_STATUS_DIRECTORY_NOT_EMPTY:                                         ERROR_DIRECTORY_NOT_EMPTY,
	NT_STATUS_FILE_CORRUPT_ERROR:                                          ERROR_FILE_CORRUPT_ERROR,
	NT_STATUS_NOT_A_DIRECTORY:                                             ERROR_NOT_A_DIRECTORY,
	NT_STATUS_BAD_LOGON_SESSION_STATE:                                     ERROR_BAD_LOGON_SESSION_STATE,
	NT_STATUS_LOGON_SESSION_COLLISION:                                     ERROR_LOGON_SESSION_COLLISION,
	NT_STATUS_NAME_TOO_LONG:                                               ERROR_NAME_TOO_LONG,
	NT_STATUS_FILES_OPEN:                                                  ERROR_FILES_OPEN,
	NT_STATUS_CONNECTION_IN_USE:                                           ERROR_CONNECTION_IN_USE,
	NT_STATUS_MESSAGE_NOT_FOUND:                                           ERROR_MESSAGE_NOT_FOUND,
	NT_STATUS_PROCESS_IS_TERMINATING:                                      ERROR_PROCESS_IS_TERMINATING,
	NT_STATUS_INVALID_LOGON_TYPE:                                          ERROR_INVALID_LOGON_TYPE,
	NT_STATUS_NO_GUID_TRANSLATION:                                         ERROR_NO_GUID_TRANSLATION,
	NT_STATUS_CANNOT_IMPERSONATE:                                          ERROR_CANNOT_IMPERSONATE,
	NT_STATUS_IMAGE_ALREADY_LOADED:                                        ERROR_IMAGE_ALREADY_LOADED,
	NT_STATUS_NO_LDT:                                                      ERROR_NO_LDT,
	NT_STATUS_INVALID_LDT_SIZE:                                            ERROR_INVALID_LDT_SIZE,
	NT_STATUS_INVALID_LDT_OFFSET:                                          ERROR_INVALID_LDT_OFFSET,
	NT_STATUS_INVALID_LDT_DESCRIPTOR:                                      ERROR_INVALID_LDT_DESCRIPTOR,
	NT_STATUS_INVALID_IMAGE_NE_FORMAT:                                     ERROR_INVALID_IMAGE_NE_FORMAT,
	NT_STATUS_RXACT_INVALID_STATE:                                         ERROR_RXACT_INVALID_STATE,
	NT_STATUS_RXACT_COMMIT_FAILURE:                                        ERROR_RXACT_COMMIT_FAILURE,
	NT_STATUS_MAPPED_FILE_SIZE_ZERO:                                       ERROR_MAPPED_FILE_SIZE_ZERO,
	NT_STATUS_TOO_MANY_OPENED_FILES:                                       ERROR_TOO_MANY_OPENED_FILES,
	NT_STATUS_CANCELLED:                                                   ERROR_CANCELLED,
	NT_STATUS_CANNOT_DELETE:                                               ERROR_CANNOT_DELETE,
	NT_STATUS_INVALID_COMPUTER_NAME:                                       ERROR_INVALID_COMPUTER_NAME,
	NT_STATUS_FILE_DELETED:                                                ERROR_FILE_DELETED,
	NT_STATUS_SPECIAL_ACCOUNT:                                             ERROR_SPECIAL_ACCOUNT,
	NT_STATUS_SPECIAL_GROUP:                                               ERROR_SPECIAL_GROUP,
	NT_STATUS_SPECIAL_USER:                                                ERROR_SPECIAL_USER,
	NT_STATUS_MEMBERS_PRIMARY_GROUP:                                       ERROR_MEMBERS_PRIMARY_GROUP,
	NT_STATUS_FILE_CLOSED:                                                 ERROR_FILE_CLOSED,
	NT_STATUS_TOO_MANY_THREADS:                                            ERROR_TOO_MANY_THREADS,
	NT_STATUS_THREAD_NOT_IN_PROCESS:                                       ERROR_THREAD_NOT_IN_PROCESS,
	NT_STATUS_TOKEN_ALREADY_IN_USE:                                        ERROR_TOKEN_ALREADY_IN_USE,
	NT_STATUS_PAGEFILE_QUOTA_EXCEEDED:                                     ERROR_PAGEFILE_QUOTA_EXCEEDED,
	NT_STATUS_COMMITMENT_LIMIT:                                            ERROR_COMMITMENT_LIMIT,
	NT_STATUS_INVALID_IMAGE_LE_FORMAT:                                     ERROR_INVALID_IMAGE_LE_FORMAT,
	NT_STATUS_INVALID_IMAGE_NOT_MZ:                                        ERROR_INVALID_IMAGE_NOT_MZ,
	NT_STATUS_INVALID_IMAGE_PROTECT:                                       ERROR_INVALID_IMAGE_PROTECT,
	NT_STATUS_INVALID_IMAGE_WIN_16:                                        ERROR_INVALID_IMAGE_WIN_16,
	NT_STATUS_LOGON_SERVER_CONFLICT:                                       ERROR_LOGON_SERVER_CONFLICT,
	NT_STATUS_TIME_DIFFERENCE_AT_DC:                                       ERROR_TIME_DIFFERENCE_AT_DC,
	NT_STATUS_SYNCHRONIZATION_REQUIRED:                                    ERROR_SYNCHRONIZATION_REQUIRED,
	NT_STATUS_DLL_NOT_FOUND:                                               ERROR_DLL_NOT_FOUND,
	NT_STATUS_OPEN_FAILED:                                                 ERROR_OPEN_FAILED,
	NT_STATUS_IO_PRIVILEGE_FAILED:                                         ERROR_IO_PRIVILEGE_FAILED,
	NT_STATUS_ORDINAL_NOT_FOUND:                                           ERROR_ORDINAL_NOT_FOUND,
	NT_STATUS_ENTRYPOINT_NOT_FOUND:                                        ERROR_ENTRYPOINT_NOT_FOUND,
	NT_STATUS_CONTROL_C_EXIT:                                              ERROR_CONTROL_C_EXIT,
	NT_STATUS_LOCAL_DISCONNECT:                                            ERROR_LOCAL_DISCONNECT,
	NT_STATUS_REMOTE_DISCONNECT:                                           ERROR_REMOTE_DISCONNECT,
	NT_STATUS_REMOTE_RESOURCES:                                            ERROR_REMOTE_RESOURCES,
	NT_STATUS_LINK_FAILED:                                                 ERROR_LINK_FAILED,
	NT_STATUS_LINK_TIMEOUT:                                                ERROR_LINK_TIMEOUT,
	NT_STATUS_INVALID_CONNECTION:                                          ERROR_INVALID_CONNECTION,
	NT_STATUS_INVALID_ADDRESS:                                             ERROR_INVALID_ADDRESS,
	NT_STATUS_DLL_INIT_FAILED:                                             ERROR_DLL_INIT_FAILED,
	NT_STATUS_MISSING_SYSTEMFILE:                                          ERROR_MISSING_SYSTEMFILE,
	NT_STATUS_UNHANDLED_EXCEPTION:                                         ERROR_UNHANDLED_EXCEPTION,
	NT_STATUS_APP_INIT_FAILURE:                                            ERROR_APP_INIT_FAILURE,
	NT_STATUS_PAGEFILE_CREATE_FAILED:                                      ERROR_PAGEFILE_CREATE_FAILED,
	NT_STATUS_NO_PAGEFILE:                                                 ERROR_NO_PAGEFILE,
	NT_STATUS_INVALID_LEVEL:                                               ERROR_INVALID_LEVEL,
	NT_STATUS_WRONG_PASSWORD_CORE:                                         ERROR_WRONG_PASSWORD_CORE,
	NT_STATUS_ILLEGAL_FLOAT_CONTEXT:                                       ERROR_ILLEGAL_FLOAT_CONTEXT,
	NT_STATUS_PIPE_BROKEN:                                                 ERROR_PIPE_BROKEN,
	NT_STATUS_REGISTRY_CORRUPT:                                            ERROR_REGISTRY_CORRUPT,
	NT_STATUS_REGISTRY_IO_FAILED:                                          ERROR_REGISTRY_IO_FAILED,
	NT_STATUS_NO_EVENT_PAIR:                                               ERROR_NO_EVENT_PAIR,
	NT_STATUS_UNRECOGNIZED_VOLUME:                                         ERROR_UNRECOGNIZED_VOLUME,
	NT_STATUS_SERIAL_NO_DEVICE_INITED:                                     ERROR_SERIAL_NO_DEVICE_INITED,
	NT_STATUS_NO_SUCH_ALIAS:                                               ERROR_NO_SUCH_ALIAS,
	NT_STATUS_MEMBER_NOT_IN_ALIAS:                                         ERROR_MEMBER_NOT_IN_ALIAS,
	NT_STATUS_MEMBER_IN_ALIAS:                                             ERROR_MEMBER_IN_ALIAS,
	NT_STATUS_ALIAS_EXISTS:                                                ERROR_ALIAS_EXISTS,
	NT_STATUS_LOGON_NOT_GRANTED:                                           ERROR_LOGON_NOT_GRANTED,
	NT_STATUS_TOO_MANY_SECRETS:                                            ERROR_TOO_MANY_SECRETS,
	NT_STATUS_SECRET_TOO_LONG:                                             ERROR_SECRET_TOO_LONG,
	NT_STATUS_INTERNAL_DB_ERROR:                                           ERROR_INTERNAL_DB_ERROR,
	NT_STATUS_FULLSCREEN_MODE:                                             ERROR_FULLSCREEN_MODE,
	NT_STATUS_TOO_MANY_CONTEXT_IDS:                                        ERROR_TOO_MANY_CONTEXT_IDS,
	NT_STATUS_LOGON_TYPE_NOT_GRANTED:                                      ERROR_LOGON_TYPE_NOT_GRANTED,
	NT_STATUS_NOT_REGISTRY_FILE:                                           ERROR_NOT_REGISTRY_FILE,
	NT_STATUS_NT_CROSS_ENCRYPTION_REQUIRED:                                ERROR_NT_CROSS_ENCRYPTION_REQUIRED,
	NT_STATUS_DOMAIN_CTRLR_CONFIG_ERROR:                                   ERROR_DOMAIN_CTRLR_CONFIG_ERROR,
	NT_STATUS_FT_MISSING_MEMBER:                                           ERROR_FT_MISSING_MEMBER,
	NT_STATUS_ILL_FORMED_SERVICE_ENTRY:                                    ERROR_ILL_FORMED_SERVICE_ENTRY,
	NT_STATUS_ILLEGAL_CHARACTER:                                           ERROR_ILLEGAL_CHARACTER,
	NT_STATUS_UNMAPPABLE_CHARACTER:                                        ERROR_UNMAPPABLE_CHARACTER,
	NT_STATUS_UNDEFINED_CHARACTER:                                         ERROR_UNDEFINED_CHARACTER,
	NT_STATUS_FLOPPY_VOLUME:                                               ERROR_FLOPPY_VOLUME,
	NT_STATUS_FLOPPY_ID_MARK_NOT_FOUND:                                    ERROR_FLOPPY_ID_MARK_NOT_FOUND,
	NT_STATUS_FLOPPY_WRONG_CYLINDER:                                       ERROR_FLOPPY_WRONG_CYLINDER,
	NT_STATUS_FLOPPY_UNKNOWN_ERROR:                                        ERROR_FLOPPY_UNKNOWN_ERROR,
	NT_STATUS_FLOPPY_BAD_REGISTERS:                                        ERROR_FLOPPY_BAD_REGISTERS,
	NT_STATUS_DISK_RECALIBRATE_FAILED:                                     ERROR_DISK_RECALIBRATE_FAILED,
	NT_STATUS_DISK_OPERATION_FAILED:                                       ERROR_DISK_OPERATION_FAILED,
	NT_STATUS_DISK_RESET_FAILED:                                           ERROR_DISK_RESET_FAILED,
	NT_STATUS_SHARED_IRQ_BUSY:                                             ERROR_SHARED_IRQ_BUSY,
	NT_STATUS_FT_ORPHANING:                                                ERROR_FT_ORPHANING,
	NT_STATUS_BIOS_FAILED_TO_CONNECT_INTERRUPT:                            ERROR_BIOS_FAILED_TO_CONNECT_INTERRUPT,
	NT_STATUS_PARTITION_FAILURE:                                           ERROR_PARTITION_FAILURE,
	NT_STATUS_INVALID_BLOCK_LENGTH:                                        ERROR_INVALID_BLOCK_LENGTH,
	NT_STATUS_DEVICE_NOT_PARTITIONED:                                      ERROR_DEVICE_NOT_PARTITIONED,
	NT_STATUS_UNABLE_TO_LOCK_MEDIA:                                        ERROR_UNABLE_TO_LOCK_MEDIA,
	NT_STATUS_UNABLE_TO_UNLOAD_MEDIA:                                      ERROR_UNABLE_TO_UNLOAD_MEDIA,
	NT_STATUS_EOM_OVERFLOW:                                                ERROR_EOM_OVERFLOW,
	NT_STATUS_NO_MEDIA:                                                    ERROR_NO_MEDIA,
	NT_STATUS_NO_SUCH_MEMBER:                                              ERROR_NO_SUCH_MEMBER,
	NT_STATUS_INVALID_MEMBER:                                              ERROR_INVALID_MEMBER,
	NT_STATUS_KEY_DELETED:                                                 ERROR_KEY_DELETED,
	NT_STATUS_NO_LOG_SPACE:                                                ERROR_NO_LOG_SPACE,
	NT_STATUS_TOO_MANY_SIDS:                                               ERROR_TOO_MANY_SIDS,
	NT_STATUS_LM_CROSS_ENCRYPTION_REQUIRED:                                ERROR_LM_CROSS_ENCRYPTION_REQUIRED,
	NT_STATUS_KEY_HAS_CHILDREN:                                            ERROR_KEY_HAS_CHILDREN,
	NT_STATUS_CHILD_MUST_BE_VOLATILE:                                      ERROR_CHILD_MUST_BE_VOLATILE,
	NT_STATUS_DEVICE_CONFIGURATION_ERROR:                                  ERROR_DEVICE_CONFIGURATION_ERROR,
	NT_STATUS_DRIVER_INTERNAL_ERROR:                                       ERROR_DRIVER_INTERNAL_ERROR,
	NT_STATUS_INVALID_DEVICE_STATE:                                        ERROR_INVALID_DEVICE_STATE,
	NT_STATUS_IO_DEVICE_ERROR:                                             ERROR_IO_DEVICE_ERROR,
	NT_STATUS_DEVICE_PROTOCOL_ERROR:                                       ERROR_DEVICE_PROTOCOL_ERROR,
	NT_STATUS_BACKUP_CONTROLLER:                                           ERROR_BACKUP_CONTROLLER,
	NT_STATUS_LOG_FILE_FULL:                                               ERROR_LOG_FILE_FULL,
	NT_STATUS_TOO_LATE:                                                    ERROR_TOO_LATE,
	NT_STATUS_NO_TRUST_LSA_SECRET:                                         ERROR_NO_TRUST_LSA_SECRET,
	NT_STATUS_NO_TRUST_SAM_ACCOUNT:                                        ERROR_NO_TRUST_SAM_ACCOUNT,
	NT_STATUS_TRUSTED_DOMAIN_FAILURE:                                      ERROR_TRUSTED_DOMAIN_FAILURE,
	NT_STATUS_TRUSTED_RELATIONSHIP_FAILURE:                                ERROR_TRUSTED_RELATIONSHIP_FAILURE,
	NT_STATUS_EVENTLOG_FILE_CORRUPT:                                       ERROR_EVENTLOG_FILE_CORRUPT,
	NT_STATUS_EVENTLOG_CANT_START:                                         ERROR_EVENTLOG_CANT_START,
	NT_STATUS_TRUST_FAILURE:                                               ERROR_TRUST_FAILURE,
	NT_STATUS_MUTANT_LIMIT_EXCEEDED:                                       ERROR_MUTANT_LIMIT_EXCEEDED,
	NT_STATUS_NETLOGON_NOT_STARTED:                                        ERROR_NETLOGON_NOT_STARTED,
	NT_STATUS_ACCOUNT_EXPIRED:                                             ERROR_ACCOUNT_EXPIRED,
	NT_STATUS_POSSIBLE_DEADLOCK:                                           ERROR_POSSIBLE_DEADLOCK,
	NT_STATUS_NETWORK_CREDENTIAL_CONFLICT:                                 ERROR_NETWORK_CREDENTIAL_CONFLICT,
	NT_STATUS_REMOTE_SESSION_LIMIT:                                        ERROR_REMOTE_SESSION_LIMIT,
	NT_STATUS_EVENTLOG_FILE_CHANGED:                                       ERROR_EVENTLOG_FILE_CHANGED,
	NT_STATUS_NOLOGON_INTERDOMAIN_TRUST_ACCOUNT:                           ERROR_NOLOGON_INTERDOMAIN_TRUST_ACCOUNT,
	NT_STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT:                           ERROR_NOLOGON_WORKSTATION_TRUST_ACCOUNT,
	NT_STATUS_NOLOGON_SERVER_TRUST_ACCOUNT:                                ERROR_NOLOGON_SERVER_TRUST_ACCOUNT,
	NT_STATUS_DOMAIN_TRUST_INCONSISTENT:                                   ERROR_DOMAIN_TRUST_INCONSISTENT,
	NT_STATUS_FS_DRIVER_REQUIRED:                                          ERROR_FS_DRIVER_REQUIRED,
	NT_STATUS_IMAGE_ALREADY_LOADED_AS_DLL:                                 ERROR_IMAGE_ALREADY_LOADED_AS_DLL,
	NT_STATUS_INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING:        ERROR_INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING,
	NT_STATUS_SHORT_NAMES_NOT_ENABLED_ON_VOLUME:                           ERROR_SHORT_NAMES_NOT_ENABLED_ON_VOLUME,
	NT_STATUS_SECURITY_STREAM_IS_INCONSISTENT:                             ERROR_SECURITY_STREAM_IS_INCONSISTENT,
	NT_STATUS_INVALID_LOCK_RANGE:                                          ERROR_INVALID_LOCK_RANGE,
	NT_STATUS_INVALID_ACE_CONDITION:                                       ERROR_INVALID_ACE_CONDITION,
	NT_STATUS_IMAGE_SUBSYSTEM_NOT_PRESENT:                                 ERROR_IMAGE_SUBSYSTEM_NOT_PRESENT,
	NT_STATUS_NOTIFICATION_GUID_ALREADY_DEFINED:                           ERROR_NOTIFICATION_GUID_ALREADY_DEFINED,
	NT_STATUS_NETWORK_OPEN_RESTRICTION:                                    ERROR_NETWORK_OPEN_RESTRICTION,
	NT_STATUS_NO_USER_SESSION_KEY:                                         ERROR_NO_USER_SESSION_KEY,
	NT_STATUS_USER_SESSION_DELETED:                                        ERROR_USER_SESSION_DELETED,
	NT_STATUS_RESOURCE_LANG_NOT_FOUND:                                     ERROR_RESOURCE_LANG_NOT_FOUND,
	NT_STATUS_INSUFF_SERVER_RESOURCES:                                     ERROR_INSUFF_SERVER_RESOURCES,
	NT_STATUS_INVALID_BUFFER_SIZE:                                         ERROR_INVALID_BUFFER_SIZE,
	NT_STATUS_INVALID_ADDRESS_COMPONENT:                                   ERROR_INVALID_ADDRESS_COMPONENT,
	NT_STATUS_INVALID_ADDRESS_WILDCARD:                                    ERROR_INVALID_ADDRESS_WILDCARD,
	NT_STATUS_TOO_MANY_ADDRESSES:                                          ERROR_TOO_MANY_ADDRESSES,
	NT_STATUS_ADDRESS_ALREADY_EXISTS:                                      ERROR_ADDRESS_ALREADY_EXISTS,
	NT_STATUS_ADDRESS_CLOSED:                                              ERROR_ADDRESS_CLOSED,
	NT_STATUS_CONNECTION_DISCONNECTED:                                     ERROR_CONNECTION_DISCONNECTED,
	NT_STATUS_CONNECTION_RESET:                                            ERROR_CONNECTION_RESET,
	NT_STATUS_TOO_MANY_NODES:                                              ERROR_TOO_MANY_NODES,
	NT_STATUS_TRANSACTION_ABORTED:                                         ERROR_TRANSACTION_ABORTED,
	NT_STATUS_TRANSACTION_TIMED_OUT:                                       ERROR_TRANSACTION_TIMED_OUT,
	NT_STATUS_TRANSACTION_NO_RELEASE:                                      ERROR_TRANSACTION_NO_RELEASE,
	NT_STATUS_TRANSACTION_NO_MATCH:                                        ERROR_TRANSACTION_NO_MATCH,
	NT_STATUS_TRANSACTION_RESPONDED:                                       ERROR_TRANSACTION_RESPONDED,
	NT_STATUS_TRANSACTION_INVALID_ID:                                      ERROR_TRANSACTION_INVALID_ID,
	NT_STATUS_TRANSACTION_INVALID_TYPE:                                    ERROR_TRANSACTION_INVALID_TYPE,
	NT_STATUS_NOT_SERVER_SESSION:                                          ERROR_NOT_SERVER_SESSION,
	NT_STATUS_NOT_CLIENT_SESSION:                                          ERROR_NOT_CLIENT_SESSION,
	NT_STATUS_CANNOT_LOAD_REGISTRY_FILE:                                   ERROR_CANNOT_LOAD_REGISTRY_FILE,
	NT_STATUS_DEBUG_ATTACH_FAILED:                                         ERROR_DEBUG_ATTACH_FAILED,
	NT_STATUS_SYSTEM_PROCESS_TERMINATED:                                   ERROR_SYSTEM_PROCESS_TERMINATED,
	NT_STATUS_DATA_NOT_ACCEPTED:                                           ERROR_DATA_NOT_ACCEPTED,
	NT_STATUS_NO_BROWSER_SERVERS_FOUND:                                    ERROR_NO_BROWSER_SERVERS_FOUND,
	NT_STATUS_VDM_HARD_ERROR:                                              ERROR_VDM_HARD_ERROR,
	NT_STATUS_DRIVER_CANCEL_TIMEOUT:                                       ERROR_DRIVER_CANCEL_TIMEOUT,
	NT_STATUS_REPLY_MESSAGE_MISMATCH:                                      ERROR_REPLY_MESSAGE_MISMATCH,
	NT_STATUS_MAPPED_ALIGNMENT:                                            ERROR_MAPPED_ALIGNMENT,
	NT_STATUS_IMAGE_CHECKSUM_MISMATCH:                                     ERROR_IMAGE_CHECKSUM_MISMATCH,
	NT_STATUS_LOST_WRITEBEHIND_DATA:                                       ERROR_LOST_WRITEBEHIND_DATA,
	NT_STATUS_CLIENT_SERVER_PARAMETERS_INVALID:                            ERROR_CLIENT_SERVER_PARAMETERS_INVALID,
	NT_STATUS_PASSWORD_MUST_CHANGE:                                        ERROR_PASSWORD_MUST_CHANGE,
	NT_STATUS_NOT_FOUND:                                                   ERROR_NOT_FOUND,
	NT_STATUS_NOT_TINY_STREAM:                                             ERROR_NOT_TINY_STREAM,
	NT_STATUS_RECOVERY_FAILURE:                                            ERROR_RECOVERY_FAILURE,
	NT_STATUS_STACK_OVERFLOW_READ:                                         ERROR_STACK_OVERFLOW_READ,
	NT_STATUS_FAIL_CHECK:                                                  ERROR_FAIL_CHECK,
	NT_STATUS_DUPLICATE_OBJECTID:                                          ERROR_DUPLICATE_OBJECTID,
	NT_STATUS_OBJECTID_EXISTS:                                             ERROR_OBJECTID_EXISTS,
	NT_STATUS_CONVERT_TO_LARGE:                                            ERROR_CONVERT_TO_LARGE,
	NT_STATUS_RETRY:                                                       ERROR_RETRY,
	NT_STATUS_FOUND_OUT_OF_SCOPE:                                          ERROR_FOUND_OUT_OF_SCOPE,
	NT_STATUS_ALLOCATE_BUCKET:                                             ERROR_ALLOCATE_BUCKET,
	NT_STATUS_PROPSET_NOT_FOUND:                                           ERROR_PROPSET_NOT_FOUND,
	NT_STATUS_MARSHALL_OVERFLOW:                                           ERROR_MARSHALL_OVERFLOW,
	NT_STATUS_INVALID_VARIANT:                                             ERROR_INVALID_VARIANT,
	NT_STATUS_DOMAIN_CONTROLLER_NOT_FOUND:                                 ERROR_DOMAIN_CONTROLLER_NOT_FOUND,
	NT_STATUS_ACCOUNT_LOCKED_OUT:                                          ERROR_ACCOUNT_LOCKED_OUT,
	NT_STATUS_HANDLE_NOT_CLOSABLE:                                         ERROR_HANDLE_NOT_CLOSABLE,
	NT_STATUS_CONNECTION_REFUSED:                                          ERROR_CONNECTION_REFUSED,
	NT_STATUS_GRACEFUL_DISCONNECT:                                         ERROR_GRACEFUL_DISCONNECT,
	NT_STATUS_ADDRESS_ALREADY_ASSOCIATED:                                  ERROR_ADDRESS_ALREADY_ASSOCIATED,
	NT_STATUS_ADDRESS_NOT_ASSOCIATED:                                      ERROR_ADDRESS_NOT_ASSOCIATED,
	NT_STATUS_CONNECTION_INVALID:                                          ERROR_CONNECTION_INVALID,
	NT_STATUS_CONNECTION_ACTIVE:                                           ERROR_CONNECTION_ACTIVE,
	NT_STATUS_NETWORK_UNREACHABLE:                                         ERROR_NETWORK_UNREACHABLE,
	NT_STATUS_HOST_UNREACHABLE:                                            ERROR_HOST_UNREACHABLE,
	NT_STATUS_PROTOCOL_UNREACHABLE:                                        ERROR_PROTOCOL_UNREACHABLE,
	NT_STATUS_PORT_UNREACHABLE:                                            ERROR_PORT_UNREACHABLE,
	NT_STATUS_REQUEST_ABORTED:                                             ERROR_REQUEST_ABORTED,
	NT_STATUS_CONNECTION_ABORTED:                                          ERROR_CONNECTION_ABORTED,
	NT_STATUS_BAD_COMPRESSION_BUFFER:                                      ERROR_BAD_COMPRESSION_BUFFER,
	NT_STATUS_USER_MAPPED_FILE:                                            ERROR_USER_MAPPED_FILE,
	NT_STATUS_AUDIT_FAILED:                                                ERROR_AUDIT_FAILED,
	NT_STATUS_TIMER_RESOLUTION_NOT_SET:                                    ERROR_TIMER_RESOLUTION_NOT_SET,
	NT_STATUS_CONNECTION_COUNT_LIMIT:                                      ERROR_CONNECTION_COUNT_LIMIT,
	NT_STATUS_LOGIN_TIME_RESTRICTION:                                      ERROR_LOGIN_TIME_RESTRICTION,
	NT_STATUS_LOGIN_WKSTA_RESTRICTION:                                     ERROR_LOGIN_WKSTA_RESTRICTION,
	NT_STATUS_IMAGE_MP_UP_MISMATCH:                                        ERROR_IMAGE_MP_UP_MISMATCH,
	NT_STATUS_INSUFFICIENT_LOGON_INFO:                                     ERROR_INSUFFICIENT_LOGON_INFO,
	NT_STATUS_BAD_DLL_ENTRYPOINT:                                          ERROR_BAD_DLL_ENTRYPOINT,
	NT_STATUS_BAD_SERVICE_ENTRYPOINT:                                      ERROR_BAD_SERVICE_ENTRYPOINT,
	NT_STATUS_LPC_REPLY_LOST:                                              ERROR_LPC_REPLY_LOST,
	NT_STATUS_IP_ADDRESS_CONFLICT1:                                        ERROR_IP_ADDRESS_CONFLICT1,
	NT_STATUS_IP_ADDRESS_CONFLICT2:                                        ERROR_IP_ADDRESS_CONFLICT2,
	NT_STATUS_REGISTRY_QUOTA_LIMIT:                                        ERROR_REGISTRY_QUOTA_LIMIT,
	NT_STATUS_PATH_NOT_COVERED:                                            ERROR_PATH_NOT_COVERED,
	NT_STATUS_NO_CALLBACK_ACTIVE:                                          ERROR_NO_CALLBACK_ACTIVE,
	NT_STATUS_LICENSE_QUOTA_EXCEEDED:                                      ERROR_LICENSE_QUOTA_EXCEEDED,
	NT_STATUS_PWD_TOO_SHORT:                                               ERROR_PWD_TOO_SHORT,
	NT_STATUS_PWD_TOO_RECENT:                                              ERROR_PWD_TOO_RECENT,
	NT_STATUS_PWD_HISTORY_CONFLICT:                                        ERROR_PWD_HISTORY_CONFLICT,
	NT_STATUS_PLUGPLAY_NO_DEVICE:                                          ERROR_PLUGPLAY_NO_DEVICE,
	NT_STATUS_UNSUPPORTED_COMPRESSION:                                     ERROR_UNSUPPORTED_COMPRESSION,
	NT_STATUS_INVALID_HW_PROFILE:                                          ERROR_INVALID_HW_PROFILE,
	NT_STATUS_INVALID_PLUGPLAY_DEVICE_PATH:                                ERROR_INVALID_PLUGPLAY_DEVICE_PATH,
	NT_STATUS_DRIVER_ORDINAL_NOT_FOUND:                                    ERROR_DRIVER_ORDINAL_NOT_FOUND,
	NT_STATUS_DRIVER_ENTRYPOINT_NOT_FOUND:                                 ERROR_DRIVER_ENTRYPOINT_NOT_FOUND,
	NT_STATUS_RESOURCE_NOT_OWNED:                                          ERROR_RESOURCE_NOT_OWNED,
	NT_STATUS_TOO_MANY_LINKS:                                              ERROR_TOO_MANY_LINKS,
	NT_STATUS_QUOTA_LIST_INCONSISTENT:                                     ERROR_QUOTA_LIST_INCONSISTENT,
	NT_STATUS_FILE_IS_OFFLINE:                                             ERROR_FILE_IS_OFFLINE,
	NT_STATUS_EVALUATION_EXPIRATION:                                       ERROR_EVALUATION_EXPIRATION,
	NT_STATUS_ILLEGAL_DLL_RELOCATION:                                      ERROR_ILLEGAL_DLL_RELOCATION,
	NT_STATUS_LICENSE_VIOLATION:                                           ERROR_LICENSE_VIOLATION,
	NT_STATUS_DLL_INIT_FAILED_LOGOFF:                                      ERROR_DLL_INIT_FAILED_LOGOFF,
	NT_STATUS_DRIVER_UNABLE_TO_LOAD:                                       ERROR_DRIVER_UNABLE_TO_LOAD,
	NT_STATUS_DFS_UNAVAILABLE:                                             ERROR_DFS_UNAVAILABLE,
	NT_STATUS_VOLUME_DISMOUNTED:                                           ERROR_VOLUME_DISMOUNTED,
	NT_STATUS_WX86_INTERNAL_ERROR:                                         ERROR_WX86_INTERNAL_ERROR,
	NT_STATUS_WX86_FLOAT_STACK_CHECK:                                      ERROR_WX86_FLOAT_STACK_CHECK,
	NT_STATUS_VALIDATE_CONTINUE:                                           ERROR_VALIDATE_CONTINUE,
	NT_STATUS_NO_MATCH:                                                    ERROR_NO_MATCH,
	NT_STATUS_NO_MORE_MATCHES:                                             ERROR_NO_MORE_MATCHES,
	NT_STATUS_NOT_A_REPARSE_POINT:                                         ERROR_NOT_A_REPARSE_POINT,
	NT_STATUS_IO_REPARSE_TAG_INVALID:                                      ERROR_IO_REPARSE_TAG_INVALID,
	NT_STATUS_IO_REPARSE_TAG_MISMATCH:                                     ERROR_IO_REPARSE_TAG_MISMATCH,
	NT_STATUS_IO_REPARSE_DATA_INVALID:                                     ERROR_IO_REPARSE_DATA_INVALID,
	NT_STATUS_IO_REPARSE_TAG_NOT_HANDLED:                                  ERROR_IO_REPARSE_TAG_NOT_HANDLED,
	NT_STATUS_REPARSE_POINT_NOT_RESOLVED:                                  ERROR_REPARSE_POINT_NOT_RESOLVED,
	NT_STATUS_DIRECTORY_IS_A_REPARSE_POINT:                                ERROR_DIRECTORY_IS_A_REPARSE_POINT,
	NT_STATUS_RANGE_LIST_CONFLICT:                                         ERROR_RANGE_LIST_CONFLICT,
	NT_STATUS_SOURCE_ELEMENT_EMPTY:                                        ERROR_SOURCE_ELEMENT_EMPTY,
	NT_STATUS_DESTINATION_ELEMENT_FULL:                                    ERROR_DESTINATION_ELEMENT_FULL,
	NT_STATUS_ILLEGAL_ELEMENT_ADDRESS:                                     ERROR_ILLEGAL_ELEMENT_ADDRESS,
	NT_STATUS_MAGAZINE_NOT_PRESENT:                                        ERROR_MAGAZINE_NOT_PRESENT,
	NT_STATUS_REINITIALIZATION_NEEDED:                                     ERROR_REINITIALIZATION_NEEDED,
	NT_STATUS_ENCRYPTION_FAILED:                                           ERROR_ENCRYPTION_FAILED,
	NT_STATUS_DECRYPTION_FAILED:                                           ERROR_DECRYPTION_FAILED,
	NT_STATUS_RANGE_NOT_FOUND:                                             ERROR_RANGE_NOT_FOUND,
	NT_STATUS_NO_RECOVERY_POLICY:                                          ERROR_NO_RECOVERY_POLICY,
	NT_STATUS_NO_EFS:                                                      ERROR_NO_EFS,
	NT_STATUS_WRONG_EFS:                                                   ERROR_WRONG_EFS,
	NT_STATUS_NO_USER_KEYS:                                                ERROR_NO_USER_KEYS,
	NT_STATUS_FILE_NOT_ENCRYPTED:                                          ERROR_FILE_NOT_ENCRYPTED,
	NT_STATUS_NOT_EXPORT_FORMAT:                                           ERROR_NOT_EXPORT_FORMAT,
	NT_STATUS_FILE_ENCRYPTED:                                              ERROR_FILE_ENCRYPTED,
	NT_STATUS_WMI_GUID_NOT_FOUND:                                          ERROR_WMI_GUID_NOT_FOUND,
	NT_STATUS_WMI_INSTANCE_NOT_FOUND:                                      ERROR_WMI_INSTANCE_NOT_FOUND,
	NT_STATUS_WMI_ITEMID_NOT_FOUND:                                        ERROR_WMI_ITEMID_NOT_FOUND,
	NT_STATUS_WMI_TRY_AGAIN:                                               ERROR_WMI_TRY_AGAIN,
	NT_STATUS_SHARED_POLICY:                                               ERROR_SHARED_POLICY,
	NT_STATUS_POLICY_OBJECT_NOT_FOUND:                                     ERROR_POLICY_OBJECT_NOT_FOUND,
	NT_STATUS_POLICY_ONLY_IN_DS:                                           ERROR_POLICY_ONLY_IN_DS,
	NT_STATUS_VOLUME_NOT_UPGRADED:                                         ERROR_VOLUME_NOT_UPGRADED,
	NT_STATUS_REMOTE_STORAGE_NOT_ACTIVE:                                   ERROR_REMOTE_STORAGE_NOT_ACTIVE,
	NT_STATUS_REMOTE_STORAGE_MEDIA_ERROR:                                  ERROR_REMOTE_STORAGE_MEDIA_ERROR,
	NT_STATUS_NO_TRACKING_SERVICE:                                         ERROR_NO_TRACKING_SERVICE,
	NT_STATUS_SERVER_SID_MISMATCH:                                         ERROR_SERVER_SID_MISMATCH,
	NT_STATUS_DS_NO_ATTRIBUTE_OR_VALUE:                                    ERROR_DS_NO_ATTRIBUTE_OR_VALUE,
	NT_STATUS_DS_INVALID_ATTRIBUTE_SYNTAX:                                 ERROR_DS_INVALID_ATTRIBUTE_SYNTAX,
	NT_STATUS_DS_ATTRIBUTE_TYPE_UNDEFINED:                                 ERROR_DS_ATTRIBUTE_TYPE_UNDEFINED,
	NT_STATUS_DS_ATTRIBUTE_OR_VALUE_EXISTS:                                ERROR_DS_ATTRIBUTE_OR_VALUE_EXISTS,
	NT_STATUS_DS_BUSY:                                                     ERROR_DS_BUSY,
	NT_STATUS_DS_UNAVAILABLE:                                              ERROR_DS_UNAVAILABLE,
	NT_STATUS_DS_NO_RIDS_ALLOCATED:                                        ERROR_DS_NO_RIDS_ALLOCATED,
	NT_STATUS_DS_NO_MORE_RIDS:                                             ERROR_DS_NO_MORE_RIDS,
	NT_STATUS_DS_INCORRECT_ROLE_OWNER:                                     ERROR_DS_INCORRECT_ROLE_OWNER,
	NT_STATUS_DS_RIDMGR_INIT_ERROR:                                        ERROR_DS_RIDMGR_INIT_ERROR,
	NT_STATUS_DS_OBJ_CLASS_VIOLATION:                                      ERROR_DS_OBJ_CLASS_VIOLATION,
	NT_STATUS_DS_CANT_ON_NON_LEAF:                                         ERROR_DS_CANT_ON_NON_LEAF,
	NT_STATUS_DS_CANT_ON_RDN:                                              ERROR_DS_CANT_ON_RDN,
	NT_STATUS_DS_CANT_MOD_OBJ_CLASS:                                       ERROR_DS_CANT_MOD_OBJ_CLASS,
	NT_STATUS_DS_CROSS_DOM_MOVE_FAILED:                                    ERROR_DS_CROSS_DOM_MOVE_FAILED,
	NT_STATUS_DS_GC_NOT_AVAILABLE:                                         ERROR_DS_GC_NOT_AVAILABLE,
	NT_STATUS_DIRECTORY_SERVICE_REQUIRED:                                  ERROR_DIRECTORY_SERVICE_REQUIRED,
	NT_STATUS_REPARSE_ATTRIBUTE_CONFLICT:                                  ERROR_REPARSE_ATTRIBUTE_CONFLICT,
	NT_STATUS_CANT_ENABLE_DENY_ONLY:                                       ERROR_CANT_ENABLE_DENY_ONLY,
	NT_STATUS_FLOAT_MULTIPLE_FAULTS:                                       ERROR_FLOAT_MULTIPLE_FAULTS,
	NT_STATUS_FLOAT_MULTIPLE_TRAPS:                                        ERROR_FLOAT_MULTIPLE_TRAPS,
	NT_STATUS_DEVICE_REMOVED:                                              ERROR_DEVICE_REMOVED,
	NT_STATUS_JOURNAL_DELETE_IN_PROGRESS:                                  ERROR_JOURNAL_DELETE_IN_PROGRESS,
	NT_STATUS_JOURNAL_NOT_ACTIVE:                                          ERROR_JOURNAL_NOT_ACTIVE,
	NT_STATUS_NOINTERFACE:                                                 ERROR_NOINTERFACE,
	NT_STATUS_DS_ADMIN_LIMIT_EXCEEDED:                                     ERROR_DS_ADMIN_LIMIT_EXCEEDED,
	NT_STATUS_DRIVER_FAILED_SLEEP:                                         ERROR_DRIVER_FAILED_SLEEP,
	NT_STATUS_MUTUAL_AUTHENTICATION_FAILED:                                ERROR_MUTUAL_AUTHENTICATION_FAILED,
	NT_STATUS_CORRUPT_SYSTEM_FILE:                                         ERROR_CORRUPT_SYSTEM_FILE,
	NT_STATUS_DATATYPE_MISALIGNMENT_ERROR:                                 ERROR_DATATYPE_MISALIGNMENT_ERROR,
	NT_STATUS_WMI_READ_ONLY:                                               ERROR_WMI_READ_ONLY,
	NT_STATUS_WMI_SET_FAILURE:                                             ERROR_WMI_SET_FAILURE,
	NT_STATUS_COMMITMENT_MINIMUM:                                          ERROR_COMMITMENT_MINIMUM,
	NT_STATUS_REG_NAT_CONSUMPTION:                                         ERROR_REG_NAT_CONSUMPTION,
	NT_STATUS_TRANSPORT_FULL:                                              ERROR_TRANSPORT_FULL,
	NT_STATUS_DS_SAM_INIT_FAILURE:                                         ERROR_DS_SAM_INIT_FAILURE,
	NT_STATUS_ONLY_IF_CONNECTED:                                           ERROR_ONLY_IF_CONNECTED,
	NT_STATUS_DS_SENSITIVE_GROUP_VIOLATION:                                ERROR_DS_SENSITIVE_GROUP_VIOLATION,
	NT_STATUS_PNP_RESTART_ENUMERATION:                                     ERROR_PNP_RESTART_ENUMERATION,
	NT_STATUS_JOURNAL_ENTRY_DELETED:                                       ERROR_JOURNAL_ENTRY_DELETED,
	NT_STATUS_DS_CANT_MOD_PRIMARYGROUPID:                                  ERROR_DS_CANT_MOD_PRIMARYGROUPID,
	NT_STATUS_SYSTEM_IMAGE_BAD_SIGNATURE:                                  ERROR_SYSTEM_IMAGE_BAD_SIGNATURE,
	NT_STATUS_PNP_REBOOT_REQUIRED:                                         ERROR_PNP_REBOOT_REQUIRED,
	NT_STATUS_POWER_STATE_INVALID:                                         ERROR_POWER_STATE_INVALID,
	NT_STATUS_DS_INVALID_GROUP_TYPE:                                       ERROR_DS_INVALID_GROUP_TYPE,
	NT_STATUS_DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN:                       ERROR_DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN,
	NT_STATUS_DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN:                        ERROR_DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN,
	NT_STATUS_DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER:                            ERROR_DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER,
	NT_STATUS_DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER:                        ERROR_DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER,
	NT_STATUS_DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER:                         ERROR_DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER,
	NT_STATUS_DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER:                      ERROR_DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER,
	NT_STATUS_DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER:                 ERROR_DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER,
	NT_STATUS_DS_HAVE_PRIMARY_MEMBERS:                                     ERROR_DS_HAVE_PRIMARY_MEMBERS,
	NT_STATUS_WMI_NOT_SUPPORTED:                                           ERROR_WMI_NOT_SUPPORTED,
	NT_STATUS_INSUFFICIENT_POWER:                                          ERROR_INSUFFICIENT_POWER,
	NT_STATUS_SAM_NEED_BOOTKEY_PASSWORD:                                   ERROR_SAM_NEED_BOOTKEY_PASSWORD,
	NT_STATUS_SAM_NEED_BOOTKEY_FLOPPY:                                     ERROR_SAM_NEED_BOOTKEY_FLOPPY,
	NT_STATUS_DS_CANT_START:                                               ERROR_DS_CANT_START,
	NT_STATUS_DS_INIT_FAILURE:                                             ERROR_DS_INIT_FAILURE,
	NT_STATUS_SAM_INIT_FAILURE:                                            ERROR_SAM_INIT_FAILURE,
	NT_STATUS_DS_GC_REQUIRED:                                              ERROR_DS_GC_REQUIRED,
	NT_STATUS_DS_LOCAL_MEMBER_OF_LOCAL_ONLY:                               ERROR_DS_LOCAL_MEMBER_OF_LOCAL_ONLY,
	NT_STATUS_DS_NO_FPO_IN_UNIVERSAL_GROUPS:                               ERROR_DS_NO_FPO_IN_UNIVERSAL_GROUPS,
	NT_STATUS_DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED:                           ERROR_DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED,
	NT_STATUS_CURRENT_DOMAIN_NOT_ALLOWED:                                  ERROR_CURRENT_DOMAIN_NOT_ALLOWED,
	NT_STATUS_CANNOT_MAKE:                                                 ERROR_CANNOT_MAKE,
	NT_STATUS_SYSTEM_SHUTDOWN:                                             ERROR_SYSTEM_SHUTDOWN,
	NT_STATUS_DS_INIT_FAILURE_CONSOLE:                                     ERROR_DS_INIT_FAILURE_CONSOLE,
	NT_STATUS_DS_SAM_INIT_FAILURE_CONSOLE:                                 ERROR_DS_SAM_INIT_FAILURE_CONSOLE,
	NT_STATUS_UNFINISHED_CONTEXT_DELETED:                                  ERROR_UNFINISHED_CONTEXT_DELETED,
	NT_STATUS_NO_TGT_REPLY:                                                ERROR_NO_TGT_REPLY,
	NT_STATUS_OBJECTID_NOT_FOUND:                                          ERROR_OBJECTID_NOT_FOUND,
	NT_STATUS_NO_IP_ADDRESSES:                                             ERROR_NO_IP_ADDRESSES,
	NT_STATUS_WRONG_CREDENTIAL_HANDLE:                                     ERROR_WRONG_CREDENTIAL_HANDLE,
	NT_STATUS_CRYPTO_SYSTEM_INVALID:                                       ERROR_CRYPTO_SYSTEM_INVALID,
	NT_STATUS_MAX_REFERRALS_EXCEEDED:                                      ERROR_MAX_REFERRALS_EXCEEDED,
	NT_STATUS_MUST_BE_KDC:                                                 ERROR_MUST_BE_KDC,
	NT_STATUS_STRONG_CRYPTO_NOT_SUPPORTED:                                 ERROR_STRONG_CRYPTO_NOT_SUPPORTED,
	NT_STATUS_TOO_MANY_PRINCIPALS:                                         ERROR_TOO_MANY_PRINCIPALS,
	NT_STATUS_NO_PA_DATA:                                                  ERROR_NO_PA_DATA,
	NT_STATUS_PKINIT_NAME_MISMATCH:                                        ERROR_PKINIT_NAME_MISMATCH,
	NT_STATUS_SMARTCARD_LOGON_REQUIRED:                                    ERROR_SMARTCARD_LOGON_REQUIRED,
	NT_STATUS_KDC_INVALID_REQUEST:                                         ERROR_KDC_INVALID_REQUEST,
	NT_STATUS_KDC_UNABLE_TO_REFER:                                         ERROR_KDC_UNABLE_TO_REFER,
	NT_STATUS_KDC_UNKNOWN_ETYPE:                                           ERROR_KDC_UNKNOWN_ETYPE,
	NT_STATUS_SHUTDOWN_IN_PROGRESS:                                        ERROR_SHUTDOWN_IN_PROGRESS,
	NT_STATUS_SERVER_SHUTDOWN_IN_PROGRESS:                                 ERROR_SERVER_SHUTDOWN_IN_PROGRESS,
	NT_STATUS_NOT_SUPPORTED_ON_SBS:                                        ERROR_NOT_SUPPORTED_ON_SBS,
	NT_STATUS_WMI_GUID_DISCONNECTED:                                       ERROR_WMI_GUID_DISCONNECTED,
	NT_STATUS_WMI_ALREADY_DISABLED:                                        ERROR_WMI_ALREADY_DISABLED,
	NT_STATUS_WMI_ALREADY_ENABLED:                                         ERROR_WMI_ALREADY_ENABLED,
	NT_STATUS_MFT_TOO_FRAGMENTED:                                          ERROR_MFT_TOO_FRAGMENTED,
	NT_STATUS_COPY_PROTECTION_FAILURE:                                     ERROR_COPY_PROTECTION_FAILURE,
	NT_STATUS_CSS_AUTHENTICATION_FAILURE:                                  ERROR_CSS_AUTHENTICATION_FAILURE,
	NT_STATUS_CSS_KEY_NOT_PRESENT:                                         ERROR_CSS_KEY_NOT_PRESENT,
	NT_STATUS_CSS_KEY_NOT_ESTABLISHED:                                     ERROR_CSS_KEY_NOT_ESTABLISHED,
	NT_STATUS_CSS_SCRAMBLED_SECTOR:                                        ERROR_CSS_SCRAMBLED_SECTOR,
	NT_STATUS_CSS_REGION_MISMATCH:                                         ERROR_CSS_REGION_MISMATCH,
	NT_STATUS_CSS_RESETS_EXHAUSTED:                                        ERROR_CSS_RESETS_EXHAUSTED,
	NT_STATUS_PKINIT_FAILURE:                                              ERROR_PKINIT_FAILURE,
	NT_STATUS_SMARTCARD_SUBSYSTEM_FAILURE:                                 ERROR_SMARTCARD_SUBSYSTEM_FAILURE,
	NT_STATUS_NO_KERB_KEY:                                                 ERROR_NO_KERB_KEY,
	NT_STATUS_HOST_DOWN:                                                   ERROR_HOST_DOWN,
	NT_STATUS_UNSUPPORTED_PREAUTH:                                         ERROR_UNSUPPORTED_PREAUTH,
	NT_STATUS_EFS_ALG_BLOB_TOO_BIG:                                        ERROR_EFS_ALG_BLOB_TOO_BIG,
	NT_STATUS_PORT_NOT_SET:                                                ERROR_PORT_NOT_SET,
	NT_STATUS_DEBUGGER_INACTIVE:                                           ERROR_DEBUGGER_INACTIVE,
	NT_STATUS_DS_VERSION_CHECK_FAILURE:                                    ERROR_DS_VERSION_CHECK_FAILURE,
	NT_STATUS_AUDITING_DISABLED:                                           ERROR_AUDITING_DISABLED,
	NT_STATUS_PRENT4_MACHINE_ACCOUNT:                                      ERROR_PRENT4_MACHINE_ACCOUNT,
	NT_STATUS_DS_AG_CANT_HAVE_UNIVERSAL_MEMBER:                            ERROR_DS_AG_CANT_HAVE_UNIVERSAL_MEMBER,
	NT_STATUS_INVALID_IMAGE_WIN_32:                                        ERROR_INVALID_IMAGE_WIN_32,
	NT_STATUS_INVALID_IMAGE_WIN_64:                                        ERROR_INVALID_IMAGE_WIN_64,
	NT_STATUS_BAD_BINDINGS:                                                ERROR_BAD_BINDINGS,
	NT_STATUS_NETWORK_SESSION_EXPIRED:                                     ERROR_NETWORK_SESSION_EXPIRED,
	NT_STATUS_APPHELP_BLOCK:                                               ERROR_APPHELP_BLOCK,
	NT_STATUS_ALL_SIDS_FILTERED:                                           ERROR_ALL_SIDS_FILTERED,
	NT_STATUS_NOT_SAFE_MODE_DRIVER:                                        ERROR_NOT_SAFE_MODE_DRIVER,
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_DEFAULT:                           ERROR_ACCESS_DISABLED_BY_POLICY_DEFAULT,
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PATH:                              ERROR_ACCESS_DISABLED_BY_POLICY_PATH,
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PUBLISHER:                         ERROR_ACCESS_DISABLED_BY_POLICY_PUBLISHER,
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_OTHER:                             ERROR_ACCESS_DISABLED_BY_POLICY_OTHER,
	NT_STATUS_FAILED_DRIVER_ENTRY:                                         ERROR_FAILED_DRIVER_ENTRY,
	NT_STATUS_DEVICE_ENUMERATION_ERROR:                                    ERROR_DEVICE_ENUMERATION_ERROR,
	NT_STATUS_MOUNT_POINT_NOT_RESOLVED:                                    ERROR_MOUNT_POINT_NOT_RESOLVED,
	NT_STATUS_INVALID_DEVICE_OBJECT_PARAMETER:                             ERROR_INVALID_DEVICE_OBJECT_PARAMETER,
	NT_STATUS_MCA_OCCURED:                                                 ERROR_MCA_OCCURED,
	NT_STATUS_DRIVER_BLOCKED_CRITICAL:                                     ERROR_DRIVER_BLOCKED_CRITICAL,
	NT_STATUS_DRIVER_BLOCKED:                                              ERROR_DRIVER_BLOCKED,
	NT_STATUS_DRIVER_DATABASE_ERROR:                                       ERROR_DRIVER_DATABASE_ERROR,
	NT_STATUS_SYSTEM_HIVE_TOO_LARGE:                                       ERROR_SYSTEM_HIVE_TOO_LARGE,
	NT_STATUS_INVALID_IMPORT_OF_NON_DLL:                                   ERROR_INVALID_IMPORT_OF_NON_DLL,
	NT_STATUS_NO_SECRETS:                                                  ERROR_NO_SECRETS,
	NT_STATUS_ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY:                       ERROR_ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY,
	NT_STATUS_FAILED_STACK_SWITCH:                                         ERROR_FAILED_STACK_SWITCH,
	NT_STATUS_HEAP_CORRUPTION:                                             ERROR_HEAP_CORRUPTION,
	NT_STATUS_SMARTCARD_WRONG_PIN:                                         ERROR_SMARTCARD_WRONG_PIN,
	NT_STATUS_SMARTCARD_CARD_BLOCKED:                                      ERROR_SMARTCARD_CARD_BLOCKED,
	NT_STATUS_SMARTCARD_CARD_NOT_AUTHENTICATED:                            ERROR_SMARTCARD_CARD_NOT_AUTHENTICATED,
	NT_STATUS_SMARTCARD_NO_CARD:                                           ERROR_SMARTCARD_NO_CARD,
	NT_STATUS_SMARTCARD_NO_KEY_CONTAINER:                                  ERROR_SMARTCARD_NO_KEY_CONTAINER,
	NT_STATUS_SMARTCARD_NO_CERTIFICATE:                                    ERROR_SMARTCARD_NO_CERTIFICATE,
	NT_STATUS_SMARTCARD_NO_KEYSET:                                         ERROR_SMARTCARD_NO_KEYSET,
	NT_STATUS_SMARTCARD_IO_ERROR:                                          ERROR_SMARTCARD_IO_ERROR,
	NT_STATUS_DOWNGRADE_DETECTED:                                          ERROR_DOWNGRADE_DETECTED,
	NT_STATUS_SMARTCARD_CERT_REVOKED:                                      ERROR_SMARTCARD_CERT_REVOKED,
	NT_STATUS_ISSUING_CA_UNTRUSTED:                                        ERROR_ISSUING_CA_UNTRUSTED,
	NT_STATUS_REVOCATION_OFFLINE_C:                                        ERROR_REVOCATION_OFFLINE_C,
	NT_STATUS_PKINIT_CLIENT_FAILURE:                                       ERROR_PKINIT_CLIENT_FAILURE,
	NT_STATUS_SMARTCARD_CERT_EXPIRED:                                      ERROR_SMARTCARD_CERT_EXPIRED,
	NT_STATUS_DRIVER_FAILED_PRIOR_UNLOAD:                                  ERROR_DRIVER_FAILED_PRIOR_UNLOAD,
	NT_STATUS_SMARTCARD_SILENT_CONTEXT:                                    ERROR_SMARTCARD_SILENT_CONTEXT,
	NT_STATUS_PER_USER_TRUST_QUOTA_EXCEEDED:                               ERROR_PER_USER_TRUST_QUOTA_EXCEEDED,
	NT_STATUS_ALL_USER_TRUST_QUOTA_EXCEEDED:                               ERROR_ALL_USER_TRUST_QUOTA_EXCEEDED,
	NT_STATUS_USER_DELETE_TRUST_QUOTA_EXCEEDED:                            ERROR_USER_DELETE_TRUST_QUOTA_EXCEEDED,
	NT_STATUS_DS_NAME_NOT_UNIQUE:                                          ERROR_DS_NAME_NOT_UNIQUE,
	NT_STATUS_DS_DUPLICATE_ID_FOUND:                                       ERROR_DS_DUPLICATE_ID_FOUND,
	NT_STATUS_DS_GROUP_CONVERSION_ERROR:                                   ERROR_DS_GROUP_CONVERSION_ERROR,
	NT_STATUS_VOLSNAP_PREPARE_HIBERNATE:                                   ERROR_VOLSNAP_PREPARE_HIBERNATE,
	NT_STATUS_USER2USER_REQUIRED:                                          ERROR_USER2USER_REQUIRED,
	NT_STATUS_STACK_BUFFER_OVERRUN:                                        ERROR_STACK_BUFFER_OVERRUN,
	NT_STATUS_NO_S4U_PROT_SUPPORT:                                         ERROR_NO_S4U_PROT_SUPPORT,
	NT_STATUS_CROSSREALM_DELEGATION_FAILURE:                               ERROR_CROSSREALM_DELEGATION_FAILURE,
	NT_STATUS_REVOCATION_OFFLINE_KDC:                                      ERROR_REVOCATION_OFFLINE_KDC,
	NT_STATUS_ISSUING_CA_UNTRUSTED_KDC:                                    ERROR_ISSUING_CA_UNTRUSTED_KDC,
	NT_STATUS_KDC_CERT_EXPIRED:                                            ERROR_KDC_CERT_EXPIRED,
	NT_STATUS_KDC_CERT_REVOKED:                                            ERROR_KDC_CERT_REVOKED,
	NT_STATUS_PARAMETER_QUOTA_EXCEEDED:                                    ERROR_PARAMETER_QUOTA_EXCEEDED,
	NT_STATUS_HIBERNATION_FAILURE:                                         ERROR_HIBERNATION_FAILURE,
	NT_STATUS_DELAY_LOAD_FAILED:                                           ERROR_DELAY_LOAD_FAILED,
	NT_STATUS_AUTHENTICATION_FIREWALL_FAILED:                              ERROR_AUTHENTICATION_FIREWALL_FAILED,
	NT_STATUS_VDM_DISALLOWED:                                              ERROR_VDM_DISALLOWED,
	NT_STATUS_HUNG_DISPLAY_DRIVER_THREAD:                                  ERROR_HUNG_DISPLAY_DRIVER_THREAD,
	NT_STATUS_INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE:     ERROR_INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE,
	NT_STATUS_INVALID_CRUNTIME_PARAMETER:                                  ERROR_INVALID_CRUNTIME_PARAMETER,
	NT_STATUS_NTLM_BLOCKED:                                                ERROR_NTLM_BLOCKED,
	NT_STATUS_DS_SRC_SID_EXISTS_IN_FOREST:                                 ERROR_DS_SRC_SID_EXISTS_IN_FOREST,
	NT_STATUS_DS_DOMAIN_NAME_EXISTS_IN_FOREST:                             ERROR_DS_DOMAIN_NAME_EXISTS_IN_FOREST,
	NT_STATUS_DS_FLAT_NAME_EXISTS_IN_FOREST:                               ERROR_DS_FLAT_NAME_EXISTS_IN_FOREST,
	NT_STATUS_INVALID_USER_PRINCIPAL_NAME:                                 ERROR_INVALID_USER_PRINCIPAL_NAME,
	NT_STATUS_ASSERTION_FAILURE:                                           ERROR_ASSERTION_FAILURE,
	NT_STATUS_VERIFIER_STOP:                                               ERROR_VERIFIER_STOP,
	NT_STATUS_CALLBACK_POP_STACK:                                          ERROR_CALLBACK_POP_STACK,
	NT_STATUS_INCOMPATIBLE_DRIVER_BLOCKED:                                 ERROR_INCOMPATIBLE_DRIVER_BLOCKED,
	NT_STATUS_HIVE_UNLOADED:                                               ERROR_HIVE_UNLOADED,
	NT_STATUS_COMPRESSION_DISABLED:                                        ERROR_COMPRESSION_DISABLED,
	NT_STATUS_FILE_SYSTEM_LIMITATION:                                      ERROR_FILE_SYSTEM_LIMITATION,
	NT_STATUS_INVALID_IMAGE_HASH:                                          ERROR_INVALID_IMAGE_HASH,
	NT_STATUS_NOT_CAPABLE:                                                 ERROR_NOT_CAPABLE,
	NT_STATUS_REQUEST_OUT_OF_SEQUENCE:                                     ERROR_REQUEST_OUT_OF_SEQUENCE,
	NT_STATUS_IMPLEMENTATION_LIMIT:                                        ERROR_IMPLEMENTATION_LIMIT,
	NT_STATUS_ELEVATION_REQUIRED:                                          ERROR_ELEVATION_REQUIRED,
	NT_STATUS_NO_SECURITY_CONTEXT:                                         ERROR_NO_SECURITY_CONTEXT,
	NT_STATUS_PKU2U_CERT_FAILURE:                                          ERROR_PKU2U_CERT_FAILURE,
	NT_STATUS_BEYOND_VDL:                                                  ERROR_BEYOND_VDL,
	NT_STATUS_ENCOUNTERED_WRITE_IN_PROGRESS:                               ERROR_ENCOUNTERED_WRITE_IN_PROGRESS,
	NT_STATUS_PTE_CHANGED:                                                 ERROR_PTE_CHANGED,
	NT_STATUS_PURGE_FAILED:                                                ERROR_PURGE_FAILED,
	NT_STATUS_CRED_REQUIRES_CONFIRMATION:                                  ERROR_CRED_REQUIRES_CONFIRMATION,
	NT_STATUS_CS_ENCRYPTION_INVALID_SERVER_RESPONSE:                       ERROR_CS_ENCRYPTION_INVALID_SERVER_RESPONSE,
	NT_STATUS_CS_ENCRYPTION_UNSUPPORTED_SERVER:                            ERROR_CS_ENCRYPTION_UNSUPPORTED_SERVER,
	NT_STATUS_CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE:                       ERROR_CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE,
	NT_STATUS_CS_ENCRYPTION_NEW_ENCRYPTED_FILE:                            ERROR_CS_ENCRYPTION_NEW_ENCRYPTED_FILE,
	NT_STATUS_CS_ENCRYPTION_FILE_NOT_CSE:                                  ERROR_CS_ENCRYPTION_FILE_NOT_CSE,
	NT_STATUS_INVALID_LABEL:                                               ERROR_INVALID_LABEL,
	NT_STATUS_DRIVER_PROCESS_TERMINATED:                                   ERROR_DRIVER_PROCESS_TERMINATED,
	NT_STATUS_AMBIGUOUS_SYSTEM_DEVICE:                                     ERROR_AMBIGUOUS_SYSTEM_DEVICE,
	NT_STATUS_SYSTEM_DEVICE_NOT_FOUND:                                     ERROR_SYSTEM_DEVICE_NOT_FOUND,
	NT_STATUS_RESTART_BOOT_APPLICATION:                                    ERROR_RESTART_BOOT_APPLICATION,
	NT_STATUS_INSUFFICIENT_NVRAM_RESOURCES:                                ERROR_INSUFFICIENT_NVRAM_RESOURCES,
	NT_STATUS_NO_RANGES_PROCESSED:                                         ERROR_NO_RANGES_PROCESSED,
	NT_STATUS_DEVICE_FEATURE_NOT_SUPPORTED:                                ERROR_DEVICE_FEATURE_NOT_SUPPORTED,
	NT_STATUS_DEVICE_UNREACHABLE:                                          ERROR_DEVICE_UNREACHABLE,
	NT_STATUS_INVALID_TOKEN:                                               ERROR_INVALID_TOKEN,
	NT_STATUS_SERVER_UNAVAILABLE:                                          ERROR_SERVER_UNAVAILABLE,
	NT_STATUS_INVALID_TASK_NAME:                                           ERROR_INVALID_TASK_NAME,
	NT_STATUS_INVALID_TASK_INDEX:                                          ERROR_INVALID_TASK_INDEX,
	NT_STATUS_THREAD_ALREADY_IN_TASK:                                      ERROR_THREAD_ALREADY_IN_TASK,
	NT_STATUS_CALLBACK_BYPASS:                                             ERROR_CALLBACK_BYPASS,
	NT_STATUS_FAIL_FAST_EXCEPTION:                                         ERROR_FAIL_FAST_EXCEPTION,
	NT_STATUS_IMAGE_CERT_REVOKED:                                          ERROR_IMAGE_CERT_REVOKED,
	NT_STATUS_PORT_CLOSED:                                                 ERROR_PORT_CLOSED,
	NT_STATUS_MESSAGE_LOST:                                                ERROR_MESSAGE_LOST,
	NT_STATUS_INVALID_MESSAGE:                                             ERROR_INVALID_MESSAGE,
	NT_STATUS_REQUEST_CANCELED:                                            ERROR_REQUEST_CANCELED,
	NT_STATUS_RECURSIVE_DISPATCH:                                          ERROR_RECURSIVE_DISPATCH,
	NT_STATUS_LPC_RECEIVE_BUFFER_EXPECTED:                                 ERROR_LPC_RECEIVE_BUFFER_EXPECTED,
	NT_STATUS_LPC_INVALID_CONNECTION_USAGE:                                ERROR_LPC_INVALID_CONNECTION_USAGE,
	NT_STATUS_LPC_REQUESTS_NOT_ALLOWED:                                    ERROR_LPC_REQUESTS_NOT_ALLOWED,
	NT_STATUS_RESOURCE_IN_USE:                                             ERROR_RESOURCE_IN_USE,
	NT_STATUS_HARDWARE_MEMORY_ERROR:                                       ERROR_HARDWARE_MEMORY_ERROR,
	NT_STATUS_THREADPOOL_HANDLE_EXCEPTION:                                 ERROR_THREADPOOL_HANDLE_EXCEPTION,
	NT_STATUS_THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED:                   ERROR_THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED,
	NT_STATUS_THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED:           ERROR_THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED,
	NT_STATUS_THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED:               ERROR_THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED,
	NT_STATUS_THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED:                ERROR_THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED,
	NT_STATUS_THREADPOOL_RELEASED_DURING_OPERATION:                        ERROR_THREADPOOL_RELEASED_DURING_OPERATION,
	NT_STATUS_CALLBACK_RETURNED_WHILE_IMPERSONATING:                       ERROR_CALLBACK_RETURNED_WHILE_IMPERSONATING,
	NT_STATUS_APC_RETURNED_WHILE_IMPERSONATING:                            ERROR_APC_RETURNED_WHILE_IMPERSONATING,
	NT_STATUS_PROCESS_IS_PROTECTED:                                        ERROR_PROCESS_IS_PROTECTED,
	NT_STATUS_MCA_EXCEPTION:                                               ERROR_MCA_EXCEPTION,
	NT_STATUS_CERTIFICATE_MAPPING_NOT_UNIQUE:                              ERROR_CERTIFICATE_MAPPING_NOT_UNIQUE,
	NT_STATUS_SYMLINK_CLASS_DISABLED:                                      ERROR_SYMLINK_CLASS_DISABLED,
	NT_STATUS_INVALID_IDN_NORMALIZATION:                                   ERROR_INVALID_IDN_NORMALIZATION,
	NT_STATUS_NO_UNICODE_TRANSLATION:                                      ERROR_NO_UNICODE_TRANSLATION,
	NT_STATUS_ALREADY_REGISTERED:                                          ERROR_ALREADY_REGISTERED,
	NT_STATUS_CONTEXT_MISMATCH:                                            ERROR_CONTEXT_MISMATCH,
	NT_STATUS_PORT_ALREADY_HAS_COMPLETION_LIST:                            ERROR_PORT_ALREADY_HAS_COMPLETION_LIST,
	NT_STATUS_CALLBACK_RETURNED_THREAD_PRIORITY:                           ERROR_CALLBACK_RETURNED_THREAD_PRIORITY,
	NT_STATUS_INVALID_THREAD:                                              ERROR_INVALID_THREAD,
	NT_STATUS_CALLBACK_RETURNED_TRANSACTION:                               ERROR_CALLBACK_RETURNED_TRANSACTION,
	NT_STATUS_CALLBACK_RETURNED_LDR_LOCK:                                  ERROR_CALLBACK_RETURNED_LDR_LOCK,
	NT_STATUS_CALLBACK_RETURNED_LANG:                                      ERROR_CALLBACK_RETURNED_LANG,
	NT_STATUS_CALLBACK_RETURNED_PRI_BACK:                                  ERROR_CALLBACK_RETURNED_PRI_BACK,
	NT_STATUS_DISK_REPAIR_DISABLED:                                        ERROR_DISK_REPAIR_DISABLED,
	NT_STATUS_DS_DOMAIN_RENAME_IN_PROGRESS:                                ERROR_DS_DOMAIN_RENAME_IN_PROGRESS,
	NT_STATUS_DISK_QUOTA_EXCEEDED:                                         ERROR_DISK_QUOTA_EXCEEDED,
	NT_STATUS_CONTENT_BLOCKED:                                             ERROR_CONTENT_BLOCKED,
	NT_STATUS_BAD_CLUSTERS:                                                ERROR_BAD_CLUSTERS,
	NT_STATUS_VOLUME_DIRTY:                                                ERROR_VOLUME_DIRTY,
	NT_STATUS_FILE_CHECKED_OUT:                                            ERROR_FILE_CHECKED_OUT,
	NT_STATUS_CHECKOUT_REQUIRED:                                           ERROR_CHECKOUT_REQUIRED,
	NT_STATUS_BAD_FILE_TYPE:                                               ERROR_BAD_FILE_TYPE,
	NT_STATUS_FILE_TOO_LARGE:                                              ERROR_FILE_TOO_LARGE,
	NT_STATUS_FORMS_AUTH_REQUIRED:                                         ERROR_FORMS_AUTH_REQUIRED,
	NT_STATUS_VIRUS_INFECTED:                                              ERROR_VIRUS_INFECTED,
	NT_STATUS_VIRUS_DELETED:                                               ERROR_VIRUS_DELETED,
	NT_STATUS_BAD_MCFG_TABLE:                                              ERROR_BAD_MCFG_TABLE,
	NT_STATUS_BAD_DATA:                                                    ERROR_BAD_DATA,
	NT_STATUS_CANNOT_BREAK_OPLOCK:                                         ERROR_CANNOT_BREAK_OPLOCK,
	NT_STATUS_WOW_ASSERTION:                                               ERROR_WOW_ASSERTION,
	NT_STATUS_INVALID_SIGNATURE:                                           ERROR_INVALID_SIGNATURE,
	NT_STATUS_HMAC_NOT_SUPPORTED:                                          ERROR_HMAC_NOT_SUPPORTED,
	NT_STATUS_AUTH_TAG_MISMATCH:                                           ERROR_AUTH_TAG_MISMATCH,
	NT_STATUS_IPSEC_QUEUE_OVERFLOW:                                        ERROR_IPSEC_QUEUE_OVERFLOW,
	NT_STATUS_ND_QUEUE_OVERFLOW:                                           ERROR_ND_QUEUE_OVERFLOW,
	NT_STATUS_HOPLIMIT_EXCEEDED:                                           ERROR_HOPLIMIT_EXCEEDED,
	NT_STATUS_PROTOCOL_NOT_SUPPORTED:                                      ERROR_PROTOCOL_NOT_SUPPORTED,
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED:                  ERROR_LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED,
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR:                  ERROR_LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR,
	NT_STATUS_LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR:                      ERROR_LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR,
	NT_STATUS_XML_PARSE_ERROR:                                             ERROR_XML_PARSE_ERROR,
	NT_STATUS_XMLDSIG_ERROR:                                               ERROR_XMLDSIG_ERROR,
	NT_STATUS_WRONG_COMPARTMENT:                                           ERROR_WRONG_COMPARTMENT,
	NT_STATUS_AUTHIP_FAILURE:                                              ERROR_AUTHIP_FAILURE,
	NT_STATUS_DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS:                       ERROR_DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS,
	NT_STATUS_DS_OID_NOT_FOUND:                                            ERROR_DS_OID_NOT_FOUND,
	NT_STATUS_HASH_NOT_SUPPORTED:                                          ERROR_HASH_NOT_SUPPORTED,
	NT_STATUS_HASH_NOT_PRESENT:                                            ERROR_HASH_NOT_PRESENT,
	NT_STATUS_OFFLOAD_READ_FLT_NOT_SUPPORTED:                              ERROR_OFFLOAD_READ_FLT_NOT_SUPPORTED,
	NT_STATUS_OFFLOAD_WRITE_FLT_NOT_SUPPORTED:                             ERROR_OFFLOAD_WRITE_FLT_NOT_SUPPORTED,
	NT_STATUS_OFFLOAD_READ_FILE_NOT_SUPPORTED:                             ERROR_OFFLOAD_READ_FILE_NOT_SUPPORTED,
	NT_STATUS_OFFLOAD_WRITE_FILE_NOT_SUPPORTED:                            ERROR_OFFLOAD_WRITE_FILE_NOT_SUPPORTED,
	NT_STATUS_DBG_NO_STATE_CHANGE:                                         ERROR_DBG_NO_STATE_CHANGE,
	NT_STATUS_DBG_APP_NOT_IDLE:                                            ERROR_DBG_APP_NOT_IDLE,
	NT_STATUS_RPC_NT_INVALID_STRING_BINDING:                               ERROR_RPC_NT_INVALID_STRING_BINDING,
	NT_STATUS_RPC_NT_WRONG_KIND_OF_BINDING:                                ERROR_RPC_NT_WRONG_KIND_OF_BINDING,
	NT_STATUS_RPC_NT_INVALID_BINDING:                                      ERROR_RPC_NT_INVALID_BINDING,
	NT_STATUS_RPC_NT_PROTSEQ_NOT_SUPPORTED:                                ERROR_RPC_NT_PROTSEQ_NOT_SUPPORTED,
	NT_STATUS_RPC_NT_INVALID_RPC_PROTSEQ:                                  ERROR_RPC_NT_INVALID_RPC_PROTSEQ,
	NT_STATUS_RPC_NT_INVALID_STRING_UUID:                                  ERROR_RPC_NT_INVALID_STRING_UUID,
	NT_STATUS_RPC_NT_INVALID_ENDPOINT_FORMAT:                              ERROR_RPC_NT_INVALID_ENDPOINT_FORMAT,
	NT_STATUS_RPC_NT_INVALID_NET_ADDR:                                     ERROR_RPC_NT_INVALID_NET_ADDR,
	NT_STATUS_RPC_NT_NO_ENDPOINT_FOUND:                                    ERROR_RPC_NT_NO_ENDPOINT_FOUND,
	NT_STATUS_RPC_NT_INVALID_TIMEOUT:                                      ERROR_RPC_NT_INVALID_TIMEOUT,
	NT_STATUS_RPC_NT_OBJECT_NOT_FOUND:                                     ERROR_RPC_NT_OBJECT_NOT_FOUND,
	NT_STATUS_RPC_NT_ALREADY_REGISTERED:                                   ERROR_RPC_NT_ALREADY_REGISTERED,
	NT_STATUS_RPC_NT_TYPE_ALREADY_REGISTERED:                              ERROR_RPC_NT_TYPE_ALREADY_REGISTERED,
	NT_STATUS_RPC_NT_ALREADY_LISTENING:                                    ERROR_RPC_NT_ALREADY_LISTENING,
	NT_STATUS_RPC_NT_NO_PROTSEQS_REGISTERED:                               ERROR_RPC_NT_NO_PROTSEQS_REGISTERED,
	NT_STATUS_RPC_NT_NOT_LISTENING:                                        ERROR_RPC_NT_NOT_LISTENING,
	NT_STATUS_RPC_NT_UNKNOWN_MGR_TYPE:                                     ERROR_RPC_NT_UNKNOWN_MGR_TYPE,
	NT_STATUS_RPC_NT_UNKNOWN_IF:                                           ERROR_RPC_NT_UNKNOWN_IF,
	NT_STATUS_RPC_NT_NO_BINDINGS:                                          ERROR_RPC_NT_NO_BINDINGS,
	NT_STATUS_RPC_NT_NO_PROTSEQS:                                          ERROR_RPC_NT_NO_PROTSEQS,
	NT_STATUS_RPC_NT_CANT_CREATE_ENDPOINT:                                 ERROR_RPC_NT_CANT_CREATE_ENDPOINT,
	NT_STATUS_RPC_NT_OUT_OF_RESOURCES:                                     ERROR_RPC_NT_OUT_OF_RESOURCES,
	NT_STATUS_RPC_NT_SERVER_UNAVAILABLE:                                   ERROR_RPC_NT_SERVER_UNAVAILABLE,
	NT_STATUS_RPC_NT_SERVER_TOO_BUSY:                                      ERROR_RPC_NT_SERVER_TOO_BUSY,
	NT_STATUS_RPC_NT_INVALID_NETWORK_OPTIONS:                              ERROR_RPC_NT_INVALID_NETWORK_OPTIONS,
	NT_STATUS_RPC_NT_NO_CALL_ACTIVE:                                       ERROR_RPC_NT_NO_CALL_ACTIVE,
	NT_STATUS_RPC_NT_CALL_FAILED:                                          ERROR_RPC_NT_CALL_FAILED,
	NT_STATUS_RPC_NT_CALL_FAILED_DNE:                                      ERROR_RPC_NT_CALL_FAILED_DNE,
	NT_STATUS_RPC_NT_PROTOCOL_ERROR:                                       ERROR_RPC_NT_PROTOCOL_ERROR,
	NT_STATUS_RPC_NT_UNSUPPORTED_TRANS_SYN:                                ERROR_RPC_NT_UNSUPPORTED_TRANS_SYN,
	NT_STATUS_RPC_NT_UNSUPPORTED_TYPE:                                     ERROR_RPC_NT_UNSUPPORTED_TYPE,
	NT_STATUS_RPC_NT_INVALID_TAG:                                          ERROR_RPC_NT_INVALID_TAG,
	NT_STATUS_RPC_NT_INVALID_BOUND:                                        ERROR_RPC_NT_INVALID_BOUND,
	NT_STATUS_RPC_NT_NO_ENTRY_NAME:                                        ERROR_RPC_NT_NO_ENTRY_NAME,
	NT_STATUS_RPC_NT_INVALID_NAME_SYNTAX:                                  ERROR_RPC_NT_INVALID_NAME_SYNTAX,
	NT_STATUS_RPC_NT_UNSUPPORTED_NAME_SYNTAX:                              ERROR_RPC_NT_UNSUPPORTED_NAME_SYNTAX,
	NT_STATUS_RPC_NT_UUID_NO_ADDRESS:                                      ERROR_RPC_NT_UUID_NO_ADDRESS,
	NT_STATUS_RPC_NT_DUPLICATE_ENDPOINT:                                   ERROR_RPC_NT_DUPLICATE_ENDPOINT,
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_TYPE:                                   ERROR_RPC_NT_UNKNOWN_AUTHN_TYPE,
	NT_STATUS_RPC_NT_MAX_CALLS_TOO_SMALL:                                  ERROR_RPC_NT_MAX_CALLS_TOO_SMALL,
	NT_STATUS_RPC_NT_STRING_TOO_LONG:                                      ERROR_RPC_NT_STRING_TOO_LONG,
	NT_STATUS_RPC_NT_PROTSEQ_NOT_FOUND:                                    ERROR_RPC_NT_PROTSEQ_NOT_FOUND,
	NT_STATUS_RPC_NT_PROCNUM_OUT_OF_RANGE:                                 ERROR_RPC_NT_PROCNUM_OUT_OF_RANGE,
	NT_STATUS_RPC_NT_BINDING_HAS_NO_AUTH:                                  ERROR_RPC_NT_BINDING_HAS_NO_AUTH,
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_SERVICE:                                ERROR_RPC_NT_UNKNOWN_AUTHN_SERVICE,
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_LEVEL:                                  ERROR_RPC_NT_UNKNOWN_AUTHN_LEVEL,
	NT_STATUS_RPC_NT_INVALID_AUTH_IDENTITY:                                ERROR_RPC_NT_INVALID_AUTH_IDENTITY,
	NT_STATUS_RPC_NT_UNKNOWN_AUTHZ_SERVICE:                                ERROR_RPC_NT_UNKNOWN_AUTHZ_SERVICE,
	NT_STATUS_EPT_NT_INVALID_ENTRY:                                        ERROR_EPT_NT_INVALID_ENTRY,
	NT_STATUS_EPT_NT_CANT_PERFORM_OP:                                      ERROR_EPT_NT_CANT_PERFORM_OP,
	NT_STATUS_EPT_NT_NOT_REGISTERED:                                       ERROR_EPT_NT_NOT_REGISTERED,
	NT_STATUS_RPC_NT_NOTHING_TO_EXPORT:                                    ERROR_RPC_NT_NOTHING_TO_EXPORT,
	NT_STATUS_RPC_NT_INCOMPLETE_NAME:                                      ERROR_RPC_NT_INCOMPLETE_NAME,
	NT_STATUS_RPC_NT_INVALID_VERS_OPTION:                                  ERROR_RPC_NT_INVALID_VERS_OPTION,
	NT_STATUS_RPC_NT_NO_MORE_MEMBERS:                                      ERROR_RPC_NT_NO_MORE_MEMBERS,
	NT_STATUS_RPC_NT_NOT_ALL_OBJS_UNEXPORTED:                              ERROR_RPC_NT_NOT_ALL_OBJS_UNEXPORTED,
	NT_STATUS_RPC_NT_INTERFACE_NOT_FOUND:                                  ERROR_RPC_NT_INTERFACE_NOT_FOUND,
	NT_STATUS_RPC_NT_ENTRY_ALREADY_EXISTS:                                 ERROR_RPC_NT_ENTRY_ALREADY_EXISTS,
	NT_STATUS_RPC_NT_ENTRY_NOT_FOUND:                                      ERROR_RPC_NT_ENTRY_NOT_FOUND,
	NT_STATUS_RPC_NT_NAME_SERVICE_UNAVAILABLE:                             ERROR_RPC_NT_NAME_SERVICE_UNAVAILABLE,
	NT_STATUS_RPC_NT_INVALID_NAF_ID:                                       ERROR_RPC_NT_INVALID_NAF_ID,
	NT_STATUS_RPC_NT_CANNOT_SUPPORT:                                       ERROR_RPC_NT_CANNOT_SUPPORT,
	NT_STATUS_RPC_NT_NO_CONTEXT_AVAILABLE:                                 ERROR_RPC_NT_NO_CONTEXT_AVAILABLE,
	NT_STATUS_RPC_NT_INTERNAL_ERROR:                                       ERROR_RPC_NT_INTERNAL_ERROR,
	NT_STATUS_RPC_NT_ZERO_DIVIDE:                                          ERROR_RPC_NT_ZERO_DIVIDE,
	NT_STATUS_RPC_NT_ADDRESS_ERROR:                                        ERROR_RPC_NT_ADDRESS_ERROR,
	NT_STATUS_RPC_NT_FP_DIV_ZERO:                                          ERROR_RPC_NT_FP_DIV_ZERO,
	NT_STATUS_RPC_NT_FP_UNDERFLOW:                                         ERROR_RPC_NT_FP_UNDERFLOW,
	NT_STATUS_RPC_NT_FP_OVERFLOW:                                          ERROR_RPC_NT_FP_OVERFLOW,
	NT_STATUS_RPC_NT_CALL_IN_PROGRESS:                                     ERROR_RPC_NT_CALL_IN_PROGRESS,
	NT_STATUS_RPC_NT_NO_MORE_BINDINGS:                                     ERROR_RPC_NT_NO_MORE_BINDINGS,
	NT_STATUS_RPC_NT_GROUP_MEMBER_NOT_FOUND:                               ERROR_RPC_NT_GROUP_MEMBER_NOT_FOUND,
	NT_STATUS_EPT_NT_CANT_CREATE:                                          ERROR_EPT_NT_CANT_CREATE,
	NT_STATUS_RPC_NT_INVALID_OBJECT:                                       ERROR_RPC_NT_INVALID_OBJECT,
	NT_STATUS_RPC_NT_NO_INTERFACES:                                        ERROR_RPC_NT_NO_INTERFACES,
	NT_STATUS_RPC_NT_CALL_CANCELLED:                                       ERROR_RPC_NT_CALL_CANCELLED,
	NT_STATUS_RPC_NT_BINDING_INCOMPLETE:                                   ERROR_RPC_NT_BINDING_INCOMPLETE,
	NT_STATUS_RPC_NT_COMM_FAILURE:                                         ERROR_RPC_NT_COMM_FAILURE,
	NT_STATUS_RPC_NT_UNSUPPORTED_AUTHN_LEVEL:                              ERROR_RPC_NT_UNSUPPORTED_AUTHN_LEVEL,
	NT_STATUS_RPC_NT_NO_PRINC_NAME:                                        ERROR_RPC_NT_NO_PRINC_NAME,
	NT_STATUS_RPC_NT_NOT_RPC_ERROR:                                        ERROR_RPC_NT_NOT_RPC_ERROR,
	NT_STATUS_RPC_NT_SEC_PKG_ERROR:                                        ERROR_RPC_NT_SEC_PKG_ERROR,
	NT_STATUS_RPC_NT_NOT_CANCELLED:                                        ERROR_RPC_NT_NOT_CANCELLED,
	NT_STATUS_RPC_NT_INVALID_ASYNC_HANDLE:                                 ERROR_RPC_NT_INVALID_ASYNC_HANDLE,
	NT_STATUS_RPC_NT_INVALID_ASYNC_CALL:                                   ERROR_RPC_NT_INVALID_ASYNC_CALL,
	NT_STATUS_RPC_NT_PROXY_ACCESS_DENIED:                                  ERROR_RPC_NT_PROXY_ACCESS_DENIED,
	NT_STATUS_RPC_NT_NO_MORE_ENTRIES:                                      ERROR_RPC_NT_NO_MORE_ENTRIES,
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_OPEN_FAIL:                              ERROR_RPC_NT_SS_CHAR_TRANS_OPEN_FAIL,
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_SHORT_FILE:                             ERROR_RPC_NT_SS_CHAR_TRANS_SHORT_FILE,
	NT_STATUS_RPC_NT_SS_IN_NULL_CONTEXT:                                   ERROR_RPC_NT_SS_IN_NULL_CONTEXT,
	NT_STATUS_RPC_NT_SS_CONTEXT_MISMATCH:                                  ERROR_RPC_NT_SS_CONTEXT_MISMATCH,
	NT_STATUS_RPC_NT_SS_CONTEXT_DAMAGED:                                   ERROR_RPC_NT_SS_CONTEXT_DAMAGED,
	NT_STATUS_RPC_NT_SS_HANDLES_MISMATCH:                                  ERROR_RPC_NT_SS_HANDLES_MISMATCH,
	NT_STATUS_RPC_NT_SS_CANNOT_GET_CALL_HANDLE:                            ERROR_RPC_NT_SS_CANNOT_GET_CALL_HANDLE,
	NT_STATUS_RPC_NT_NULL_REF_POINTER:                                     ERROR_RPC_NT_NULL_REF_POINTER,
	NT_STATUS_RPC_NT_ENUM_VALUE_OUT_OF_RANGE:                              ERROR_RPC_NT_ENUM_VALUE_OUT_OF_RANGE,
	NT_STATUS_RPC_NT_BYTE_COUNT_TOO_SMALL:                                 ERROR_RPC_NT_BYTE_COUNT_TOO_SMALL,
	NT_STATUS_RPC_NT_BAD_STUB_DATA:                                        ERROR_RPC_NT_BAD_STUB_DATA,
	NT_STATUS_RPC_NT_INVALID_ES_ACTION:                                    ERROR_RPC_NT_INVALID_ES_ACTION,
	NT_STATUS_RPC_NT_WRONG_ES_VERSION:                                     ERROR_RPC_NT_WRONG_ES_VERSION,
	NT_STATUS_RPC_NT_WRONG_STUB_VERSION:                                   ERROR_RPC_NT_WRONG_STUB_VERSION,
	NT_STATUS_RPC_NT_INVALID_PIPE_OBJECT:                                  ERROR_RPC_NT_INVALID_PIPE_OBJECT,
	NT_STATUS_RPC_NT_INVALID_PIPE_OPERATION:                               ERROR_RPC_NT_INVALID_PIPE_OPERATION,
	NT_STATUS_RPC_NT_WRONG_PIPE_VERSION:                                   ERROR_RPC_NT_WRONG_PIPE_VERSION,
	NT_STATUS_RPC_NT_PIPE_CLOSED:                                          ERROR_RPC_NT_PIPE_CLOSED,
	NT_STATUS_RPC_NT_PIPE_DISCIPLINE_ERROR:                                ERROR_RPC_NT_PIPE_DISCIPLINE_ERROR,
	NT_STATUS_RPC_NT_PIPE_EMPTY:                                           ERROR_RPC_NT_PIPE_EMPTY,
	NT_STATUS_PNP_BAD_MPS_TABLE:                                           ERROR_PNP_BAD_MPS_TABLE,
	NT_STATUS_PNP_TRANSLATION_FAILED:                                      ERROR_PNP_TRANSLATION_FAILED,
	NT_STATUS_PNP_IRQ_TRANSLATION_FAILED:                                  ERROR_PNP_IRQ_TRANSLATION_FAILED,
	NT_STATUS_PNP_INVALID_ID:                                              ERROR_PNP_INVALID_ID,
	NT_STATUS_IO_REISSUE_AS_CACHED:                                        ERROR_IO_REISSUE_AS_CACHED,
	NT_STATUS_CTX_WINSTATION_NAME_INVALID:                                 ERROR_CTX_WINSTATION_NAME_INVALID,
	NT_STATUS_CTX_INVALID_PD:                                              ERROR_CTX_INVALID_PD,
	NT_STATUS_CTX_PD_NOT_FOUND:                                            ERROR_CTX_PD_NOT_FOUND,
	NT_STATUS_CTX_CLOSE_PENDING:                                           ERROR_CTX_CLOSE_PENDING,
	NT_STATUS_CTX_NO_OUTBUF:                                               ERROR_CTX_NO_OUTBUF,
	NT_STATUS_CTX_MODEM_INF_NOT_FOUND:                                     ERROR_CTX_MODEM_INF_NOT_FOUND,
	NT_STATUS_CTX_INVALID_MODEMNAME:                                       ERROR_CTX_INVALID_MODEMNAME,
	NT_STATUS_CTX_RESPONSE_ERROR:                                          ERROR_CTX_RESPONSE_ERROR,
	NT_STATUS_CTX_MODEM_RESPONSE_TIMEOUT:                                  ERROR_CTX_MODEM_RESPONSE_TIMEOUT,
	NT_STATUS_CTX_MODEM_RESPONSE_NO_CARRIER:                               ERROR_CTX_MODEM_RESPONSE_NO_CARRIER,
	NT_STATUS_CTX_MODEM_RESPONSE_NO_DIALTONE:                              ERROR_CTX_MODEM_RESPONSE_NO_DIALTONE,
	NT_STATUS_CTX_MODEM_RESPONSE_BUSY:                                     ERROR_CTX_MODEM_RESPONSE_BUSY,
	NT_STATUS_CTX_MODEM_RESPONSE_VOICE:                                    ERROR_CTX_MODEM_RESPONSE_VOICE,
	NT_STATUS_CTX_TD_ERROR:                                                ERROR_CTX_TD_ERROR,
	NT_STATUS_CTX_LICENSE_CLIENT_INVALID:                                  ERROR_CTX_LICENSE_CLIENT_INVALID,
	NT_STATUS_CTX_LICENSE_NOT_AVAILABLE:                                   ERROR_CTX_LICENSE_NOT_AVAILABLE,
	NT_STATUS_CTX_LICENSE_EXPIRED:                                         ERROR_CTX_LICENSE_EXPIRED,
	NT_STATUS_CTX_WINSTATION_NOT_FOUND:                                    ERROR_CTX_WINSTATION_NOT_FOUND,
	NT_STATUS_CTX_WINSTATION_NAME_COLLISION:                               ERROR_CTX_WINSTATION_NAME_COLLISION,
	NT_STATUS_CTX_WINSTATION_BUSY:                                         ERROR_CTX_WINSTATION_BUSY,
	NT_STATUS_CTX_BAD_VIDEO_MODE:                                          ERROR_CTX_BAD_VIDEO_MODE,
	NT_STATUS_CTX_GRAPHICS_INVALID:                                        ERROR_CTX_GRAPHICS_INVALID,
	NT_STATUS_CTX_NOT_CONSOLE:                                             ERROR_CTX_NOT_CONSOLE,
	NT_STATUS_CTX_CLIENT_QUERY_TIMEOUT:                                    ERROR_CTX_CLIENT_QUERY_TIMEOUT,
	NT_STATUS_CTX_CONSOLE_DISCONNECT:                                      ERROR_CTX_CONSOLE_DISCONNECT,
	NT_STATUS_CTX_CONSOLE_CONNECT:                                         ERROR_CTX_CONSOLE_CONNECT,
	NT_STATUS_CTX_SHADOW_DENIED:                                           ERROR_CTX_SHADOW_DENIED,
	NT_STATUS_CTX_WINSTATION_ACCESS_DENIED:                                ERROR_CTX_WINSTATION_ACCESS_DENIED,
	NT_STATUS_CTX_INVALID_WD:                                              ERROR_CTX_INVALID_WD,
	NT_STATUS_CTX_WD_NOT_FOUND:                                            ERROR_CTX_WD_NOT_FOUND,
	NT_STATUS_CTX_SHADOW_INVALID:                                          ERROR_CTX_SHADOW_INVALID,
	NT_STATUS_CTX_SHADOW_DISABLED:                                         ERROR_CTX_SHADOW_DISABLED,
	NT_STATUS_RDP_PROTOCOL_ERROR:                                          ERROR_RDP_PROTOCOL_ERROR,
	NT_STATUS_CTX_CLIENT_LICENSE_NOT_SET:                                  ERROR_CTX_CLIENT_LICENSE_NOT_SET,
	NT_STATUS_CTX_CLIENT_LICENSE_IN_USE:                                   ERROR_CTX_CLIENT_LICENSE_IN_USE,
	NT_STATUS_CTX_SHADOW_ENDED_BY_MODE_CHANGE:                             ERROR_CTX_SHADOW_ENDED_BY_MODE_CHANGE,
	NT_STATUS_CTX_SHADOW_NOT_RUNNING:                                      ERROR_CTX_SHADOW_NOT_RUNNING,
	NT_STATUS_CTX_LOGON_DISABLED:                                          ERROR_CTX_LOGON_DISABLED,
	NT_STATUS_CTX_SECURITY_LAYER_ERROR:                                    ERROR_CTX_SECURITY_LAYER_ERROR,
	NT_STATUS_TS_INCOMPATIBLE_SESSIONS:                                    ERROR_TS_INCOMPATIBLE_SESSIONS,
	NT_STATUS_MUI_FILE_NOT_FOUND:                                          ERROR_MUI_FILE_NOT_FOUND,
	NT_STATUS_MUI_INVALID_FILE:                                            ERROR_MUI_INVALID_FILE,
	NT_STATUS_MUI_INVALID_RC_CONFIG:                                       ERROR_MUI_INVALID_RC_CONFIG,
	NT_STATUS_MUI_INVALID_LOCALE_NAME:                                     ERROR_MUI_INVALID_LOCALE_NAME,
	NT_STATUS_MUI_INVALID_ULTIMATEFALLBACK_NAME:                           ERROR_MUI_INVALID_ULTIMATEFALLBACK_NAME,
	NT_STATUS_MUI_FILE_NOT_LOADED:                                         ERROR_MUI_FILE_NOT_LOADED,
	NT_STATUS_RESOURCE_ENUM_USER_STOP:                                     ERROR_RESOURCE_ENUM_USER_STOP,
	NT_STATUS_CLUSTER_INVALID_NODE:                                        ERROR_CLUSTER_INVALID_NODE,
	NT_STATUS_CLUSTER_NODE_EXISTS:                                         ERROR_CLUSTER_NODE_EXISTS,
	NT_STATUS_CLUSTER_JOIN_IN_PROGRESS:                                    ERROR_CLUSTER_JOIN_IN_PROGRESS,
	NT_STATUS_CLUSTER_NODE_NOT_FOUND:                                      ERROR_CLUSTER_NODE_NOT_FOUND,
	NT_STATUS_CLUSTER_LOCAL_NODE_NOT_FOUND:                                ERROR_CLUSTER_LOCAL_NODE_NOT_FOUND,
	NT_STATUS_CLUSTER_NETWORK_EXISTS:                                      ERROR_CLUSTER_NETWORK_EXISTS,
	NT_STATUS_CLUSTER_NETWORK_NOT_FOUND:                                   ERROR_CLUSTER_NETWORK_NOT_FOUND,
	NT_STATUS_CLUSTER_NETINTERFACE_EXISTS:                                 ERROR_CLUSTER_NETINTERFACE_EXISTS,
	NT_STATUS_CLUSTER_NETINTERFACE_NOT_FOUND:                              ERROR_CLUSTER_NETINTERFACE_NOT_FOUND,
	NT_STATUS_CLUSTER_INVALID_REQUEST:                                     ERROR_CLUSTER_INVALID_REQUEST,
	NT_STATUS_CLUSTER_INVALID_NETWORK_PROVIDER:                            ERROR_CLUSTER_INVALID_NETWORK_PROVIDER,
	NT_STATUS_CLUSTER_NODE_DOWN:                                           ERROR_CLUSTER_NODE_DOWN,
	NT_STATUS_CLUSTER_NODE_UNREACHABLE:                                    ERROR_CLUSTER_NODE_UNREACHABLE,
	NT_STATUS_CLUSTER_NODE_NOT_MEMBER:                                     ERROR_CLUSTER_NODE_NOT_MEMBER,
	NT_STATUS_CLUSTER_JOIN_NOT_IN_PROGRESS:                                ERROR_CLUSTER_JOIN_NOT_IN_PROGRESS,
	NT_STATUS_CLUSTER_INVALID_NETWORK:                                     ERROR_CLUSTER_INVALID_NETWORK,
	NT_STATUS_CLUSTER_NO_NET_ADAPTERS:                                     ERROR_CLUSTER_NO_NET_ADAPTERS,
	NT_STATUS_CLUSTER_NODE_UP:                                             ERROR_CLUSTER_NODE_UP,
	NT_STATUS_CLUSTER_NODE_PAUSED:                                         ERROR_CLUSTER_NODE_PAUSED,
	NT_STATUS_CLUSTER_NODE_NOT_PAUSED:                                     ERROR_CLUSTER_NODE_NOT_PAUSED,
	NT_STATUS_CLUSTER_NO_SECURITY_CONTEXT:                                 ERROR_CLUSTER_NO_SECURITY_CONTEXT,
	NT_STATUS_CLUSTER_NETWORK_NOT_INTERNAL:                                ERROR_CLUSTER_NETWORK_NOT_INTERNAL,
	NT_STATUS_CLUSTER_POISONED:                                            ERROR_CLUSTER_POISONED,
	NT_STATUS_ACPI_INVALID_OPCODE:                                         ERROR_ACPI_INVALID_OPCODE,
	NT_STATUS_ACPI_STACK_OVERFLOW:                                         ERROR_ACPI_STACK_OVERFLOW,
	NT_STATUS_ACPI_ASSERT_FAILED:                                          ERROR_ACPI_ASSERT_FAILED,
	NT_STATUS_ACPI_INVALID_INDEX:                                          ERROR_ACPI_INVALID_INDEX,
	NT_STATUS_ACPI_INVALID_ARGUMENT:                                       ERROR_ACPI_INVALID_ARGUMENT,
	NT_STATUS_ACPI_FATAL:                                                  ERROR_ACPI_FATAL,
	NT_STATUS_ACPI_INVALID_SUPERNAME:                                      ERROR_ACPI_INVALID_SUPERNAME,
	NT_STATUS_ACPI_INVALID_ARGTYPE:                                        ERROR_ACPI_INVALID_ARGTYPE,
	NT_STATUS_ACPI_INVALID_OBJTYPE:                                        ERROR_ACPI_INVALID_OBJTYPE,
	NT_STATUS_ACPI_INVALID_TARGETTYPE:                                     ERROR_ACPI_INVALID_TARGETTYPE,
	NT_STATUS_ACPI_INCORRECT_ARGUMENT_COUNT:                               ERROR_ACPI_INCORRECT_ARGUMENT_COUNT,
	NT_STATUS_ACPI_ADDRESS_NOT_MAPPED:                                     ERROR_ACPI_ADDRESS_NOT_MAPPED,
	NT_STATUS_ACPI_INVALID_EVENTTYPE:                                      ERROR_ACPI_INVALID_EVENTTYPE,
	NT_STATUS_ACPI_HANDLER_COLLISION:                                      ERROR_ACPI_HANDLER_COLLISION,
	NT_STATUS_ACPI_INVALID_DATA:                                           ERROR_ACPI_INVALID_DATA,
	NT_STATUS_ACPI_INVALID_REGION:                                         ERROR_ACPI_INVALID_REGION,
	NT_STATUS_ACPI_INVALID_ACCESS_SIZE:                                    ERROR_ACPI_INVALID_ACCESS_SIZE,
	NT_STATUS_ACPI_ACQUIRE_GLOBAL_LOCK:                                    ERROR_ACPI_ACQUIRE_GLOBAL_LOCK,
	NT_STATUS_ACPI_ALREADY_INITIALIZED:                                    ERROR_ACPI_ALREADY_INITIALIZED,
	NT_STATUS_ACPI_NOT_INITIALIZED:                                        ERROR_ACPI_NOT_INITIALIZED,
	NT_STATUS_ACPI_INVALID_MUTEX_LEVEL:                                    ERROR_ACPI_INVALID_MUTEX_LEVEL,
	NT_STATUS_ACPI_MUTEX_NOT_OWNED:                                        ERROR_ACPI_MUTEX_NOT_OWNED,
	NT_STATUS_ACPI_MUTEX_NOT_OWNER:                                        ERROR_ACPI_MUTEX_NOT_OWNER,
	NT_STATUS_ACPI_RS_ACCESS:                                              ERROR_ACPI_RS_ACCESS,
	NT_STATUS_ACPI_INVALID_TABLE:                                          ERROR_ACPI_INVALID_TABLE,
	NT_STATUS_ACPI_REG_HANDLER_FAILED:                                     ERROR_ACPI_REG_HANDLER_FAILED,
	NT_STATUS_ACPI_POWER_REQUEST_FAILED:                                   ERROR_ACPI_POWER_REQUEST_FAILED,
	NT_STATUS_SXS_SECTION_NOT_FOUND:                                       ERROR_SXS_SECTION_NOT_FOUND,
	NT_STATUS_SXS_CANT_GEN_ACTCTX:                                         ERROR_SXS_CANT_GEN_ACTCTX,
	NT_STATUS_SXS_INVALID_ACTCTXDATA_FORMAT:                               ERROR_SXS_INVALID_ACTCTXDATA_FORMAT,
	NT_STATUS_SXS_ASSEMBLY_NOT_FOUND:                                      ERROR_SXS_ASSEMBLY_NOT_FOUND,
	NT_STATUS_SXS_MANIFEST_FORMAT_ERROR:                                   ERROR_SXS_MANIFEST_FORMAT_ERROR,
	NT_STATUS_SXS_MANIFEST_PARSE_ERROR:                                    ERROR_SXS_MANIFEST_PARSE_ERROR,
	NT_STATUS_SXS_ACTIVATION_CONTEXT_DISABLED:                             ERROR_SXS_ACTIVATION_CONTEXT_DISABLED,
	NT_STATUS_SXS_KEY_NOT_FOUND:                                           ERROR_SXS_KEY_NOT_FOUND,
	NT_STATUS_SXS_VERSION_CONFLICT:                                        ERROR_SXS_VERSION_CONFLICT,
	NT_STATUS_SXS_WRONG_SECTION_TYPE:                                      ERROR_SXS_WRONG_SECTION_TYPE,
	NT_STATUS_SXS_THREAD_QUERIES_DISABLED:                                 ERROR_SXS_THREAD_QUERIES_DISABLED,
	NT_STATUS_SXS_ASSEMBLY_MISSING:                                        ERROR_SXS_ASSEMBLY_MISSING,
	NT_STATUS_SXS_PROCESS_DEFAULT_ALREADY_SET:                             ERROR_SXS_PROCESS_DEFAULT_ALREADY_SET,
	NT_STATUS_SXS_EARLY_DEACTIVATION:                                      ERROR_SXS_EARLY_DEACTIVATION,
	NT_STATUS_SXS_INVALID_DEACTIVATION:                                    ERROR_SXS_INVALID_DEACTIVATION,
	NT_STATUS_SXS_MULTIPLE_DEACTIVATION:                                   ERROR_SXS_MULTIPLE_DEACTIVATION,
	NT_STATUS_SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY:                 ERROR_SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY,
	NT_STATUS_SXS_PROCESS_TERMINATION_REQUESTED:                           ERROR_SXS_PROCESS_TERMINATION_REQUESTED,
	NT_STATUS_SXS_CORRUPT_ACTIVATION_STACK:                                ERROR_SXS_CORRUPT_ACTIVATION_STACK,
	NT_STATUS_SXS_CORRUPTION:                                              ERROR_SXS_CORRUPTION,
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE:                        ERROR_SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE,
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_NAME:                         ERROR_SXS_INVALID_IDENTITY_ATTRIBUTE_NAME,
	NT_STATUS_SXS_IDENTITY_DUPLICATE_ATTRIBUTE:                            ERROR_SXS_IDENTITY_DUPLICATE_ATTRIBUTE,
	NT_STATUS_SXS_IDENTITY_PARSE_ERROR:                                    ERROR_SXS_IDENTITY_PARSE_ERROR,
	NT_STATUS_SXS_COMPONENT_STORE_CORRUPT:                                 ERROR_SXS_COMPONENT_STORE_CORRUPT,
	NT_STATUS_SXS_FILE_HASH_MISMATCH:                                      ERROR_SXS_FILE_HASH_MISMATCH,
	NT_STATUS_SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT:           ERROR_SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT,
	NT_STATUS_SXS_IDENTITIES_DIFFERENT:                                    ERROR_SXS_IDENTITIES_DIFFERENT,
	NT_STATUS_SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT:                            ERROR_SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT,
	NT_STATUS_SXS_FILE_NOT_PART_OF_ASSEMBLY:                               ERROR_SXS_FILE_NOT_PART_OF_ASSEMBLY,
	NT_STATUS_ADVANCED_INSTALLER_FAILED:                                   ERROR_ADVANCED_INSTALLER_FAILED,
	NT_STATUS_XML_ENCODING_MISMATCH:                                       ERROR_XML_ENCODING_MISMATCH,
	NT_STATUS_SXS_MANIFEST_TOO_BIG:                                        ERROR_SXS_MANIFEST_TOO_BIG,
	NT_STATUS_SXS_SETTING_NOT_REGISTERED:                                  ERROR_SXS_SETTING_NOT_REGISTERED,
	NT_STATUS_SXS_TRANSACTION_CLOSURE_INCOMPLETE:                          ERROR_SXS_TRANSACTION_CLOSURE_INCOMPLETE,
	NT_STATUS_SMI_PRIMITIVE_INSTALLER_FAILED:                              ERROR_SMI_PRIMITIVE_INSTALLER_FAILED,
	NT_STATUS_GENERIC_COMMAND_FAILED:                                      ERROR_GENERIC_COMMAND_FAILED,
	NT_STATUS_SXS_FILE_HASH_MISSING:                                       ERROR_SXS_FILE_HASH_MISSING,
	NT_STATUS_TRANSACTIONAL_CONFLICT:                                      ERROR_TRANSACTIONAL_CONFLICT,
	NT_STATUS_INVALID_TRANSACTION:                                         ERROR_INVALID_TRANSACTION,
	NT_STATUS_TRANSACTION_NOT_ACTIVE:                                      ERROR_TRANSACTION_NOT_ACTIVE,
	NT_STATUS_TM_INITIALIZATION_FAILED:                                    ERROR_TM_INITIALIZATION_FAILED,
	NT_STATUS_RM_NOT_ACTIVE:                                               ERROR_RM_NOT_ACTIVE,
	NT_STATUS_RM_METADATA_CORRUPT:                                         ERROR_RM_METADATA_CORRUPT,
	NT_STATUS_TRANSACTION_NOT_JOINED:                                      ERROR_TRANSACTION_NOT_JOINED,
	NT_STATUS_DIRECTORY_NOT_RM:                                            ERROR_DIRECTORY_NOT_RM,
	NT_STATUS_TRANSACTIONS_UNSUPPORTED_REMOTE:                             ERROR_TRANSACTIONS_UNSUPPORTED_REMOTE,
	NT_STATUS_LOG_RESIZE_INVALID_SIZE:                                     ERROR_LOG_RESIZE_INVALID_SIZE,
	NT_STATUS_REMOTE_FILE_VERSION_MISMATCH:                                ERROR_REMOTE_FILE_VERSION_MISMATCH,
	NT_STATUS_CRM_PROTOCOL_ALREADY_EXISTS:                                 ERROR_CRM_PROTOCOL_ALREADY_EXISTS,
	NT_STATUS_TRANSACTION_PROPAGATION_FAILED:                              ERROR_TRANSACTION_PROPAGATION_FAILED,
	NT_STATUS_CRM_PROTOCOL_NOT_FOUND:                                      ERROR_CRM_PROTOCOL_NOT_FOUND,
	NT_STATUS_TRANSACTION_SUPERIOR_EXISTS:                                 ERROR_TRANSACTION_SUPERIOR_EXISTS,
	NT_STATUS_TRANSACTION_REQUEST_NOT_VALID:                               ERROR_TRANSACTION_REQUEST_NOT_VALID,
	NT_STATUS_TRANSACTION_NOT_REQUESTED:                                   ERROR_TRANSACTION_NOT_REQUESTED,
	NT_STATUS_TRANSACTION_ALREADY_ABORTED:                                 ERROR_TRANSACTION_ALREADY_ABORTED,
	NT_STATUS_TRANSACTION_ALREADY_COMMITTED:                               ERROR_TRANSACTION_ALREADY_COMMITTED,
	NT_STATUS_TRANSACTION_INVALID_MARSHALL_BUFFER:                         ERROR_TRANSACTION_INVALID_MARSHALL_BUFFER,
	NT_STATUS_CURRENT_TRANSACTION_NOT_VALID:                               ERROR_CURRENT_TRANSACTION_NOT_VALID,
	NT_STATUS_LOG_GROWTH_FAILED:                                           ERROR_LOG_GROWTH_FAILED,
	NT_STATUS_OBJECT_NO_LONGER_EXISTS:                                     ERROR_OBJECT_NO_LONGER_EXISTS,
	NT_STATUS_STREAM_MINIVERSION_NOT_FOUND:                                ERROR_STREAM_MINIVERSION_NOT_FOUND,
	NT_STATUS_STREAM_MINIVERSION_NOT_VALID:                                ERROR_STREAM_MINIVERSION_NOT_VALID,
	NT_STATUS_MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION:         ERROR_MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION,
	NT_STATUS_CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT:                    ERROR_CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT,
	NT_STATUS_CANT_CREATE_MORE_STREAM_MINIVERSIONS:                        ERROR_CANT_CREATE_MORE_STREAM_MINIVERSIONS,
	NT_STATUS_HANDLE_NO_LONGER_VALID:                                      ERROR_HANDLE_NO_LONGER_VALID,
	NT_STATUS_LOG_CORRUPTION_DETECTED:                                     ERROR_LOG_CORRUPTION_DETECTED,
	NT_STATUS_RM_DISCONNECTED:                                             ERROR_RM_DISCONNECTED,
	NT_STATUS_ENLISTMENT_NOT_SUPERIOR:                                     ERROR_ENLISTMENT_NOT_SUPERIOR,
	NT_STATUS_FILE_IDENTITY_NOT_PERSISTENT:                                ERROR_FILE_IDENTITY_NOT_PERSISTENT,
	NT_STATUS_CANT_BREAK_TRANSACTIONAL_DEPENDENCY:                         ERROR_CANT_BREAK_TRANSACTIONAL_DEPENDENCY,
	NT_STATUS_CANT_CROSS_RM_BOUNDARY:                                      ERROR_CANT_CROSS_RM_BOUNDARY,
	NT_STATUS_TXF_DIR_NOT_EMPTY:                                           ERROR_TXF_DIR_NOT_EMPTY,
	NT_STATUS_INDOUBT_TRANSACTIONS_EXIST:                                  ERROR_INDOUBT_TRANSACTIONS_EXIST,
	NT_STATUS_TM_VOLATILE:                                                 ERROR_TM_VOLATILE,
	NT_STATUS_ROLLBACK_TIMER_EXPIRED:                                      ERROR_ROLLBACK_TIMER_EXPIRED,
	NT_STATUS_TXF_ATTRIBUTE_CORRUPT:                                       ERROR_TXF_ATTRIBUTE_CORRUPT,
	NT_STATUS_EFS_NOT_ALLOWED_IN_TRANSACTION:                              ERROR_EFS_NOT_ALLOWED_IN_TRANSACTION,
	NT_STATUS_TRANSACTIONAL_OPEN_NOT_ALLOWED:                              ERROR_TRANSACTIONAL_OPEN_NOT_ALLOWED,
	NT_STATUS_TRANSACTED_MAPPING_UNSUPPORTED_REMOTE:                       ERROR_TRANSACTED_MAPPING_UNSUPPORTED_REMOTE,
	NT_STATUS_TRANSACTION_REQUIRED_PROMOTION:                              ERROR_TRANSACTION_REQUIRED_PROMOTION,
	NT_STATUS_CANNOT_EXECUTE_FILE_IN_TRANSACTION:                          ERROR_CANNOT_EXECUTE_FILE_IN_TRANSACTION,
	NT_STATUS_TRANSACTIONS_NOT_FROZEN:                                     ERROR_TRANSACTIONS_NOT_FROZEN,
	NT_STATUS_TRANSACTION_FREEZE_IN_PROGRESS:                              ERROR_TRANSACTION_FREEZE_IN_PROGRESS,
	NT_STATUS_NOT_SNAPSHOT_VOLUME:                                         ERROR_NOT_SNAPSHOT_VOLUME,
	NT_STATUS_NO_SAVEPOINT_WITH_OPEN_FILES:                                ERROR_NO_SAVEPOINT_WITH_OPEN_FILES,
	NT_STATUS_SPARSE_NOT_ALLOWED_IN_TRANSACTION:                           ERROR_SPARSE_NOT_ALLOWED_IN_TRANSACTION,
	NT_STATUS_TM_IDENTITY_MISMATCH:                                        ERROR_TM_IDENTITY_MISMATCH,
	NT_STATUS_FLOATED_SECTION:                                             ERROR_FLOATED_SECTION,
	NT_STATUS_CANNOT_ACCEPT_TRANSACTED_WORK:                               ERROR_CANNOT_ACCEPT_TRANSACTED_WORK,
	NT_STATUS_CANNOT_ABORT_TRANSACTIONS:                                   ERROR_CANNOT_ABORT_TRANSACTIONS,
	NT_STATUS_TRANSACTION_NOT_FOUND:                                       ERROR_TRANSACTION_NOT_FOUND,
	NT_STATUS_RESOURCEMANAGER_NOT_FOUND:                                   ERROR_RESOURCEMANAGER_NOT_FOUND,
	NT_STATUS_ENLISTMENT_NOT_FOUND:                                        ERROR_ENLISTMENT_NOT_FOUND,
	NT_STATUS_TRANSACTIONMANAGER_NOT_FOUND:                                ERROR_TRANSACTIONMANAGER_NOT_FOUND,
	NT_STATUS_TRANSACTIONMANAGER_NOT_ONLINE:                               ERROR_TRANSACTIONMANAGER_NOT_ONLINE,
	NT_STATUS_TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION:                  ERROR_TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION,
	NT_STATUS_TRANSACTION_NOT_ROOT:                                        ERROR_TRANSACTION_NOT_ROOT,
	NT_STATUS_TRANSACTION_OBJECT_EXPIRED:                                  ERROR_TRANSACTION_OBJECT_EXPIRED,
	NT_STATUS_COMPRESSION_NOT_ALLOWED_IN_TRANSACTION:                      ERROR_COMPRESSION_NOT_ALLOWED_IN_TRANSACTION,
	NT_STATUS_TRANSACTION_RESPONSE_NOT_ENLISTED:                           ERROR_TRANSACTION_RESPONSE_NOT_ENLISTED,
	NT_STATUS_TRANSACTION_RECORD_TOO_LONG:                                 ERROR_TRANSACTION_RECORD_TOO_LONG,
	NT_STATUS_NO_LINK_TRACKING_IN_TRANSACTION:                             ERROR_NO_LINK_TRACKING_IN_TRANSACTION,
	NT_STATUS_OPERATION_NOT_SUPPORTED_IN_TRANSACTION:                      ERROR_OPERATION_NOT_SUPPORTED_IN_TRANSACTION,
	NT_STATUS_TRANSACTION_INTEGRITY_VIOLATED:                              ERROR_TRANSACTION_INTEGRITY_VIOLATED,
	NT_STATUS_EXPIRED_HANDLE:                                              ERROR_EXPIRED_HANDLE,
	NT_STATUS_TRANSACTION_NOT_ENLISTED:                                    ERROR_TRANSACTION_NOT_ENLISTED,
	NT_STATUS_LOG_SECTOR_INVALID:                                          ERROR_LOG_SECTOR_INVALID,
	NT_STATUS_LOG_SECTOR_PARITY_INVALID:                                   ERROR_LOG_SECTOR_PARITY_INVALID,
	NT_STATUS_LOG_SECTOR_REMAPPED:                                         ERROR_LOG_SECTOR_REMAPPED,
	NT_STATUS_LOG_BLOCK_INCOMPLETE:                                        ERROR_LOG_BLOCK_INCOMPLETE,
	NT_STATUS_LOG_INVALID_RANGE:                                           ERROR_LOG_INVALID_RANGE,
	NT_STATUS_LOG_BLOCKS_EXHAUSTED:                                        ERROR_LOG_BLOCKS_EXHAUSTED,
	NT_STATUS_LOG_READ_CONTEXT_INVALID:                                    ERROR_LOG_READ_CONTEXT_INVALID,
	NT_STATUS_LOG_RESTART_INVALID:                                         ERROR_LOG_RESTART_INVALID,
	NT_STATUS_LOG_BLOCK_VERSION:                                           ERROR_LOG_BLOCK_VERSION,
	NT_STATUS_LOG_BLOCK_INVALID:                                           ERROR_LOG_BLOCK_INVALID,
	NT_STATUS_LOG_READ_MODE_INVALID:                                       ERROR_LOG_READ_MODE_INVALID,
	NT_STATUS_LOG_METADATA_CORRUPT:                                        ERROR_LOG_METADATA_CORRUPT,
	NT_STATUS_LOG_METADATA_INVALID:                                        ERROR_LOG_METADATA_INVALID,
	NT_STATUS_LOG_METADATA_INCONSISTENT:                                   ERROR_LOG_METADATA_INCONSISTENT,
	NT_STATUS_LOG_RESERVATION_INVALID:                                     ERROR_LOG_RESERVATION_INVALID,
	NT_STATUS_LOG_CANT_DELETE:                                             ERROR_LOG_CANT_DELETE,
	NT_STATUS_LOG_CONTAINER_LIMIT_EXCEEDED:                                ERROR_LOG_CONTAINER_LIMIT_EXCEEDED,
	NT_STATUS_LOG_START_OF_LOG:                                            ERROR_LOG_START_OF_LOG,
	NT_STATUS_LOG_POLICY_ALREADY_INSTALLED:                                ERROR_LOG_POLICY_ALREADY_INSTALLED,
	NT_STATUS_LOG_POLICY_NOT_INSTALLED:                                    ERROR_LOG_POLICY_NOT_INSTALLED,
	NT_STATUS_LOG_POLICY_INVALID:                                          ERROR_LOG_POLICY_INVALID,
	NT_STATUS_LOG_POLICY_CONFLICT:                                         ERROR_LOG_POLICY_CONFLICT,
	NT_STATUS_LOG_PINNED_ARCHIVE_TAIL:                                     ERROR_LOG_PINNED_ARCHIVE_TAIL,
	NT_STATUS_LOG_RECORD_NONEXISTENT:                                      ERROR_LOG_RECORD_NONEXISTENT,
	NT_STATUS_LOG_RECORDS_RESERVED_INVALID:                                ERROR_LOG_RECORDS_RESERVED_INVALID,
	NT_STATUS_LOG_SPACE_RESERVED_INVALID:                                  ERROR_LOG_SPACE_RESERVED_INVALID,
	NT_STATUS_LOG_TAIL_INVALID:                                            ERROR_LOG_TAIL_INVALID,
	NT_STATUS_LOG_FULL:                                                    ERROR_LOG_FULL,
	NT_STATUS_LOG_MULTIPLEXED:                                             ERROR_LOG_MULTIPLEXED,
	NT_STATUS_LOG_DEDICATED:                                               ERROR_LOG_DEDICATED,
	NT_STATUS_LOG_ARCHIVE_NOT_IN_PROGRESS:                                 ERROR_LOG_ARCHIVE_NOT_IN_PROGRESS,
	NT_STATUS_LOG_ARCHIVE_IN_PROGRESS:                                     ERROR_LOG_ARCHIVE_IN_PROGRESS,
	NT_STATUS_LOG_EPHEMERAL:                                               ERROR_LOG_EPHEMERAL,
	NT_STATUS_LOG_NOT_ENOUGH_CONTAINERS:                                   ERROR_LOG_NOT_ENOUGH_CONTAINERS,
	NT_STATUS_LOG_CLIENT_ALREADY_REGISTERED:                               ERROR_LOG_CLIENT_ALREADY_REGISTERED,
	NT_STATUS_LOG_CLIENT_NOT_REGISTERED:                                   ERROR_LOG_CLIENT_NOT_REGISTERED,
	NT_STATUS_LOG_FULL_HANDLER_IN_PROGRESS:                                ERROR_LOG_FULL_HANDLER_IN_PROGRESS,
	NT_STATUS_LOG_CONTAINER_READ_FAILED:                                   ERROR_LOG_CONTAINER_READ_FAILED,
	NT_STATUS_LOG_CONTAINER_WRITE_FAILED:                                  ERROR_LOG_CONTAINER_WRITE_FAILED,
	NT_STATUS_LOG_CONTAINER_OPEN_FAILED:                                   ERROR_LOG_CONTAINER_OPEN_FAILED,
	NT_STATUS_LOG_CONTAINER_STATE_INVALID:                                 ERROR_LOG_CONTAINER_STATE_INVALID,
	NT_STATUS_LOG_STATE_INVALID:                                           ERROR_LOG_STATE_INVALID,
	NT_STATUS_LOG_PINNED:                                                  ERROR_LOG_PINNED,
	NT_STATUS_LOG_METADATA_FLUSH_FAILED:                                   ERROR_LOG_METADATA_FLUSH_FAILED,
	NT_STATUS_LOG_INCONSISTENT_SECURITY:                                   ERROR_LOG_INCONSISTENT_SECURITY,
	NT_STATUS_LOG_APPENDED_FLUSH_FAILED:                                   ERROR_LOG_APPENDED_FLUSH_FAILED,
	NT_STATUS_LOG_PINNED_RESERVATION:                                      ERROR_LOG_PINNED_RESERVATION,
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD:                            ERROR_VIDEO_HUNG_DISPLAY_DRIVER_THREAD,
	NT_STATUS_FLT_NO_HANDLER_DEFINED:                                      ERROR_FLT_NO_HANDLER_DEFINED,
	NT_STATUS_FLT_CONTEXT_ALREADY_DEFINED:                                 ERROR_FLT_CONTEXT_ALREADY_DEFINED,
	NT_STATUS_FLT_INVALID_ASYNCHRONOUS_REQUEST:                            ERROR_FLT_INVALID_ASYNCHRONOUS_REQUEST,
	NT_STATUS_FLT_DISALLOW_FAST_IO:                                        ERROR_FLT_DISALLOW_FAST_IO,
	NT_STATUS_FLT_INVALID_NAME_REQUEST:                                    ERROR_FLT_INVALID_NAME_REQUEST,
	NT_STATUS_FLT_NOT_SAFE_TO_POST_OPERATION:                              ERROR_FLT_NOT_SAFE_TO_POST_OPERATION,
	NT_STATUS_FLT_NOT_INITIALIZED:                                         ERROR_FLT_NOT_INITIALIZED,
	NT_STATUS_FLT_FILTER_NOT_READY:                                        ERROR_FLT_FILTER_NOT_READY,
	NT_STATUS_FLT_POST_OPERATION_CLEANUP:                                  ERROR_FLT_POST_OPERATION_CLEANUP,
	NT_STATUS_FLT_INTERNAL_ERROR:                                          ERROR_FLT_INTERNAL_ERROR,
	NT_STATUS_FLT_DELETING_OBJECT:                                         ERROR_FLT_DELETING_OBJECT,
	NT_STATUS_FLT_MUST_BE_NONPAGED_POOL:                                   ERROR_FLT_MUST_BE_NONPAGED_POOL,
	NT_STATUS_FLT_DUPLICATE_ENTRY:                                         ERROR_FLT_DUPLICATE_ENTRY,
	NT_STATUS_FLT_CBDQ_DISABLED:                                           ERROR_FLT_CBDQ_DISABLED,
	NT_STATUS_FLT_DO_NOT_ATTACH:                                           ERROR_FLT_DO_NOT_ATTACH,
	NT_STATUS_FLT_DO_NOT_DETACH:                                           ERROR_FLT_DO_NOT_DETACH,
	NT_STATUS_FLT_INSTANCE_ALTITUDE_COLLISION:                             ERROR_FLT_INSTANCE_ALTITUDE_COLLISION,
	NT_STATUS_FLT_INSTANCE_NAME_COLLISION:                                 ERROR_FLT_INSTANCE_NAME_COLLISION,
	NT_STATUS_FLT_FILTER_NOT_FOUND:                                        ERROR_FLT_FILTER_NOT_FOUND,
	NT_STATUS_FLT_VOLUME_NOT_FOUND:                                        ERROR_FLT_VOLUME_NOT_FOUND,
	NT_STATUS_FLT_INSTANCE_NOT_FOUND:                                      ERROR_FLT_INSTANCE_NOT_FOUND,
	NT_STATUS_FLT_CONTEXT_ALLOCATION_NOT_FOUND:                            ERROR_FLT_CONTEXT_ALLOCATION_NOT_FOUND,
	NT_STATUS_FLT_INVALID_CONTEXT_REGISTRATION:                            ERROR_FLT_INVALID_CONTEXT_REGISTRATION,
	NT_STATUS_FLT_NAME_CACHE_MISS:                                         ERROR_FLT_NAME_CACHE_MISS,
	NT_STATUS_FLT_NO_DEVICE_OBJECT:                                        ERROR_FLT_NO_DEVICE_OBJECT,
	NT_STATUS_FLT_VOLUME_ALREADY_MOUNTED:                                  ERROR_FLT_VOLUME_ALREADY_MOUNTED,
	NT_STATUS_FLT_ALREADY_ENLISTED:                                        ERROR_FLT_ALREADY_ENLISTED,
	NT_STATUS_FLT_CONTEXT_ALREADY_LINKED:                                  ERROR_FLT_CONTEXT_ALREADY_LINKED,
	NT_STATUS_FLT_NO_WAITER_FOR_REPLY:                                     ERROR_FLT_NO_WAITER_FOR_REPLY,
	NT_STATUS_MONITOR_NO_DESCRIPTOR:                                       ERROR_MONITOR_NO_DESCRIPTOR,
	NT_STATUS_MONITOR_UNKNOWN_DESCRIPTOR_FORMAT:                           ERROR_MONITOR_UNKNOWN_DESCRIPTOR_FORMAT,
	NT_STATUS_MONITOR_INVALID_DESCRIPTOR_CHECKSUM:                         ERROR_MONITOR_INVALID_DESCRIPTOR_CHECKSUM,
	NT_STATUS_MONITOR_INVALID_STANDARD_TIMING_BLOCK:                       ERROR_MONITOR_INVALID_STANDARD_TIMING_BLOCK,
	NT_STATUS_MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED:                   ERROR_MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED,
	NT_STATUS_MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK:                  ERROR_MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK,
	NT_STATUS_MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK:                  ERROR_MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK,
	NT_STATUS_MONITOR_NO_MORE_DESCRIPTOR_DATA:                             ERROR_MONITOR_NO_MORE_DESCRIPTOR_DATA,
	NT_STATUS_MONITOR_INVALID_DETAILED_TIMING_BLOCK:                       ERROR_MONITOR_INVALID_DETAILED_TIMING_BLOCK,
	NT_STATUS_MONITOR_INVALID_MANUFACTURE_DATE:                            ERROR_MONITOR_INVALID_MANUFACTURE_DATE,
	NT_STATUS_GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER:                           ERROR_GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER,
	NT_STATUS_GRAPHICS_INSUFFICIENT_DMA_BUFFER:                            ERROR_GRAPHICS_INSUFFICIENT_DMA_BUFFER,
	NT_STATUS_GRAPHICS_INVALID_DISPLAY_ADAPTER:                            ERROR_GRAPHICS_INVALID_DISPLAY_ADAPTER,
	NT_STATUS_GRAPHICS_ADAPTER_WAS_RESET:                                  ERROR_GRAPHICS_ADAPTER_WAS_RESET,
	NT_STATUS_GRAPHICS_INVALID_DRIVER_MODEL:                               ERROR_GRAPHICS_INVALID_DRIVER_MODEL,
	NT_STATUS_GRAPHICS_PRESENT_MODE_CHANGED:                               ERROR_GRAPHICS_PRESENT_MODE_CHANGED,
	NT_STATUS_GRAPHICS_PRESENT_OCCLUDED:                                   ERROR_GRAPHICS_PRESENT_OCCLUDED,
	NT_STATUS_GRAPHICS_PRESENT_DENIED:                                     ERROR_GRAPHICS_PRESENT_DENIED,
	NT_STATUS_GRAPHICS_CANNOTCOLORCONVERT:                                 ERROR_GRAPHICS_CANNOTCOLORCONVERT,
	NT_STATUS_GRAPHICS_PRESENT_REDIRECTION_DISABLED:                       ERROR_GRAPHICS_PRESENT_REDIRECTION_DISABLED,
	NT_STATUS_GRAPHICS_PRESENT_UNOCCLUDED:                                 ERROR_GRAPHICS_PRESENT_UNOCCLUDED,
	NT_STATUS_GRAPHICS_NO_VIDEO_MEMORY:                                    ERROR_GRAPHICS_NO_VIDEO_MEMORY,
	NT_STATUS_GRAPHICS_CANT_LOCK_MEMORY:                                   ERROR_GRAPHICS_CANT_LOCK_MEMORY,
	NT_STATUS_GRAPHICS_ALLOCATION_BUSY:                                    ERROR_GRAPHICS_ALLOCATION_BUSY,
	NT_STATUS_GRAPHICS_TOO_MANY_REFERENCES:                                ERROR_GRAPHICS_TOO_MANY_REFERENCES,
	NT_STATUS_GRAPHICS_TRY_AGAIN_LATER:                                    ERROR_GRAPHICS_TRY_AGAIN_LATER,
	NT_STATUS_GRAPHICS_TRY_AGAIN_NOW:                                      ERROR_GRAPHICS_TRY_AGAIN_NOW,
	NT_STATUS_GRAPHICS_ALLOCATION_INVALID:                                 ERROR_GRAPHICS_ALLOCATION_INVALID,
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE:                   ERROR_GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE,
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED:                   ERROR_GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED,
	NT_STATUS_GRAPHICS_CANT_EVICT_PINNED_ALLOCATION:                       ERROR_GRAPHICS_CANT_EVICT_PINNED_ALLOCATION,
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_USAGE:                           ERROR_GRAPHICS_INVALID_ALLOCATION_USAGE,
	NT_STATUS_GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION:                      ERROR_GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION,
	NT_STATUS_GRAPHICS_ALLOCATION_CLOSED:                                  ERROR_GRAPHICS_ALLOCATION_CLOSED,
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_INSTANCE:                        ERROR_GRAPHICS_INVALID_ALLOCATION_INSTANCE,
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_HANDLE:                          ERROR_GRAPHICS_INVALID_ALLOCATION_HANDLE,
	NT_STATUS_GRAPHICS_WRONG_ALLOCATION_DEVICE:                            ERROR_GRAPHICS_WRONG_ALLOCATION_DEVICE,
	NT_STATUS_GRAPHICS_ALLOCATION_CONTENT_LOST:                            ERROR_GRAPHICS_ALLOCATION_CONTENT_LOST,
	NT_STATUS_GRAPHICS_GPU_EXCEPTION_ON_DEVICE:                            ERROR_GRAPHICS_GPU_EXCEPTION_ON_DEVICE,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY:                             ERROR_GRAPHICS_INVALID_VIDPN_TOPOLOGY,
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED:                       ERROR_GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED:             ERROR_GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_INVALID_VIDPN:                                      ERROR_GRAPHICS_INVALID_VIDPN,
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE:                       ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE,
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET:                       ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET,
	NT_STATUS_GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED:                       ERROR_GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_SOURCEMODESET:                        ERROR_GRAPHICS_INVALID_VIDPN_SOURCEMODESET,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGETMODESET:                        ERROR_GRAPHICS_INVALID_VIDPN_TARGETMODESET,
	NT_STATUS_GRAPHICS_INVALID_FREQUENCY:                                  ERROR_GRAPHICS_INVALID_FREQUENCY,
	NT_STATUS_GRAPHICS_INVALID_ACTIVE_REGION:                              ERROR_GRAPHICS_INVALID_ACTIVE_REGION,
	NT_STATUS_GRAPHICS_INVALID_TOTAL_REGION:                               ERROR_GRAPHICS_INVALID_TOTAL_REGION,
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE:                  ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE,
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE:                  ERROR_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE,
	NT_STATUS_GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET:                     ERROR_GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET,
	NT_STATUS_GRAPHICS_PATH_ALREADY_IN_TOPOLOGY:                           ERROR_GRAPHICS_PATH_ALREADY_IN_TOPOLOGY,
	NT_STATUS_GRAPHICS_MODE_ALREADY_IN_MODESET:                            ERROR_GRAPHICS_MODE_ALREADY_IN_MODESET,
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTSOURCESET:                      ERROR_GRAPHICS_INVALID_VIDEOPRESENTSOURCESET,
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTTARGETSET:                      ERROR_GRAPHICS_INVALID_VIDEOPRESENTTARGETSET,
	NT_STATUS_GRAPHICS_SOURCE_ALREADY_IN_SET:                              ERROR_GRAPHICS_SOURCE_ALREADY_IN_SET,
	NT_STATUS_GRAPHICS_TARGET_ALREADY_IN_SET:                              ERROR_GRAPHICS_TARGET_ALREADY_IN_SET,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_PRESENT_PATH:                         ERROR_GRAPHICS_INVALID_VIDPN_PRESENT_PATH,
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY:                      ERROR_GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET:                  ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE:                     ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE,
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_NOT_IN_SET:                          ERROR_GRAPHICS_FREQUENCYRANGE_NOT_IN_SET,
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET:                      ERROR_GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET,
	NT_STATUS_GRAPHICS_STALE_MODESET:                                      ERROR_GRAPHICS_STALE_MODESET,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCEMODESET:                      ERROR_GRAPHICS_INVALID_MONITOR_SOURCEMODESET,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCE_MODE:                        ERROR_GRAPHICS_INVALID_MONITOR_SOURCE_MODE,
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN:                    ERROR_GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN,
	NT_STATUS_GRAPHICS_MODE_ID_MUST_BE_UNIQUE:                             ERROR_GRAPHICS_MODE_ID_MUST_BE_UNIQUE,
	NT_STATUS_GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION:    ERROR_GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION,
	NT_STATUS_GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES:            ERROR_GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES,
	NT_STATUS_GRAPHICS_PATH_NOT_IN_TOPOLOGY:                               ERROR_GRAPHICS_PATH_NOT_IN_TOPOLOGY,
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE:              ERROR_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE,
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET:              ERROR_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET,
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTORSET:                       ERROR_GRAPHICS_INVALID_MONITORDESCRIPTORSET,
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTOR:                          ERROR_GRAPHICS_INVALID_MONITORDESCRIPTOR,
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET:                       ERROR_GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET,
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET:                   ERROR_GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET,
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE:                ERROR_GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE:                   ERROR_GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE,
	NT_STATUS_GRAPHICS_RESOURCES_NOT_RELATED:                              ERROR_GRAPHICS_RESOURCES_NOT_RELATED,
	NT_STATUS_GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE:                           ERROR_GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE,
	NT_STATUS_GRAPHICS_TARGET_ID_MUST_BE_UNIQUE:                           ERROR_GRAPHICS_TARGET_ID_MUST_BE_UNIQUE,
	NT_STATUS_GRAPHICS_NO_AVAILABLE_VIDPN_TARGET:                          ERROR_GRAPHICS_NO_AVAILABLE_VIDPN_TARGET,
	NT_STATUS_GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER:       ERROR_GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER,
	NT_STATUS_GRAPHICS_NO_VIDPNMGR:                                        ERROR_GRAPHICS_NO_VIDPNMGR,
	NT_STATUS_GRAPHICS_NO_ACTIVE_VIDPN:                                    ERROR_GRAPHICS_NO_ACTIVE_VIDPN,
	NT_STATUS_GRAPHICS_STALE_VIDPN_TOPOLOGY:                               ERROR_GRAPHICS_STALE_VIDPN_TOPOLOGY,
	NT_STATUS_GRAPHICS_MONITOR_NOT_CONNECTED:                              ERROR_GRAPHICS_MONITOR_NOT_CONNECTED,
	NT_STATUS_GRAPHICS_SOURCE_NOT_IN_TOPOLOGY:                             ERROR_GRAPHICS_SOURCE_NOT_IN_TOPOLOGY,
	NT_STATUS_GRAPHICS_INVALID_PRIMARYSURFACE_SIZE:                        ERROR_GRAPHICS_INVALID_PRIMARYSURFACE_SIZE,
	NT_STATUS_GRAPHICS_INVALID_VISIBLEREGION_SIZE:                         ERROR_GRAPHICS_INVALID_VISIBLEREGION_SIZE,
	NT_STATUS_GRAPHICS_INVALID_STRIDE:                                     ERROR_GRAPHICS_INVALID_STRIDE,
	NT_STATUS_GRAPHICS_INVALID_PIXELFORMAT:                                ERROR_GRAPHICS_INVALID_PIXELFORMAT,
	NT_STATUS_GRAPHICS_INVALID_COLORBASIS:                                 ERROR_GRAPHICS_INVALID_COLORBASIS,
	NT_STATUS_GRAPHICS_INVALID_PIXELVALUEACCESSMODE:                       ERROR_GRAPHICS_INVALID_PIXELVALUEACCESSMODE,
	NT_STATUS_GRAPHICS_TARGET_NOT_IN_TOPOLOGY:                             ERROR_GRAPHICS_TARGET_NOT_IN_TOPOLOGY,
	NT_STATUS_GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT:                 ERROR_GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT,
	NT_STATUS_GRAPHICS_VIDPN_SOURCE_IN_USE:                                ERROR_GRAPHICS_VIDPN_SOURCE_IN_USE,
	NT_STATUS_GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN:                           ERROR_GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN,
	NT_STATUS_GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL:                    ERROR_GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL,
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION:       ERROR_GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION,
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED: ERROR_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_INVALID_GAMMA_RAMP:                                 ERROR_GRAPHICS_INVALID_GAMMA_RAMP,
	NT_STATUS_GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED:                           ERROR_GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_MULTISAMPLING_NOT_SUPPORTED:                        ERROR_GRAPHICS_MULTISAMPLING_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_MODE_NOT_IN_MODESET:                                ERROR_GRAPHICS_MODE_NOT_IN_MODESET,
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON:       ERROR_GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON,
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_TYPE:                          ERROR_GRAPHICS_INVALID_PATH_CONTENT_TYPE,
	NT_STATUS_GRAPHICS_INVALID_COPYPROTECTION_TYPE:                        ERROR_GRAPHICS_INVALID_COPYPROTECTION_TYPE,
	NT_STATUS_GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS:                  ERROR_GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS,
	NT_STATUS_GRAPHICS_INVALID_SCANLINE_ORDERING:                          ERROR_GRAPHICS_INVALID_SCANLINE_ORDERING,
	NT_STATUS_GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED:                       ERROR_GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED,
	NT_STATUS_GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS:                   ERROR_GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS,
	NT_STATUS_GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT:                        ERROR_GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT,
	NT_STATUS_GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM:                     ERROR_GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN:                  ERROR_GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN,
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT:          ERROR_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT,
	NT_STATUS_GRAPHICS_MAX_NUM_PATHS_REACHED:                              ERROR_GRAPHICS_MAX_NUM_PATHS_REACHED,
	NT_STATUS_GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION:                 ERROR_GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION,
	NT_STATUS_GRAPHICS_INVALID_CLIENT_TYPE:                                ERROR_GRAPHICS_INVALID_CLIENT_TYPE,
	NT_STATUS_GRAPHICS_CLIENTVIDPN_NOT_SET:                                ERROR_GRAPHICS_CLIENTVIDPN_NOT_SET,
	NT_STATUS_GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED:                  ERROR_GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED,
	NT_STATUS_GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED:                     ERROR_GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_NOT_A_LINKED_ADAPTER:                               ERROR_GRAPHICS_NOT_A_LINKED_ADAPTER,
	NT_STATUS_GRAPHICS_LEADLINK_NOT_ENUMERATED:                            ERROR_GRAPHICS_LEADLINK_NOT_ENUMERATED,
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_ENUMERATED:                          ERROR_GRAPHICS_CHAINLINKS_NOT_ENUMERATED,
	NT_STATUS_GRAPHICS_ADAPTER_CHAIN_NOT_READY:                            ERROR_GRAPHICS_ADAPTER_CHAIN_NOT_READY,
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_STARTED:                             ERROR_GRAPHICS_CHAINLINKS_NOT_STARTED,
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_POWERED_ON:                          ERROR_GRAPHICS_CHAINLINKS_NOT_POWERED_ON,
	NT_STATUS_GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE:                     ERROR_GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE,
	NT_STATUS_GRAPHICS_NOT_POST_DEVICE_DRIVER:                             ERROR_GRAPHICS_NOT_POST_DEVICE_DRIVER,
	NT_STATUS_GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED:                        ERROR_GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED,
	NT_STATUS_GRAPHICS_OPM_NOT_SUPPORTED:                                  ERROR_GRAPHICS_OPM_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_COPP_NOT_SUPPORTED:                                 ERROR_GRAPHICS_COPP_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_UAB_NOT_SUPPORTED:                                  ERROR_GRAPHICS_UAB_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS:                   ERROR_GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS,
	NT_STATUS_GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL:                      ERROR_GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL,
	NT_STATUS_GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST:                     ERROR_GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST,
	NT_STATUS_GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME:          ERROR_GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME,
	NT_STATUS_GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP:         ERROR_GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP,
	NT_STATUS_GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED:                ERROR_GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_OPM_INVALID_POINTER:                                ERROR_GRAPHICS_OPM_INVALID_POINTER,
	NT_STATUS_GRAPHICS_OPM_INTERNAL_ERROR:                                 ERROR_GRAPHICS_OPM_INTERNAL_ERROR,
	NT_STATUS_GRAPHICS_OPM_INVALID_HANDLE:                                 ERROR_GRAPHICS_OPM_INVALID_HANDLE,
	NT_STATUS_GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE:       ERROR_GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE,
	NT_STATUS_GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH:                     ERROR_GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH,
	NT_STATUS_GRAPHICS_OPM_SPANNING_MODE_ENABLED:                          ERROR_GRAPHICS_OPM_SPANNING_MODE_ENABLED,
	NT_STATUS_GRAPHICS_OPM_THEATER_MODE_ENABLED:                           ERROR_GRAPHICS_OPM_THEATER_MODE_ENABLED,
	NT_STATUS_GRAPHICS_PVP_HFS_FAILED:                                     ERROR_GRAPHICS_PVP_HFS_FAILED,
	NT_STATUS_GRAPHICS_OPM_INVALID_SRM:                                    ERROR_GRAPHICS_OPM_INVALID_SRM,
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP:                   ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP,
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP:                    ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP,
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA:                  ERROR_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA,
	NT_STATUS_GRAPHICS_OPM_HDCP_SRM_NEVER_SET:                             ERROR_GRAPHICS_OPM_HDCP_SRM_NEVER_SET,
	NT_STATUS_GRAPHICS_OPM_RESOLUTION_TOO_HIGH:                            ERROR_GRAPHICS_OPM_RESOLUTION_TOO_HIGH,
	NT_STATUS_GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE:               ERROR_GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE,
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS:              ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS,
	NT_STATUS_GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS:                ERROR_GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS,
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS:  ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS,
	NT_STATUS_GRAPHICS_OPM_INVALID_INFORMATION_REQUEST:                    ERROR_GRAPHICS_OPM_INVALID_INFORMATION_REQUEST,
	NT_STATUS_GRAPHICS_OPM_DRIVER_INTERNAL_ERROR:                          ERROR_GRAPHICS_OPM_DRIVER_INTERNAL_ERROR,
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS:   ERROR_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS,
	NT_STATUS_GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED:                        ERROR_GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST:                  ERROR_GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST,
	NT_STATUS_GRAPHICS_I2C_NOT_SUPPORTED:                                  ERROR_GRAPHICS_I2C_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST:                          ERROR_GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST,
	NT_STATUS_GRAPHICS_I2C_ERROR_TRANSMITTING_DATA:                        ERROR_GRAPHICS_I2C_ERROR_TRANSMITTING_DATA,
	NT_STATUS_GRAPHICS_I2C_ERROR_RECEIVING_DATA:                           ERROR_GRAPHICS_I2C_ERROR_RECEIVING_DATA,
	NT_STATUS_GRAPHICS_DDCCI_VCP_NOT_SUPPORTED:                            ERROR_GRAPHICS_DDCCI_VCP_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_DDCCI_INVALID_DATA:                                 ERROR_GRAPHICS_DDCCI_INVALID_DATA,
	NT_STATUS_GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE:  ERROR_GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE,
	NT_STATUS_GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING:                  ERROR_GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING,
	NT_STATUS_GRAPHICS_MCA_INTERNAL_ERROR:                                 ERROR_GRAPHICS_MCA_INTERNAL_ERROR,
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND:                      ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND,
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH:                       ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH,
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM:                     ERROR_GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM,
	NT_STATUS_GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE:                    ERROR_GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE,
	NT_STATUS_GRAPHICS_MONITOR_NO_LONGER_EXISTS:                           ERROR_GRAPHICS_MONITOR_NO_LONGER_EXISTS,
	NT_STATUS_GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED:                     ERROR_GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED,
	NT_STATUS_GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME:              ERROR_GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME,
	NT_STATUS_GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP:             ERROR_GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP,
	NT_STATUS_GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED:                    ERROR_GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED,
	NT_STATUS_GRAPHICS_INVALID_POINTER:                                    ERROR_GRAPHICS_INVALID_POINTER,
	NT_STATUS_GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE:           ERROR_GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE,
	NT_STATUS_GRAPHICS_PARAMETER_ARRAY_TOO_SMALL:                          ERROR_GRAPHICS_PARAMETER_ARRAY_TOO_SMALL,
	NT_STATUS_GRAPHICS_INTERNAL_ERROR:                                     ERROR_GRAPHICS_INTERNAL_ERROR,
	NT_STATUS_GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS:                    ERROR_GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS,
	NT_STATUS_FVE_LOCKED_VOLUME:                                           ERROR_FVE_LOCKED_VOLUME,
	NT_STATUS_FVE_NOT_ENCRYPTED:                                           ERROR_FVE_NOT_ENCRYPTED,
	NT_STATUS_FVE_BAD_INFORMATION:                                         ERROR_FVE_BAD_INFORMATION,
	NT_STATUS_FVE_TOO_SMALL:                                               ERROR_FVE_TOO_SMALL,
	NT_STATUS_FVE_FAILED_WRONG_FS:                                         ERROR_FVE_FAILED_WRONG_FS,
	NT_STATUS_FVE_FAILED_BAD_FS:                                           ERROR_FVE_FAILED_BAD_FS,
	NT_STATUS_FVE_FS_NOT_EXTENDED:                                         ERROR_FVE_FS_NOT_EXTENDED,
	NT_STATUS_FVE_FS_MOUNTED:                                              ERROR_FVE_FS_MOUNTED,
	NT_STATUS_FVE_NO_LICENSE:                                              ERROR_FVE_NO_LICENSE,
	NT_STATUS_FVE_ACTION_NOT_ALLOWED:                                      ERROR_FVE_ACTION_NOT_ALLOWED,
	NT_STATUS_FVE_BAD_DATA:                                                ERROR_FVE_BAD_DATA,
	NT_STATUS_FVE_VOLUME_NOT_BOUND:                                        ERROR_FVE_VOLUME_NOT_BOUND,
	NT_STATUS_FVE_NOT_DATA_VOLUME:                                         ERROR_FVE_NOT_DATA_VOLUME,
	NT_STATUS_FVE_CONV_READ_ERROR:                                         ERROR_FVE_CONV_READ_ERROR,
	NT_STATUS_FVE_CONV_WRITE_ERROR:                                        ERROR_FVE_CONV_WRITE_ERROR,
	NT_STATUS_FVE_OVERLAPPED_UPDATE:                                       ERROR_FVE_OVERLAPPED_UPDATE,
	NT_STATUS_FVE_FAILED_SECTOR_SIZE:                                      ERROR_FVE_FAILED_SECTOR_SIZE,
	NT_STATUS_FVE_FAILED_AUTHENTICATION:                                   ERROR_FVE_FAILED_AUTHENTICATION,
	NT_STATUS_FVE_NOT_OS_VOLUME:                                           ERROR_FVE_NOT_OS_VOLUME,
	NT_STATUS_FVE_KEYFILE_NOT_FOUND:                                       ERROR_FVE_KEYFILE_NOT_FOUND,
	NT_STATUS_FVE_KEYFILE_INVALID:                                         ERROR_FVE_KEYFILE_INVALID,
	NT_STATUS_FVE_KEYFILE_NO_VMK:                                          ERROR_FVE_KEYFILE_NO_VMK,
	NT_STATUS_FVE_TPM_DISABLED:                                            ERROR_FVE_TPM_DISABLED,
	NT_STATUS_FVE_TPM_SRK_AUTH_NOT_ZERO:                                   ERROR_FVE_TPM_SRK_AUTH_NOT_ZERO,
	NT_STATUS_FVE_TPM_INVALID_PCR:                                         ERROR_FVE_TPM_INVALID_PCR,
	NT_STATUS_FVE_TPM_NO_VMK:                                              ERROR_FVE_TPM_NO_VMK,
	NT_STATUS_FVE_PIN_INVALID:                                             ERROR_FVE_PIN_INVALID,
	NT_STATUS_FVE_AUTH_INVALID_APPLICATION:                                ERROR_FVE_AUTH_INVALID_APPLICATION,
	NT_STATUS_FVE_AUTH_INVALID_CONFIG:                                     ERROR_FVE_AUTH_INVALID_CONFIG,
	NT_STATUS_FVE_DEBUGGER_ENABLED:                                        ERROR_FVE_DEBUGGER_ENABLED,
	NT_STATUS_FVE_DRY_RUN_FAILED:                                          ERROR_FVE_DRY_RUN_FAILED,
	NT_STATUS_FVE_BAD_METADATA_POINTER:                                    ERROR_FVE_BAD_METADATA_POINTER,
	NT_STATUS_FVE_OLD_METADATA_COPY:                                       ERROR_FVE_OLD_METADATA_COPY,
	NT_STATUS_FVE_REBOOT_REQUIRED:                                         ERROR_FVE_REBOOT_REQUIRED,
	NT_STATUS_FVE_RAW_ACCESS:                                              ERROR_FVE_RAW_ACCESS,
	NT_STATUS_FVE_RAW_BLOCKED:                                             ERROR_FVE_RAW_BLOCKED,
	NT_STATUS_FVE_NO_FEATURE_LICENSE:                                      ERROR_FVE_NO_FEATURE_LICENSE,
	NT_STATUS_FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED:                     ERROR_FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED,
	NT_STATUS_FVE_CONV_RECOVERY_FAILED:                                    ERROR_FVE_CONV_RECOVERY_FAILED,
	NT_STATUS_FVE_VIRTUALIZED_SPACE_TOO_BIG:                               ERROR_FVE_VIRTUALIZED_SPACE_TOO_BIG,
	NT_STATUS_FVE_VOLUME_TOO_SMALL:                                        ERROR_FVE_VOLUME_TOO_SMALL,
	NT_STATUS_FWP_CALLOUT_NOT_FOUND:                                       ERROR_FWP_CALLOUT_NOT_FOUND,
	NT_STATUS_FWP_CONDITION_NOT_FOUND:                                     ERROR_FWP_CONDITION_NOT_FOUND,
	NT_STATUS_FWP_FILTER_NOT_FOUND:                                        ERROR_FWP_FILTER_NOT_FOUND,
	NT_STATUS_FWP_LAYER_NOT_FOUND:                                         ERROR_FWP_LAYER_NOT_FOUND,
	NT_STATUS_FWP_PROVIDER_NOT_FOUND:                                      ERROR_FWP_PROVIDER_NOT_FOUND,
	NT_STATUS_FWP_PROVIDER_CONTEXT_NOT_FOUND:                              ERROR_FWP_PROVIDER_CONTEXT_NOT_FOUND,
	NT_STATUS_FWP_SUBLAYER_NOT_FOUND:                                      ERROR_FWP_SUBLAYER_NOT_FOUND,
	NT_STATUS_FWP_NOT_FOUND:                                               ERROR_FWP_NOT_FOUND,
	NT_STATUS_FWP_ALREADY_EXISTS:                                          ERROR_FWP_ALREADY_EXISTS,
	NT_STATUS_FWP_IN_USE:                                                  ERROR_FWP_IN_USE,
	NT_STATUS_FWP_DYNAMIC_SESSION_IN_PROGRESS:                             ERROR_FWP_DYNAMIC_SESSION_IN_PROGRESS,
	NT_STATUS_FWP_WRONG_SESSION:                                           ERROR_FWP_WRONG_SESSION,
	NT_STATUS_FWP_NO_TXN_IN_PROGRESS:                                      ERROR_FWP_NO_TXN_IN_PROGRESS,
	NT_STATUS_FWP_TXN_IN_PROGRESS:                                         ERROR_FWP_TXN_IN_PROGRESS,
	NT_STATUS_FWP_TXN_ABORTED:                                             ERROR_FWP_TXN_ABORTED,
	NT_STATUS_FWP_SESSION_ABORTED:                                         ERROR_FWP_SESSION_ABORTED,
	NT_STATUS_FWP_INCOMPATIBLE_TXN:                                        ERROR_FWP_INCOMPATIBLE_TXN,
	NT_STATUS_FWP_TIMEOUT:                                                 ERROR_FWP_TIMEOUT,
	NT_STATUS_FWP_NET_EVENTS_DISABLED:                                     ERROR_FWP_NET_EVENTS_DISABLED,
	NT_STATUS_FWP_INCOMPATIBLE_LAYER:                                      ERROR_FWP_INCOMPATIBLE_LAYER,
	NT_STATUS_FWP_KM_CLIENTS_ONLY:                                         ERROR_FWP_KM_CLIENTS_ONLY,
	NT_STATUS_FWP_LIFETIME_MISMATCH:                                       ERROR_FWP_LIFETIME_MISMATCH,
	NT_STATUS_FWP_BUILTIN_OBJECT:                                          ERROR_FWP_BUILTIN_OBJECT,
	NT_STATUS_FWP_TOO_MANY_BOOTTIME_FILTERS:                               ERROR_FWP_TOO_MANY_BOOTTIME_FILTERS,
	// NT_STATUS_FWP_TOO_MANY_CALLOUTS: // NT_STATUS_FWP_TOO_MANY_CALLOUTS,
	NT_STATUS_FWP_NOTIFICATION_DROPPED:               ERROR_FWP_NOTIFICATION_DROPPED,
	NT_STATUS_FWP_TRAFFIC_MISMATCH:                   ERROR_FWP_TRAFFIC_MISMATCH,
	NT_STATUS_FWP_INCOMPATIBLE_SA_STATE:              ERROR_FWP_INCOMPATIBLE_SA_STATE,
	NT_STATUS_FWP_NULL_POINTER:                       ERROR_FWP_NULL_POINTER,
	NT_STATUS_FWP_INVALID_ENUMERATOR:                 ERROR_FWP_INVALID_ENUMERATOR,
	NT_STATUS_FWP_INVALID_FLAGS:                      ERROR_FWP_INVALID_FLAGS,
	NT_STATUS_FWP_INVALID_NET_MASK:                   ERROR_FWP_INVALID_NET_MASK,
	NT_STATUS_FWP_INVALID_RANGE:                      ERROR_FWP_INVALID_RANGE,
	NT_STATUS_FWP_INVALID_INTERVAL:                   ERROR_FWP_INVALID_INTERVAL,
	NT_STATUS_FWP_ZERO_LENGTH_ARRAY:                  ERROR_FWP_ZERO_LENGTH_ARRAY,
	NT_STATUS_FWP_NULL_DISPLAY_NAME:                  ERROR_FWP_NULL_DISPLAY_NAME,
	NT_STATUS_FWP_INVALID_ACTION_TYPE:                ERROR_FWP_INVALID_ACTION_TYPE,
	NT_STATUS_FWP_INVALID_WEIGHT:                     ERROR_FWP_INVALID_WEIGHT,
	NT_STATUS_FWP_MATCH_TYPE_MISMATCH:                ERROR_FWP_MATCH_TYPE_MISMATCH,
	NT_STATUS_FWP_TYPE_MISMATCH:                      ERROR_FWP_TYPE_MISMATCH,
	NT_STATUS_FWP_OUT_OF_BOUNDS:                      ERROR_FWP_OUT_OF_BOUNDS,
	NT_STATUS_FWP_RESERVED:                           ERROR_FWP_RESERVED,
	NT_STATUS_FWP_DUPLICATE_CONDITION:                ERROR_FWP_DUPLICATE_CONDITION,
	NT_STATUS_FWP_DUPLICATE_KEYMOD:                   ERROR_FWP_DUPLICATE_KEYMOD,
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_LAYER:     ERROR_FWP_ACTION_INCOMPATIBLE_WITH_LAYER,
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER:  ERROR_FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER,
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER:    ERROR_FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER,
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT:  ERROR_FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT,
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_METHOD:           ERROR_FWP_INCOMPATIBLE_AUTH_METHOD,
	NT_STATUS_FWP_INCOMPATIBLE_DH_GROUP:              ERROR_FWP_INCOMPATIBLE_DH_GROUP,
	NT_STATUS_FWP_EM_NOT_SUPPORTED:                   ERROR_FWP_EM_NOT_SUPPORTED,
	NT_STATUS_FWP_NEVER_MATCH:                        ERROR_FWP_NEVER_MATCH,
	NT_STATUS_FWP_PROVIDER_CONTEXT_MISMATCH:          ERROR_FWP_PROVIDER_CONTEXT_MISMATCH,
	NT_STATUS_FWP_INVALID_PARAMETER:                  ERROR_FWP_INVALID_PARAMETER,
	NT_STATUS_FWP_TOO_MANY_SUBLAYERS:                 ERROR_FWP_TOO_MANY_SUBLAYERS,
	NT_STATUS_FWP_CALLOUT_NOTIFICATION_FAILED:        ERROR_FWP_CALLOUT_NOTIFICATION_FAILED,
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_CONFIG:           ERROR_FWP_INCOMPATIBLE_AUTH_CONFIG,
	NT_STATUS_FWP_INCOMPATIBLE_CIPHER_CONFIG:         ERROR_FWP_INCOMPATIBLE_CIPHER_CONFIG,
	NT_STATUS_FWP_DUPLICATE_AUTH_METHOD:              ERROR_FWP_DUPLICATE_AUTH_METHOD,
	NT_STATUS_FWP_TCPIP_NOT_READY:                    ERROR_FWP_TCPIP_NOT_READY,
	NT_STATUS_FWP_INJECT_HANDLE_CLOSING:              ERROR_FWP_INJECT_HANDLE_CLOSING,
	NT_STATUS_FWP_INJECT_HANDLE_STALE:                ERROR_FWP_INJECT_HANDLE_STALE,
	NT_STATUS_FWP_CANNOT_PEND:                        ERROR_FWP_CANNOT_PEND,
	NT_STATUS_NDIS_CLOSING:                           ERROR_NDIS_CLOSING,
	NT_STATUS_NDIS_BAD_VERSION:                       ERROR_NDIS_BAD_VERSION,
	NT_STATUS_NDIS_BAD_CHARACTERISTICS:               ERROR_NDIS_BAD_CHARACTERISTICS,
	NT_STATUS_NDIS_ADAPTER_NOT_FOUND:                 ERROR_NDIS_ADAPTER_NOT_FOUND,
	NT_STATUS_NDIS_OPEN_FAILED:                       ERROR_NDIS_OPEN_FAILED,
	NT_STATUS_NDIS_DEVICE_FAILED:                     ERROR_NDIS_DEVICE_FAILED,
	NT_STATUS_NDIS_MULTICAST_FULL:                    ERROR_NDIS_MULTICAST_FULL,
	NT_STATUS_NDIS_MULTICAST_EXISTS:                  ERROR_NDIS_MULTICAST_EXISTS,
	NT_STATUS_NDIS_MULTICAST_NOT_FOUND:               ERROR_NDIS_MULTICAST_NOT_FOUND,
	NT_STATUS_NDIS_REQUEST_ABORTED:                   ERROR_NDIS_REQUEST_ABORTED,
	NT_STATUS_NDIS_RESET_IN_PROGRESS:                 ERROR_NDIS_RESET_IN_PROGRESS,
	NT_STATUS_NDIS_INVALID_PACKET:                    ERROR_NDIS_INVALID_PACKET,
	NT_STATUS_NDIS_INVALID_DEVICE_REQUEST:            ERROR_NDIS_INVALID_DEVICE_REQUEST,
	NT_STATUS_NDIS_ADAPTER_NOT_READY:                 ERROR_NDIS_ADAPTER_NOT_READY,
	NT_STATUS_NDIS_INVALID_LENGTH:                    ERROR_NDIS_INVALID_LENGTH,
	NT_STATUS_NDIS_INVALID_DATA:                      ERROR_NDIS_INVALID_DATA,
	NT_STATUS_NDIS_BUFFER_TOO_SHORT:                  ERROR_NDIS_BUFFER_TOO_SHORT,
	NT_STATUS_NDIS_INVALID_OID:                       ERROR_NDIS_INVALID_OID,
	NT_STATUS_NDIS_ADAPTER_REMOVED:                   ERROR_NDIS_ADAPTER_REMOVED,
	NT_STATUS_NDIS_UNSUPPORTED_MEDIA:                 ERROR_NDIS_UNSUPPORTED_MEDIA,
	NT_STATUS_NDIS_GROUP_ADDRESS_IN_USE:              ERROR_NDIS_GROUP_ADDRESS_IN_USE,
	NT_STATUS_NDIS_FILE_NOT_FOUND:                    ERROR_NDIS_FILE_NOT_FOUND,
	NT_STATUS_NDIS_ERROR_READING_FILE:                ERROR_NDIS_ERROR_READING_FILE,
	NT_STATUS_NDIS_ALREADY_MAPPED:                    ERROR_NDIS_ALREADY_MAPPED,
	NT_STATUS_NDIS_RESOURCE_CONFLICT:                 ERROR_NDIS_RESOURCE_CONFLICT,
	NT_STATUS_NDIS_MEDIA_DISCONNECTED:                ERROR_NDIS_MEDIA_DISCONNECTED,
	NT_STATUS_NDIS_INVALID_ADDRESS:                   ERROR_NDIS_INVALID_ADDRESS,
	NT_STATUS_NDIS_PAUSED:                            ERROR_NDIS_PAUSED,
	NT_STATUS_NDIS_INTERFACE_NOT_FOUND:               ERROR_NDIS_INTERFACE_NOT_FOUND,
	NT_STATUS_NDIS_UNSUPPORTED_REVISION:              ERROR_NDIS_UNSUPPORTED_REVISION,
	NT_STATUS_NDIS_INVALID_PORT:                      ERROR_NDIS_INVALID_PORT,
	NT_STATUS_NDIS_INVALID_PORT_STATE:                ERROR_NDIS_INVALID_PORT_STATE,
	NT_STATUS_NDIS_LOW_POWER_STATE:                   ERROR_NDIS_LOW_POWER_STATE,
	NT_STATUS_NDIS_NOT_SUPPORTED:                     ERROR_NDIS_NOT_SUPPORTED,
	NT_STATUS_NDIS_OFFLOAD_POLICY:                    ERROR_NDIS_OFFLOAD_POLICY,
	NT_STATUS_NDIS_OFFLOAD_CONNECTION_REJECTED:       ERROR_NDIS_OFFLOAD_CONNECTION_REJECTED,
	NT_STATUS_NDIS_OFFLOAD_PATH_REJECTED:             ERROR_NDIS_OFFLOAD_PATH_REJECTED,
	NT_STATUS_NDIS_DOT11_AUTO_CONFIG_ENABLED:         ERROR_NDIS_DOT11_AUTO_CONFIG_ENABLED,
	NT_STATUS_NDIS_DOT11_MEDIA_IN_USE:                ERROR_NDIS_DOT11_MEDIA_IN_USE,
	NT_STATUS_NDIS_DOT11_POWER_STATE_INVALID:         ERROR_NDIS_DOT11_POWER_STATE_INVALID,
	NT_STATUS_NDIS_PM_WOL_PATTERN_LIST_FULL:          ERROR_NDIS_PM_WOL_PATTERN_LIST_FULL,
	NT_STATUS_NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL:     ERROR_NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL,
	NT_STATUS_IPSEC_BAD_SPI:                          ERROR_IPSEC_BAD_SPI,
	NT_STATUS_IPSEC_SA_LIFETIME_EXPIRED:              ERROR_IPSEC_SA_LIFETIME_EXPIRED,
	NT_STATUS_IPSEC_WRONG_SA:                         ERROR_IPSEC_WRONG_SA,
	NT_STATUS_IPSEC_REPLAY_CHECK_FAILED:              ERROR_IPSEC_REPLAY_CHECK_FAILED,
	NT_STATUS_IPSEC_INVALID_PACKET:                   ERROR_IPSEC_INVALID_PACKET,
	NT_STATUS_IPSEC_INTEGRITY_CHECK_FAILED:           ERROR_IPSEC_INTEGRITY_CHECK_FAILED,
	NT_STATUS_IPSEC_CLEAR_TEXT_DROP:                  ERROR_IPSEC_CLEAR_TEXT_DROP,
	NT_STATUS_IPSEC_AUTH_FIREWALL_DROP:               ERROR_IPSEC_AUTH_FIREWALL_DROP,
	NT_STATUS_IPSEC_THROTTLE_DROP:                    ERROR_IPSEC_THROTTLE_DROP,
	NT_STATUS_IPSEC_DOSP_BLOCK:                       ERROR_IPSEC_DOSP_BLOCK,
	NT_STATUS_IPSEC_DOSP_RECEIVED_MULTICAST:          ERROR_IPSEC_DOSP_RECEIVED_MULTICAST,
	NT_STATUS_IPSEC_DOSP_INVALID_PACKET:              ERROR_IPSEC_DOSP_INVALID_PACKET,
	NT_STATUS_IPSEC_DOSP_STATE_LOOKUP_FAILED:         ERROR_IPSEC_DOSP_STATE_LOOKUP_FAILED,
	NT_STATUS_IPSEC_DOSP_MAX_ENTRIES:                 ERROR_IPSEC_DOSP_MAX_ENTRIES,
	NT_STATUS_IPSEC_DOSP_KEYMOD_NOT_ALLOWED:          ERROR_IPSEC_DOSP_KEYMOD_NOT_ALLOWED,
	NT_STATUS_IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES: ERROR_IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES,
	NT_STATUS_VOLMGR_MIRROR_NOT_SUPPORTED:            ERROR_VOLMGR_MIRROR_NOT_SUPPORTED,
	NT_STATUS_VOLMGR_RAID5_NOT_SUPPORTED:             ERROR_VOLMGR_RAID5_NOT_SUPPORTED,
	NT_STATUS_VIRTDISK_PROVIDER_NOT_FOUND:            ERROR_VIRTDISK_PROVIDER_NOT_FOUND,
	NT_STATUS_VIRTDISK_NOT_VIRTUAL_DISK:              ERROR_VIRTDISK_NOT_VIRTUAL_DISK,
	NT_STATUS_VHD_PARENT_VHD_ACCESS_DENIED:           ERROR_VHD_PARENT_VHD_ACCESS_DENIED,
	NT_STATUS_VHD_CHILD_PARENT_SIZE_MISMATCH:         ERROR_VHD_CHILD_PARENT_SIZE_MISMATCH,
	NT_STATUS_VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED:  ERROR_VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED,
	NT_STATUS_VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT: ERROR_VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT,
	NT_STATUS_SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP:  ERROR_SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP,
	NT_STATUS_SMB_BAD_CLUSTER_DIALECT:                ERROR_SMB_BAD_CLUSTER_DIALECT,
}

var NTStatusToStringName = map[NT_STATUS]string{
	NT_STATUS_SUCCESS: "SUCCESS",
	// NT_STATUS_WAIT_0: // NT_STATUS_WAIT_0",
	NT_STATUS_WAIT_1:    "WAIT_1",
	NT_STATUS_WAIT_2:    "WAIT_2",
	NT_STATUS_WAIT_3:    "WAIT_3",
	NT_STATUS_WAIT_63:   "WAIT_63",
	NT_STATUS_ABANDONED: "ABANDONED",
	// NT_STATUS_ABANDONED_WAIT_0: // NT_STATUS_ABANDONED_WAIT_0",
	NT_STATUS_ABANDONED_WAIT_63:                                           "ABANDONED_WAIT_63",
	NT_STATUS_USER_APC:                                                    "USER_APC",
	NT_STATUS_ALERTED:                                                     "ALERTED",
	NT_STATUS_TIMEOUT:                                                     "TIMEOUT",
	NT_STATUS_PENDING:                                                     "PENDING",
	NT_STATUS_REPARSE:                                                     "REPARSE",
	NT_STATUS_MORE_ENTRIES:                                                "MORE_ENTRIES",
	NT_STATUS_NOT_ALL_ASSIGNED:                                            "NOT_ALL_ASSIGNED",
	NT_STATUS_SOME_NOT_MAPPED:                                             "SOME_NOT_MAPPED",
	NT_STATUS_OPLOCK_BREAK_IN_PROGRESS:                                    "OPLOCK_BREAK_IN_PROGRESS",
	NT_STATUS_VOLUME_MOUNTED:                                              "VOLUME_MOUNTED",
	NT_STATUS_RXACT_COMMITTED:                                             "RXACT_COMMITTED",
	NT_STATUS_NOTIFY_CLEANUP:                                              "NOTIFY_CLEANUP",
	NT_STATUS_NOTIFY_ENUM_DIR:                                             "NOTIFY_ENUM_DIR",
	NT_STATUS_NO_QUOTAS_FOR_ACCOUNT:                                       "NO_QUOTAS_FOR_ACCOUNT",
	NT_STATUS_PRIMARY_TRANSPORT_CONNECT_FAILED:                            "PRIMARY_TRANSPORT_CONNECT_FAILED",
	NT_STATUS_PAGE_FAULT_TRANSITION:                                       "PAGE_FAULT_TRANSITION",
	NT_STATUS_PAGE_FAULT_DEMAND_ZERO:                                      "PAGE_FAULT_DEMAND_ZERO",
	NT_STATUS_PAGE_FAULT_COPY_ON_WRITE:                                    "PAGE_FAULT_COPY_ON_WRITE",
	NT_STATUS_PAGE_FAULT_GUARD_PAGE:                                       "PAGE_FAULT_GUARD_PAGE",
	NT_STATUS_PAGE_FAULT_PAGING_FILE:                                      "PAGE_FAULT_PAGING_FILE",
	NT_STATUS_CACHE_PAGE_LOCKED:                                           "CACHE_PAGE_LOCKED",
	NT_STATUS_CRASH_DUMP:                                                  "CRASH_DUMP",
	NT_STATUS_BUFFER_ALL_ZEROS:                                            "BUFFER_ALL_ZEROS",
	NT_STATUS_REPARSE_OBJECT:                                              "REPARSE_OBJECT",
	NT_STATUS_RESOURCE_REQUIREMENTS_CHANGED:                               "RESOURCE_REQUIREMENTS_CHANGED",
	NT_STATUS_TRANSLATION_COMPLETE:                                        "TRANSLATION_COMPLETE",
	NT_STATUS_DS_MEMBERSHIP_EVALUATED_LOCALLY:                             "DS_MEMBERSHIP_EVALUATED_LOCALLY",
	NT_STATUS_NOTHING_TO_TERMINATE:                                        "NOTHING_TO_TERMINATE",
	NT_STATUS_PROCESS_NOT_IN_JOB:                                          "PROCESS_NOT_IN_JOB",
	NT_STATUS_PROCESS_IN_JOB:                                              "PROCESS_IN_JOB",
	NT_STATUS_VOLSNAP_HIBERNATE_READY:                                     "VOLSNAP_HIBERNATE_READY",
	NT_STATUS_FSFILTER_OP_COMPLETED_SUCCESSFULLY:                          "FSFILTER_OP_COMPLETED_SUCCESSFULLY",
	NT_STATUS_INTERRUPT_VECTOR_ALREADY_CONNECTED:                          "INTERRUPT_VECTOR_ALREADY_CONNECTED",
	NT_STATUS_INTERRUPT_STILL_CONNECTED:                                   "INTERRUPT_STILL_CONNECTED",
	NT_STATUS_PROCESS_CLONED:                                              "PROCESS_CLONED",
	NT_STATUS_FILE_LOCKED_WITH_ONLY_READERS:                               "FILE_LOCKED_WITH_ONLY_READERS",
	NT_STATUS_FILE_LOCKED_WITH_WRITERS:                                    "FILE_LOCKED_WITH_WRITERS",
	NT_STATUS_RESOURCEMANAGER_READ_ONLY:                                   "RESOURCEMANAGER_READ_ONLY",
	NT_STATUS_WAIT_FOR_OPLOCK:                                             "WAIT_FOR_OPLOCK",
	NT_STATUS_DBG_EXCEPTION_HANDLED:                                       "DBG_EXCEPTION_HANDLED",
	NT_STATUS_DBG_CONTINUE:                                                "DBG_CONTINUE",
	NT_STATUS_FLT_IO_COMPLETE:                                             "FLT_IO_COMPLETE",
	NT_STATUS_FILE_NOT_AVAILABLE:                                          "FILE_NOT_AVAILABLE",
	NT_STATUS_SHARE_UNAVAILABLE:                                           "SHARE_UNAVAILABLE",
	NT_STATUS_CALLBACK_RETURNED_THREAD_AFFINITY:                           "CALLBACK_RETURNED_THREAD_AFFINITY",
	NT_STATUS_OBJECT_NAME_EXISTS:                                          "OBJECT_NAME_EXISTS",
	NT_STATUS_THREAD_WAS_SUSPENDED:                                        "THREAD_WAS_SUSPENDED",
	NT_STATUS_WORKING_SET_LIMIT_RANGE:                                     "WORKING_SET_LIMIT_RANGE",
	NT_STATUS_IMAGE_NOT_AT_BASE:                                           "IMAGE_NOT_AT_BASE",
	NT_STATUS_RXACT_STATE_CREATED:                                         "RXACT_STATE_CREATED",
	NT_STATUS_SEGMENT_NOTIFICATION:                                        "SEGMENT_NOTIFICATION",
	NT_STATUS_LOCAL_USER_SESSION_KEY:                                      "LOCAL_USER_SESSION_KEY",
	NT_STATUS_BAD_CURRENT_DIRECTORY:                                       "BAD_CURRENT_DIRECTORY",
	NT_STATUS_SERIAL_MORE_WRITES:                                          "SERIAL_MORE_WRITES",
	NT_STATUS_REGISTRY_RECOVERED:                                          "REGISTRY_RECOVERED",
	NT_STATUS_FT_READ_RECOVERY_FROM_BACKUP:                                "FT_READ_RECOVERY_FROM_BACKUP",
	NT_STATUS_FT_WRITE_RECOVERY:                                           "FT_WRITE_RECOVERY",
	NT_STATUS_SERIAL_COUNTER_TIMEOUT:                                      "SERIAL_COUNTER_TIMEOUT",
	NT_STATUS_NULL_LM_PASSWORD:                                            "NULL_LM_PASSWORD",
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH:                                 "IMAGE_MACHINE_TYPE_MISMATCH",
	NT_STATUS_RECEIVE_PARTIAL:                                             "RECEIVE_PARTIAL",
	NT_STATUS_RECEIVE_EXPEDITED:                                           "RECEIVE_EXPEDITED",
	NT_STATUS_RECEIVE_PARTIAL_EXPEDITED:                                   "RECEIVE_PARTIAL_EXPEDITED",
	NT_STATUS_EVENT_DONE:                                                  "EVENT_DONE",
	NT_STATUS_EVENT_PENDING:                                               "EVENT_PENDING",
	NT_STATUS_CHECKING_FILE_SYSTEM:                                        "CHECKING_FILE_SYSTEM",
	NT_STATUS_FATAL_APP_EXIT:                                              "FATAL_APP_EXIT",
	NT_STATUS_PREDEFINED_HANDLE:                                           "PREDEFINED_HANDLE",
	NT_STATUS_WAS_UNLOCKED:                                                "WAS_UNLOCKED",
	NT_STATUS_SERVICE_NOTIFICATION:                                        "SERVICE_NOTIFICATION",
	NT_STATUS_WAS_LOCKED:                                                  "WAS_LOCKED",
	NT_STATUS_LOG_HARD_ERROR:                                              "LOG_HARD_ERROR",
	NT_STATUS_ALREADY_WIN32:                                               "ALREADY_WIN32",
	NT_STATUS_WX86_UNSIMULATE:                                             "WX86_UNSIMULATE",
	NT_STATUS_WX86_CONTINUE:                                               "WX86_CONTINUE",
	NT_STATUS_WX86_SINGLE_STEP:                                            "WX86_SINGLE_STEP",
	NT_STATUS_WX86_BREAKPOINT:                                             "WX86_BREAKPOINT",
	NT_STATUS_WX86_EXCEPTION_CONTINUE:                                     "WX86_EXCEPTION_CONTINUE",
	NT_STATUS_WX86_EXCEPTION_LASTCHANCE:                                   "WX86_EXCEPTION_LASTCHANCE",
	NT_STATUS_WX86_EXCEPTION_CHAIN:                                        "WX86_EXCEPTION_CHAIN",
	NT_STATUS_IMAGE_MACHINE_TYPE_MISMATCH_EXE:                             "IMAGE_MACHINE_TYPE_MISMATCH_EXE",
	NT_STATUS_NO_YIELD_PERFORMED:                                          "NO_YIELD_PERFORMED",
	NT_STATUS_TIMER_RESUME_IGNORED:                                        "TIMER_RESUME_IGNORED",
	NT_STATUS_ARBITRATION_UNHANDLED:                                       "ARBITRATION_UNHANDLED",
	NT_STATUS_CARDBUS_NOT_SUPPORTED:                                       "CARDBUS_NOT_SUPPORTED",
	NT_STATUS_WX86_CREATEWX86TIB:                                          "WX86_CREATEWX86TIB",
	NT_STATUS_MP_PROCESSOR_MISMATCH:                                       "MP_PROCESSOR_MISMATCH",
	NT_STATUS_HIBERNATED:                                                  "HIBERNATED",
	NT_STATUS_RESUME_HIBERNATION:                                          "RESUME_HIBERNATION",
	NT_STATUS_FIRMWARE_UPDATED:                                            "FIRMWARE_UPDATED",
	NT_STATUS_DRIVERS_LEAKING_LOCKED_PAGES:                                "DRIVERS_LEAKING_LOCKED_PAGES",
	NT_STATUS_MESSAGE_RETRIEVED:                                           "MESSAGE_RETRIEVED",
	NT_STATUS_SYSTEM_POWERSTATE_TRANSITION:                                "SYSTEM_POWERSTATE_TRANSITION",
	NT_STATUS_ALPC_CHECK_COMPLETION_LIST:                                  "ALPC_CHECK_COMPLETION_LIST",
	NT_STATUS_SYSTEM_POWERSTATE_COMPLEX_TRANSITION:                        "SYSTEM_POWERSTATE_COMPLEX_TRANSITION",
	NT_STATUS_ACCESS_AUDIT_BY_POLICY:                                      "ACCESS_AUDIT_BY_POLICY",
	NT_STATUS_ABANDON_HIBERFILE:                                           "ABANDON_HIBERFILE",
	NT_STATUS_BIZRULES_NOT_ENABLED:                                        "BIZRULES_NOT_ENABLED",
	NT_STATUS_WAKE_SYSTEM:                                                 "WAKE_SYSTEM",
	NT_STATUS_DS_SHUTTING_DOWN:                                            "DS_SHUTTING_DOWN",
	NT_STATUS_DBG_REPLY_LATER:                                             "DBG_REPLY_LATER",
	NT_STATUS_DBG_UNABLE_TO_PROVIDE_HANDLE:                                "DBG_UNABLE_TO_PROVIDE_HANDLE",
	NT_STATUS_DBG_TERMINATE_THREAD:                                        "DBG_TERMINATE_THREAD",
	NT_STATUS_DBG_TERMINATE_PROCESS:                                       "DBG_TERMINATE_PROCESS",
	NT_STATUS_DBG_CONTROL_C:                                               "DBG_CONTROL_C",
	NT_STATUS_DBG_PRINTEXCEPTION_C:                                        "DBG_PRINTEXCEPTION_C",
	NT_STATUS_DBG_RIPEXCEPTION:                                            "DBG_RIPEXCEPTION",
	NT_STATUS_DBG_CONTROL_BREAK:                                           "DBG_CONTROL_BREAK",
	NT_STATUS_DBG_COMMAND_EXCEPTION:                                       "DBG_COMMAND_EXCEPTION",
	NT_STATUS_RPC_NT_UUID_LOCAL_ONLY:                                      "RPC_NT_UUID_LOCAL_ONLY",
	NT_STATUS_RPC_NT_SEND_INCOMPLETE:                                      "RPC_NT_SEND_INCOMPLETE",
	NT_STATUS_CTX_CDM_CONNECT:                                             "CTX_CDM_CONNECT",
	NT_STATUS_CTX_CDM_DISCONNECT:                                          "CTX_CDM_DISCONNECT",
	NT_STATUS_SXS_RELEASE_ACTIVATION_CONTEXT:                              "SXS_RELEASE_ACTIVATION_CONTEXT",
	NT_STATUS_RECOVERY_NOT_NEEDED:                                         "RECOVERY_NOT_NEEDED",
	NT_STATUS_RM_ALREADY_STARTED:                                          "RM_ALREADY_STARTED",
	NT_STATUS_LOG_NO_RESTART:                                              "LOG_NO_RESTART",
	NT_STATUS_VIDEO_DRIVER_DEBUG_REPORT_REQUEST:                           "VIDEO_DRIVER_DEBUG_REPORT_REQUEST",
	NT_STATUS_GRAPHICS_PARTIAL_DATA_POPULATED:                             "GRAPHICS_PARTIAL_DATA_POPULATED",
	NT_STATUS_GRAPHICS_DRIVER_MISMATCH:                                    "GRAPHICS_DRIVER_MISMATCH",
	NT_STATUS_GRAPHICS_MODE_NOT_PINNED:                                    "GRAPHICS_MODE_NOT_PINNED",
	NT_STATUS_GRAPHICS_NO_PREFERRED_MODE:                                  "GRAPHICS_NO_PREFERRED_MODE",
	NT_STATUS_GRAPHICS_DATASET_IS_EMPTY:                                   "GRAPHICS_DATASET_IS_EMPTY",
	NT_STATUS_GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET:                        "GRAPHICS_NO_MORE_ELEMENTS_IN_DATASET",
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED:    "GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_PINNED",
	NT_STATUS_GRAPHICS_UNKNOWN_CHILD_STATUS:                               "GRAPHICS_UNKNOWN_CHILD_STATUS",
	NT_STATUS_GRAPHICS_LEADLINK_START_DEFERRED:                            "GRAPHICS_LEADLINK_START_DEFERRED",
	NT_STATUS_GRAPHICS_POLLING_TOO_FREQUENTLY:                             "GRAPHICS_POLLING_TOO_FREQUENTLY",
	NT_STATUS_GRAPHICS_START_DEFERRED:                                     "GRAPHICS_START_DEFERRED",
	NT_STATUS_NDIS_INDICATION_REQUIRED:                                    "NDIS_INDICATION_REQUIRED",
	NT_STATUS_GUARD_PAGE_VIOLATION:                                        "GUARD_PAGE_VIOLATION",
	NT_STATUS_DATATYPE_MISALIGNMENT:                                       "DATATYPE_MISALIGNMENT",
	NT_STATUS_BREAKPOINT:                                                  "BREAKPOINT",
	NT_STATUS_SINGLE_STEP:                                                 "SINGLE_STEP",
	NT_STATUS_BUFFER_OVERFLOW:                                             "BUFFER_OVERFLOW",
	NT_STATUS_NO_MORE_FILES:                                               "NO_MORE_FILES",
	NT_STATUS_WAKE_SYSTEM_DEBUGGER:                                        "WAKE_SYSTEM_DEBUGGER",
	NT_STATUS_HANDLES_CLOSED:                                              "HANDLES_CLOSED",
	NT_STATUS_NO_INHERITANCE:                                              "NO_INHERITANCE",
	NT_STATUS_GUID_SUBSTITUTION_MADE:                                      "GUID_SUBSTITUTION_MADE",
	NT_STATUS_PARTIAL_COPY:                                                "PARTIAL_COPY",
	NT_STATUS_DEVICE_PAPER_EMPTY:                                          "DEVICE_PAPER_EMPTY",
	NT_STATUS_DEVICE_POWERED_OFF:                                          "DEVICE_POWERED_OFF",
	NT_STATUS_DEVICE_OFF_LINE:                                             "DEVICE_OFF_LINE",
	NT_STATUS_DEVICE_BUSY:                                                 "DEVICE_BUSY",
	NT_STATUS_NO_MORE_EAS:                                                 "NO_MORE_EAS",
	NT_STATUS_INVALID_EA_NAME:                                             "INVALID_EA_NAME",
	NT_STATUS_EA_LIST_INCONSISTENT:                                        "EA_LIST_INCONSISTENT",
	NT_STATUS_INVALID_EA_FLAG:                                             "INVALID_EA_FLAG",
	NT_STATUS_VERIFY_REQUIRED:                                             "VERIFY_REQUIRED",
	NT_STATUS_EXTRANEOUS_INFORMATION:                                      "EXTRANEOUS_INFORMATION",
	NT_STATUS_RXACT_COMMIT_NECESSARY:                                      "RXACT_COMMIT_NECESSARY",
	NT_STATUS_NO_MORE_ENTRIES:                                             "NO_MORE_ENTRIES",
	NT_STATUS_FILEMARK_DETECTED:                                           "FILEMARK_DETECTED",
	NT_STATUS_MEDIA_CHANGED:                                               "MEDIA_CHANGED",
	NT_STATUS_BUS_RESET:                                                   "BUS_RESET",
	NT_STATUS_END_OF_MEDIA:                                                "END_OF_MEDIA",
	NT_STATUS_BEGINNING_OF_MEDIA:                                          "BEGINNING_OF_MEDIA",
	NT_STATUS_MEDIA_CHECK:                                                 "MEDIA_CHECK",
	NT_STATUS_SETMARK_DETECTED:                                            "SETMARK_DETECTED",
	NT_STATUS_NO_DATA_DETECTED:                                            "NO_DATA_DETECTED",
	NT_STATUS_REDIRECTOR_HAS_OPEN_HANDLES:                                 "REDIRECTOR_HAS_OPEN_HANDLES",
	NT_STATUS_SERVER_HAS_OPEN_HANDLES:                                     "SERVER_HAS_OPEN_HANDLES",
	NT_STATUS_ALREADY_DISCONNECTED:                                        "ALREADY_DISCONNECTED",
	NT_STATUS_LONGJUMP:                                                    "LONGJUMP",
	NT_STATUS_CLEANER_CARTRIDGE_INSTALLED:                                 "CLEANER_CARTRIDGE_INSTALLED",
	NT_STATUS_PLUGPLAY_QUERY_VETOED:                                       "PLUGPLAY_QUERY_VETOED",
	NT_STATUS_UNWIND_CONSOLIDATE:                                          "UNWIND_CONSOLIDATE",
	NT_STATUS_REGISTRY_HIVE_RECOVERED:                                     "REGISTRY_HIVE_RECOVERED",
	NT_STATUS_DLL_MIGHT_BE_INSECURE:                                       "DLL_MIGHT_BE_INSECURE",
	NT_STATUS_DLL_MIGHT_BE_INCOMPATIBLE:                                   "DLL_MIGHT_BE_INCOMPATIBLE",
	NT_STATUS_STOPPED_ON_SYMLINK:                                          "STOPPED_ON_SYMLINK",
	NT_STATUS_DEVICE_REQUIRES_CLEANING:                                    "DEVICE_REQUIRES_CLEANING",
	NT_STATUS_DEVICE_DOOR_OPEN:                                            "DEVICE_DOOR_OPEN",
	NT_STATUS_DATA_LOST_REPAIR:                                            "DATA_LOST_REPAIR",
	NT_STATUS_DBG_EXCEPTION_NOT_HANDLED:                                   "DBG_EXCEPTION_NOT_HANDLED",
	NT_STATUS_CLUSTER_NODE_ALREADY_UP:                                     "CLUSTER_NODE_ALREADY_UP",
	NT_STATUS_CLUSTER_NODE_ALREADY_DOWN:                                   "CLUSTER_NODE_ALREADY_DOWN",
	NT_STATUS_CLUSTER_NETWORK_ALREADY_ONLINE:                              "CLUSTER_NETWORK_ALREADY_ONLINE",
	NT_STATUS_CLUSTER_NETWORK_ALREADY_OFFLINE:                             "CLUSTER_NETWORK_ALREADY_OFFLINE",
	NT_STATUS_CLUSTER_NODE_ALREADY_MEMBER:                                 "CLUSTER_NODE_ALREADY_MEMBER",
	NT_STATUS_COULD_NOT_RESIZE_LOG:                                        "COULD_NOT_RESIZE_LOG",
	NT_STATUS_NO_TXF_METADATA:                                             "NO_TXF_METADATA",
	NT_STATUS_CANT_RECOVER_WITH_HANDLE_OPEN:                               "CANT_RECOVER_WITH_HANDLE_OPEN",
	NT_STATUS_TXF_METADATA_ALREADY_PRESENT:                                "TXF_METADATA_ALREADY_PRESENT",
	NT_STATUS_TRANSACTION_SCOPE_CALLBACKS_NOT_SET:                         "TRANSACTION_SCOPE_CALLBACKS_NOT_SET",
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED:                  "VIDEO_HUNG_DISPLAY_DRIVER_THREAD_RECOVERED",
	NT_STATUS_FLT_BUFFER_TOO_SMALL:                                        "FLT_BUFFER_TOO_SMALL",
	NT_STATUS_FVE_PARTIAL_METADATA:                                        "FVE_PARTIAL_METADATA",
	NT_STATUS_FVE_TRANSIENT_STATE:                                         "FVE_TRANSIENT_STATE",
	NT_STATUS_UNSUCCESSFUL:                                                "UNSUCCESSFUL",
	NT_STATUS_NOT_IMPLEMENTED:                                             "NOT_IMPLEMENTED",
	NT_STATUS_INVALID_INFO_CLASS:                                          "INVALID_INFO_CLASS",
	NT_STATUS_INFO_LENGTH_MISMATCH:                                        "INFO_LENGTH_MISMATCH",
	NT_STATUS_ACCESS_VIOLATION:                                            "ACCESS_VIOLATION",
	NT_STATUS_IN_PAGE_ERROR:                                               "IN_PAGE_ERROR",
	NT_STATUS_PAGEFILE_QUOTA:                                              "PAGEFILE_QUOTA",
	NT_STATUS_INVALID_HANDLE:                                              "INVALID_HANDLE",
	NT_STATUS_BAD_INITIAL_STACK:                                           "BAD_INITIAL_STACK",
	NT_STATUS_BAD_INITIAL_PC:                                              "BAD_INITIAL_PC",
	NT_STATUS_INVALID_CID:                                                 "INVALID_CID",
	NT_STATUS_TIMER_NOT_CANCELED:                                          "TIMER_NOT_CANCELED",
	NT_STATUS_INVALID_PARAMETER:                                           "INVALID_PARAMETER",
	NT_STATUS_NO_SUCH_DEVICE:                                              "NO_SUCH_DEVICE",
	NT_STATUS_NO_SUCH_FILE:                                                "NO_SUCH_FILE",
	NT_STATUS_INVALID_DEVICE_REQUEST:                                      "INVALID_DEVICE_REQUEST",
	NT_STATUS_END_OF_FILE:                                                 "END_OF_FILE",
	NT_STATUS_WRONG_VOLUME:                                                "WRONG_VOLUME",
	NT_STATUS_NO_MEDIA_IN_DEVICE:                                          "NO_MEDIA_IN_DEVICE",
	NT_STATUS_UNRECOGNIZED_MEDIA:                                          "UNRECOGNIZED_MEDIA",
	NT_STATUS_NONEXISTENT_SECTOR:                                          "NONEXISTENT_SECTOR",
	NT_STATUS_MORE_PROCESSING_REQUIRED:                                    "MORE_PROCESSING_REQUIRED",
	NT_STATUS_NO_MEMORY:                                                   "NO_MEMORY",
	NT_STATUS_CONFLICTING_ADDRESSES:                                       "CONFLICTING_ADDRESSES",
	NT_STATUS_NOT_MAPPED_VIEW:                                             "NOT_MAPPED_VIEW",
	NT_STATUS_UNABLE_TO_FREE_VM:                                           "UNABLE_TO_FREE_VM",
	NT_STATUS_UNABLE_TO_DELETE_SECTION:                                    "UNABLE_TO_DELETE_SECTION",
	NT_STATUS_INVALID_SYSTEM_SERVICE:                                      "INVALID_SYSTEM_SERVICE",
	NT_STATUS_ILLEGAL_INSTRUCTION:                                         "ILLEGAL_INSTRUCTION",
	NT_STATUS_INVALID_LOCK_SEQUENCE:                                       "INVALID_LOCK_SEQUENCE",
	NT_STATUS_INVALID_VIEW_SIZE:                                           "INVALID_VIEW_SIZE",
	NT_STATUS_INVALID_FILE_FOR_SECTION:                                    "INVALID_FILE_FOR_SECTION",
	NT_STATUS_ALREADY_COMMITTED:                                           "ALREADY_COMMITTED",
	NT_STATUS_ACCESS_DENIED:                                               "ACCESS_DENIED",
	NT_STATUS_BUFFER_TOO_SMALL:                                            "BUFFER_TOO_SMALL",
	NT_STATUS_OBJECT_TYPE_MISMATCH:                                        "OBJECT_TYPE_MISMATCH",
	NT_STATUS_NONCONTINUABLE_EXCEPTION:                                    "NONCONTINUABLE_EXCEPTION",
	NT_STATUS_INVALID_DISPOSITION:                                         "INVALID_DISPOSITION",
	NT_STATUS_UNWIND:                                                      "UNWIND",
	NT_STATUS_BAD_STACK:                                                   "BAD_STACK",
	NT_STATUS_INVALID_UNWIND_TARGET:                                       "INVALID_UNWIND_TARGET",
	NT_STATUS_NOT_LOCKED:                                                  "NOT_LOCKED",
	NT_STATUS_PARITY_ERROR:                                                "PARITY_ERROR",
	NT_STATUS_UNABLE_TO_DECOMMIT_VM:                                       "UNABLE_TO_DECOMMIT_VM",
	NT_STATUS_NOT_COMMITTED:                                               "NOT_COMMITTED",
	NT_STATUS_INVALID_PORT_ATTRIBUTES:                                     "INVALID_PORT_ATTRIBUTES",
	NT_STATUS_PORT_MESSAGE_TOO_LONG:                                       "PORT_MESSAGE_TOO_LONG",
	NT_STATUS_INVALID_PARAMETER_MIX:                                       "INVALID_PARAMETER_MIX",
	NT_STATUS_INVALID_QUOTA_LOWER:                                         "INVALID_QUOTA_LOWER",
	NT_STATUS_DISK_CORRUPT_ERROR:                                          "DISK_CORRUPT_ERROR",
	NT_STATUS_OBJECT_NAME_INVALID:                                         "OBJECT_NAME_INVALID",
	NT_STATUS_OBJECT_NAME_NOT_FOUND:                                       "OBJECT_NAME_NOT_FOUND",
	NT_STATUS_OBJECT_NAME_COLLISION:                                       "OBJECT_NAME_COLLISION",
	NT_STATUS_PORT_DISCONNECTED:                                           "PORT_DISCONNECTED",
	NT_STATUS_DEVICE_ALREADY_ATTACHED:                                     "DEVICE_ALREADY_ATTACHED",
	NT_STATUS_OBJECT_PATH_INVALID:                                         "OBJECT_PATH_INVALID",
	NT_STATUS_OBJECT_PATH_NOT_FOUND:                                       "OBJECT_PATH_NOT_FOUND",
	NT_STATUS_OBJECT_PATH_SYNTAX_BAD:                                      "OBJECT_PATH_SYNTAX_BAD",
	NT_STATUS_DATA_OVERRUN:                                                "DATA_OVERRUN",
	NT_STATUS_DATA_LATE_ERROR:                                             "DATA_LATE_ERROR",
	NT_STATUS_DATA_ERROR:                                                  "DATA_ERROR",
	NT_STATUS_CRC_ERROR:                                                   "CRC_ERROR",
	NT_STATUS_SECTION_TOO_BIG:                                             "SECTION_TOO_BIG",
	NT_STATUS_PORT_CONNECTION_REFUSED:                                     "PORT_CONNECTION_REFUSED",
	NT_STATUS_INVALID_PORT_HANDLE:                                         "INVALID_PORT_HANDLE",
	NT_STATUS_SHARING_VIOLATION:                                           "SHARING_VIOLATION",
	NT_STATUS_QUOTA_EXCEEDED:                                              "QUOTA_EXCEEDED",
	NT_STATUS_INVALID_PAGE_PROTECTION:                                     "INVALID_PAGE_PROTECTION",
	NT_STATUS_MUTANT_NOT_OWNED:                                            "MUTANT_NOT_OWNED",
	NT_STATUS_SEMAPHORE_LIMIT_EXCEEDED:                                    "SEMAPHORE_LIMIT_EXCEEDED",
	NT_STATUS_PORT_ALREADY_SET:                                            "PORT_ALREADY_SET",
	NT_STATUS_SECTION_NOT_IMAGE:                                           "SECTION_NOT_IMAGE",
	NT_STATUS_SUSPEND_COUNT_EXCEEDED:                                      "SUSPEND_COUNT_EXCEEDED",
	NT_STATUS_THREAD_IS_TERMINATING:                                       "THREAD_IS_TERMINATING",
	NT_STATUS_BAD_WORKING_SET_LIMIT:                                       "BAD_WORKING_SET_LIMIT",
	NT_STATUS_INCOMPATIBLE_FILE_MAP:                                       "INCOMPATIBLE_FILE_MAP",
	NT_STATUS_SECTION_PROTECTION:                                          "SECTION_PROTECTION",
	NT_STATUS_EAS_NOT_SUPPORTED:                                           "EAS_NOT_SUPPORTED",
	NT_STATUS_EA_TOO_LARGE:                                                "EA_TOO_LARGE",
	NT_STATUS_NONEXISTENT_EA_ENTRY:                                        "NONEXISTENT_EA_ENTRY",
	NT_STATUS_NO_EAS_ON_FILE:                                              "NO_EAS_ON_FILE",
	NT_STATUS_EA_CORRUPT_ERROR:                                            "EA_CORRUPT_ERROR",
	NT_STATUS_FILE_LOCK_CONFLICT:                                          "FILE_LOCK_CONFLICT",
	NT_STATUS_LOCK_NOT_GRANTED:                                            "LOCK_NOT_GRANTED",
	NT_STATUS_DELETE_PENDING:                                              "DELETE_PENDING",
	NT_STATUS_CTL_FILE_NOT_SUPPORTED:                                      "CTL_FILE_NOT_SUPPORTED",
	NT_STATUS_UNKNOWN_REVISION:                                            "UNKNOWN_REVISION",
	NT_STATUS_REVISION_MISMATCH:                                           "REVISION_MISMATCH",
	NT_STATUS_INVALID_OWNER:                                               "INVALID_OWNER",
	NT_STATUS_INVALID_PRIMARY_GROUP:                                       "INVALID_PRIMARY_GROUP",
	NT_STATUS_NO_IMPERSONATION_TOKEN:                                      "NO_IMPERSONATION_TOKEN",
	NT_STATUS_CANT_DISABLE_MANDATORY:                                      "CANT_DISABLE_MANDATORY",
	NT_STATUS_NO_LOGON_SERVERS:                                            "NO_LOGON_SERVERS",
	NT_STATUS_NO_SUCH_LOGON_SESSION:                                       "NO_SUCH_LOGON_SESSION",
	NT_STATUS_NO_SUCH_PRIVILEGE:                                           "NO_SUCH_PRIVILEGE",
	NT_STATUS_PRIVILEGE_NOT_HELD:                                          "PRIVILEGE_NOT_HELD",
	NT_STATUS_INVALID_ACCOUNT_NAME:                                        "INVALID_ACCOUNT_NAME",
	NT_STATUS_USER_EXISTS:                                                 "USER_EXISTS",
	NT_STATUS_NO_SUCH_USER:                                                "NO_SUCH_USER",
	NT_STATUS_GROUP_EXISTS:                                                "GROUP_EXISTS",
	NT_STATUS_NO_SUCH_GROUP:                                               "NO_SUCH_GROUP",
	NT_STATUS_MEMBER_IN_GROUP:                                             "MEMBER_IN_GROUP",
	NT_STATUS_MEMBER_NOT_IN_GROUP:                                         "MEMBER_NOT_IN_GROUP",
	NT_STATUS_LAST_ADMIN:                                                  "LAST_ADMIN",
	NT_STATUS_WRONG_PASSWORD:                                              "WRONG_PASSWORD",
	NT_STATUS_ILL_FORMED_PASSWORD:                                         "ILL_FORMED_PASSWORD",
	NT_STATUS_PASSWORD_RESTRICTION:                                        "PASSWORD_RESTRICTION",
	NT_STATUS_LOGON_FAILURE:                                               "LOGON_FAILURE",
	NT_STATUS_ACCOUNT_RESTRICTION:                                         "ACCOUNT_RESTRICTION",
	NT_STATUS_INVALID_LOGON_HOURS:                                         "INVALID_LOGON_HOURS",
	NT_STATUS_INVALID_WORKSTATION:                                         "INVALID_WORKSTATION",
	NT_STATUS_PASSWORD_EXPIRED:                                            "PASSWORD_EXPIRED",
	NT_STATUS_ACCOUNT_DISABLED:                                            "ACCOUNT_DISABLED",
	NT_STATUS_NONE_MAPPED:                                                 "NONE_MAPPED",
	NT_STATUS_TOO_MANY_LUIDS_REQUESTED:                                    "TOO_MANY_LUIDS_REQUESTED",
	NT_STATUS_LUIDS_EXHAUSTED:                                             "LUIDS_EXHAUSTED",
	NT_STATUS_INVALID_SUB_AUTHORITY:                                       "INVALID_SUB_AUTHORITY",
	NT_STATUS_INVALID_ACL:                                                 "INVALID_ACL",
	NT_STATUS_INVALID_SID:                                                 "INVALID_SID",
	NT_STATUS_INVALID_SECURITY_DESCR:                                      "INVALID_SECURITY_DESCR",
	NT_STATUS_PROCEDURE_NOT_FOUND:                                         "PROCEDURE_NOT_FOUND",
	NT_STATUS_INVALID_IMAGE_FORMAT:                                        "INVALID_IMAGE_FORMAT",
	NT_STATUS_NO_TOKEN:                                                    "NO_TOKEN",
	NT_STATUS_BAD_INHERITANCE_ACL:                                         "BAD_INHERITANCE_ACL",
	NT_STATUS_RANGE_NOT_LOCKED:                                            "RANGE_NOT_LOCKED",
	NT_STATUS_DISK_FULL:                                                   "DISK_FULL",
	NT_STATUS_SERVER_DISABLED:                                             "SERVER_DISABLED",
	NT_STATUS_SERVER_NOT_DISABLED:                                         "SERVER_NOT_DISABLED",
	NT_STATUS_TOO_MANY_GUIDS_REQUESTED:                                    "TOO_MANY_GUIDS_REQUESTED",
	NT_STATUS_GUIDS_EXHAUSTED:                                             "GUIDS_EXHAUSTED",
	NT_STATUS_INVALID_ID_AUTHORITY:                                        "INVALID_ID_AUTHORITY",
	NT_STATUS_AGENTS_EXHAUSTED:                                            "AGENTS_EXHAUSTED",
	NT_STATUS_INVALID_VOLUME_LABEL:                                        "INVALID_VOLUME_LABEL",
	NT_STATUS_SECTION_NOT_EXTENDED:                                        "SECTION_NOT_EXTENDED",
	NT_STATUS_NOT_MAPPED_DATA:                                             "NOT_MAPPED_DATA",
	NT_STATUS_RESOURCE_DATA_NOT_FOUND:                                     "RESOURCE_DATA_NOT_FOUND",
	NT_STATUS_RESOURCE_TYPE_NOT_FOUND:                                     "RESOURCE_TYPE_NOT_FOUND",
	NT_STATUS_RESOURCE_NAME_NOT_FOUND:                                     "RESOURCE_NAME_NOT_FOUND",
	NT_STATUS_ARRAY_BOUNDS_EXCEEDED:                                       "ARRAY_BOUNDS_EXCEEDED",
	NT_STATUS_FLOAT_DENORMAL_OPERAND:                                      "FLOAT_DENORMAL_OPERAND",
	NT_STATUS_FLOAT_DIVIDE_BY_ZERO:                                        "FLOAT_DIVIDE_BY_ZERO",
	NT_STATUS_FLOAT_INEXACT_RESULT:                                        "FLOAT_INEXACT_RESULT",
	NT_STATUS_FLOAT_INVALID_OPERATION:                                     "FLOAT_INVALID_OPERATION",
	NT_STATUS_FLOAT_OVERFLOW:                                              "FLOAT_OVERFLOW",
	NT_STATUS_FLOAT_STACK_CHECK:                                           "FLOAT_STACK_CHECK",
	NT_STATUS_FLOAT_UNDERFLOW:                                             "FLOAT_UNDERFLOW",
	NT_STATUS_INTEGER_DIVIDE_BY_ZERO:                                      "INTEGER_DIVIDE_BY_ZERO",
	NT_STATUS_INTEGER_OVERFLOW:                                            "INTEGER_OVERFLOW",
	NT_STATUS_PRIVILEGED_INSTRUCTION:                                      "PRIVILEGED_INSTRUCTION",
	NT_STATUS_TOO_MANY_PAGING_FILES:                                       "TOO_MANY_PAGING_FILES",
	NT_STATUS_FILE_INVALID:                                                "FILE_INVALID",
	NT_STATUS_ALLOTTED_SPACE_EXCEEDED:                                     "ALLOTTED_SPACE_EXCEEDED",
	NT_STATUS_INSUFFICIENT_RESOURCES:                                      "INSUFFICIENT_RESOURCES",
	NT_STATUS_DFS_EXIT_PATH_FOUND:                                         "DFS_EXIT_PATH_FOUND",
	NT_STATUS_DEVICE_DATA_ERROR:                                           "DEVICE_DATA_ERROR",
	NT_STATUS_DEVICE_NOT_CONNECTED:                                        "DEVICE_NOT_CONNECTED",
	NT_STATUS_FREE_VM_NOT_AT_BASE:                                         "FREE_VM_NOT_AT_BASE",
	NT_STATUS_MEMORY_NOT_ALLOCATED:                                        "MEMORY_NOT_ALLOCATED",
	NT_STATUS_WORKING_SET_QUOTA:                                           "WORKING_SET_QUOTA",
	NT_STATUS_MEDIA_WRITE_PROTECTED:                                       "MEDIA_WRITE_PROTECTED",
	NT_STATUS_DEVICE_NOT_READY:                                            "DEVICE_NOT_READY",
	NT_STATUS_INVALID_GROUP_ATTRIBUTES:                                    "INVALID_GROUP_ATTRIBUTES",
	NT_STATUS_BAD_IMPERSONATION_LEVEL:                                     "BAD_IMPERSONATION_LEVEL",
	NT_STATUS_CANT_OPEN_ANONYMOUS:                                         "CANT_OPEN_ANONYMOUS",
	NT_STATUS_BAD_VALIDATION_CLASS:                                        "BAD_VALIDATION_CLASS",
	NT_STATUS_BAD_TOKEN_TYPE:                                              "BAD_TOKEN_TYPE",
	NT_STATUS_BAD_MASTER_BOOT_RECORD:                                      "BAD_MASTER_BOOT_RECORD",
	NT_STATUS_INSTRUCTION_MISALIGNMENT:                                    "INSTRUCTION_MISALIGNMENT",
	NT_STATUS_INSTANCE_NOT_AVAILABLE:                                      "INSTANCE_NOT_AVAILABLE",
	NT_STATUS_PIPE_NOT_AVAILABLE:                                          "PIPE_NOT_AVAILABLE",
	NT_STATUS_INVALID_PIPE_STATE:                                          "INVALID_PIPE_STATE",
	NT_STATUS_PIPE_BUSY:                                                   "PIPE_BUSY",
	NT_STATUS_ILLEGAL_FUNCTION:                                            "ILLEGAL_FUNCTION",
	NT_STATUS_PIPE_DISCONNECTED:                                           "PIPE_DISCONNECTED",
	NT_STATUS_PIPE_CLOSING:                                                "PIPE_CLOSING",
	NT_STATUS_PIPE_CONNECTED:                                              "PIPE_CONNECTED",
	NT_STATUS_PIPE_LISTENING:                                              "PIPE_LISTENING",
	NT_STATUS_INVALID_READ_MODE:                                           "INVALID_READ_MODE",
	NT_STATUS_IO_TIMEOUT:                                                  "IO_TIMEOUT",
	NT_STATUS_FILE_FORCED_CLOSED:                                          "FILE_FORCED_CLOSED",
	NT_STATUS_PROFILING_NOT_STARTED:                                       "PROFILING_NOT_STARTED",
	NT_STATUS_PROFILING_NOT_STOPPED:                                       "PROFILING_NOT_STOPPED",
	NT_STATUS_COULD_NOT_INTERPRET:                                         "COULD_NOT_INTERPRET",
	NT_STATUS_FILE_IS_A_DIRECTORY:                                         "FILE_IS_A_DIRECTORY",
	NT_STATUS_NOT_SUPPORTED:                                               "NOT_SUPPORTED",
	NT_STATUS_REMOTE_NOT_LISTENING:                                        "REMOTE_NOT_LISTENING",
	NT_STATUS_DUPLICATE_NAME:                                              "DUPLICATE_NAME",
	NT_STATUS_BAD_NETWORK_PATH:                                            "BAD_NETWORK_PATH",
	NT_STATUS_NETWORK_BUSY:                                                "NETWORK_BUSY",
	NT_STATUS_DEVICE_DOES_NOT_EXIST:                                       "DEVICE_DOES_NOT_EXIST",
	NT_STATUS_TOO_MANY_COMMANDS:                                           "TOO_MANY_COMMANDS",
	NT_STATUS_ADAPTER_HARDWARE_ERROR:                                      "ADAPTER_HARDWARE_ERROR",
	NT_STATUS_INVALID_NETWORK_RESPONSE:                                    "INVALID_NETWORK_RESPONSE",
	NT_STATUS_UNEXPECTED_NETWORK_ERROR:                                    "UNEXPECTED_NETWORK_ERROR",
	NT_STATUS_BAD_REMOTE_ADAPTER:                                          "BAD_REMOTE_ADAPTER",
	NT_STATUS_PRINT_QUEUE_FULL:                                            "PRINT_QUEUE_FULL",
	NT_STATUS_NO_SPOOL_SPACE:                                              "NO_SPOOL_SPACE",
	NT_STATUS_PRINT_CANCELLED:                                             "PRINT_CANCELLED",
	NT_STATUS_NETWORK_NAME_DELETED:                                        "NETWORK_NAME_DELETED",
	NT_STATUS_NETWORK_ACCESS_DENIED:                                       "NETWORK_ACCESS_DENIED",
	NT_STATUS_BAD_DEVICE_TYPE:                                             "BAD_DEVICE_TYPE",
	NT_STATUS_BAD_NETWORK_NAME:                                            "BAD_NETWORK_NAME",
	NT_STATUS_TOO_MANY_NAMES:                                              "TOO_MANY_NAMES",
	NT_STATUS_TOO_MANY_SESSIONS:                                           "TOO_MANY_SESSIONS",
	NT_STATUS_SHARING_PAUSED:                                              "SHARING_PAUSED",
	NT_STATUS_REQUEST_NOT_ACCEPTED:                                        "REQUEST_NOT_ACCEPTED",
	NT_STATUS_REDIRECTOR_PAUSED:                                           "REDIRECTOR_PAUSED",
	NT_STATUS_NET_WRITE_FAULT:                                             "NET_WRITE_FAULT",
	NT_STATUS_PROFILING_AT_LIMIT:                                          "PROFILING_AT_LIMIT",
	NT_STATUS_NOT_SAME_DEVICE:                                             "NOT_SAME_DEVICE",
	NT_STATUS_FILE_RENAMED:                                                "FILE_RENAMED",
	NT_STATUS_VIRTUAL_CIRCUIT_CLOSED:                                      "VIRTUAL_CIRCUIT_CLOSED",
	NT_STATUS_NO_SECURITY_ON_OBJECT:                                       "NO_SECURITY_ON_OBJECT",
	NT_STATUS_CANT_WAIT:                                                   "CANT_WAIT",
	NT_STATUS_PIPE_EMPTY:                                                  "PIPE_EMPTY",
	NT_STATUS_CANT_ACCESS_DOMAIN_INFO:                                     "CANT_ACCESS_DOMAIN_INFO",
	NT_STATUS_CANT_TERMINATE_SELF:                                         "CANT_TERMINATE_SELF",
	NT_STATUS_INVALID_SERVER_STATE:                                        "INVALID_SERVER_STATE",
	NT_STATUS_INVALID_DOMAIN_STATE:                                        "INVALID_DOMAIN_STATE",
	NT_STATUS_INVALID_DOMAIN_ROLE:                                         "INVALID_DOMAIN_ROLE",
	NT_STATUS_NO_SUCH_DOMAIN:                                              "NO_SUCH_DOMAIN",
	NT_STATUS_DOMAIN_EXISTS:                                               "DOMAIN_EXISTS",
	NT_STATUS_DOMAIN_LIMIT_EXCEEDED:                                       "DOMAIN_LIMIT_EXCEEDED",
	NT_STATUS_OPLOCK_NOT_GRANTED:                                          "OPLOCK_NOT_GRANTED",
	NT_STATUS_INVALID_OPLOCK_PROTOCOL:                                     "INVALID_OPLOCK_PROTOCOL",
	NT_STATUS_INTERNAL_DB_CORRUPTION:                                      "INTERNAL_DB_CORRUPTION",
	NT_STATUS_INTERNAL_ERROR:                                              "INTERNAL_ERROR",
	NT_STATUS_GENERIC_NOT_MAPPED:                                          "GENERIC_NOT_MAPPED",
	NT_STATUS_BAD_DESCRIPTOR_FORMAT:                                       "BAD_DESCRIPTOR_FORMAT",
	NT_STATUS_INVALID_USER_BUFFER:                                         "INVALID_USER_BUFFER",
	NT_STATUS_UNEXPECTED_IO_ERROR:                                         "UNEXPECTED_IO_ERROR",
	NT_STATUS_UNEXPECTED_MM_CREATE_ERR:                                    "UNEXPECTED_MM_CREATE_ERR",
	NT_STATUS_UNEXPECTED_MM_MAP_ERROR:                                     "UNEXPECTED_MM_MAP_ERROR",
	NT_STATUS_UNEXPECTED_MM_EXTEND_ERR:                                    "UNEXPECTED_MM_EXTEND_ERR",
	NT_STATUS_NOT_LOGON_PROCESS:                                           "NOT_LOGON_PROCESS",
	NT_STATUS_LOGON_SESSION_EXISTS:                                        "LOGON_SESSION_EXISTS",
	NT_STATUS_INVALID_PARAMETER_1:                                         "INVALID_PARAMETER_1",
	NT_STATUS_INVALID_PARAMETER_2:                                         "INVALID_PARAMETER_2",
	NT_STATUS_INVALID_PARAMETER_3:                                         "INVALID_PARAMETER_3",
	NT_STATUS_INVALID_PARAMETER_4:                                         "INVALID_PARAMETER_4",
	NT_STATUS_INVALID_PARAMETER_5:                                         "INVALID_PARAMETER_5",
	NT_STATUS_INVALID_PARAMETER_6:                                         "INVALID_PARAMETER_6",
	NT_STATUS_INVALID_PARAMETER_7:                                         "INVALID_PARAMETER_7",
	NT_STATUS_INVALID_PARAMETER_8:                                         "INVALID_PARAMETER_8",
	NT_STATUS_INVALID_PARAMETER_9:                                         "INVALID_PARAMETER_9",
	NT_STATUS_INVALID_PARAMETER_10:                                        "INVALID_PARAMETER_10",
	NT_STATUS_INVALID_PARAMETER_11:                                        "INVALID_PARAMETER_11",
	NT_STATUS_INVALID_PARAMETER_12:                                        "INVALID_PARAMETER_12",
	NT_STATUS_REDIRECTOR_NOT_STARTED:                                      "REDIRECTOR_NOT_STARTED",
	NT_STATUS_REDIRECTOR_STARTED:                                          "REDIRECTOR_STARTED",
	NT_STATUS_STACK_OVERFLOW:                                              "STACK_OVERFLOW",
	NT_STATUS_NO_SUCH_PACKAGE:                                             "NO_SUCH_PACKAGE",
	NT_STATUS_BAD_FUNCTION_TABLE:                                          "BAD_FUNCTION_TABLE",
	NT_STATUS_VARIABLE_NOT_FOUND:                                          "VARIABLE_NOT_FOUND",
	NT_STATUS_DIRECTORY_NOT_EMPTY:                                         "DIRECTORY_NOT_EMPTY",
	NT_STATUS_FILE_CORRUPT_ERROR:                                          "FILE_CORRUPT_ERROR",
	NT_STATUS_NOT_A_DIRECTORY:                                             "NOT_A_DIRECTORY",
	NT_STATUS_BAD_LOGON_SESSION_STATE:                                     "BAD_LOGON_SESSION_STATE",
	NT_STATUS_LOGON_SESSION_COLLISION:                                     "LOGON_SESSION_COLLISION",
	NT_STATUS_NAME_TOO_LONG:                                               "NAME_TOO_LONG",
	NT_STATUS_FILES_OPEN:                                                  "FILES_OPEN",
	NT_STATUS_CONNECTION_IN_USE:                                           "CONNECTION_IN_USE",
	NT_STATUS_MESSAGE_NOT_FOUND:                                           "MESSAGE_NOT_FOUND",
	NT_STATUS_PROCESS_IS_TERMINATING:                                      "PROCESS_IS_TERMINATING",
	NT_STATUS_INVALID_LOGON_TYPE:                                          "INVALID_LOGON_TYPE",
	NT_STATUS_NO_GUID_TRANSLATION:                                         "NO_GUID_TRANSLATION",
	NT_STATUS_CANNOT_IMPERSONATE:                                          "CANNOT_IMPERSONATE",
	NT_STATUS_IMAGE_ALREADY_LOADED:                                        "IMAGE_ALREADY_LOADED",
	NT_STATUS_NO_LDT:                                                      "NO_LDT",
	NT_STATUS_INVALID_LDT_SIZE:                                            "INVALID_LDT_SIZE",
	NT_STATUS_INVALID_LDT_OFFSET:                                          "INVALID_LDT_OFFSET",
	NT_STATUS_INVALID_LDT_DESCRIPTOR:                                      "INVALID_LDT_DESCRIPTOR",
	NT_STATUS_INVALID_IMAGE_NE_FORMAT:                                     "INVALID_IMAGE_NE_FORMAT",
	NT_STATUS_RXACT_INVALID_STATE:                                         "RXACT_INVALID_STATE",
	NT_STATUS_RXACT_COMMIT_FAILURE:                                        "RXACT_COMMIT_FAILURE",
	NT_STATUS_MAPPED_FILE_SIZE_ZERO:                                       "MAPPED_FILE_SIZE_ZERO",
	NT_STATUS_TOO_MANY_OPENED_FILES:                                       "TOO_MANY_OPENED_FILES",
	NT_STATUS_CANCELLED:                                                   "CANCELLED",
	NT_STATUS_CANNOT_DELETE:                                               "CANNOT_DELETE",
	NT_STATUS_INVALID_COMPUTER_NAME:                                       "INVALID_COMPUTER_NAME",
	NT_STATUS_FILE_DELETED:                                                "FILE_DELETED",
	NT_STATUS_SPECIAL_ACCOUNT:                                             "SPECIAL_ACCOUNT",
	NT_STATUS_SPECIAL_GROUP:                                               "SPECIAL_GROUP",
	NT_STATUS_SPECIAL_USER:                                                "SPECIAL_USER",
	NT_STATUS_MEMBERS_PRIMARY_GROUP:                                       "MEMBERS_PRIMARY_GROUP",
	NT_STATUS_FILE_CLOSED:                                                 "FILE_CLOSED",
	NT_STATUS_TOO_MANY_THREADS:                                            "TOO_MANY_THREADS",
	NT_STATUS_THREAD_NOT_IN_PROCESS:                                       "THREAD_NOT_IN_PROCESS",
	NT_STATUS_TOKEN_ALREADY_IN_USE:                                        "TOKEN_ALREADY_IN_USE",
	NT_STATUS_PAGEFILE_QUOTA_EXCEEDED:                                     "PAGEFILE_QUOTA_EXCEEDED",
	NT_STATUS_COMMITMENT_LIMIT:                                            "COMMITMENT_LIMIT",
	NT_STATUS_INVALID_IMAGE_LE_FORMAT:                                     "INVALID_IMAGE_LE_FORMAT",
	NT_STATUS_INVALID_IMAGE_NOT_MZ:                                        "INVALID_IMAGE_NOT_MZ",
	NT_STATUS_INVALID_IMAGE_PROTECT:                                       "INVALID_IMAGE_PROTECT",
	NT_STATUS_INVALID_IMAGE_WIN_16:                                        "INVALID_IMAGE_WIN_16",
	NT_STATUS_LOGON_SERVER_CONFLICT:                                       "LOGON_SERVER_CONFLICT",
	NT_STATUS_TIME_DIFFERENCE_AT_DC:                                       "TIME_DIFFERENCE_AT_DC",
	NT_STATUS_SYNCHRONIZATION_REQUIRED:                                    "SYNCHRONIZATION_REQUIRED",
	NT_STATUS_DLL_NOT_FOUND:                                               "DLL_NOT_FOUND",
	NT_STATUS_OPEN_FAILED:                                                 "OPEN_FAILED",
	NT_STATUS_IO_PRIVILEGE_FAILED:                                         "IO_PRIVILEGE_FAILED",
	NT_STATUS_ORDINAL_NOT_FOUND:                                           "ORDINAL_NOT_FOUND",
	NT_STATUS_ENTRYPOINT_NOT_FOUND:                                        "ENTRYPOINT_NOT_FOUND",
	NT_STATUS_CONTROL_C_EXIT:                                              "CONTROL_C_EXIT",
	NT_STATUS_LOCAL_DISCONNECT:                                            "LOCAL_DISCONNECT",
	NT_STATUS_REMOTE_DISCONNECT:                                           "REMOTE_DISCONNECT",
	NT_STATUS_REMOTE_RESOURCES:                                            "REMOTE_RESOURCES",
	NT_STATUS_LINK_FAILED:                                                 "LINK_FAILED",
	NT_STATUS_LINK_TIMEOUT:                                                "LINK_TIMEOUT",
	NT_STATUS_INVALID_CONNECTION:                                          "INVALID_CONNECTION",
	NT_STATUS_INVALID_ADDRESS:                                             "INVALID_ADDRESS",
	NT_STATUS_DLL_INIT_FAILED:                                             "DLL_INIT_FAILED",
	NT_STATUS_MISSING_SYSTEMFILE:                                          "MISSING_SYSTEMFILE",
	NT_STATUS_UNHANDLED_EXCEPTION:                                         "UNHANDLED_EXCEPTION",
	NT_STATUS_APP_INIT_FAILURE:                                            "APP_INIT_FAILURE",
	NT_STATUS_PAGEFILE_CREATE_FAILED:                                      "PAGEFILE_CREATE_FAILED",
	NT_STATUS_NO_PAGEFILE:                                                 "NO_PAGEFILE",
	NT_STATUS_INVALID_LEVEL:                                               "INVALID_LEVEL",
	NT_STATUS_WRONG_PASSWORD_CORE:                                         "WRONG_PASSWORD_CORE",
	NT_STATUS_ILLEGAL_FLOAT_CONTEXT:                                       "ILLEGAL_FLOAT_CONTEXT",
	NT_STATUS_PIPE_BROKEN:                                                 "PIPE_BROKEN",
	NT_STATUS_REGISTRY_CORRUPT:                                            "REGISTRY_CORRUPT",
	NT_STATUS_REGISTRY_IO_FAILED:                                          "REGISTRY_IO_FAILED",
	NT_STATUS_NO_EVENT_PAIR:                                               "NO_EVENT_PAIR",
	NT_STATUS_UNRECOGNIZED_VOLUME:                                         "UNRECOGNIZED_VOLUME",
	NT_STATUS_SERIAL_NO_DEVICE_INITED:                                     "SERIAL_NO_DEVICE_INITED",
	NT_STATUS_NO_SUCH_ALIAS:                                               "NO_SUCH_ALIAS",
	NT_STATUS_MEMBER_NOT_IN_ALIAS:                                         "MEMBER_NOT_IN_ALIAS",
	NT_STATUS_MEMBER_IN_ALIAS:                                             "MEMBER_IN_ALIAS",
	NT_STATUS_ALIAS_EXISTS:                                                "ALIAS_EXISTS",
	NT_STATUS_LOGON_NOT_GRANTED:                                           "LOGON_NOT_GRANTED",
	NT_STATUS_TOO_MANY_SECRETS:                                            "TOO_MANY_SECRETS",
	NT_STATUS_SECRET_TOO_LONG:                                             "SECRET_TOO_LONG",
	NT_STATUS_INTERNAL_DB_ERROR:                                           "INTERNAL_DB_ERROR",
	NT_STATUS_FULLSCREEN_MODE:                                             "FULLSCREEN_MODE",
	NT_STATUS_TOO_MANY_CONTEXT_IDS:                                        "TOO_MANY_CONTEXT_IDS",
	NT_STATUS_LOGON_TYPE_NOT_GRANTED:                                      "LOGON_TYPE_NOT_GRANTED",
	NT_STATUS_NOT_REGISTRY_FILE:                                           "NOT_REGISTRY_FILE",
	NT_STATUS_NT_CROSS_ENCRYPTION_REQUIRED:                                "NT_CROSS_ENCRYPTION_REQUIRED",
	NT_STATUS_DOMAIN_CTRLR_CONFIG_ERROR:                                   "DOMAIN_CTRLR_CONFIG_ERROR",
	NT_STATUS_FT_MISSING_MEMBER:                                           "FT_MISSING_MEMBER",
	NT_STATUS_ILL_FORMED_SERVICE_ENTRY:                                    "ILL_FORMED_SERVICE_ENTRY",
	NT_STATUS_ILLEGAL_CHARACTER:                                           "ILLEGAL_CHARACTER",
	NT_STATUS_UNMAPPABLE_CHARACTER:                                        "UNMAPPABLE_CHARACTER",
	NT_STATUS_UNDEFINED_CHARACTER:                                         "UNDEFINED_CHARACTER",
	NT_STATUS_FLOPPY_VOLUME:                                               "FLOPPY_VOLUME",
	NT_STATUS_FLOPPY_ID_MARK_NOT_FOUND:                                    "FLOPPY_ID_MARK_NOT_FOUND",
	NT_STATUS_FLOPPY_WRONG_CYLINDER:                                       "FLOPPY_WRONG_CYLINDER",
	NT_STATUS_FLOPPY_UNKNOWN_ERROR:                                        "FLOPPY_UNKNOWN_ERROR",
	NT_STATUS_FLOPPY_BAD_REGISTERS:                                        "FLOPPY_BAD_REGISTERS",
	NT_STATUS_DISK_RECALIBRATE_FAILED:                                     "DISK_RECALIBRATE_FAILED",
	NT_STATUS_DISK_OPERATION_FAILED:                                       "DISK_OPERATION_FAILED",
	NT_STATUS_DISK_RESET_FAILED:                                           "DISK_RESET_FAILED",
	NT_STATUS_SHARED_IRQ_BUSY:                                             "SHARED_IRQ_BUSY",
	NT_STATUS_FT_ORPHANING:                                                "FT_ORPHANING",
	NT_STATUS_BIOS_FAILED_TO_CONNECT_INTERRUPT:                            "BIOS_FAILED_TO_CONNECT_INTERRUPT",
	NT_STATUS_PARTITION_FAILURE:                                           "PARTITION_FAILURE",
	NT_STATUS_INVALID_BLOCK_LENGTH:                                        "INVALID_BLOCK_LENGTH",
	NT_STATUS_DEVICE_NOT_PARTITIONED:                                      "DEVICE_NOT_PARTITIONED",
	NT_STATUS_UNABLE_TO_LOCK_MEDIA:                                        "UNABLE_TO_LOCK_MEDIA",
	NT_STATUS_UNABLE_TO_UNLOAD_MEDIA:                                      "UNABLE_TO_UNLOAD_MEDIA",
	NT_STATUS_EOM_OVERFLOW:                                                "EOM_OVERFLOW",
	NT_STATUS_NO_MEDIA:                                                    "NO_MEDIA",
	NT_STATUS_NO_SUCH_MEMBER:                                              "NO_SUCH_MEMBER",
	NT_STATUS_INVALID_MEMBER:                                              "INVALID_MEMBER",
	NT_STATUS_KEY_DELETED:                                                 "KEY_DELETED",
	NT_STATUS_NO_LOG_SPACE:                                                "NO_LOG_SPACE",
	NT_STATUS_TOO_MANY_SIDS:                                               "TOO_MANY_SIDS",
	NT_STATUS_LM_CROSS_ENCRYPTION_REQUIRED:                                "LM_CROSS_ENCRYPTION_REQUIRED",
	NT_STATUS_KEY_HAS_CHILDREN:                                            "KEY_HAS_CHILDREN",
	NT_STATUS_CHILD_MUST_BE_VOLATILE:                                      "CHILD_MUST_BE_VOLATILE",
	NT_STATUS_DEVICE_CONFIGURATION_ERROR:                                  "DEVICE_CONFIGURATION_ERROR",
	NT_STATUS_DRIVER_INTERNAL_ERROR:                                       "DRIVER_INTERNAL_ERROR",
	NT_STATUS_INVALID_DEVICE_STATE:                                        "INVALID_DEVICE_STATE",
	NT_STATUS_IO_DEVICE_ERROR:                                             "IO_DEVICE_ERROR",
	NT_STATUS_DEVICE_PROTOCOL_ERROR:                                       "DEVICE_PROTOCOL_ERROR",
	NT_STATUS_BACKUP_CONTROLLER:                                           "BACKUP_CONTROLLER",
	NT_STATUS_LOG_FILE_FULL:                                               "LOG_FILE_FULL",
	NT_STATUS_TOO_LATE:                                                    "TOO_LATE",
	NT_STATUS_NO_TRUST_LSA_SECRET:                                         "NO_TRUST_LSA_SECRET",
	NT_STATUS_NO_TRUST_SAM_ACCOUNT:                                        "NO_TRUST_SAM_ACCOUNT",
	NT_STATUS_TRUSTED_DOMAIN_FAILURE:                                      "TRUSTED_DOMAIN_FAILURE",
	NT_STATUS_TRUSTED_RELATIONSHIP_FAILURE:                                "TRUSTED_RELATIONSHIP_FAILURE",
	NT_STATUS_EVENTLOG_FILE_CORRUPT:                                       "EVENTLOG_FILE_CORRUPT",
	NT_STATUS_EVENTLOG_CANT_START:                                         "EVENTLOG_CANT_START",
	NT_STATUS_TRUST_FAILURE:                                               "TRUST_FAILURE",
	NT_STATUS_MUTANT_LIMIT_EXCEEDED:                                       "MUTANT_LIMIT_EXCEEDED",
	NT_STATUS_NETLOGON_NOT_STARTED:                                        "NETLOGON_NOT_STARTED",
	NT_STATUS_ACCOUNT_EXPIRED:                                             "ACCOUNT_EXPIRED",
	NT_STATUS_POSSIBLE_DEADLOCK:                                           "POSSIBLE_DEADLOCK",
	NT_STATUS_NETWORK_CREDENTIAL_CONFLICT:                                 "NETWORK_CREDENTIAL_CONFLICT",
	NT_STATUS_REMOTE_SESSION_LIMIT:                                        "REMOTE_SESSION_LIMIT",
	NT_STATUS_EVENTLOG_FILE_CHANGED:                                       "EVENTLOG_FILE_CHANGED",
	NT_STATUS_NOLOGON_INTERDOMAIN_TRUST_ACCOUNT:                           "NOLOGON_INTERDOMAIN_TRUST_ACCOUNT",
	NT_STATUS_NOLOGON_WORKSTATION_TRUST_ACCOUNT:                           "NOLOGON_WORKSTATION_TRUST_ACCOUNT",
	NT_STATUS_NOLOGON_SERVER_TRUST_ACCOUNT:                                "NOLOGON_SERVER_TRUST_ACCOUNT",
	NT_STATUS_DOMAIN_TRUST_INCONSISTENT:                                   "DOMAIN_TRUST_INCONSISTENT",
	NT_STATUS_FS_DRIVER_REQUIRED:                                          "FS_DRIVER_REQUIRED",
	NT_STATUS_IMAGE_ALREADY_LOADED_AS_DLL:                                 "IMAGE_ALREADY_LOADED_AS_DLL",
	NT_STATUS_INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING:        "INCOMPATIBLE_WITH_GLOBAL_SHORT_NAME_REGISTRY_SETTING",
	NT_STATUS_SHORT_NAMES_NOT_ENABLED_ON_VOLUME:                           "SHORT_NAMES_NOT_ENABLED_ON_VOLUME",
	NT_STATUS_SECURITY_STREAM_IS_INCONSISTENT:                             "SECURITY_STREAM_IS_INCONSISTENT",
	NT_STATUS_INVALID_LOCK_RANGE:                                          "INVALID_LOCK_RANGE",
	NT_STATUS_INVALID_ACE_CONDITION:                                       "INVALID_ACE_CONDITION",
	NT_STATUS_IMAGE_SUBSYSTEM_NOT_PRESENT:                                 "IMAGE_SUBSYSTEM_NOT_PRESENT",
	NT_STATUS_NOTIFICATION_GUID_ALREADY_DEFINED:                           "NOTIFICATION_GUID_ALREADY_DEFINED",
	NT_STATUS_NETWORK_OPEN_RESTRICTION:                                    "NETWORK_OPEN_RESTRICTION",
	NT_STATUS_NO_USER_SESSION_KEY:                                         "NO_USER_SESSION_KEY",
	NT_STATUS_USER_SESSION_DELETED:                                        "USER_SESSION_DELETED",
	NT_STATUS_RESOURCE_LANG_NOT_FOUND:                                     "RESOURCE_LANG_NOT_FOUND",
	NT_STATUS_INSUFF_SERVER_RESOURCES:                                     "INSUFF_SERVER_RESOURCES",
	NT_STATUS_INVALID_BUFFER_SIZE:                                         "INVALID_BUFFER_SIZE",
	NT_STATUS_INVALID_ADDRESS_COMPONENT:                                   "INVALID_ADDRESS_COMPONENT",
	NT_STATUS_INVALID_ADDRESS_WILDCARD:                                    "INVALID_ADDRESS_WILDCARD",
	NT_STATUS_TOO_MANY_ADDRESSES:                                          "TOO_MANY_ADDRESSES",
	NT_STATUS_ADDRESS_ALREADY_EXISTS:                                      "ADDRESS_ALREADY_EXISTS",
	NT_STATUS_ADDRESS_CLOSED:                                              "ADDRESS_CLOSED",
	NT_STATUS_CONNECTION_DISCONNECTED:                                     "CONNECTION_DISCONNECTED",
	NT_STATUS_CONNECTION_RESET:                                            "CONNECTION_RESET",
	NT_STATUS_TOO_MANY_NODES:                                              "TOO_MANY_NODES",
	NT_STATUS_TRANSACTION_ABORTED:                                         "TRANSACTION_ABORTED",
	NT_STATUS_TRANSACTION_TIMED_OUT:                                       "TRANSACTION_TIMED_OUT",
	NT_STATUS_TRANSACTION_NO_RELEASE:                                      "TRANSACTION_NO_RELEASE",
	NT_STATUS_TRANSACTION_NO_MATCH:                                        "TRANSACTION_NO_MATCH",
	NT_STATUS_TRANSACTION_RESPONDED:                                       "TRANSACTION_RESPONDED",
	NT_STATUS_TRANSACTION_INVALID_ID:                                      "TRANSACTION_INVALID_ID",
	NT_STATUS_TRANSACTION_INVALID_TYPE:                                    "TRANSACTION_INVALID_TYPE",
	NT_STATUS_NOT_SERVER_SESSION:                                          "NOT_SERVER_SESSION",
	NT_STATUS_NOT_CLIENT_SESSION:                                          "NOT_CLIENT_SESSION",
	NT_STATUS_CANNOT_LOAD_REGISTRY_FILE:                                   "CANNOT_LOAD_REGISTRY_FILE",
	NT_STATUS_DEBUG_ATTACH_FAILED:                                         "DEBUG_ATTACH_FAILED",
	NT_STATUS_SYSTEM_PROCESS_TERMINATED:                                   "SYSTEM_PROCESS_TERMINATED",
	NT_STATUS_DATA_NOT_ACCEPTED:                                           "DATA_NOT_ACCEPTED",
	NT_STATUS_NO_BROWSER_SERVERS_FOUND:                                    "NO_BROWSER_SERVERS_FOUND",
	NT_STATUS_VDM_HARD_ERROR:                                              "VDM_HARD_ERROR",
	NT_STATUS_DRIVER_CANCEL_TIMEOUT:                                       "DRIVER_CANCEL_TIMEOUT",
	NT_STATUS_REPLY_MESSAGE_MISMATCH:                                      "REPLY_MESSAGE_MISMATCH",
	NT_STATUS_MAPPED_ALIGNMENT:                                            "MAPPED_ALIGNMENT",
	NT_STATUS_IMAGE_CHECKSUM_MISMATCH:                                     "IMAGE_CHECKSUM_MISMATCH",
	NT_STATUS_LOST_WRITEBEHIND_DATA:                                       "LOST_WRITEBEHIND_DATA",
	NT_STATUS_CLIENT_SERVER_PARAMETERS_INVALID:                            "CLIENT_SERVER_PARAMETERS_INVALID",
	NT_STATUS_PASSWORD_MUST_CHANGE:                                        "PASSWORD_MUST_CHANGE",
	NT_STATUS_NOT_FOUND:                                                   "NOT_FOUND",
	NT_STATUS_NOT_TINY_STREAM:                                             "NOT_TINY_STREAM",
	NT_STATUS_RECOVERY_FAILURE:                                            "RECOVERY_FAILURE",
	NT_STATUS_STACK_OVERFLOW_READ:                                         "STACK_OVERFLOW_READ",
	NT_STATUS_FAIL_CHECK:                                                  "FAIL_CHECK",
	NT_STATUS_DUPLICATE_OBJECTID:                                          "DUPLICATE_OBJECTID",
	NT_STATUS_OBJECTID_EXISTS:                                             "OBJECTID_EXISTS",
	NT_STATUS_CONVERT_TO_LARGE:                                            "CONVERT_TO_LARGE",
	NT_STATUS_RETRY:                                                       "RETRY",
	NT_STATUS_FOUND_OUT_OF_SCOPE:                                          "FOUND_OUT_OF_SCOPE",
	NT_STATUS_ALLOCATE_BUCKET:                                             "ALLOCATE_BUCKET",
	NT_STATUS_PROPSET_NOT_FOUND:                                           "PROPSET_NOT_FOUND",
	NT_STATUS_MARSHALL_OVERFLOW:                                           "MARSHALL_OVERFLOW",
	NT_STATUS_INVALID_VARIANT:                                             "INVALID_VARIANT",
	NT_STATUS_DOMAIN_CONTROLLER_NOT_FOUND:                                 "DOMAIN_CONTROLLER_NOT_FOUND",
	NT_STATUS_ACCOUNT_LOCKED_OUT:                                          "ACCOUNT_LOCKED_OUT",
	NT_STATUS_HANDLE_NOT_CLOSABLE:                                         "HANDLE_NOT_CLOSABLE",
	NT_STATUS_CONNECTION_REFUSED:                                          "CONNECTION_REFUSED",
	NT_STATUS_GRACEFUL_DISCONNECT:                                         "GRACEFUL_DISCONNECT",
	NT_STATUS_ADDRESS_ALREADY_ASSOCIATED:                                  "ADDRESS_ALREADY_ASSOCIATED",
	NT_STATUS_ADDRESS_NOT_ASSOCIATED:                                      "ADDRESS_NOT_ASSOCIATED",
	NT_STATUS_CONNECTION_INVALID:                                          "CONNECTION_INVALID",
	NT_STATUS_CONNECTION_ACTIVE:                                           "CONNECTION_ACTIVE",
	NT_STATUS_NETWORK_UNREACHABLE:                                         "NETWORK_UNREACHABLE",
	NT_STATUS_HOST_UNREACHABLE:                                            "HOST_UNREACHABLE",
	NT_STATUS_PROTOCOL_UNREACHABLE:                                        "PROTOCOL_UNREACHABLE",
	NT_STATUS_PORT_UNREACHABLE:                                            "PORT_UNREACHABLE",
	NT_STATUS_REQUEST_ABORTED:                                             "REQUEST_ABORTED",
	NT_STATUS_CONNECTION_ABORTED:                                          "CONNECTION_ABORTED",
	NT_STATUS_BAD_COMPRESSION_BUFFER:                                      "BAD_COMPRESSION_BUFFER",
	NT_STATUS_USER_MAPPED_FILE:                                            "USER_MAPPED_FILE",
	NT_STATUS_AUDIT_FAILED:                                                "AUDIT_FAILED",
	NT_STATUS_TIMER_RESOLUTION_NOT_SET:                                    "TIMER_RESOLUTION_NOT_SET",
	NT_STATUS_CONNECTION_COUNT_LIMIT:                                      "CONNECTION_COUNT_LIMIT",
	NT_STATUS_LOGIN_TIME_RESTRICTION:                                      "LOGIN_TIME_RESTRICTION",
	NT_STATUS_LOGIN_WKSTA_RESTRICTION:                                     "LOGIN_WKSTA_RESTRICTION",
	NT_STATUS_IMAGE_MP_UP_MISMATCH:                                        "IMAGE_MP_UP_MISMATCH",
	NT_STATUS_INSUFFICIENT_LOGON_INFO:                                     "INSUFFICIENT_LOGON_INFO",
	NT_STATUS_BAD_DLL_ENTRYPOINT:                                          "BAD_DLL_ENTRYPOINT",
	NT_STATUS_BAD_SERVICE_ENTRYPOINT:                                      "BAD_SERVICE_ENTRYPOINT",
	NT_STATUS_LPC_REPLY_LOST:                                              "LPC_REPLY_LOST",
	NT_STATUS_IP_ADDRESS_CONFLICT1:                                        "IP_ADDRESS_CONFLICT1",
	NT_STATUS_IP_ADDRESS_CONFLICT2:                                        "IP_ADDRESS_CONFLICT2",
	NT_STATUS_REGISTRY_QUOTA_LIMIT:                                        "REGISTRY_QUOTA_LIMIT",
	NT_STATUS_PATH_NOT_COVERED:                                            "PATH_NOT_COVERED",
	NT_STATUS_NO_CALLBACK_ACTIVE:                                          "NO_CALLBACK_ACTIVE",
	NT_STATUS_LICENSE_QUOTA_EXCEEDED:                                      "LICENSE_QUOTA_EXCEEDED",
	NT_STATUS_PWD_TOO_SHORT:                                               "PWD_TOO_SHORT",
	NT_STATUS_PWD_TOO_RECENT:                                              "PWD_TOO_RECENT",
	NT_STATUS_PWD_HISTORY_CONFLICT:                                        "PWD_HISTORY_CONFLICT",
	NT_STATUS_PLUGPLAY_NO_DEVICE:                                          "PLUGPLAY_NO_DEVICE",
	NT_STATUS_UNSUPPORTED_COMPRESSION:                                     "UNSUPPORTED_COMPRESSION",
	NT_STATUS_INVALID_HW_PROFILE:                                          "INVALID_HW_PROFILE",
	NT_STATUS_INVALID_PLUGPLAY_DEVICE_PATH:                                "INVALID_PLUGPLAY_DEVICE_PATH",
	NT_STATUS_DRIVER_ORDINAL_NOT_FOUND:                                    "DRIVER_ORDINAL_NOT_FOUND",
	NT_STATUS_DRIVER_ENTRYPOINT_NOT_FOUND:                                 "DRIVER_ENTRYPOINT_NOT_FOUND",
	NT_STATUS_RESOURCE_NOT_OWNED:                                          "RESOURCE_NOT_OWNED",
	NT_STATUS_TOO_MANY_LINKS:                                              "TOO_MANY_LINKS",
	NT_STATUS_QUOTA_LIST_INCONSISTENT:                                     "QUOTA_LIST_INCONSISTENT",
	NT_STATUS_FILE_IS_OFFLINE:                                             "FILE_IS_OFFLINE",
	NT_STATUS_EVALUATION_EXPIRATION:                                       "EVALUATION_EXPIRATION",
	NT_STATUS_ILLEGAL_DLL_RELOCATION:                                      "ILLEGAL_DLL_RELOCATION",
	NT_STATUS_LICENSE_VIOLATION:                                           "LICENSE_VIOLATION",
	NT_STATUS_DLL_INIT_FAILED_LOGOFF:                                      "DLL_INIT_FAILED_LOGOFF",
	NT_STATUS_DRIVER_UNABLE_TO_LOAD:                                       "DRIVER_UNABLE_TO_LOAD",
	NT_STATUS_DFS_UNAVAILABLE:                                             "DFS_UNAVAILABLE",
	NT_STATUS_VOLUME_DISMOUNTED:                                           "VOLUME_DISMOUNTED",
	NT_STATUS_WX86_INTERNAL_ERROR:                                         "WX86_INTERNAL_ERROR",
	NT_STATUS_WX86_FLOAT_STACK_CHECK:                                      "WX86_FLOAT_STACK_CHECK",
	NT_STATUS_VALIDATE_CONTINUE:                                           "VALIDATE_CONTINUE",
	NT_STATUS_NO_MATCH:                                                    "NO_MATCH",
	NT_STATUS_NO_MORE_MATCHES:                                             "NO_MORE_MATCHES",
	NT_STATUS_NOT_A_REPARSE_POINT:                                         "NOT_A_REPARSE_POINT",
	NT_STATUS_IO_REPARSE_TAG_INVALID:                                      "IO_REPARSE_TAG_INVALID",
	NT_STATUS_IO_REPARSE_TAG_MISMATCH:                                     "IO_REPARSE_TAG_MISMATCH",
	NT_STATUS_IO_REPARSE_DATA_INVALID:                                     "IO_REPARSE_DATA_INVALID",
	NT_STATUS_IO_REPARSE_TAG_NOT_HANDLED:                                  "IO_REPARSE_TAG_NOT_HANDLED",
	NT_STATUS_REPARSE_POINT_NOT_RESOLVED:                                  "REPARSE_POINT_NOT_RESOLVED",
	NT_STATUS_DIRECTORY_IS_A_REPARSE_POINT:                                "DIRECTORY_IS_A_REPARSE_POINT",
	NT_STATUS_RANGE_LIST_CONFLICT:                                         "RANGE_LIST_CONFLICT",
	NT_STATUS_SOURCE_ELEMENT_EMPTY:                                        "SOURCE_ELEMENT_EMPTY",
	NT_STATUS_DESTINATION_ELEMENT_FULL:                                    "DESTINATION_ELEMENT_FULL",
	NT_STATUS_ILLEGAL_ELEMENT_ADDRESS:                                     "ILLEGAL_ELEMENT_ADDRESS",
	NT_STATUS_MAGAZINE_NOT_PRESENT:                                        "MAGAZINE_NOT_PRESENT",
	NT_STATUS_REINITIALIZATION_NEEDED:                                     "REINITIALIZATION_NEEDED",
	NT_STATUS_ENCRYPTION_FAILED:                                           "ENCRYPTION_FAILED",
	NT_STATUS_DECRYPTION_FAILED:                                           "DECRYPTION_FAILED",
	NT_STATUS_RANGE_NOT_FOUND:                                             "RANGE_NOT_FOUND",
	NT_STATUS_NO_RECOVERY_POLICY:                                          "NO_RECOVERY_POLICY",
	NT_STATUS_NO_EFS:                                                      "NO_EFS",
	NT_STATUS_WRONG_EFS:                                                   "WRONG_EFS",
	NT_STATUS_NO_USER_KEYS:                                                "NO_USER_KEYS",
	NT_STATUS_FILE_NOT_ENCRYPTED:                                          "FILE_NOT_ENCRYPTED",
	NT_STATUS_NOT_EXPORT_FORMAT:                                           "NOT_EXPORT_FORMAT",
	NT_STATUS_FILE_ENCRYPTED:                                              "FILE_ENCRYPTED",
	NT_STATUS_WMI_GUID_NOT_FOUND:                                          "WMI_GUID_NOT_FOUND",
	NT_STATUS_WMI_INSTANCE_NOT_FOUND:                                      "WMI_INSTANCE_NOT_FOUND",
	NT_STATUS_WMI_ITEMID_NOT_FOUND:                                        "WMI_ITEMID_NOT_FOUND",
	NT_STATUS_WMI_TRY_AGAIN:                                               "WMI_TRY_AGAIN",
	NT_STATUS_SHARED_POLICY:                                               "SHARED_POLICY",
	NT_STATUS_POLICY_OBJECT_NOT_FOUND:                                     "POLICY_OBJECT_NOT_FOUND",
	NT_STATUS_POLICY_ONLY_IN_DS:                                           "POLICY_ONLY_IN_DS",
	NT_STATUS_VOLUME_NOT_UPGRADED:                                         "VOLUME_NOT_UPGRADED",
	NT_STATUS_REMOTE_STORAGE_NOT_ACTIVE:                                   "REMOTE_STORAGE_NOT_ACTIVE",
	NT_STATUS_REMOTE_STORAGE_MEDIA_ERROR:                                  "REMOTE_STORAGE_MEDIA_ERROR",
	NT_STATUS_NO_TRACKING_SERVICE:                                         "NO_TRACKING_SERVICE",
	NT_STATUS_SERVER_SID_MISMATCH:                                         "SERVER_SID_MISMATCH",
	NT_STATUS_DS_NO_ATTRIBUTE_OR_VALUE:                                    "DS_NO_ATTRIBUTE_OR_VALUE",
	NT_STATUS_DS_INVALID_ATTRIBUTE_SYNTAX:                                 "DS_INVALID_ATTRIBUTE_SYNTAX",
	NT_STATUS_DS_ATTRIBUTE_TYPE_UNDEFINED:                                 "DS_ATTRIBUTE_TYPE_UNDEFINED",
	NT_STATUS_DS_ATTRIBUTE_OR_VALUE_EXISTS:                                "DS_ATTRIBUTE_OR_VALUE_EXISTS",
	NT_STATUS_DS_BUSY:                                                     "DS_BUSY",
	NT_STATUS_DS_UNAVAILABLE:                                              "DS_UNAVAILABLE",
	NT_STATUS_DS_NO_RIDS_ALLOCATED:                                        "DS_NO_RIDS_ALLOCATED",
	NT_STATUS_DS_NO_MORE_RIDS:                                             "DS_NO_MORE_RIDS",
	NT_STATUS_DS_INCORRECT_ROLE_OWNER:                                     "DS_INCORRECT_ROLE_OWNER",
	NT_STATUS_DS_RIDMGR_INIT_ERROR:                                        "DS_RIDMGR_INIT_ERROR",
	NT_STATUS_DS_OBJ_CLASS_VIOLATION:                                      "DS_OBJ_CLASS_VIOLATION",
	NT_STATUS_DS_CANT_ON_NON_LEAF:                                         "DS_CANT_ON_NON_LEAF",
	NT_STATUS_DS_CANT_ON_RDN:                                              "DS_CANT_ON_RDN",
	NT_STATUS_DS_CANT_MOD_OBJ_CLASS:                                       "DS_CANT_MOD_OBJ_CLASS",
	NT_STATUS_DS_CROSS_DOM_MOVE_FAILED:                                    "DS_CROSS_DOM_MOVE_FAILED",
	NT_STATUS_DS_GC_NOT_AVAILABLE:                                         "DS_GC_NOT_AVAILABLE",
	NT_STATUS_DIRECTORY_SERVICE_REQUIRED:                                  "DIRECTORY_SERVICE_REQUIRED",
	NT_STATUS_REPARSE_ATTRIBUTE_CONFLICT:                                  "REPARSE_ATTRIBUTE_CONFLICT",
	NT_STATUS_CANT_ENABLE_DENY_ONLY:                                       "CANT_ENABLE_DENY_ONLY",
	NT_STATUS_FLOAT_MULTIPLE_FAULTS:                                       "FLOAT_MULTIPLE_FAULTS",
	NT_STATUS_FLOAT_MULTIPLE_TRAPS:                                        "FLOAT_MULTIPLE_TRAPS",
	NT_STATUS_DEVICE_REMOVED:                                              "DEVICE_REMOVED",
	NT_STATUS_JOURNAL_DELETE_IN_PROGRESS:                                  "JOURNAL_DELETE_IN_PROGRESS",
	NT_STATUS_JOURNAL_NOT_ACTIVE:                                          "JOURNAL_NOT_ACTIVE",
	NT_STATUS_NOINTERFACE:                                                 "NOINTERFACE",
	NT_STATUS_DS_ADMIN_LIMIT_EXCEEDED:                                     "DS_ADMIN_LIMIT_EXCEEDED",
	NT_STATUS_DRIVER_FAILED_SLEEP:                                         "DRIVER_FAILED_SLEEP",
	NT_STATUS_MUTUAL_AUTHENTICATION_FAILED:                                "MUTUAL_AUTHENTICATION_FAILED",
	NT_STATUS_CORRUPT_SYSTEM_FILE:                                         "CORRUPT_SYSTEM_FILE",
	NT_STATUS_DATATYPE_MISALIGNMENT_ERROR:                                 "DATATYPE_MISALIGNMENT_ERROR",
	NT_STATUS_WMI_READ_ONLY:                                               "WMI_READ_ONLY",
	NT_STATUS_WMI_SET_FAILURE:                                             "WMI_SET_FAILURE",
	NT_STATUS_COMMITMENT_MINIMUM:                                          "COMMITMENT_MINIMUM",
	NT_STATUS_REG_NAT_CONSUMPTION:                                         "REG_NAT_CONSUMPTION",
	NT_STATUS_TRANSPORT_FULL:                                              "TRANSPORT_FULL",
	NT_STATUS_DS_SAM_INIT_FAILURE:                                         "DS_SAM_INIT_FAILURE",
	NT_STATUS_ONLY_IF_CONNECTED:                                           "ONLY_IF_CONNECTED",
	NT_STATUS_DS_SENSITIVE_GROUP_VIOLATION:                                "DS_SENSITIVE_GROUP_VIOLATION",
	NT_STATUS_PNP_RESTART_ENUMERATION:                                     "PNP_RESTART_ENUMERATION",
	NT_STATUS_JOURNAL_ENTRY_DELETED:                                       "JOURNAL_ENTRY_DELETED",
	NT_STATUS_DS_CANT_MOD_PRIMARYGROUPID:                                  "DS_CANT_MOD_PRIMARYGROUPID",
	NT_STATUS_SYSTEM_IMAGE_BAD_SIGNATURE:                                  "SYSTEM_IMAGE_BAD_SIGNATURE",
	NT_STATUS_PNP_REBOOT_REQUIRED:                                         "PNP_REBOOT_REQUIRED",
	NT_STATUS_POWER_STATE_INVALID:                                         "POWER_STATE_INVALID",
	NT_STATUS_DS_INVALID_GROUP_TYPE:                                       "DS_INVALID_GROUP_TYPE",
	NT_STATUS_DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN:                       "DS_NO_NEST_GLOBALGROUP_IN_MIXEDDOMAIN",
	NT_STATUS_DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN:                        "DS_NO_NEST_LOCALGROUP_IN_MIXEDDOMAIN",
	NT_STATUS_DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER:                            "DS_GLOBAL_CANT_HAVE_LOCAL_MEMBER",
	NT_STATUS_DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER:                        "DS_GLOBAL_CANT_HAVE_UNIVERSAL_MEMBER",
	NT_STATUS_DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER:                         "DS_UNIVERSAL_CANT_HAVE_LOCAL_MEMBER",
	NT_STATUS_DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER:                      "DS_GLOBAL_CANT_HAVE_CROSSDOMAIN_MEMBER",
	NT_STATUS_DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER:                 "DS_LOCAL_CANT_HAVE_CROSSDOMAIN_LOCAL_MEMBER",
	NT_STATUS_DS_HAVE_PRIMARY_MEMBERS:                                     "DS_HAVE_PRIMARY_MEMBERS",
	NT_STATUS_WMI_NOT_SUPPORTED:                                           "WMI_NOT_SUPPORTED",
	NT_STATUS_INSUFFICIENT_POWER:                                          "INSUFFICIENT_POWER",
	NT_STATUS_SAM_NEED_BOOTKEY_PASSWORD:                                   "SAM_NEED_BOOTKEY_PASSWORD",
	NT_STATUS_SAM_NEED_BOOTKEY_FLOPPY:                                     "SAM_NEED_BOOTKEY_FLOPPY",
	NT_STATUS_DS_CANT_START:                                               "DS_CANT_START",
	NT_STATUS_DS_INIT_FAILURE:                                             "DS_INIT_FAILURE",
	NT_STATUS_SAM_INIT_FAILURE:                                            "SAM_INIT_FAILURE",
	NT_STATUS_DS_GC_REQUIRED:                                              "DS_GC_REQUIRED",
	NT_STATUS_DS_LOCAL_MEMBER_OF_LOCAL_ONLY:                               "DS_LOCAL_MEMBER_OF_LOCAL_ONLY",
	NT_STATUS_DS_NO_FPO_IN_UNIVERSAL_GROUPS:                               "DS_NO_FPO_IN_UNIVERSAL_GROUPS",
	NT_STATUS_DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED:                           "DS_MACHINE_ACCOUNT_QUOTA_EXCEEDED",
	NT_STATUS_CURRENT_DOMAIN_NOT_ALLOWED:                                  "CURRENT_DOMAIN_NOT_ALLOWED",
	NT_STATUS_CANNOT_MAKE:                                                 "CANNOT_MAKE",
	NT_STATUS_SYSTEM_SHUTDOWN:                                             "SYSTEM_SHUTDOWN",
	NT_STATUS_DS_INIT_FAILURE_CONSOLE:                                     "DS_INIT_FAILURE_CONSOLE",
	NT_STATUS_DS_SAM_INIT_FAILURE_CONSOLE:                                 "DS_SAM_INIT_FAILURE_CONSOLE",
	NT_STATUS_UNFINISHED_CONTEXT_DELETED:                                  "UNFINISHED_CONTEXT_DELETED",
	NT_STATUS_NO_TGT_REPLY:                                                "NO_TGT_REPLY",
	NT_STATUS_OBJECTID_NOT_FOUND:                                          "OBJECTID_NOT_FOUND",
	NT_STATUS_NO_IP_ADDRESSES:                                             "NO_IP_ADDRESSES",
	NT_STATUS_WRONG_CREDENTIAL_HANDLE:                                     "WRONG_CREDENTIAL_HANDLE",
	NT_STATUS_CRYPTO_SYSTEM_INVALID:                                       "CRYPTO_SYSTEM_INVALID",
	NT_STATUS_MAX_REFERRALS_EXCEEDED:                                      "MAX_REFERRALS_EXCEEDED",
	NT_STATUS_MUST_BE_KDC:                                                 "MUST_BE_KDC",
	NT_STATUS_STRONG_CRYPTO_NOT_SUPPORTED:                                 "STRONG_CRYPTO_NOT_SUPPORTED",
	NT_STATUS_TOO_MANY_PRINCIPALS:                                         "TOO_MANY_PRINCIPALS",
	NT_STATUS_NO_PA_DATA:                                                  "NO_PA_DATA",
	NT_STATUS_PKINIT_NAME_MISMATCH:                                        "PKINIT_NAME_MISMATCH",
	NT_STATUS_SMARTCARD_LOGON_REQUIRED:                                    "SMARTCARD_LOGON_REQUIRED",
	NT_STATUS_KDC_INVALID_REQUEST:                                         "KDC_INVALID_REQUEST",
	NT_STATUS_KDC_UNABLE_TO_REFER:                                         "KDC_UNABLE_TO_REFER",
	NT_STATUS_KDC_UNKNOWN_ETYPE:                                           "KDC_UNKNOWN_ETYPE",
	NT_STATUS_SHUTDOWN_IN_PROGRESS:                                        "SHUTDOWN_IN_PROGRESS",
	NT_STATUS_SERVER_SHUTDOWN_IN_PROGRESS:                                 "SERVER_SHUTDOWN_IN_PROGRESS",
	NT_STATUS_NOT_SUPPORTED_ON_SBS:                                        "NOT_SUPPORTED_ON_SBS",
	NT_STATUS_WMI_GUID_DISCONNECTED:                                       "WMI_GUID_DISCONNECTED",
	NT_STATUS_WMI_ALREADY_DISABLED:                                        "WMI_ALREADY_DISABLED",
	NT_STATUS_WMI_ALREADY_ENABLED:                                         "WMI_ALREADY_ENABLED",
	NT_STATUS_MFT_TOO_FRAGMENTED:                                          "MFT_TOO_FRAGMENTED",
	NT_STATUS_COPY_PROTECTION_FAILURE:                                     "COPY_PROTECTION_FAILURE",
	NT_STATUS_CSS_AUTHENTICATION_FAILURE:                                  "CSS_AUTHENTICATION_FAILURE",
	NT_STATUS_CSS_KEY_NOT_PRESENT:                                         "CSS_KEY_NOT_PRESENT",
	NT_STATUS_CSS_KEY_NOT_ESTABLISHED:                                     "CSS_KEY_NOT_ESTABLISHED",
	NT_STATUS_CSS_SCRAMBLED_SECTOR:                                        "CSS_SCRAMBLED_SECTOR",
	NT_STATUS_CSS_REGION_MISMATCH:                                         "CSS_REGION_MISMATCH",
	NT_STATUS_CSS_RESETS_EXHAUSTED:                                        "CSS_RESETS_EXHAUSTED",
	NT_STATUS_PKINIT_FAILURE:                                              "PKINIT_FAILURE",
	NT_STATUS_SMARTCARD_SUBSYSTEM_FAILURE:                                 "SMARTCARD_SUBSYSTEM_FAILURE",
	NT_STATUS_NO_KERB_KEY:                                                 "NO_KERB_KEY",
	NT_STATUS_HOST_DOWN:                                                   "HOST_DOWN",
	NT_STATUS_UNSUPPORTED_PREAUTH:                                         "UNSUPPORTED_PREAUTH",
	NT_STATUS_EFS_ALG_BLOB_TOO_BIG:                                        "EFS_ALG_BLOB_TOO_BIG",
	NT_STATUS_PORT_NOT_SET:                                                "PORT_NOT_SET",
	NT_STATUS_DEBUGGER_INACTIVE:                                           "DEBUGGER_INACTIVE",
	NT_STATUS_DS_VERSION_CHECK_FAILURE:                                    "DS_VERSION_CHECK_FAILURE",
	NT_STATUS_AUDITING_DISABLED:                                           "AUDITING_DISABLED",
	NT_STATUS_PRENT4_MACHINE_ACCOUNT:                                      "PRENT4_MACHINE_ACCOUNT",
	NT_STATUS_DS_AG_CANT_HAVE_UNIVERSAL_MEMBER:                            "DS_AG_CANT_HAVE_UNIVERSAL_MEMBER",
	NT_STATUS_INVALID_IMAGE_WIN_32:                                        "INVALID_IMAGE_WIN_32",
	NT_STATUS_INVALID_IMAGE_WIN_64:                                        "INVALID_IMAGE_WIN_64",
	NT_STATUS_BAD_BINDINGS:                                                "BAD_BINDINGS",
	NT_STATUS_NETWORK_SESSION_EXPIRED:                                     "NETWORK_SESSION_EXPIRED",
	NT_STATUS_APPHELP_BLOCK:                                               "APPHELP_BLOCK",
	NT_STATUS_ALL_SIDS_FILTERED:                                           "ALL_SIDS_FILTERED",
	NT_STATUS_NOT_SAFE_MODE_DRIVER:                                        "NOT_SAFE_MODE_DRIVER",
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_DEFAULT:                           "ACCESS_DISABLED_BY_POLICY_DEFAULT",
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PATH:                              "ACCESS_DISABLED_BY_POLICY_PATH",
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_PUBLISHER:                         "ACCESS_DISABLED_BY_POLICY_PUBLISHER",
	NT_STATUS_ACCESS_DISABLED_BY_POLICY_OTHER:                             "ACCESS_DISABLED_BY_POLICY_OTHER",
	NT_STATUS_FAILED_DRIVER_ENTRY:                                         "FAILED_DRIVER_ENTRY",
	NT_STATUS_DEVICE_ENUMERATION_ERROR:                                    "DEVICE_ENUMERATION_ERROR",
	NT_STATUS_MOUNT_POINT_NOT_RESOLVED:                                    "MOUNT_POINT_NOT_RESOLVED",
	NT_STATUS_INVALID_DEVICE_OBJECT_PARAMETER:                             "INVALID_DEVICE_OBJECT_PARAMETER",
	NT_STATUS_MCA_OCCURED:                                                 "MCA_OCCURED",
	NT_STATUS_DRIVER_BLOCKED_CRITICAL:                                     "DRIVER_BLOCKED_CRITICAL",
	NT_STATUS_DRIVER_BLOCKED:                                              "DRIVER_BLOCKED",
	NT_STATUS_DRIVER_DATABASE_ERROR:                                       "DRIVER_DATABASE_ERROR",
	NT_STATUS_SYSTEM_HIVE_TOO_LARGE:                                       "SYSTEM_HIVE_TOO_LARGE",
	NT_STATUS_INVALID_IMPORT_OF_NON_DLL:                                   "INVALID_IMPORT_OF_NON_DLL",
	NT_STATUS_NO_SECRETS:                                                  "NO_SECRETS",
	NT_STATUS_ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY:                       "ACCESS_DISABLED_NO_SAFER_UI_BY_POLICY",
	NT_STATUS_FAILED_STACK_SWITCH:                                         "FAILED_STACK_SWITCH",
	NT_STATUS_HEAP_CORRUPTION:                                             "HEAP_CORRUPTION",
	NT_STATUS_SMARTCARD_WRONG_PIN:                                         "SMARTCARD_WRONG_PIN",
	NT_STATUS_SMARTCARD_CARD_BLOCKED:                                      "SMARTCARD_CARD_BLOCKED",
	NT_STATUS_SMARTCARD_CARD_NOT_AUTHENTICATED:                            "SMARTCARD_CARD_NOT_AUTHENTICATED",
	NT_STATUS_SMARTCARD_NO_CARD:                                           "SMARTCARD_NO_CARD",
	NT_STATUS_SMARTCARD_NO_KEY_CONTAINER:                                  "SMARTCARD_NO_KEY_CONTAINER",
	NT_STATUS_SMARTCARD_NO_CERTIFICATE:                                    "SMARTCARD_NO_CERTIFICATE",
	NT_STATUS_SMARTCARD_NO_KEYSET:                                         "SMARTCARD_NO_KEYSET",
	NT_STATUS_SMARTCARD_IO_ERROR:                                          "SMARTCARD_IO_ERROR",
	NT_STATUS_DOWNGRADE_DETECTED:                                          "DOWNGRADE_DETECTED",
	NT_STATUS_SMARTCARD_CERT_REVOKED:                                      "SMARTCARD_CERT_REVOKED",
	NT_STATUS_ISSUING_CA_UNTRUSTED:                                        "ISSUING_CA_UNTRUSTED",
	NT_STATUS_REVOCATION_OFFLINE_C:                                        "REVOCATION_OFFLINE_C",
	NT_STATUS_PKINIT_CLIENT_FAILURE:                                       "PKINIT_CLIENT_FAILURE",
	NT_STATUS_SMARTCARD_CERT_EXPIRED:                                      "SMARTCARD_CERT_EXPIRED",
	NT_STATUS_DRIVER_FAILED_PRIOR_UNLOAD:                                  "DRIVER_FAILED_PRIOR_UNLOAD",
	NT_STATUS_SMARTCARD_SILENT_CONTEXT:                                    "SMARTCARD_SILENT_CONTEXT",
	NT_STATUS_PER_USER_TRUST_QUOTA_EXCEEDED:                               "PER_USER_TRUST_QUOTA_EXCEEDED",
	NT_STATUS_ALL_USER_TRUST_QUOTA_EXCEEDED:                               "ALL_USER_TRUST_QUOTA_EXCEEDED",
	NT_STATUS_USER_DELETE_TRUST_QUOTA_EXCEEDED:                            "USER_DELETE_TRUST_QUOTA_EXCEEDED",
	NT_STATUS_DS_NAME_NOT_UNIQUE:                                          "DS_NAME_NOT_UNIQUE",
	NT_STATUS_DS_DUPLICATE_ID_FOUND:                                       "DS_DUPLICATE_ID_FOUND",
	NT_STATUS_DS_GROUP_CONVERSION_ERROR:                                   "DS_GROUP_CONVERSION_ERROR",
	NT_STATUS_VOLSNAP_PREPARE_HIBERNATE:                                   "VOLSNAP_PREPARE_HIBERNATE",
	NT_STATUS_USER2USER_REQUIRED:                                          "USER2USER_REQUIRED",
	NT_STATUS_STACK_BUFFER_OVERRUN:                                        "STACK_BUFFER_OVERRUN",
	NT_STATUS_NO_S4U_PROT_SUPPORT:                                         "NO_S4U_PROT_SUPPORT",
	NT_STATUS_CROSSREALM_DELEGATION_FAILURE:                               "CROSSREALM_DELEGATION_FAILURE",
	NT_STATUS_REVOCATION_OFFLINE_KDC:                                      "REVOCATION_OFFLINE_KDC",
	NT_STATUS_ISSUING_CA_UNTRUSTED_KDC:                                    "ISSUING_CA_UNTRUSTED_KDC",
	NT_STATUS_KDC_CERT_EXPIRED:                                            "KDC_CERT_EXPIRED",
	NT_STATUS_KDC_CERT_REVOKED:                                            "KDC_CERT_REVOKED",
	NT_STATUS_PARAMETER_QUOTA_EXCEEDED:                                    "PARAMETER_QUOTA_EXCEEDED",
	NT_STATUS_HIBERNATION_FAILURE:                                         "HIBERNATION_FAILURE",
	NT_STATUS_DELAY_LOAD_FAILED:                                           "DELAY_LOAD_FAILED",
	NT_STATUS_AUTHENTICATION_FIREWALL_FAILED:                              "AUTHENTICATION_FIREWALL_FAILED",
	NT_STATUS_VDM_DISALLOWED:                                              "VDM_DISALLOWED",
	NT_STATUS_HUNG_DISPLAY_DRIVER_THREAD:                                  "HUNG_DISPLAY_DRIVER_THREAD",
	NT_STATUS_INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE:     "INSUFFICIENT_RESOURCE_FOR_SPECIFIED_SHARED_SECTION_SIZE",
	NT_STATUS_INVALID_CRUNTIME_PARAMETER:                                  "INVALID_CRUNTIME_PARAMETER",
	NT_STATUS_NTLM_BLOCKED:                                                "NTLM_BLOCKED",
	NT_STATUS_DS_SRC_SID_EXISTS_IN_FOREST:                                 "DS_SRC_SID_EXISTS_IN_FOREST",
	NT_STATUS_DS_DOMAIN_NAME_EXISTS_IN_FOREST:                             "DS_DOMAIN_NAME_EXISTS_IN_FOREST",
	NT_STATUS_DS_FLAT_NAME_EXISTS_IN_FOREST:                               "DS_FLAT_NAME_EXISTS_IN_FOREST",
	NT_STATUS_INVALID_USER_PRINCIPAL_NAME:                                 "INVALID_USER_PRINCIPAL_NAME",
	NT_STATUS_ASSERTION_FAILURE:                                           "ASSERTION_FAILURE",
	NT_STATUS_VERIFIER_STOP:                                               "VERIFIER_STOP",
	NT_STATUS_CALLBACK_POP_STACK:                                          "CALLBACK_POP_STACK",
	NT_STATUS_INCOMPATIBLE_DRIVER_BLOCKED:                                 "INCOMPATIBLE_DRIVER_BLOCKED",
	NT_STATUS_HIVE_UNLOADED:                                               "HIVE_UNLOADED",
	NT_STATUS_COMPRESSION_DISABLED:                                        "COMPRESSION_DISABLED",
	NT_STATUS_FILE_SYSTEM_LIMITATION:                                      "FILE_SYSTEM_LIMITATION",
	NT_STATUS_INVALID_IMAGE_HASH:                                          "INVALID_IMAGE_HASH",
	NT_STATUS_NOT_CAPABLE:                                                 "NOT_CAPABLE",
	NT_STATUS_REQUEST_OUT_OF_SEQUENCE:                                     "REQUEST_OUT_OF_SEQUENCE",
	NT_STATUS_IMPLEMENTATION_LIMIT:                                        "IMPLEMENTATION_LIMIT",
	NT_STATUS_ELEVATION_REQUIRED:                                          "ELEVATION_REQUIRED",
	NT_STATUS_NO_SECURITY_CONTEXT:                                         "NO_SECURITY_CONTEXT",
	NT_STATUS_PKU2U_CERT_FAILURE:                                          "PKU2U_CERT_FAILURE",
	NT_STATUS_BEYOND_VDL:                                                  "BEYOND_VDL",
	NT_STATUS_ENCOUNTERED_WRITE_IN_PROGRESS:                               "ENCOUNTERED_WRITE_IN_PROGRESS",
	NT_STATUS_PTE_CHANGED:                                                 "PTE_CHANGED",
	NT_STATUS_PURGE_FAILED:                                                "PURGE_FAILED",
	NT_STATUS_CRED_REQUIRES_CONFIRMATION:                                  "CRED_REQUIRES_CONFIRMATION",
	NT_STATUS_CS_ENCRYPTION_INVALID_SERVER_RESPONSE:                       "CS_ENCRYPTION_INVALID_SERVER_RESPONSE",
	NT_STATUS_CS_ENCRYPTION_UNSUPPORTED_SERVER:                            "CS_ENCRYPTION_UNSUPPORTED_SERVER",
	NT_STATUS_CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE:                       "CS_ENCRYPTION_EXISTING_ENCRYPTED_FILE",
	NT_STATUS_CS_ENCRYPTION_NEW_ENCRYPTED_FILE:                            "CS_ENCRYPTION_NEW_ENCRYPTED_FILE",
	NT_STATUS_CS_ENCRYPTION_FILE_NOT_CSE:                                  "CS_ENCRYPTION_FILE_NOT_CSE",
	NT_STATUS_INVALID_LABEL:                                               "INVALID_LABEL",
	NT_STATUS_DRIVER_PROCESS_TERMINATED:                                   "DRIVER_PROCESS_TERMINATED",
	NT_STATUS_AMBIGUOUS_SYSTEM_DEVICE:                                     "AMBIGUOUS_SYSTEM_DEVICE",
	NT_STATUS_SYSTEM_DEVICE_NOT_FOUND:                                     "SYSTEM_DEVICE_NOT_FOUND",
	NT_STATUS_RESTART_BOOT_APPLICATION:                                    "RESTART_BOOT_APPLICATION",
	NT_STATUS_INSUFFICIENT_NVRAM_RESOURCES:                                "INSUFFICIENT_NVRAM_RESOURCES",
	NT_STATUS_NO_RANGES_PROCESSED:                                         "NO_RANGES_PROCESSED",
	NT_STATUS_DEVICE_FEATURE_NOT_SUPPORTED:                                "DEVICE_FEATURE_NOT_SUPPORTED",
	NT_STATUS_DEVICE_UNREACHABLE:                                          "DEVICE_UNREACHABLE",
	NT_STATUS_INVALID_TOKEN:                                               "INVALID_TOKEN",
	NT_STATUS_SERVER_UNAVAILABLE:                                          "SERVER_UNAVAILABLE",
	NT_STATUS_INVALID_TASK_NAME:                                           "INVALID_TASK_NAME",
	NT_STATUS_INVALID_TASK_INDEX:                                          "INVALID_TASK_INDEX",
	NT_STATUS_THREAD_ALREADY_IN_TASK:                                      "THREAD_ALREADY_IN_TASK",
	NT_STATUS_CALLBACK_BYPASS:                                             "CALLBACK_BYPASS",
	NT_STATUS_FAIL_FAST_EXCEPTION:                                         "FAIL_FAST_EXCEPTION",
	NT_STATUS_IMAGE_CERT_REVOKED:                                          "IMAGE_CERT_REVOKED",
	NT_STATUS_PORT_CLOSED:                                                 "PORT_CLOSED",
	NT_STATUS_MESSAGE_LOST:                                                "MESSAGE_LOST",
	NT_STATUS_INVALID_MESSAGE:                                             "INVALID_MESSAGE",
	NT_STATUS_REQUEST_CANCELED:                                            "REQUEST_CANCELED",
	NT_STATUS_RECURSIVE_DISPATCH:                                          "RECURSIVE_DISPATCH",
	NT_STATUS_LPC_RECEIVE_BUFFER_EXPECTED:                                 "LPC_RECEIVE_BUFFER_EXPECTED",
	NT_STATUS_LPC_INVALID_CONNECTION_USAGE:                                "LPC_INVALID_CONNECTION_USAGE",
	NT_STATUS_LPC_REQUESTS_NOT_ALLOWED:                                    "LPC_REQUESTS_NOT_ALLOWED",
	NT_STATUS_RESOURCE_IN_USE:                                             "RESOURCE_IN_USE",
	NT_STATUS_HARDWARE_MEMORY_ERROR:                                       "HARDWARE_MEMORY_ERROR",
	NT_STATUS_THREADPOOL_HANDLE_EXCEPTION:                                 "THREADPOOL_HANDLE_EXCEPTION",
	NT_STATUS_THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED:                   "THREADPOOL_SET_EVENT_ON_COMPLETION_FAILED",
	NT_STATUS_THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED:           "THREADPOOL_RELEASE_SEMAPHORE_ON_COMPLETION_FAILED",
	NT_STATUS_THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED:               "THREADPOOL_RELEASE_MUTEX_ON_COMPLETION_FAILED",
	NT_STATUS_THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED:                "THREADPOOL_FREE_LIBRARY_ON_COMPLETION_FAILED",
	NT_STATUS_THREADPOOL_RELEASED_DURING_OPERATION:                        "THREADPOOL_RELEASED_DURING_OPERATION",
	NT_STATUS_CALLBACK_RETURNED_WHILE_IMPERSONATING:                       "CALLBACK_RETURNED_WHILE_IMPERSONATING",
	NT_STATUS_APC_RETURNED_WHILE_IMPERSONATING:                            "APC_RETURNED_WHILE_IMPERSONATING",
	NT_STATUS_PROCESS_IS_PROTECTED:                                        "PROCESS_IS_PROTECTED",
	NT_STATUS_MCA_EXCEPTION:                                               "MCA_EXCEPTION",
	NT_STATUS_CERTIFICATE_MAPPING_NOT_UNIQUE:                              "CERTIFICATE_MAPPING_NOT_UNIQUE",
	NT_STATUS_SYMLINK_CLASS_DISABLED:                                      "SYMLINK_CLASS_DISABLED",
	NT_STATUS_INVALID_IDN_NORMALIZATION:                                   "INVALID_IDN_NORMALIZATION",
	NT_STATUS_NO_UNICODE_TRANSLATION:                                      "NO_UNICODE_TRANSLATION",
	NT_STATUS_ALREADY_REGISTERED:                                          "ALREADY_REGISTERED",
	NT_STATUS_CONTEXT_MISMATCH:                                            "CONTEXT_MISMATCH",
	NT_STATUS_PORT_ALREADY_HAS_COMPLETION_LIST:                            "PORT_ALREADY_HAS_COMPLETION_LIST",
	NT_STATUS_CALLBACK_RETURNED_THREAD_PRIORITY:                           "CALLBACK_RETURNED_THREAD_PRIORITY",
	NT_STATUS_INVALID_THREAD:                                              "INVALID_THREAD",
	NT_STATUS_CALLBACK_RETURNED_TRANSACTION:                               "CALLBACK_RETURNED_TRANSACTION",
	NT_STATUS_CALLBACK_RETURNED_LDR_LOCK:                                  "CALLBACK_RETURNED_LDR_LOCK",
	NT_STATUS_CALLBACK_RETURNED_LANG:                                      "CALLBACK_RETURNED_LANG",
	NT_STATUS_CALLBACK_RETURNED_PRI_BACK:                                  "CALLBACK_RETURNED_PRI_BACK",
	NT_STATUS_DISK_REPAIR_DISABLED:                                        "DISK_REPAIR_DISABLED",
	NT_STATUS_DS_DOMAIN_RENAME_IN_PROGRESS:                                "DS_DOMAIN_RENAME_IN_PROGRESS",
	NT_STATUS_DISK_QUOTA_EXCEEDED:                                         "DISK_QUOTA_EXCEEDED",
	NT_STATUS_CONTENT_BLOCKED:                                             "CONTENT_BLOCKED",
	NT_STATUS_BAD_CLUSTERS:                                                "BAD_CLUSTERS",
	NT_STATUS_VOLUME_DIRTY:                                                "VOLUME_DIRTY",
	NT_STATUS_FILE_CHECKED_OUT:                                            "FILE_CHECKED_OUT",
	NT_STATUS_CHECKOUT_REQUIRED:                                           "CHECKOUT_REQUIRED",
	NT_STATUS_BAD_FILE_TYPE:                                               "BAD_FILE_TYPE",
	NT_STATUS_FILE_TOO_LARGE:                                              "FILE_TOO_LARGE",
	NT_STATUS_FORMS_AUTH_REQUIRED:                                         "FORMS_AUTH_REQUIRED",
	NT_STATUS_VIRUS_INFECTED:                                              "VIRUS_INFECTED",
	NT_STATUS_VIRUS_DELETED:                                               "VIRUS_DELETED",
	NT_STATUS_BAD_MCFG_TABLE:                                              "BAD_MCFG_TABLE",
	NT_STATUS_BAD_DATA:                                                    "BAD_DATA",
	NT_STATUS_CANNOT_BREAK_OPLOCK:                                         "CANNOT_BREAK_OPLOCK",
	NT_STATUS_WOW_ASSERTION:                                               "WOW_ASSERTION",
	NT_STATUS_INVALID_SIGNATURE:                                           "INVALID_SIGNATURE",
	NT_STATUS_HMAC_NOT_SUPPORTED:                                          "HMAC_NOT_SUPPORTED",
	NT_STATUS_AUTH_TAG_MISMATCH:                                           "AUTH_TAG_MISMATCH",
	NT_STATUS_IPSEC_QUEUE_OVERFLOW:                                        "IPSEC_QUEUE_OVERFLOW",
	NT_STATUS_ND_QUEUE_OVERFLOW:                                           "ND_QUEUE_OVERFLOW",
	NT_STATUS_HOPLIMIT_EXCEEDED:                                           "HOPLIMIT_EXCEEDED",
	NT_STATUS_PROTOCOL_NOT_SUPPORTED:                                      "PROTOCOL_NOT_SUPPORTED",
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED:                  "LOST_WRITEBEHIND_DATA_NETWORK_DISCONNECTED",
	NT_STATUS_LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR:                  "LOST_WRITEBEHIND_DATA_NETWORK_SERVER_ERROR",
	NT_STATUS_LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR:                      "LOST_WRITEBEHIND_DATA_LOCAL_DISK_ERROR",
	NT_STATUS_XML_PARSE_ERROR:                                             "XML_PARSE_ERROR",
	NT_STATUS_XMLDSIG_ERROR:                                               "XMLDSIG_ERROR",
	NT_STATUS_WRONG_COMPARTMENT:                                           "WRONG_COMPARTMENT",
	NT_STATUS_AUTHIP_FAILURE:                                              "AUTHIP_FAILURE",
	NT_STATUS_DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS:                       "DS_OID_MAPPED_GROUP_CANT_HAVE_MEMBERS",
	NT_STATUS_DS_OID_NOT_FOUND:                                            "DS_OID_NOT_FOUND",
	NT_STATUS_HASH_NOT_SUPPORTED:                                          "HASH_NOT_SUPPORTED",
	NT_STATUS_HASH_NOT_PRESENT:                                            "HASH_NOT_PRESENT",
	NT_STATUS_OFFLOAD_READ_FLT_NOT_SUPPORTED:                              "OFFLOAD_READ_FLT_NOT_SUPPORTED",
	NT_STATUS_OFFLOAD_WRITE_FLT_NOT_SUPPORTED:                             "OFFLOAD_WRITE_FLT_NOT_SUPPORTED",
	NT_STATUS_OFFLOAD_READ_FILE_NOT_SUPPORTED:                             "OFFLOAD_READ_FILE_NOT_SUPPORTED",
	NT_STATUS_OFFLOAD_WRITE_FILE_NOT_SUPPORTED:                            "OFFLOAD_WRITE_FILE_NOT_SUPPORTED",
	NT_STATUS_DBG_NO_STATE_CHANGE:                                         "DBG_NO_STATE_CHANGE",
	NT_STATUS_DBG_APP_NOT_IDLE:                                            "DBG_APP_NOT_IDLE",
	NT_STATUS_RPC_NT_INVALID_STRING_BINDING:                               "RPC_NT_INVALID_STRING_BINDING",
	NT_STATUS_RPC_NT_WRONG_KIND_OF_BINDING:                                "RPC_NT_WRONG_KIND_OF_BINDING",
	NT_STATUS_RPC_NT_INVALID_BINDING:                                      "RPC_NT_INVALID_BINDING",
	NT_STATUS_RPC_NT_PROTSEQ_NOT_SUPPORTED:                                "RPC_NT_PROTSEQ_NOT_SUPPORTED",
	NT_STATUS_RPC_NT_INVALID_RPC_PROTSEQ:                                  "RPC_NT_INVALID_RPC_PROTSEQ",
	NT_STATUS_RPC_NT_INVALID_STRING_UUID:                                  "RPC_NT_INVALID_STRING_UUID",
	NT_STATUS_RPC_NT_INVALID_ENDPOINT_FORMAT:                              "RPC_NT_INVALID_ENDPOINT_FORMAT",
	NT_STATUS_RPC_NT_INVALID_NET_ADDR:                                     "RPC_NT_INVALID_NET_ADDR",
	NT_STATUS_RPC_NT_NO_ENDPOINT_FOUND:                                    "RPC_NT_NO_ENDPOINT_FOUND",
	NT_STATUS_RPC_NT_INVALID_TIMEOUT:                                      "RPC_NT_INVALID_TIMEOUT",
	NT_STATUS_RPC_NT_OBJECT_NOT_FOUND:                                     "RPC_NT_OBJECT_NOT_FOUND",
	NT_STATUS_RPC_NT_ALREADY_REGISTERED:                                   "RPC_NT_ALREADY_REGISTERED",
	NT_STATUS_RPC_NT_TYPE_ALREADY_REGISTERED:                              "RPC_NT_TYPE_ALREADY_REGISTERED",
	NT_STATUS_RPC_NT_ALREADY_LISTENING:                                    "RPC_NT_ALREADY_LISTENING",
	NT_STATUS_RPC_NT_NO_PROTSEQS_REGISTERED:                               "RPC_NT_NO_PROTSEQS_REGISTERED",
	NT_STATUS_RPC_NT_NOT_LISTENING:                                        "RPC_NT_NOT_LISTENING",
	NT_STATUS_RPC_NT_UNKNOWN_MGR_TYPE:                                     "RPC_NT_UNKNOWN_MGR_TYPE",
	NT_STATUS_RPC_NT_UNKNOWN_IF:                                           "RPC_NT_UNKNOWN_IF",
	NT_STATUS_RPC_NT_NO_BINDINGS:                                          "RPC_NT_NO_BINDINGS",
	NT_STATUS_RPC_NT_NO_PROTSEQS:                                          "RPC_NT_NO_PROTSEQS",
	NT_STATUS_RPC_NT_CANT_CREATE_ENDPOINT:                                 "RPC_NT_CANT_CREATE_ENDPOINT",
	NT_STATUS_RPC_NT_OUT_OF_RESOURCES:                                     "RPC_NT_OUT_OF_RESOURCES",
	NT_STATUS_RPC_NT_SERVER_UNAVAILABLE:                                   "RPC_NT_SERVER_UNAVAILABLE",
	NT_STATUS_RPC_NT_SERVER_TOO_BUSY:                                      "RPC_NT_SERVER_TOO_BUSY",
	NT_STATUS_RPC_NT_INVALID_NETWORK_OPTIONS:                              "RPC_NT_INVALID_NETWORK_OPTIONS",
	NT_STATUS_RPC_NT_NO_CALL_ACTIVE:                                       "RPC_NT_NO_CALL_ACTIVE",
	NT_STATUS_RPC_NT_CALL_FAILED:                                          "RPC_NT_CALL_FAILED",
	NT_STATUS_RPC_NT_CALL_FAILED_DNE:                                      "RPC_NT_CALL_FAILED_DNE",
	NT_STATUS_RPC_NT_PROTOCOL_ERROR:                                       "RPC_NT_PROTOCOL_ERROR",
	NT_STATUS_RPC_NT_UNSUPPORTED_TRANS_SYN:                                "RPC_NT_UNSUPPORTED_TRANS_SYN",
	NT_STATUS_RPC_NT_UNSUPPORTED_TYPE:                                     "RPC_NT_UNSUPPORTED_TYPE",
	NT_STATUS_RPC_NT_INVALID_TAG:                                          "RPC_NT_INVALID_TAG",
	NT_STATUS_RPC_NT_INVALID_BOUND:                                        "RPC_NT_INVALID_BOUND",
	NT_STATUS_RPC_NT_NO_ENTRY_NAME:                                        "RPC_NT_NO_ENTRY_NAME",
	NT_STATUS_RPC_NT_INVALID_NAME_SYNTAX:                                  "RPC_NT_INVALID_NAME_SYNTAX",
	NT_STATUS_RPC_NT_UNSUPPORTED_NAME_SYNTAX:                              "RPC_NT_UNSUPPORTED_NAME_SYNTAX",
	NT_STATUS_RPC_NT_UUID_NO_ADDRESS:                                      "RPC_NT_UUID_NO_ADDRESS",
	NT_STATUS_RPC_NT_DUPLICATE_ENDPOINT:                                   "RPC_NT_DUPLICATE_ENDPOINT",
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_TYPE:                                   "RPC_NT_UNKNOWN_AUTHN_TYPE",
	NT_STATUS_RPC_NT_MAX_CALLS_TOO_SMALL:                                  "RPC_NT_MAX_CALLS_TOO_SMALL",
	NT_STATUS_RPC_NT_STRING_TOO_LONG:                                      "RPC_NT_STRING_TOO_LONG",
	NT_STATUS_RPC_NT_PROTSEQ_NOT_FOUND:                                    "RPC_NT_PROTSEQ_NOT_FOUND",
	NT_STATUS_RPC_NT_PROCNUM_OUT_OF_RANGE:                                 "RPC_NT_PROCNUM_OUT_OF_RANGE",
	NT_STATUS_RPC_NT_BINDING_HAS_NO_AUTH:                                  "RPC_NT_BINDING_HAS_NO_AUTH",
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_SERVICE:                                "RPC_NT_UNKNOWN_AUTHN_SERVICE",
	NT_STATUS_RPC_NT_UNKNOWN_AUTHN_LEVEL:                                  "RPC_NT_UNKNOWN_AUTHN_LEVEL",
	NT_STATUS_RPC_NT_INVALID_AUTH_IDENTITY:                                "RPC_NT_INVALID_AUTH_IDENTITY",
	NT_STATUS_RPC_NT_UNKNOWN_AUTHZ_SERVICE:                                "RPC_NT_UNKNOWN_AUTHZ_SERVICE",
	NT_STATUS_EPT_NT_INVALID_ENTRY:                                        "EPT_NT_INVALID_ENTRY",
	NT_STATUS_EPT_NT_CANT_PERFORM_OP:                                      "EPT_NT_CANT_PERFORM_OP",
	NT_STATUS_EPT_NT_NOT_REGISTERED:                                       "EPT_NT_NOT_REGISTERED",
	NT_STATUS_RPC_NT_NOTHING_TO_EXPORT:                                    "RPC_NT_NOTHING_TO_EXPORT",
	NT_STATUS_RPC_NT_INCOMPLETE_NAME:                                      "RPC_NT_INCOMPLETE_NAME",
	NT_STATUS_RPC_NT_INVALID_VERS_OPTION:                                  "RPC_NT_INVALID_VERS_OPTION",
	NT_STATUS_RPC_NT_NO_MORE_MEMBERS:                                      "RPC_NT_NO_MORE_MEMBERS",
	NT_STATUS_RPC_NT_NOT_ALL_OBJS_UNEXPORTED:                              "RPC_NT_NOT_ALL_OBJS_UNEXPORTED",
	NT_STATUS_RPC_NT_INTERFACE_NOT_FOUND:                                  "RPC_NT_INTERFACE_NOT_FOUND",
	NT_STATUS_RPC_NT_ENTRY_ALREADY_EXISTS:                                 "RPC_NT_ENTRY_ALREADY_EXISTS",
	NT_STATUS_RPC_NT_ENTRY_NOT_FOUND:                                      "RPC_NT_ENTRY_NOT_FOUND",
	NT_STATUS_RPC_NT_NAME_SERVICE_UNAVAILABLE:                             "RPC_NT_NAME_SERVICE_UNAVAILABLE",
	NT_STATUS_RPC_NT_INVALID_NAF_ID:                                       "RPC_NT_INVALID_NAF_ID",
	NT_STATUS_RPC_NT_CANNOT_SUPPORT:                                       "RPC_NT_CANNOT_SUPPORT",
	NT_STATUS_RPC_NT_NO_CONTEXT_AVAILABLE:                                 "RPC_NT_NO_CONTEXT_AVAILABLE",
	NT_STATUS_RPC_NT_INTERNAL_ERROR:                                       "RPC_NT_INTERNAL_ERROR",
	NT_STATUS_RPC_NT_ZERO_DIVIDE:                                          "RPC_NT_ZERO_DIVIDE",
	NT_STATUS_RPC_NT_ADDRESS_ERROR:                                        "RPC_NT_ADDRESS_ERROR",
	NT_STATUS_RPC_NT_FP_DIV_ZERO:                                          "RPC_NT_FP_DIV_ZERO",
	NT_STATUS_RPC_NT_FP_UNDERFLOW:                                         "RPC_NT_FP_UNDERFLOW",
	NT_STATUS_RPC_NT_FP_OVERFLOW:                                          "RPC_NT_FP_OVERFLOW",
	NT_STATUS_RPC_NT_CALL_IN_PROGRESS:                                     "RPC_NT_CALL_IN_PROGRESS",
	NT_STATUS_RPC_NT_NO_MORE_BINDINGS:                                     "RPC_NT_NO_MORE_BINDINGS",
	NT_STATUS_RPC_NT_GROUP_MEMBER_NOT_FOUND:                               "RPC_NT_GROUP_MEMBER_NOT_FOUND",
	NT_STATUS_EPT_NT_CANT_CREATE:                                          "EPT_NT_CANT_CREATE",
	NT_STATUS_RPC_NT_INVALID_OBJECT:                                       "RPC_NT_INVALID_OBJECT",
	NT_STATUS_RPC_NT_NO_INTERFACES:                                        "RPC_NT_NO_INTERFACES",
	NT_STATUS_RPC_NT_CALL_CANCELLED:                                       "RPC_NT_CALL_CANCELLED",
	NT_STATUS_RPC_NT_BINDING_INCOMPLETE:                                   "RPC_NT_BINDING_INCOMPLETE",
	NT_STATUS_RPC_NT_COMM_FAILURE:                                         "RPC_NT_COMM_FAILURE",
	NT_STATUS_RPC_NT_UNSUPPORTED_AUTHN_LEVEL:                              "RPC_NT_UNSUPPORTED_AUTHN_LEVEL",
	NT_STATUS_RPC_NT_NO_PRINC_NAME:                                        "RPC_NT_NO_PRINC_NAME",
	NT_STATUS_RPC_NT_NOT_RPC_ERROR:                                        "RPC_NT_NOT_RPC_ERROR",
	NT_STATUS_RPC_NT_SEC_PKG_ERROR:                                        "RPC_NT_SEC_PKG_ERROR",
	NT_STATUS_RPC_NT_NOT_CANCELLED:                                        "RPC_NT_NOT_CANCELLED",
	NT_STATUS_RPC_NT_INVALID_ASYNC_HANDLE:                                 "RPC_NT_INVALID_ASYNC_HANDLE",
	NT_STATUS_RPC_NT_INVALID_ASYNC_CALL:                                   "RPC_NT_INVALID_ASYNC_CALL",
	NT_STATUS_RPC_NT_PROXY_ACCESS_DENIED:                                  "RPC_NT_PROXY_ACCESS_DENIED",
	NT_STATUS_RPC_NT_NO_MORE_ENTRIES:                                      "RPC_NT_NO_MORE_ENTRIES",
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_OPEN_FAIL:                              "RPC_NT_SS_CHAR_TRANS_OPEN_FAIL",
	NT_STATUS_RPC_NT_SS_CHAR_TRANS_SHORT_FILE:                             "RPC_NT_SS_CHAR_TRANS_SHORT_FILE",
	NT_STATUS_RPC_NT_SS_IN_NULL_CONTEXT:                                   "RPC_NT_SS_IN_NULL_CONTEXT",
	NT_STATUS_RPC_NT_SS_CONTEXT_MISMATCH:                                  "RPC_NT_SS_CONTEXT_MISMATCH",
	NT_STATUS_RPC_NT_SS_CONTEXT_DAMAGED:                                   "RPC_NT_SS_CONTEXT_DAMAGED",
	NT_STATUS_RPC_NT_SS_HANDLES_MISMATCH:                                  "RPC_NT_SS_HANDLES_MISMATCH",
	NT_STATUS_RPC_NT_SS_CANNOT_GET_CALL_HANDLE:                            "RPC_NT_SS_CANNOT_GET_CALL_HANDLE",
	NT_STATUS_RPC_NT_NULL_REF_POINTER:                                     "RPC_NT_NULL_REF_POINTER",
	NT_STATUS_RPC_NT_ENUM_VALUE_OUT_OF_RANGE:                              "RPC_NT_ENUM_VALUE_OUT_OF_RANGE",
	NT_STATUS_RPC_NT_BYTE_COUNT_TOO_SMALL:                                 "RPC_NT_BYTE_COUNT_TOO_SMALL",
	NT_STATUS_RPC_NT_BAD_STUB_DATA:                                        "RPC_NT_BAD_STUB_DATA",
	NT_STATUS_RPC_NT_INVALID_ES_ACTION:                                    "RPC_NT_INVALID_ES_ACTION",
	NT_STATUS_RPC_NT_WRONG_ES_VERSION:                                     "RPC_NT_WRONG_ES_VERSION",
	NT_STATUS_RPC_NT_WRONG_STUB_VERSION:                                   "RPC_NT_WRONG_STUB_VERSION",
	NT_STATUS_RPC_NT_INVALID_PIPE_OBJECT:                                  "RPC_NT_INVALID_PIPE_OBJECT",
	NT_STATUS_RPC_NT_INVALID_PIPE_OPERATION:                               "RPC_NT_INVALID_PIPE_OPERATION",
	NT_STATUS_RPC_NT_WRONG_PIPE_VERSION:                                   "RPC_NT_WRONG_PIPE_VERSION",
	NT_STATUS_RPC_NT_PIPE_CLOSED:                                          "RPC_NT_PIPE_CLOSED",
	NT_STATUS_RPC_NT_PIPE_DISCIPLINE_ERROR:                                "RPC_NT_PIPE_DISCIPLINE_ERROR",
	NT_STATUS_RPC_NT_PIPE_EMPTY:                                           "RPC_NT_PIPE_EMPTY",
	NT_STATUS_PNP_BAD_MPS_TABLE:                                           "PNP_BAD_MPS_TABLE",
	NT_STATUS_PNP_TRANSLATION_FAILED:                                      "PNP_TRANSLATION_FAILED",
	NT_STATUS_PNP_IRQ_TRANSLATION_FAILED:                                  "PNP_IRQ_TRANSLATION_FAILED",
	NT_STATUS_PNP_INVALID_ID:                                              "PNP_INVALID_ID",
	NT_STATUS_IO_REISSUE_AS_CACHED:                                        "IO_REISSUE_AS_CACHED",
	NT_STATUS_CTX_WINSTATION_NAME_INVALID:                                 "CTX_WINSTATION_NAME_INVALID",
	NT_STATUS_CTX_INVALID_PD:                                              "CTX_INVALID_PD",
	NT_STATUS_CTX_PD_NOT_FOUND:                                            "CTX_PD_NOT_FOUND",
	NT_STATUS_CTX_CLOSE_PENDING:                                           "CTX_CLOSE_PENDING",
	NT_STATUS_CTX_NO_OUTBUF:                                               "CTX_NO_OUTBUF",
	NT_STATUS_CTX_MODEM_INF_NOT_FOUND:                                     "CTX_MODEM_INF_NOT_FOUND",
	NT_STATUS_CTX_INVALID_MODEMNAME:                                       "CTX_INVALID_MODEMNAME",
	NT_STATUS_CTX_RESPONSE_ERROR:                                          "CTX_RESPONSE_ERROR",
	NT_STATUS_CTX_MODEM_RESPONSE_TIMEOUT:                                  "CTX_MODEM_RESPONSE_TIMEOUT",
	NT_STATUS_CTX_MODEM_RESPONSE_NO_CARRIER:                               "CTX_MODEM_RESPONSE_NO_CARRIER",
	NT_STATUS_CTX_MODEM_RESPONSE_NO_DIALTONE:                              "CTX_MODEM_RESPONSE_NO_DIALTONE",
	NT_STATUS_CTX_MODEM_RESPONSE_BUSY:                                     "CTX_MODEM_RESPONSE_BUSY",
	NT_STATUS_CTX_MODEM_RESPONSE_VOICE:                                    "CTX_MODEM_RESPONSE_VOICE",
	NT_STATUS_CTX_TD_ERROR:                                                "CTX_TD_ERROR",
	NT_STATUS_CTX_LICENSE_CLIENT_INVALID:                                  "CTX_LICENSE_CLIENT_INVALID",
	NT_STATUS_CTX_LICENSE_NOT_AVAILABLE:                                   "CTX_LICENSE_NOT_AVAILABLE",
	NT_STATUS_CTX_LICENSE_EXPIRED:                                         "CTX_LICENSE_EXPIRED",
	NT_STATUS_CTX_WINSTATION_NOT_FOUND:                                    "CTX_WINSTATION_NOT_FOUND",
	NT_STATUS_CTX_WINSTATION_NAME_COLLISION:                               "CTX_WINSTATION_NAME_COLLISION",
	NT_STATUS_CTX_WINSTATION_BUSY:                                         "CTX_WINSTATION_BUSY",
	NT_STATUS_CTX_BAD_VIDEO_MODE:                                          "CTX_BAD_VIDEO_MODE",
	NT_STATUS_CTX_GRAPHICS_INVALID:                                        "CTX_GRAPHICS_INVALID",
	NT_STATUS_CTX_NOT_CONSOLE:                                             "CTX_NOT_CONSOLE",
	NT_STATUS_CTX_CLIENT_QUERY_TIMEOUT:                                    "CTX_CLIENT_QUERY_TIMEOUT",
	NT_STATUS_CTX_CONSOLE_DISCONNECT:                                      "CTX_CONSOLE_DISCONNECT",
	NT_STATUS_CTX_CONSOLE_CONNECT:                                         "CTX_CONSOLE_CONNECT",
	NT_STATUS_CTX_SHADOW_DENIED:                                           "CTX_SHADOW_DENIED",
	NT_STATUS_CTX_WINSTATION_ACCESS_DENIED:                                "CTX_WINSTATION_ACCESS_DENIED",
	NT_STATUS_CTX_INVALID_WD:                                              "CTX_INVALID_WD",
	NT_STATUS_CTX_WD_NOT_FOUND:                                            "CTX_WD_NOT_FOUND",
	NT_STATUS_CTX_SHADOW_INVALID:                                          "CTX_SHADOW_INVALID",
	NT_STATUS_CTX_SHADOW_DISABLED:                                         "CTX_SHADOW_DISABLED",
	NT_STATUS_RDP_PROTOCOL_ERROR:                                          "RDP_PROTOCOL_ERROR",
	NT_STATUS_CTX_CLIENT_LICENSE_NOT_SET:                                  "CTX_CLIENT_LICENSE_NOT_SET",
	NT_STATUS_CTX_CLIENT_LICENSE_IN_USE:                                   "CTX_CLIENT_LICENSE_IN_USE",
	NT_STATUS_CTX_SHADOW_ENDED_BY_MODE_CHANGE:                             "CTX_SHADOW_ENDED_BY_MODE_CHANGE",
	NT_STATUS_CTX_SHADOW_NOT_RUNNING:                                      "CTX_SHADOW_NOT_RUNNING",
	NT_STATUS_CTX_LOGON_DISABLED:                                          "CTX_LOGON_DISABLED",
	NT_STATUS_CTX_SECURITY_LAYER_ERROR:                                    "CTX_SECURITY_LAYER_ERROR",
	NT_STATUS_TS_INCOMPATIBLE_SESSIONS:                                    "TS_INCOMPATIBLE_SESSIONS",
	NT_STATUS_MUI_FILE_NOT_FOUND:                                          "MUI_FILE_NOT_FOUND",
	NT_STATUS_MUI_INVALID_FILE:                                            "MUI_INVALID_FILE",
	NT_STATUS_MUI_INVALID_RC_CONFIG:                                       "MUI_INVALID_RC_CONFIG",
	NT_STATUS_MUI_INVALID_LOCALE_NAME:                                     "MUI_INVALID_LOCALE_NAME",
	NT_STATUS_MUI_INVALID_ULTIMATEFALLBACK_NAME:                           "MUI_INVALID_ULTIMATEFALLBACK_NAME",
	NT_STATUS_MUI_FILE_NOT_LOADED:                                         "MUI_FILE_NOT_LOADED",
	NT_STATUS_RESOURCE_ENUM_USER_STOP:                                     "RESOURCE_ENUM_USER_STOP",
	NT_STATUS_CLUSTER_INVALID_NODE:                                        "CLUSTER_INVALID_NODE",
	NT_STATUS_CLUSTER_NODE_EXISTS:                                         "CLUSTER_NODE_EXISTS",
	NT_STATUS_CLUSTER_JOIN_IN_PROGRESS:                                    "CLUSTER_JOIN_IN_PROGRESS",
	NT_STATUS_CLUSTER_NODE_NOT_FOUND:                                      "CLUSTER_NODE_NOT_FOUND",
	NT_STATUS_CLUSTER_LOCAL_NODE_NOT_FOUND:                                "CLUSTER_LOCAL_NODE_NOT_FOUND",
	NT_STATUS_CLUSTER_NETWORK_EXISTS:                                      "CLUSTER_NETWORK_EXISTS",
	NT_STATUS_CLUSTER_NETWORK_NOT_FOUND:                                   "CLUSTER_NETWORK_NOT_FOUND",
	NT_STATUS_CLUSTER_NETINTERFACE_EXISTS:                                 "CLUSTER_NETINTERFACE_EXISTS",
	NT_STATUS_CLUSTER_NETINTERFACE_NOT_FOUND:                              "CLUSTER_NETINTERFACE_NOT_FOUND",
	NT_STATUS_CLUSTER_INVALID_REQUEST:                                     "CLUSTER_INVALID_REQUEST",
	NT_STATUS_CLUSTER_INVALID_NETWORK_PROVIDER:                            "CLUSTER_INVALID_NETWORK_PROVIDER",
	NT_STATUS_CLUSTER_NODE_DOWN:                                           "CLUSTER_NODE_DOWN",
	NT_STATUS_CLUSTER_NODE_UNREACHABLE:                                    "CLUSTER_NODE_UNREACHABLE",
	NT_STATUS_CLUSTER_NODE_NOT_MEMBER:                                     "CLUSTER_NODE_NOT_MEMBER",
	NT_STATUS_CLUSTER_JOIN_NOT_IN_PROGRESS:                                "CLUSTER_JOIN_NOT_IN_PROGRESS",
	NT_STATUS_CLUSTER_INVALID_NETWORK:                                     "CLUSTER_INVALID_NETWORK",
	NT_STATUS_CLUSTER_NO_NET_ADAPTERS:                                     "CLUSTER_NO_NET_ADAPTERS",
	NT_STATUS_CLUSTER_NODE_UP:                                             "CLUSTER_NODE_UP",
	NT_STATUS_CLUSTER_NODE_PAUSED:                                         "CLUSTER_NODE_PAUSED",
	NT_STATUS_CLUSTER_NODE_NOT_PAUSED:                                     "CLUSTER_NODE_NOT_PAUSED",
	NT_STATUS_CLUSTER_NO_SECURITY_CONTEXT:                                 "CLUSTER_NO_SECURITY_CONTEXT",
	NT_STATUS_CLUSTER_NETWORK_NOT_INTERNAL:                                "CLUSTER_NETWORK_NOT_INTERNAL",
	NT_STATUS_CLUSTER_POISONED:                                            "CLUSTER_POISONED",
	NT_STATUS_ACPI_INVALID_OPCODE:                                         "ACPI_INVALID_OPCODE",
	NT_STATUS_ACPI_STACK_OVERFLOW:                                         "ACPI_STACK_OVERFLOW",
	NT_STATUS_ACPI_ASSERT_FAILED:                                          "ACPI_ASSERT_FAILED",
	NT_STATUS_ACPI_INVALID_INDEX:                                          "ACPI_INVALID_INDEX",
	NT_STATUS_ACPI_INVALID_ARGUMENT:                                       "ACPI_INVALID_ARGUMENT",
	NT_STATUS_ACPI_FATAL:                                                  "ACPI_FATAL",
	NT_STATUS_ACPI_INVALID_SUPERNAME:                                      "ACPI_INVALID_SUPERNAME",
	NT_STATUS_ACPI_INVALID_ARGTYPE:                                        "ACPI_INVALID_ARGTYPE",
	NT_STATUS_ACPI_INVALID_OBJTYPE:                                        "ACPI_INVALID_OBJTYPE",
	NT_STATUS_ACPI_INVALID_TARGETTYPE:                                     "ACPI_INVALID_TARGETTYPE",
	NT_STATUS_ACPI_INCORRECT_ARGUMENT_COUNT:                               "ACPI_INCORRECT_ARGUMENT_COUNT",
	NT_STATUS_ACPI_ADDRESS_NOT_MAPPED:                                     "ACPI_ADDRESS_NOT_MAPPED",
	NT_STATUS_ACPI_INVALID_EVENTTYPE:                                      "ACPI_INVALID_EVENTTYPE",
	NT_STATUS_ACPI_HANDLER_COLLISION:                                      "ACPI_HANDLER_COLLISION",
	NT_STATUS_ACPI_INVALID_DATA:                                           "ACPI_INVALID_DATA",
	NT_STATUS_ACPI_INVALID_REGION:                                         "ACPI_INVALID_REGION",
	NT_STATUS_ACPI_INVALID_ACCESS_SIZE:                                    "ACPI_INVALID_ACCESS_SIZE",
	NT_STATUS_ACPI_ACQUIRE_GLOBAL_LOCK:                                    "ACPI_ACQUIRE_GLOBAL_LOCK",
	NT_STATUS_ACPI_ALREADY_INITIALIZED:                                    "ACPI_ALREADY_INITIALIZED",
	NT_STATUS_ACPI_NOT_INITIALIZED:                                        "ACPI_NOT_INITIALIZED",
	NT_STATUS_ACPI_INVALID_MUTEX_LEVEL:                                    "ACPI_INVALID_MUTEX_LEVEL",
	NT_STATUS_ACPI_MUTEX_NOT_OWNED:                                        "ACPI_MUTEX_NOT_OWNED",
	NT_STATUS_ACPI_MUTEX_NOT_OWNER:                                        "ACPI_MUTEX_NOT_OWNER",
	NT_STATUS_ACPI_RS_ACCESS:                                              "ACPI_RS_ACCESS",
	NT_STATUS_ACPI_INVALID_TABLE:                                          "ACPI_INVALID_TABLE",
	NT_STATUS_ACPI_REG_HANDLER_FAILED:                                     "ACPI_REG_HANDLER_FAILED",
	NT_STATUS_ACPI_POWER_REQUEST_FAILED:                                   "ACPI_POWER_REQUEST_FAILED",
	NT_STATUS_SXS_SECTION_NOT_FOUND:                                       "SXS_SECTION_NOT_FOUND",
	NT_STATUS_SXS_CANT_GEN_ACTCTX:                                         "SXS_CANT_GEN_ACTCTX",
	NT_STATUS_SXS_INVALID_ACTCTXDATA_FORMAT:                               "SXS_INVALID_ACTCTXDATA_FORMAT",
	NT_STATUS_SXS_ASSEMBLY_NOT_FOUND:                                      "SXS_ASSEMBLY_NOT_FOUND",
	NT_STATUS_SXS_MANIFEST_FORMAT_ERROR:                                   "SXS_MANIFEST_FORMAT_ERROR",
	NT_STATUS_SXS_MANIFEST_PARSE_ERROR:                                    "SXS_MANIFEST_PARSE_ERROR",
	NT_STATUS_SXS_ACTIVATION_CONTEXT_DISABLED:                             "SXS_ACTIVATION_CONTEXT_DISABLED",
	NT_STATUS_SXS_KEY_NOT_FOUND:                                           "SXS_KEY_NOT_FOUND",
	NT_STATUS_SXS_VERSION_CONFLICT:                                        "SXS_VERSION_CONFLICT",
	NT_STATUS_SXS_WRONG_SECTION_TYPE:                                      "SXS_WRONG_SECTION_TYPE",
	NT_STATUS_SXS_THREAD_QUERIES_DISABLED:                                 "SXS_THREAD_QUERIES_DISABLED",
	NT_STATUS_SXS_ASSEMBLY_MISSING:                                        "SXS_ASSEMBLY_MISSING",
	NT_STATUS_SXS_PROCESS_DEFAULT_ALREADY_SET:                             "SXS_PROCESS_DEFAULT_ALREADY_SET",
	NT_STATUS_SXS_EARLY_DEACTIVATION:                                      "SXS_EARLY_DEACTIVATION",
	NT_STATUS_SXS_INVALID_DEACTIVATION:                                    "SXS_INVALID_DEACTIVATION",
	NT_STATUS_SXS_MULTIPLE_DEACTIVATION:                                   "SXS_MULTIPLE_DEACTIVATION",
	NT_STATUS_SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY:                 "SXS_SYSTEM_DEFAULT_ACTIVATION_CONTEXT_EMPTY",
	NT_STATUS_SXS_PROCESS_TERMINATION_REQUESTED:                           "SXS_PROCESS_TERMINATION_REQUESTED",
	NT_STATUS_SXS_CORRUPT_ACTIVATION_STACK:                                "SXS_CORRUPT_ACTIVATION_STACK",
	NT_STATUS_SXS_CORRUPTION:                                              "SXS_CORRUPTION",
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE:                        "SXS_INVALID_IDENTITY_ATTRIBUTE_VALUE",
	NT_STATUS_SXS_INVALID_IDENTITY_ATTRIBUTE_NAME:                         "SXS_INVALID_IDENTITY_ATTRIBUTE_NAME",
	NT_STATUS_SXS_IDENTITY_DUPLICATE_ATTRIBUTE:                            "SXS_IDENTITY_DUPLICATE_ATTRIBUTE",
	NT_STATUS_SXS_IDENTITY_PARSE_ERROR:                                    "SXS_IDENTITY_PARSE_ERROR",
	NT_STATUS_SXS_COMPONENT_STORE_CORRUPT:                                 "SXS_COMPONENT_STORE_CORRUPT",
	NT_STATUS_SXS_FILE_HASH_MISMATCH:                                      "SXS_FILE_HASH_MISMATCH",
	NT_STATUS_SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT:           "SXS_MANIFEST_IDENTITY_SAME_BUT_CONTENTS_DIFFERENT",
	NT_STATUS_SXS_IDENTITIES_DIFFERENT:                                    "SXS_IDENTITIES_DIFFERENT",
	NT_STATUS_SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT:                            "SXS_ASSEMBLY_IS_NOT_A_DEPLOYMENT",
	NT_STATUS_SXS_FILE_NOT_PART_OF_ASSEMBLY:                               "SXS_FILE_NOT_PART_OF_ASSEMBLY",
	NT_STATUS_ADVANCED_INSTALLER_FAILED:                                   "ADVANCED_INSTALLER_FAILED",
	NT_STATUS_XML_ENCODING_MISMATCH:                                       "XML_ENCODING_MISMATCH",
	NT_STATUS_SXS_MANIFEST_TOO_BIG:                                        "SXS_MANIFEST_TOO_BIG",
	NT_STATUS_SXS_SETTING_NOT_REGISTERED:                                  "SXS_SETTING_NOT_REGISTERED",
	NT_STATUS_SXS_TRANSACTION_CLOSURE_INCOMPLETE:                          "SXS_TRANSACTION_CLOSURE_INCOMPLETE",
	NT_STATUS_SMI_PRIMITIVE_INSTALLER_FAILED:                              "SMI_PRIMITIVE_INSTALLER_FAILED",
	NT_STATUS_GENERIC_COMMAND_FAILED:                                      "GENERIC_COMMAND_FAILED",
	NT_STATUS_SXS_FILE_HASH_MISSING:                                       "SXS_FILE_HASH_MISSING",
	NT_STATUS_TRANSACTIONAL_CONFLICT:                                      "TRANSACTIONAL_CONFLICT",
	NT_STATUS_INVALID_TRANSACTION:                                         "INVALID_TRANSACTION",
	NT_STATUS_TRANSACTION_NOT_ACTIVE:                                      "TRANSACTION_NOT_ACTIVE",
	NT_STATUS_TM_INITIALIZATION_FAILED:                                    "TM_INITIALIZATION_FAILED",
	NT_STATUS_RM_NOT_ACTIVE:                                               "RM_NOT_ACTIVE",
	NT_STATUS_RM_METADATA_CORRUPT:                                         "RM_METADATA_CORRUPT",
	NT_STATUS_TRANSACTION_NOT_JOINED:                                      "TRANSACTION_NOT_JOINED",
	NT_STATUS_DIRECTORY_NOT_RM:                                            "DIRECTORY_NOT_RM",
	NT_STATUS_TRANSACTIONS_UNSUPPORTED_REMOTE:                             "TRANSACTIONS_UNSUPPORTED_REMOTE",
	NT_STATUS_LOG_RESIZE_INVALID_SIZE:                                     "LOG_RESIZE_INVALID_SIZE",
	NT_STATUS_REMOTE_FILE_VERSION_MISMATCH:                                "REMOTE_FILE_VERSION_MISMATCH",
	NT_STATUS_CRM_PROTOCOL_ALREADY_EXISTS:                                 "CRM_PROTOCOL_ALREADY_EXISTS",
	NT_STATUS_TRANSACTION_PROPAGATION_FAILED:                              "TRANSACTION_PROPAGATION_FAILED",
	NT_STATUS_CRM_PROTOCOL_NOT_FOUND:                                      "CRM_PROTOCOL_NOT_FOUND",
	NT_STATUS_TRANSACTION_SUPERIOR_EXISTS:                                 "TRANSACTION_SUPERIOR_EXISTS",
	NT_STATUS_TRANSACTION_REQUEST_NOT_VALID:                               "TRANSACTION_REQUEST_NOT_VALID",
	NT_STATUS_TRANSACTION_NOT_REQUESTED:                                   "TRANSACTION_NOT_REQUESTED",
	NT_STATUS_TRANSACTION_ALREADY_ABORTED:                                 "TRANSACTION_ALREADY_ABORTED",
	NT_STATUS_TRANSACTION_ALREADY_COMMITTED:                               "TRANSACTION_ALREADY_COMMITTED",
	NT_STATUS_TRANSACTION_INVALID_MARSHALL_BUFFER:                         "TRANSACTION_INVALID_MARSHALL_BUFFER",
	NT_STATUS_CURRENT_TRANSACTION_NOT_VALID:                               "CURRENT_TRANSACTION_NOT_VALID",
	NT_STATUS_LOG_GROWTH_FAILED:                                           "LOG_GROWTH_FAILED",
	NT_STATUS_OBJECT_NO_LONGER_EXISTS:                                     "OBJECT_NO_LONGER_EXISTS",
	NT_STATUS_STREAM_MINIVERSION_NOT_FOUND:                                "STREAM_MINIVERSION_NOT_FOUND",
	NT_STATUS_STREAM_MINIVERSION_NOT_VALID:                                "STREAM_MINIVERSION_NOT_VALID",
	NT_STATUS_MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION:         "MINIVERSION_INACCESSIBLE_FROM_SPECIFIED_TRANSACTION",
	NT_STATUS_CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT:                    "CANT_OPEN_MINIVERSION_WITH_MODIFY_INTENT",
	NT_STATUS_CANT_CREATE_MORE_STREAM_MINIVERSIONS:                        "CANT_CREATE_MORE_STREAM_MINIVERSIONS",
	NT_STATUS_HANDLE_NO_LONGER_VALID:                                      "HANDLE_NO_LONGER_VALID",
	NT_STATUS_LOG_CORRUPTION_DETECTED:                                     "LOG_CORRUPTION_DETECTED",
	NT_STATUS_RM_DISCONNECTED:                                             "RM_DISCONNECTED",
	NT_STATUS_ENLISTMENT_NOT_SUPERIOR:                                     "ENLISTMENT_NOT_SUPERIOR",
	NT_STATUS_FILE_IDENTITY_NOT_PERSISTENT:                                "FILE_IDENTITY_NOT_PERSISTENT",
	NT_STATUS_CANT_BREAK_TRANSACTIONAL_DEPENDENCY:                         "CANT_BREAK_TRANSACTIONAL_DEPENDENCY",
	NT_STATUS_CANT_CROSS_RM_BOUNDARY:                                      "CANT_CROSS_RM_BOUNDARY",
	NT_STATUS_TXF_DIR_NOT_EMPTY:                                           "TXF_DIR_NOT_EMPTY",
	NT_STATUS_INDOUBT_TRANSACTIONS_EXIST:                                  "INDOUBT_TRANSACTIONS_EXIST",
	NT_STATUS_TM_VOLATILE:                                                 "TM_VOLATILE",
	NT_STATUS_ROLLBACK_TIMER_EXPIRED:                                      "ROLLBACK_TIMER_EXPIRED",
	NT_STATUS_TXF_ATTRIBUTE_CORRUPT:                                       "TXF_ATTRIBUTE_CORRUPT",
	NT_STATUS_EFS_NOT_ALLOWED_IN_TRANSACTION:                              "EFS_NOT_ALLOWED_IN_TRANSACTION",
	NT_STATUS_TRANSACTIONAL_OPEN_NOT_ALLOWED:                              "TRANSACTIONAL_OPEN_NOT_ALLOWED",
	NT_STATUS_TRANSACTED_MAPPING_UNSUPPORTED_REMOTE:                       "TRANSACTED_MAPPING_UNSUPPORTED_REMOTE",
	NT_STATUS_TRANSACTION_REQUIRED_PROMOTION:                              "TRANSACTION_REQUIRED_PROMOTION",
	NT_STATUS_CANNOT_EXECUTE_FILE_IN_TRANSACTION:                          "CANNOT_EXECUTE_FILE_IN_TRANSACTION",
	NT_STATUS_TRANSACTIONS_NOT_FROZEN:                                     "TRANSACTIONS_NOT_FROZEN",
	NT_STATUS_TRANSACTION_FREEZE_IN_PROGRESS:                              "TRANSACTION_FREEZE_IN_PROGRESS",
	NT_STATUS_NOT_SNAPSHOT_VOLUME:                                         "NOT_SNAPSHOT_VOLUME",
	NT_STATUS_NO_SAVEPOINT_WITH_OPEN_FILES:                                "NO_SAVEPOINT_WITH_OPEN_FILES",
	NT_STATUS_SPARSE_NOT_ALLOWED_IN_TRANSACTION:                           "SPARSE_NOT_ALLOWED_IN_TRANSACTION",
	NT_STATUS_TM_IDENTITY_MISMATCH:                                        "TM_IDENTITY_MISMATCH",
	NT_STATUS_FLOATED_SECTION:                                             "FLOATED_SECTION",
	NT_STATUS_CANNOT_ACCEPT_TRANSACTED_WORK:                               "CANNOT_ACCEPT_TRANSACTED_WORK",
	NT_STATUS_CANNOT_ABORT_TRANSACTIONS:                                   "CANNOT_ABORT_TRANSACTIONS",
	NT_STATUS_TRANSACTION_NOT_FOUND:                                       "TRANSACTION_NOT_FOUND",
	NT_STATUS_RESOURCEMANAGER_NOT_FOUND:                                   "RESOURCEMANAGER_NOT_FOUND",
	NT_STATUS_ENLISTMENT_NOT_FOUND:                                        "ENLISTMENT_NOT_FOUND",
	NT_STATUS_TRANSACTIONMANAGER_NOT_FOUND:                                "TRANSACTIONMANAGER_NOT_FOUND",
	NT_STATUS_TRANSACTIONMANAGER_NOT_ONLINE:                               "TRANSACTIONMANAGER_NOT_ONLINE",
	NT_STATUS_TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION:                  "TRANSACTIONMANAGER_RECOVERY_NAME_COLLISION",
	NT_STATUS_TRANSACTION_NOT_ROOT:                                        "TRANSACTION_NOT_ROOT",
	NT_STATUS_TRANSACTION_OBJECT_EXPIRED:                                  "TRANSACTION_OBJECT_EXPIRED",
	NT_STATUS_COMPRESSION_NOT_ALLOWED_IN_TRANSACTION:                      "COMPRESSION_NOT_ALLOWED_IN_TRANSACTION",
	NT_STATUS_TRANSACTION_RESPONSE_NOT_ENLISTED:                           "TRANSACTION_RESPONSE_NOT_ENLISTED",
	NT_STATUS_TRANSACTION_RECORD_TOO_LONG:                                 "TRANSACTION_RECORD_TOO_LONG",
	NT_STATUS_NO_LINK_TRACKING_IN_TRANSACTION:                             "NO_LINK_TRACKING_IN_TRANSACTION",
	NT_STATUS_OPERATION_NOT_SUPPORTED_IN_TRANSACTION:                      "OPERATION_NOT_SUPPORTED_IN_TRANSACTION",
	NT_STATUS_TRANSACTION_INTEGRITY_VIOLATED:                              "TRANSACTION_INTEGRITY_VIOLATED",
	NT_STATUS_EXPIRED_HANDLE:                                              "EXPIRED_HANDLE",
	NT_STATUS_TRANSACTION_NOT_ENLISTED:                                    "TRANSACTION_NOT_ENLISTED",
	NT_STATUS_LOG_SECTOR_INVALID:                                          "LOG_SECTOR_INVALID",
	NT_STATUS_LOG_SECTOR_PARITY_INVALID:                                   "LOG_SECTOR_PARITY_INVALID",
	NT_STATUS_LOG_SECTOR_REMAPPED:                                         "LOG_SECTOR_REMAPPED",
	NT_STATUS_LOG_BLOCK_INCOMPLETE:                                        "LOG_BLOCK_INCOMPLETE",
	NT_STATUS_LOG_INVALID_RANGE:                                           "LOG_INVALID_RANGE",
	NT_STATUS_LOG_BLOCKS_EXHAUSTED:                                        "LOG_BLOCKS_EXHAUSTED",
	NT_STATUS_LOG_READ_CONTEXT_INVALID:                                    "LOG_READ_CONTEXT_INVALID",
	NT_STATUS_LOG_RESTART_INVALID:                                         "LOG_RESTART_INVALID",
	NT_STATUS_LOG_BLOCK_VERSION:                                           "LOG_BLOCK_VERSION",
	NT_STATUS_LOG_BLOCK_INVALID:                                           "LOG_BLOCK_INVALID",
	NT_STATUS_LOG_READ_MODE_INVALID:                                       "LOG_READ_MODE_INVALID",
	NT_STATUS_LOG_METADATA_CORRUPT:                                        "LOG_METADATA_CORRUPT",
	NT_STATUS_LOG_METADATA_INVALID:                                        "LOG_METADATA_INVALID",
	NT_STATUS_LOG_METADATA_INCONSISTENT:                                   "LOG_METADATA_INCONSISTENT",
	NT_STATUS_LOG_RESERVATION_INVALID:                                     "LOG_RESERVATION_INVALID",
	NT_STATUS_LOG_CANT_DELETE:                                             "LOG_CANT_DELETE",
	NT_STATUS_LOG_CONTAINER_LIMIT_EXCEEDED:                                "LOG_CONTAINER_LIMIT_EXCEEDED",
	NT_STATUS_LOG_START_OF_LOG:                                            "LOG_START_OF_LOG",
	NT_STATUS_LOG_POLICY_ALREADY_INSTALLED:                                "LOG_POLICY_ALREADY_INSTALLED",
	NT_STATUS_LOG_POLICY_NOT_INSTALLED:                                    "LOG_POLICY_NOT_INSTALLED",
	NT_STATUS_LOG_POLICY_INVALID:                                          "LOG_POLICY_INVALID",
	NT_STATUS_LOG_POLICY_CONFLICT:                                         "LOG_POLICY_CONFLICT",
	NT_STATUS_LOG_PINNED_ARCHIVE_TAIL:                                     "LOG_PINNED_ARCHIVE_TAIL",
	NT_STATUS_LOG_RECORD_NONEXISTENT:                                      "LOG_RECORD_NONEXISTENT",
	NT_STATUS_LOG_RECORDS_RESERVED_INVALID:                                "LOG_RECORDS_RESERVED_INVALID",
	NT_STATUS_LOG_SPACE_RESERVED_INVALID:                                  "LOG_SPACE_RESERVED_INVALID",
	NT_STATUS_LOG_TAIL_INVALID:                                            "LOG_TAIL_INVALID",
	NT_STATUS_LOG_FULL:                                                    "LOG_FULL",
	NT_STATUS_LOG_MULTIPLEXED:                                             "LOG_MULTIPLEXED",
	NT_STATUS_LOG_DEDICATED:                                               "LOG_DEDICATED",
	NT_STATUS_LOG_ARCHIVE_NOT_IN_PROGRESS:                                 "LOG_ARCHIVE_NOT_IN_PROGRESS",
	NT_STATUS_LOG_ARCHIVE_IN_PROGRESS:                                     "LOG_ARCHIVE_IN_PROGRESS",
	NT_STATUS_LOG_EPHEMERAL:                                               "LOG_EPHEMERAL",
	NT_STATUS_LOG_NOT_ENOUGH_CONTAINERS:                                   "LOG_NOT_ENOUGH_CONTAINERS",
	NT_STATUS_LOG_CLIENT_ALREADY_REGISTERED:                               "LOG_CLIENT_ALREADY_REGISTERED",
	NT_STATUS_LOG_CLIENT_NOT_REGISTERED:                                   "LOG_CLIENT_NOT_REGISTERED",
	NT_STATUS_LOG_FULL_HANDLER_IN_PROGRESS:                                "LOG_FULL_HANDLER_IN_PROGRESS",
	NT_STATUS_LOG_CONTAINER_READ_FAILED:                                   "LOG_CONTAINER_READ_FAILED",
	NT_STATUS_LOG_CONTAINER_WRITE_FAILED:                                  "LOG_CONTAINER_WRITE_FAILED",
	NT_STATUS_LOG_CONTAINER_OPEN_FAILED:                                   "LOG_CONTAINER_OPEN_FAILED",
	NT_STATUS_LOG_CONTAINER_STATE_INVALID:                                 "LOG_CONTAINER_STATE_INVALID",
	NT_STATUS_LOG_STATE_INVALID:                                           "LOG_STATE_INVALID",
	NT_STATUS_LOG_PINNED:                                                  "LOG_PINNED",
	NT_STATUS_LOG_METADATA_FLUSH_FAILED:                                   "LOG_METADATA_FLUSH_FAILED",
	NT_STATUS_LOG_INCONSISTENT_SECURITY:                                   "LOG_INCONSISTENT_SECURITY",
	NT_STATUS_LOG_APPENDED_FLUSH_FAILED:                                   "LOG_APPENDED_FLUSH_FAILED",
	NT_STATUS_LOG_PINNED_RESERVATION:                                      "LOG_PINNED_RESERVATION",
	NT_STATUS_VIDEO_HUNG_DISPLAY_DRIVER_THREAD:                            "VIDEO_HUNG_DISPLAY_DRIVER_THREAD",
	NT_STATUS_FLT_NO_HANDLER_DEFINED:                                      "FLT_NO_HANDLER_DEFINED",
	NT_STATUS_FLT_CONTEXT_ALREADY_DEFINED:                                 "FLT_CONTEXT_ALREADY_DEFINED",
	NT_STATUS_FLT_INVALID_ASYNCHRONOUS_REQUEST:                            "FLT_INVALID_ASYNCHRONOUS_REQUEST",
	NT_STATUS_FLT_DISALLOW_FAST_IO:                                        "FLT_DISALLOW_FAST_IO",
	NT_STATUS_FLT_INVALID_NAME_REQUEST:                                    "FLT_INVALID_NAME_REQUEST",
	NT_STATUS_FLT_NOT_SAFE_TO_POST_OPERATION:                              "FLT_NOT_SAFE_TO_POST_OPERATION",
	NT_STATUS_FLT_NOT_INITIALIZED:                                         "FLT_NOT_INITIALIZED",
	NT_STATUS_FLT_FILTER_NOT_READY:                                        "FLT_FILTER_NOT_READY",
	NT_STATUS_FLT_POST_OPERATION_CLEANUP:                                  "FLT_POST_OPERATION_CLEANUP",
	NT_STATUS_FLT_INTERNAL_ERROR:                                          "FLT_INTERNAL_ERROR",
	NT_STATUS_FLT_DELETING_OBJECT:                                         "FLT_DELETING_OBJECT",
	NT_STATUS_FLT_MUST_BE_NONPAGED_POOL:                                   "FLT_MUST_BE_NONPAGED_POOL",
	NT_STATUS_FLT_DUPLICATE_ENTRY:                                         "FLT_DUPLICATE_ENTRY",
	NT_STATUS_FLT_CBDQ_DISABLED:                                           "FLT_CBDQ_DISABLED",
	NT_STATUS_FLT_DO_NOT_ATTACH:                                           "FLT_DO_NOT_ATTACH",
	NT_STATUS_FLT_DO_NOT_DETACH:                                           "FLT_DO_NOT_DETACH",
	NT_STATUS_FLT_INSTANCE_ALTITUDE_COLLISION:                             "FLT_INSTANCE_ALTITUDE_COLLISION",
	NT_STATUS_FLT_INSTANCE_NAME_COLLISION:                                 "FLT_INSTANCE_NAME_COLLISION",
	NT_STATUS_FLT_FILTER_NOT_FOUND:                                        "FLT_FILTER_NOT_FOUND",
	NT_STATUS_FLT_VOLUME_NOT_FOUND:                                        "FLT_VOLUME_NOT_FOUND",
	NT_STATUS_FLT_INSTANCE_NOT_FOUND:                                      "FLT_INSTANCE_NOT_FOUND",
	NT_STATUS_FLT_CONTEXT_ALLOCATION_NOT_FOUND:                            "FLT_CONTEXT_ALLOCATION_NOT_FOUND",
	NT_STATUS_FLT_INVALID_CONTEXT_REGISTRATION:                            "FLT_INVALID_CONTEXT_REGISTRATION",
	NT_STATUS_FLT_NAME_CACHE_MISS:                                         "FLT_NAME_CACHE_MISS",
	NT_STATUS_FLT_NO_DEVICE_OBJECT:                                        "FLT_NO_DEVICE_OBJECT",
	NT_STATUS_FLT_VOLUME_ALREADY_MOUNTED:                                  "FLT_VOLUME_ALREADY_MOUNTED",
	NT_STATUS_FLT_ALREADY_ENLISTED:                                        "FLT_ALREADY_ENLISTED",
	NT_STATUS_FLT_CONTEXT_ALREADY_LINKED:                                  "FLT_CONTEXT_ALREADY_LINKED",
	NT_STATUS_FLT_NO_WAITER_FOR_REPLY:                                     "FLT_NO_WAITER_FOR_REPLY",
	NT_STATUS_MONITOR_NO_DESCRIPTOR:                                       "MONITOR_NO_DESCRIPTOR",
	NT_STATUS_MONITOR_UNKNOWN_DESCRIPTOR_FORMAT:                           "MONITOR_UNKNOWN_DESCRIPTOR_FORMAT",
	NT_STATUS_MONITOR_INVALID_DESCRIPTOR_CHECKSUM:                         "MONITOR_INVALID_DESCRIPTOR_CHECKSUM",
	NT_STATUS_MONITOR_INVALID_STANDARD_TIMING_BLOCK:                       "MONITOR_INVALID_STANDARD_TIMING_BLOCK",
	NT_STATUS_MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED:                   "MONITOR_WMI_DATABLOCK_REGISTRATION_FAILED",
	NT_STATUS_MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK:                  "MONITOR_INVALID_SERIAL_NUMBER_MONDSC_BLOCK",
	NT_STATUS_MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK:                  "MONITOR_INVALID_USER_FRIENDLY_MONDSC_BLOCK",
	NT_STATUS_MONITOR_NO_MORE_DESCRIPTOR_DATA:                             "MONITOR_NO_MORE_DESCRIPTOR_DATA",
	NT_STATUS_MONITOR_INVALID_DETAILED_TIMING_BLOCK:                       "MONITOR_INVALID_DETAILED_TIMING_BLOCK",
	NT_STATUS_MONITOR_INVALID_MANUFACTURE_DATE:                            "MONITOR_INVALID_MANUFACTURE_DATE",
	NT_STATUS_GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER:                           "GRAPHICS_NOT_EXCLUSIVE_MODE_OWNER",
	NT_STATUS_GRAPHICS_INSUFFICIENT_DMA_BUFFER:                            "GRAPHICS_INSUFFICIENT_DMA_BUFFER",
	NT_STATUS_GRAPHICS_INVALID_DISPLAY_ADAPTER:                            "GRAPHICS_INVALID_DISPLAY_ADAPTER",
	NT_STATUS_GRAPHICS_ADAPTER_WAS_RESET:                                  "GRAPHICS_ADAPTER_WAS_RESET",
	NT_STATUS_GRAPHICS_INVALID_DRIVER_MODEL:                               "GRAPHICS_INVALID_DRIVER_MODEL",
	NT_STATUS_GRAPHICS_PRESENT_MODE_CHANGED:                               "GRAPHICS_PRESENT_MODE_CHANGED",
	NT_STATUS_GRAPHICS_PRESENT_OCCLUDED:                                   "GRAPHICS_PRESENT_OCCLUDED",
	NT_STATUS_GRAPHICS_PRESENT_DENIED:                                     "GRAPHICS_PRESENT_DENIED",
	NT_STATUS_GRAPHICS_CANNOTCOLORCONVERT:                                 "GRAPHICS_CANNOTCOLORCONVERT",
	NT_STATUS_GRAPHICS_PRESENT_REDIRECTION_DISABLED:                       "GRAPHICS_PRESENT_REDIRECTION_DISABLED",
	NT_STATUS_GRAPHICS_PRESENT_UNOCCLUDED:                                 "GRAPHICS_PRESENT_UNOCCLUDED",
	NT_STATUS_GRAPHICS_NO_VIDEO_MEMORY:                                    "GRAPHICS_NO_VIDEO_MEMORY",
	NT_STATUS_GRAPHICS_CANT_LOCK_MEMORY:                                   "GRAPHICS_CANT_LOCK_MEMORY",
	NT_STATUS_GRAPHICS_ALLOCATION_BUSY:                                    "GRAPHICS_ALLOCATION_BUSY",
	NT_STATUS_GRAPHICS_TOO_MANY_REFERENCES:                                "GRAPHICS_TOO_MANY_REFERENCES",
	NT_STATUS_GRAPHICS_TRY_AGAIN_LATER:                                    "GRAPHICS_TRY_AGAIN_LATER",
	NT_STATUS_GRAPHICS_TRY_AGAIN_NOW:                                      "GRAPHICS_TRY_AGAIN_NOW",
	NT_STATUS_GRAPHICS_ALLOCATION_INVALID:                                 "GRAPHICS_ALLOCATION_INVALID",
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE:                   "GRAPHICS_UNSWIZZLING_APERTURE_UNAVAILABLE",
	NT_STATUS_GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED:                   "GRAPHICS_UNSWIZZLING_APERTURE_UNSUPPORTED",
	NT_STATUS_GRAPHICS_CANT_EVICT_PINNED_ALLOCATION:                       "GRAPHICS_CANT_EVICT_PINNED_ALLOCATION",
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_USAGE:                           "GRAPHICS_INVALID_ALLOCATION_USAGE",
	NT_STATUS_GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION:                      "GRAPHICS_CANT_RENDER_LOCKED_ALLOCATION",
	NT_STATUS_GRAPHICS_ALLOCATION_CLOSED:                                  "GRAPHICS_ALLOCATION_CLOSED",
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_INSTANCE:                        "GRAPHICS_INVALID_ALLOCATION_INSTANCE",
	NT_STATUS_GRAPHICS_INVALID_ALLOCATION_HANDLE:                          "GRAPHICS_INVALID_ALLOCATION_HANDLE",
	NT_STATUS_GRAPHICS_WRONG_ALLOCATION_DEVICE:                            "GRAPHICS_WRONG_ALLOCATION_DEVICE",
	NT_STATUS_GRAPHICS_ALLOCATION_CONTENT_LOST:                            "GRAPHICS_ALLOCATION_CONTENT_LOST",
	NT_STATUS_GRAPHICS_GPU_EXCEPTION_ON_DEVICE:                            "GRAPHICS_GPU_EXCEPTION_ON_DEVICE",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY:                             "GRAPHICS_INVALID_VIDPN_TOPOLOGY",
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED:                       "GRAPHICS_VIDPN_TOPOLOGY_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED:             "GRAPHICS_VIDPN_TOPOLOGY_CURRENTLY_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_INVALID_VIDPN:                                      "GRAPHICS_INVALID_VIDPN",
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE:                       "GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE",
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET:                       "GRAPHICS_INVALID_VIDEO_PRESENT_TARGET",
	NT_STATUS_GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED:                       "GRAPHICS_VIDPN_MODALITY_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_SOURCEMODESET:                        "GRAPHICS_INVALID_VIDPN_SOURCEMODESET",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGETMODESET:                        "GRAPHICS_INVALID_VIDPN_TARGETMODESET",
	NT_STATUS_GRAPHICS_INVALID_FREQUENCY:                                  "GRAPHICS_INVALID_FREQUENCY",
	NT_STATUS_GRAPHICS_INVALID_ACTIVE_REGION:                              "GRAPHICS_INVALID_ACTIVE_REGION",
	NT_STATUS_GRAPHICS_INVALID_TOTAL_REGION:                               "GRAPHICS_INVALID_TOTAL_REGION",
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE:                  "GRAPHICS_INVALID_VIDEO_PRESENT_SOURCE_MODE",
	NT_STATUS_GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE:                  "GRAPHICS_INVALID_VIDEO_PRESENT_TARGET_MODE",
	NT_STATUS_GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET:                     "GRAPHICS_PINNED_MODE_MUST_REMAIN_IN_SET",
	NT_STATUS_GRAPHICS_PATH_ALREADY_IN_TOPOLOGY:                           "GRAPHICS_PATH_ALREADY_IN_TOPOLOGY",
	NT_STATUS_GRAPHICS_MODE_ALREADY_IN_MODESET:                            "GRAPHICS_MODE_ALREADY_IN_MODESET",
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTSOURCESET:                      "GRAPHICS_INVALID_VIDEOPRESENTSOURCESET",
	NT_STATUS_GRAPHICS_INVALID_VIDEOPRESENTTARGETSET:                      "GRAPHICS_INVALID_VIDEOPRESENTTARGETSET",
	NT_STATUS_GRAPHICS_SOURCE_ALREADY_IN_SET:                              "GRAPHICS_SOURCE_ALREADY_IN_SET",
	NT_STATUS_GRAPHICS_TARGET_ALREADY_IN_SET:                              "GRAPHICS_TARGET_ALREADY_IN_SET",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_PRESENT_PATH:                         "GRAPHICS_INVALID_VIDPN_PRESENT_PATH",
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY:                      "GRAPHICS_NO_RECOMMENDED_VIDPN_TOPOLOGY",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET:                  "GRAPHICS_INVALID_MONITOR_FREQUENCYRANGESET",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE:                     "GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE",
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_NOT_IN_SET:                          "GRAPHICS_FREQUENCYRANGE_NOT_IN_SET",
	NT_STATUS_GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET:                      "GRAPHICS_FREQUENCYRANGE_ALREADY_IN_SET",
	NT_STATUS_GRAPHICS_STALE_MODESET:                                      "GRAPHICS_STALE_MODESET",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCEMODESET:                      "GRAPHICS_INVALID_MONITOR_SOURCEMODESET",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_SOURCE_MODE:                        "GRAPHICS_INVALID_MONITOR_SOURCE_MODE",
	NT_STATUS_GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN:                    "GRAPHICS_NO_RECOMMENDED_FUNCTIONAL_VIDPN",
	NT_STATUS_GRAPHICS_MODE_ID_MUST_BE_UNIQUE:                             "GRAPHICS_MODE_ID_MUST_BE_UNIQUE",
	NT_STATUS_GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION:    "GRAPHICS_EMPTY_ADAPTER_MONITOR_MODE_SUPPORT_INTERSECTION",
	NT_STATUS_GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES:            "GRAPHICS_VIDEO_PRESENT_TARGETS_LESS_THAN_SOURCES",
	NT_STATUS_GRAPHICS_PATH_NOT_IN_TOPOLOGY:                               "GRAPHICS_PATH_NOT_IN_TOPOLOGY",
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE:              "GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_SOURCE",
	NT_STATUS_GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET:              "GRAPHICS_ADAPTER_MUST_HAVE_AT_LEAST_ONE_TARGET",
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTORSET:                       "GRAPHICS_INVALID_MONITORDESCRIPTORSET",
	NT_STATUS_GRAPHICS_INVALID_MONITORDESCRIPTOR:                          "GRAPHICS_INVALID_MONITORDESCRIPTOR",
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET:                       "GRAPHICS_MONITORDESCRIPTOR_NOT_IN_SET",
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET:                   "GRAPHICS_MONITORDESCRIPTOR_ALREADY_IN_SET",
	NT_STATUS_GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE:                "GRAPHICS_MONITORDESCRIPTOR_ID_MUST_BE_UNIQUE",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE:                   "GRAPHICS_INVALID_VIDPN_TARGET_SUBSET_TYPE",
	NT_STATUS_GRAPHICS_RESOURCES_NOT_RELATED:                              "GRAPHICS_RESOURCES_NOT_RELATED",
	NT_STATUS_GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE:                           "GRAPHICS_SOURCE_ID_MUST_BE_UNIQUE",
	NT_STATUS_GRAPHICS_TARGET_ID_MUST_BE_UNIQUE:                           "GRAPHICS_TARGET_ID_MUST_BE_UNIQUE",
	NT_STATUS_GRAPHICS_NO_AVAILABLE_VIDPN_TARGET:                          "GRAPHICS_NO_AVAILABLE_VIDPN_TARGET",
	NT_STATUS_GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER:       "GRAPHICS_MONITOR_COULD_NOT_BE_ASSOCIATED_WITH_ADAPTER",
	NT_STATUS_GRAPHICS_NO_VIDPNMGR:                                        "GRAPHICS_NO_VIDPNMGR",
	NT_STATUS_GRAPHICS_NO_ACTIVE_VIDPN:                                    "GRAPHICS_NO_ACTIVE_VIDPN",
	NT_STATUS_GRAPHICS_STALE_VIDPN_TOPOLOGY:                               "GRAPHICS_STALE_VIDPN_TOPOLOGY",
	NT_STATUS_GRAPHICS_MONITOR_NOT_CONNECTED:                              "GRAPHICS_MONITOR_NOT_CONNECTED",
	NT_STATUS_GRAPHICS_SOURCE_NOT_IN_TOPOLOGY:                             "GRAPHICS_SOURCE_NOT_IN_TOPOLOGY",
	NT_STATUS_GRAPHICS_INVALID_PRIMARYSURFACE_SIZE:                        "GRAPHICS_INVALID_PRIMARYSURFACE_SIZE",
	NT_STATUS_GRAPHICS_INVALID_VISIBLEREGION_SIZE:                         "GRAPHICS_INVALID_VISIBLEREGION_SIZE",
	NT_STATUS_GRAPHICS_INVALID_STRIDE:                                     "GRAPHICS_INVALID_STRIDE",
	NT_STATUS_GRAPHICS_INVALID_PIXELFORMAT:                                "GRAPHICS_INVALID_PIXELFORMAT",
	NT_STATUS_GRAPHICS_INVALID_COLORBASIS:                                 "GRAPHICS_INVALID_COLORBASIS",
	NT_STATUS_GRAPHICS_INVALID_PIXELVALUEACCESSMODE:                       "GRAPHICS_INVALID_PIXELVALUEACCESSMODE",
	NT_STATUS_GRAPHICS_TARGET_NOT_IN_TOPOLOGY:                             "GRAPHICS_TARGET_NOT_IN_TOPOLOGY",
	NT_STATUS_GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT:                 "GRAPHICS_NO_DISPLAY_MODE_MANAGEMENT_SUPPORT",
	NT_STATUS_GRAPHICS_VIDPN_SOURCE_IN_USE:                                "GRAPHICS_VIDPN_SOURCE_IN_USE",
	NT_STATUS_GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN:                           "GRAPHICS_CANT_ACCESS_ACTIVE_VIDPN",
	NT_STATUS_GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL:                    "GRAPHICS_INVALID_PATH_IMPORTANCE_ORDINAL",
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION:       "GRAPHICS_INVALID_PATH_CONTENT_GEOMETRY_TRANSFORMATION",
	NT_STATUS_GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED: "GRAPHICS_PATH_CONTENT_GEOMETRY_TRANSFORMATION_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_INVALID_GAMMA_RAMP:                                 "GRAPHICS_INVALID_GAMMA_RAMP",
	NT_STATUS_GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED:                           "GRAPHICS_GAMMA_RAMP_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_MULTISAMPLING_NOT_SUPPORTED:                        "GRAPHICS_MULTISAMPLING_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_MODE_NOT_IN_MODESET:                                "GRAPHICS_MODE_NOT_IN_MODESET",
	NT_STATUS_GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON:       "GRAPHICS_INVALID_VIDPN_TOPOLOGY_RECOMMENDATION_REASON",
	NT_STATUS_GRAPHICS_INVALID_PATH_CONTENT_TYPE:                          "GRAPHICS_INVALID_PATH_CONTENT_TYPE",
	NT_STATUS_GRAPHICS_INVALID_COPYPROTECTION_TYPE:                        "GRAPHICS_INVALID_COPYPROTECTION_TYPE",
	NT_STATUS_GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS:                  "GRAPHICS_UNASSIGNED_MODESET_ALREADY_EXISTS",
	NT_STATUS_GRAPHICS_INVALID_SCANLINE_ORDERING:                          "GRAPHICS_INVALID_SCANLINE_ORDERING",
	NT_STATUS_GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED:                       "GRAPHICS_TOPOLOGY_CHANGES_NOT_ALLOWED",
	NT_STATUS_GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS:                   "GRAPHICS_NO_AVAILABLE_IMPORTANCE_ORDINALS",
	NT_STATUS_GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT:                        "GRAPHICS_INCOMPATIBLE_PRIVATE_FORMAT",
	NT_STATUS_GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM:                     "GRAPHICS_INVALID_MODE_PRUNING_ALGORITHM",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN:                  "GRAPHICS_INVALID_MONITOR_CAPABILITY_ORIGIN",
	NT_STATUS_GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT:          "GRAPHICS_INVALID_MONITOR_FREQUENCYRANGE_CONSTRAINT",
	NT_STATUS_GRAPHICS_MAX_NUM_PATHS_REACHED:                              "GRAPHICS_MAX_NUM_PATHS_REACHED",
	NT_STATUS_GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION:                 "GRAPHICS_CANCEL_VIDPN_TOPOLOGY_AUGMENTATION",
	NT_STATUS_GRAPHICS_INVALID_CLIENT_TYPE:                                "GRAPHICS_INVALID_CLIENT_TYPE",
	NT_STATUS_GRAPHICS_CLIENTVIDPN_NOT_SET:                                "GRAPHICS_CLIENTVIDPN_NOT_SET",
	NT_STATUS_GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED:                  "GRAPHICS_SPECIFIED_CHILD_ALREADY_CONNECTED",
	NT_STATUS_GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED:                     "GRAPHICS_CHILD_DESCRIPTOR_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_NOT_A_LINKED_ADAPTER:                               "GRAPHICS_NOT_A_LINKED_ADAPTER",
	NT_STATUS_GRAPHICS_LEADLINK_NOT_ENUMERATED:                            "GRAPHICS_LEADLINK_NOT_ENUMERATED",
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_ENUMERATED:                          "GRAPHICS_CHAINLINKS_NOT_ENUMERATED",
	NT_STATUS_GRAPHICS_ADAPTER_CHAIN_NOT_READY:                            "GRAPHICS_ADAPTER_CHAIN_NOT_READY",
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_STARTED:                             "GRAPHICS_CHAINLINKS_NOT_STARTED",
	NT_STATUS_GRAPHICS_CHAINLINKS_NOT_POWERED_ON:                          "GRAPHICS_CHAINLINKS_NOT_POWERED_ON",
	NT_STATUS_GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE:                     "GRAPHICS_INCONSISTENT_DEVICE_LINK_STATE",
	NT_STATUS_GRAPHICS_NOT_POST_DEVICE_DRIVER:                             "GRAPHICS_NOT_POST_DEVICE_DRIVER",
	NT_STATUS_GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED:                        "GRAPHICS_ADAPTER_ACCESS_NOT_EXCLUDED",
	NT_STATUS_GRAPHICS_OPM_NOT_SUPPORTED:                                  "GRAPHICS_OPM_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_COPP_NOT_SUPPORTED:                                 "GRAPHICS_COPP_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_UAB_NOT_SUPPORTED:                                  "GRAPHICS_UAB_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS:                   "GRAPHICS_OPM_INVALID_ENCRYPTED_PARAMETERS",
	NT_STATUS_GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL:                      "GRAPHICS_OPM_PARAMETER_ARRAY_TOO_SMALL",
	NT_STATUS_GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST:                     "GRAPHICS_OPM_NO_PROTECTED_OUTPUTS_EXIST",
	NT_STATUS_GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME:          "GRAPHICS_PVP_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME",
	NT_STATUS_GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP:         "GRAPHICS_PVP_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP",
	NT_STATUS_GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED:                "GRAPHICS_PVP_MIRRORING_DEVICES_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_OPM_INVALID_POINTER:                                "GRAPHICS_OPM_INVALID_POINTER",
	NT_STATUS_GRAPHICS_OPM_INTERNAL_ERROR:                                 "GRAPHICS_OPM_INTERNAL_ERROR",
	NT_STATUS_GRAPHICS_OPM_INVALID_HANDLE:                                 "GRAPHICS_OPM_INVALID_HANDLE",
	NT_STATUS_GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE:       "GRAPHICS_PVP_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE",
	NT_STATUS_GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH:                     "GRAPHICS_PVP_INVALID_CERTIFICATE_LENGTH",
	NT_STATUS_GRAPHICS_OPM_SPANNING_MODE_ENABLED:                          "GRAPHICS_OPM_SPANNING_MODE_ENABLED",
	NT_STATUS_GRAPHICS_OPM_THEATER_MODE_ENABLED:                           "GRAPHICS_OPM_THEATER_MODE_ENABLED",
	NT_STATUS_GRAPHICS_PVP_HFS_FAILED:                                     "GRAPHICS_PVP_HFS_FAILED",
	NT_STATUS_GRAPHICS_OPM_INVALID_SRM:                                    "GRAPHICS_OPM_INVALID_SRM",
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP:                   "GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_HDCP",
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP:                    "GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_ACP",
	NT_STATUS_GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA:                  "GRAPHICS_OPM_OUTPUT_DOES_NOT_SUPPORT_CGMSA",
	NT_STATUS_GRAPHICS_OPM_HDCP_SRM_NEVER_SET:                             "GRAPHICS_OPM_HDCP_SRM_NEVER_SET",
	NT_STATUS_GRAPHICS_OPM_RESOLUTION_TOO_HIGH:                            "GRAPHICS_OPM_RESOLUTION_TOO_HIGH",
	NT_STATUS_GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE:               "GRAPHICS_OPM_ALL_HDCP_HARDWARE_ALREADY_IN_USE",
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS:              "GRAPHICS_OPM_PROTECTED_OUTPUT_NO_LONGER_EXISTS",
	NT_STATUS_GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS:                "GRAPHICS_OPM_SESSION_TYPE_CHANGE_IN_PROGRESS",
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS:  "GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_COPP_SEMANTICS",
	NT_STATUS_GRAPHICS_OPM_INVALID_INFORMATION_REQUEST:                    "GRAPHICS_OPM_INVALID_INFORMATION_REQUEST",
	NT_STATUS_GRAPHICS_OPM_DRIVER_INTERNAL_ERROR:                          "GRAPHICS_OPM_DRIVER_INTERNAL_ERROR",
	NT_STATUS_GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS:   "GRAPHICS_OPM_PROTECTED_OUTPUT_DOES_NOT_HAVE_OPM_SEMANTICS",
	NT_STATUS_GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED:                        "GRAPHICS_OPM_SIGNALING_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST:                  "GRAPHICS_OPM_INVALID_CONFIGURATION_REQUEST",
	NT_STATUS_GRAPHICS_I2C_NOT_SUPPORTED:                                  "GRAPHICS_I2C_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST:                          "GRAPHICS_I2C_DEVICE_DOES_NOT_EXIST",
	NT_STATUS_GRAPHICS_I2C_ERROR_TRANSMITTING_DATA:                        "GRAPHICS_I2C_ERROR_TRANSMITTING_DATA",
	NT_STATUS_GRAPHICS_I2C_ERROR_RECEIVING_DATA:                           "GRAPHICS_I2C_ERROR_RECEIVING_DATA",
	NT_STATUS_GRAPHICS_DDCCI_VCP_NOT_SUPPORTED:                            "GRAPHICS_DDCCI_VCP_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_DDCCI_INVALID_DATA:                                 "GRAPHICS_DDCCI_INVALID_DATA",
	NT_STATUS_GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE:  "GRAPHICS_DDCCI_MONITOR_RETURNED_INVALID_TIMING_STATUS_BYTE",
	NT_STATUS_GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING:                  "GRAPHICS_DDCCI_INVALID_CAPABILITIES_STRING",
	NT_STATUS_GRAPHICS_MCA_INTERNAL_ERROR:                                 "GRAPHICS_MCA_INTERNAL_ERROR",
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND:                      "GRAPHICS_DDCCI_INVALID_MESSAGE_COMMAND",
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH:                       "GRAPHICS_DDCCI_INVALID_MESSAGE_LENGTH",
	NT_STATUS_GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM:                     "GRAPHICS_DDCCI_INVALID_MESSAGE_CHECKSUM",
	NT_STATUS_GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE:                    "GRAPHICS_INVALID_PHYSICAL_MONITOR_HANDLE",
	NT_STATUS_GRAPHICS_MONITOR_NO_LONGER_EXISTS:                           "GRAPHICS_MONITOR_NO_LONGER_EXISTS",
	NT_STATUS_GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED:                     "GRAPHICS_ONLY_CONSOLE_SESSION_SUPPORTED",
	NT_STATUS_GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME:              "GRAPHICS_NO_DISPLAY_DEVICE_CORRESPONDS_TO_NAME",
	NT_STATUS_GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP:             "GRAPHICS_DISPLAY_DEVICE_NOT_ATTACHED_TO_DESKTOP",
	NT_STATUS_GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED:                    "GRAPHICS_MIRRORING_DEVICES_NOT_SUPPORTED",
	NT_STATUS_GRAPHICS_INVALID_POINTER:                                    "GRAPHICS_INVALID_POINTER",
	NT_STATUS_GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE:           "GRAPHICS_NO_MONITORS_CORRESPOND_TO_DISPLAY_DEVICE",
	NT_STATUS_GRAPHICS_PARAMETER_ARRAY_TOO_SMALL:                          "GRAPHICS_PARAMETER_ARRAY_TOO_SMALL",
	NT_STATUS_GRAPHICS_INTERNAL_ERROR:                                     "GRAPHICS_INTERNAL_ERROR",
	NT_STATUS_GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS:                    "GRAPHICS_SESSION_TYPE_CHANGE_IN_PROGRESS",
	NT_STATUS_FVE_LOCKED_VOLUME:                                           "FVE_LOCKED_VOLUME",
	NT_STATUS_FVE_NOT_ENCRYPTED:                                           "FVE_NOT_ENCRYPTED",
	NT_STATUS_FVE_BAD_INFORMATION:                                         "FVE_BAD_INFORMATION",
	NT_STATUS_FVE_TOO_SMALL:                                               "FVE_TOO_SMALL",
	NT_STATUS_FVE_FAILED_WRONG_FS:                                         "FVE_FAILED_WRONG_FS",
	NT_STATUS_FVE_FAILED_BAD_FS:                                           "FVE_FAILED_BAD_FS",
	NT_STATUS_FVE_FS_NOT_EXTENDED:                                         "FVE_FS_NOT_EXTENDED",
	NT_STATUS_FVE_FS_MOUNTED:                                              "FVE_FS_MOUNTED",
	NT_STATUS_FVE_NO_LICENSE:                                              "FVE_NO_LICENSE",
	NT_STATUS_FVE_ACTION_NOT_ALLOWED:                                      "FVE_ACTION_NOT_ALLOWED",
	NT_STATUS_FVE_BAD_DATA:                                                "FVE_BAD_DATA",
	NT_STATUS_FVE_VOLUME_NOT_BOUND:                                        "FVE_VOLUME_NOT_BOUND",
	NT_STATUS_FVE_NOT_DATA_VOLUME:                                         "FVE_NOT_DATA_VOLUME",
	NT_STATUS_FVE_CONV_READ_ERROR:                                         "FVE_CONV_READ_ERROR",
	NT_STATUS_FVE_CONV_WRITE_ERROR:                                        "FVE_CONV_WRITE_ERROR",
	NT_STATUS_FVE_OVERLAPPED_UPDATE:                                       "FVE_OVERLAPPED_UPDATE",
	NT_STATUS_FVE_FAILED_SECTOR_SIZE:                                      "FVE_FAILED_SECTOR_SIZE",
	NT_STATUS_FVE_FAILED_AUTHENTICATION:                                   "FVE_FAILED_AUTHENTICATION",
	NT_STATUS_FVE_NOT_OS_VOLUME:                                           "FVE_NOT_OS_VOLUME",
	NT_STATUS_FVE_KEYFILE_NOT_FOUND:                                       "FVE_KEYFILE_NOT_FOUND",
	NT_STATUS_FVE_KEYFILE_INVALID:                                         "FVE_KEYFILE_INVALID",
	NT_STATUS_FVE_KEYFILE_NO_VMK:                                          "FVE_KEYFILE_NO_VMK",
	NT_STATUS_FVE_TPM_DISABLED:                                            "FVE_TPM_DISABLED",
	NT_STATUS_FVE_TPM_SRK_AUTH_NOT_ZERO:                                   "FVE_TPM_SRK_AUTH_NOT_ZERO",
	NT_STATUS_FVE_TPM_INVALID_PCR:                                         "FVE_TPM_INVALID_PCR",
	NT_STATUS_FVE_TPM_NO_VMK:                                              "FVE_TPM_NO_VMK",
	NT_STATUS_FVE_PIN_INVALID:                                             "FVE_PIN_INVALID",
	NT_STATUS_FVE_AUTH_INVALID_APPLICATION:                                "FVE_AUTH_INVALID_APPLICATION",
	NT_STATUS_FVE_AUTH_INVALID_CONFIG:                                     "FVE_AUTH_INVALID_CONFIG",
	NT_STATUS_FVE_DEBUGGER_ENABLED:                                        "FVE_DEBUGGER_ENABLED",
	NT_STATUS_FVE_DRY_RUN_FAILED:                                          "FVE_DRY_RUN_FAILED",
	NT_STATUS_FVE_BAD_METADATA_POINTER:                                    "FVE_BAD_METADATA_POINTER",
	NT_STATUS_FVE_OLD_METADATA_COPY:                                       "FVE_OLD_METADATA_COPY",
	NT_STATUS_FVE_REBOOT_REQUIRED:                                         "FVE_REBOOT_REQUIRED",
	NT_STATUS_FVE_RAW_ACCESS:                                              "FVE_RAW_ACCESS",
	NT_STATUS_FVE_RAW_BLOCKED:                                             "FVE_RAW_BLOCKED",
	NT_STATUS_FVE_NO_FEATURE_LICENSE:                                      "FVE_NO_FEATURE_LICENSE",
	NT_STATUS_FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED:                     "FVE_POLICY_USER_DISABLE_RDV_NOT_ALLOWED",
	NT_STATUS_FVE_CONV_RECOVERY_FAILED:                                    "FVE_CONV_RECOVERY_FAILED",
	NT_STATUS_FVE_VIRTUALIZED_SPACE_TOO_BIG:                               "FVE_VIRTUALIZED_SPACE_TOO_BIG",
	NT_STATUS_FVE_VOLUME_TOO_SMALL:                                        "FVE_VOLUME_TOO_SMALL",
	NT_STATUS_FWP_CALLOUT_NOT_FOUND:                                       "FWP_CALLOUT_NOT_FOUND",
	NT_STATUS_FWP_CONDITION_NOT_FOUND:                                     "FWP_CONDITION_NOT_FOUND",
	NT_STATUS_FWP_FILTER_NOT_FOUND:                                        "FWP_FILTER_NOT_FOUND",
	NT_STATUS_FWP_LAYER_NOT_FOUND:                                         "FWP_LAYER_NOT_FOUND",
	NT_STATUS_FWP_PROVIDER_NOT_FOUND:                                      "FWP_PROVIDER_NOT_FOUND",
	NT_STATUS_FWP_PROVIDER_CONTEXT_NOT_FOUND:                              "FWP_PROVIDER_CONTEXT_NOT_FOUND",
	NT_STATUS_FWP_SUBLAYER_NOT_FOUND:                                      "FWP_SUBLAYER_NOT_FOUND",
	NT_STATUS_FWP_NOT_FOUND:                                               "FWP_NOT_FOUND",
	NT_STATUS_FWP_ALREADY_EXISTS:                                          "FWP_ALREADY_EXISTS",
	NT_STATUS_FWP_IN_USE:                                                  "FWP_IN_USE",
	NT_STATUS_FWP_DYNAMIC_SESSION_IN_PROGRESS:                             "FWP_DYNAMIC_SESSION_IN_PROGRESS",
	NT_STATUS_FWP_WRONG_SESSION:                                           "FWP_WRONG_SESSION",
	NT_STATUS_FWP_NO_TXN_IN_PROGRESS:                                      "FWP_NO_TXN_IN_PROGRESS",
	NT_STATUS_FWP_TXN_IN_PROGRESS:                                         "FWP_TXN_IN_PROGRESS",
	NT_STATUS_FWP_TXN_ABORTED:                                             "FWP_TXN_ABORTED",
	NT_STATUS_FWP_SESSION_ABORTED:                                         "FWP_SESSION_ABORTED",
	NT_STATUS_FWP_INCOMPATIBLE_TXN:                                        "FWP_INCOMPATIBLE_TXN",
	NT_STATUS_FWP_TIMEOUT:                                                 "FWP_TIMEOUT",
	NT_STATUS_FWP_NET_EVENTS_DISABLED:                                     "FWP_NET_EVENTS_DISABLED",
	NT_STATUS_FWP_INCOMPATIBLE_LAYER:                                      "FWP_INCOMPATIBLE_LAYER",
	NT_STATUS_FWP_KM_CLIENTS_ONLY:                                         "FWP_KM_CLIENTS_ONLY",
	NT_STATUS_FWP_LIFETIME_MISMATCH:                                       "FWP_LIFETIME_MISMATCH",
	NT_STATUS_FWP_BUILTIN_OBJECT:                                          "FWP_BUILTIN_OBJECT",
	NT_STATUS_FWP_TOO_MANY_BOOTTIME_FILTERS:                               "FWP_TOO_MANY_BOOTTIME_FILTERS",
	// NT_STATUS_FWP_TOO_MANY_CALLOUTS: // NT_STATUS_FWP_TOO_MANY_CALLOUTS",
	NT_STATUS_FWP_NOTIFICATION_DROPPED:               "FWP_NOTIFICATION_DROPPED",
	NT_STATUS_FWP_TRAFFIC_MISMATCH:                   "FWP_TRAFFIC_MISMATCH",
	NT_STATUS_FWP_INCOMPATIBLE_SA_STATE:              "FWP_INCOMPATIBLE_SA_STATE",
	NT_STATUS_FWP_NULL_POINTER:                       "FWP_NULL_POINTER",
	NT_STATUS_FWP_INVALID_ENUMERATOR:                 "FWP_INVALID_ENUMERATOR",
	NT_STATUS_FWP_INVALID_FLAGS:                      "FWP_INVALID_FLAGS",
	NT_STATUS_FWP_INVALID_NET_MASK:                   "FWP_INVALID_NET_MASK",
	NT_STATUS_FWP_INVALID_RANGE:                      "FWP_INVALID_RANGE",
	NT_STATUS_FWP_INVALID_INTERVAL:                   "FWP_INVALID_INTERVAL",
	NT_STATUS_FWP_ZERO_LENGTH_ARRAY:                  "FWP_ZERO_LENGTH_ARRAY",
	NT_STATUS_FWP_NULL_DISPLAY_NAME:                  "FWP_NULL_DISPLAY_NAME",
	NT_STATUS_FWP_INVALID_ACTION_TYPE:                "FWP_INVALID_ACTION_TYPE",
	NT_STATUS_FWP_INVALID_WEIGHT:                     "FWP_INVALID_WEIGHT",
	NT_STATUS_FWP_MATCH_TYPE_MISMATCH:                "FWP_MATCH_TYPE_MISMATCH",
	NT_STATUS_FWP_TYPE_MISMATCH:                      "FWP_TYPE_MISMATCH",
	NT_STATUS_FWP_OUT_OF_BOUNDS:                      "FWP_OUT_OF_BOUNDS",
	NT_STATUS_FWP_RESERVED:                           "FWP_RESERVED",
	NT_STATUS_FWP_DUPLICATE_CONDITION:                "FWP_DUPLICATE_CONDITION",
	NT_STATUS_FWP_DUPLICATE_KEYMOD:                   "FWP_DUPLICATE_KEYMOD",
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_LAYER:     "FWP_ACTION_INCOMPATIBLE_WITH_LAYER",
	NT_STATUS_FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER:  "FWP_ACTION_INCOMPATIBLE_WITH_SUBLAYER",
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER:    "FWP_CONTEXT_INCOMPATIBLE_WITH_LAYER",
	NT_STATUS_FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT:  "FWP_CONTEXT_INCOMPATIBLE_WITH_CALLOUT",
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_METHOD:           "FWP_INCOMPATIBLE_AUTH_METHOD",
	NT_STATUS_FWP_INCOMPATIBLE_DH_GROUP:              "FWP_INCOMPATIBLE_DH_GROUP",
	NT_STATUS_FWP_EM_NOT_SUPPORTED:                   "FWP_EM_NOT_SUPPORTED",
	NT_STATUS_FWP_NEVER_MATCH:                        "FWP_NEVER_MATCH",
	NT_STATUS_FWP_PROVIDER_CONTEXT_MISMATCH:          "FWP_PROVIDER_CONTEXT_MISMATCH",
	NT_STATUS_FWP_INVALID_PARAMETER:                  "FWP_INVALID_PARAMETER",
	NT_STATUS_FWP_TOO_MANY_SUBLAYERS:                 "FWP_TOO_MANY_SUBLAYERS",
	NT_STATUS_FWP_CALLOUT_NOTIFICATION_FAILED:        "FWP_CALLOUT_NOTIFICATION_FAILED",
	NT_STATUS_FWP_INCOMPATIBLE_AUTH_CONFIG:           "FWP_INCOMPATIBLE_AUTH_CONFIG",
	NT_STATUS_FWP_INCOMPATIBLE_CIPHER_CONFIG:         "FWP_INCOMPATIBLE_CIPHER_CONFIG",
	NT_STATUS_FWP_DUPLICATE_AUTH_METHOD:              "FWP_DUPLICATE_AUTH_METHOD",
	NT_STATUS_FWP_TCPIP_NOT_READY:                    "FWP_TCPIP_NOT_READY",
	NT_STATUS_FWP_INJECT_HANDLE_CLOSING:              "FWP_INJECT_HANDLE_CLOSING",
	NT_STATUS_FWP_INJECT_HANDLE_STALE:                "FWP_INJECT_HANDLE_STALE",
	NT_STATUS_FWP_CANNOT_PEND:                        "FWP_CANNOT_PEND",
	NT_STATUS_NDIS_CLOSING:                           "NDIS_CLOSING",
	NT_STATUS_NDIS_BAD_VERSION:                       "NDIS_BAD_VERSION",
	NT_STATUS_NDIS_BAD_CHARACTERISTICS:               "NDIS_BAD_CHARACTERISTICS",
	NT_STATUS_NDIS_ADAPTER_NOT_FOUND:                 "NDIS_ADAPTER_NOT_FOUND",
	NT_STATUS_NDIS_OPEN_FAILED:                       "NDIS_OPEN_FAILED",
	NT_STATUS_NDIS_DEVICE_FAILED:                     "NDIS_DEVICE_FAILED",
	NT_STATUS_NDIS_MULTICAST_FULL:                    "NDIS_MULTICAST_FULL",
	NT_STATUS_NDIS_MULTICAST_EXISTS:                  "NDIS_MULTICAST_EXISTS",
	NT_STATUS_NDIS_MULTICAST_NOT_FOUND:               "NDIS_MULTICAST_NOT_FOUND",
	NT_STATUS_NDIS_REQUEST_ABORTED:                   "NDIS_REQUEST_ABORTED",
	NT_STATUS_NDIS_RESET_IN_PROGRESS:                 "NDIS_RESET_IN_PROGRESS",
	NT_STATUS_NDIS_INVALID_PACKET:                    "NDIS_INVALID_PACKET",
	NT_STATUS_NDIS_INVALID_DEVICE_REQUEST:            "NDIS_INVALID_DEVICE_REQUEST",
	NT_STATUS_NDIS_ADAPTER_NOT_READY:                 "NDIS_ADAPTER_NOT_READY",
	NT_STATUS_NDIS_INVALID_LENGTH:                    "NDIS_INVALID_LENGTH",
	NT_STATUS_NDIS_INVALID_DATA:                      "NDIS_INVALID_DATA",
	NT_STATUS_NDIS_BUFFER_TOO_SHORT:                  "NDIS_BUFFER_TOO_SHORT",
	NT_STATUS_NDIS_INVALID_OID:                       "NDIS_INVALID_OID",
	NT_STATUS_NDIS_ADAPTER_REMOVED:                   "NDIS_ADAPTER_REMOVED",
	NT_STATUS_NDIS_UNSUPPORTED_MEDIA:                 "NDIS_UNSUPPORTED_MEDIA",
	NT_STATUS_NDIS_GROUP_ADDRESS_IN_USE:              "NDIS_GROUP_ADDRESS_IN_USE",
	NT_STATUS_NDIS_FILE_NOT_FOUND:                    "NDIS_FILE_NOT_FOUND",
	NT_STATUS_NDIS_ERROR_READING_FILE:                "NDIS_ERROR_READING_FILE",
	NT_STATUS_NDIS_ALREADY_MAPPED:                    "NDIS_ALREADY_MAPPED",
	NT_STATUS_NDIS_RESOURCE_CONFLICT:                 "NDIS_RESOURCE_CONFLICT",
	NT_STATUS_NDIS_MEDIA_DISCONNECTED:                "NDIS_MEDIA_DISCONNECTED",
	NT_STATUS_NDIS_INVALID_ADDRESS:                   "NDIS_INVALID_ADDRESS",
	NT_STATUS_NDIS_PAUSED:                            "NDIS_PAUSED",
	NT_STATUS_NDIS_INTERFACE_NOT_FOUND:               "NDIS_INTERFACE_NOT_FOUND",
	NT_STATUS_NDIS_UNSUPPORTED_REVISION:              "NDIS_UNSUPPORTED_REVISION",
	NT_STATUS_NDIS_INVALID_PORT:                      "NDIS_INVALID_PORT",
	NT_STATUS_NDIS_INVALID_PORT_STATE:                "NDIS_INVALID_PORT_STATE",
	NT_STATUS_NDIS_LOW_POWER_STATE:                   "NDIS_LOW_POWER_STATE",
	NT_STATUS_NDIS_NOT_SUPPORTED:                     "NDIS_NOT_SUPPORTED",
	NT_STATUS_NDIS_OFFLOAD_POLICY:                    "NDIS_OFFLOAD_POLICY",
	NT_STATUS_NDIS_OFFLOAD_CONNECTION_REJECTED:       "NDIS_OFFLOAD_CONNECTION_REJECTED",
	NT_STATUS_NDIS_OFFLOAD_PATH_REJECTED:             "NDIS_OFFLOAD_PATH_REJECTED",
	NT_STATUS_NDIS_DOT11_AUTO_CONFIG_ENABLED:         "NDIS_DOT11_AUTO_CONFIG_ENABLED",
	NT_STATUS_NDIS_DOT11_MEDIA_IN_USE:                "NDIS_DOT11_MEDIA_IN_USE",
	NT_STATUS_NDIS_DOT11_POWER_STATE_INVALID:         "NDIS_DOT11_POWER_STATE_INVALID",
	NT_STATUS_NDIS_PM_WOL_PATTERN_LIST_FULL:          "NDIS_PM_WOL_PATTERN_LIST_FULL",
	NT_STATUS_NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL:     "NDIS_PM_PROTOCOL_OFFLOAD_LIST_FULL",
	NT_STATUS_IPSEC_BAD_SPI:                          "IPSEC_BAD_SPI",
	NT_STATUS_IPSEC_SA_LIFETIME_EXPIRED:              "IPSEC_SA_LIFETIME_EXPIRED",
	NT_STATUS_IPSEC_WRONG_SA:                         "IPSEC_WRONG_SA",
	NT_STATUS_IPSEC_REPLAY_CHECK_FAILED:              "IPSEC_REPLAY_CHECK_FAILED",
	NT_STATUS_IPSEC_INVALID_PACKET:                   "IPSEC_INVALID_PACKET",
	NT_STATUS_IPSEC_INTEGRITY_CHECK_FAILED:           "IPSEC_INTEGRITY_CHECK_FAILED",
	NT_STATUS_IPSEC_CLEAR_TEXT_DROP:                  "IPSEC_CLEAR_TEXT_DROP",
	NT_STATUS_IPSEC_AUTH_FIREWALL_DROP:               "IPSEC_AUTH_FIREWALL_DROP",
	NT_STATUS_IPSEC_THROTTLE_DROP:                    "IPSEC_THROTTLE_DROP",
	NT_STATUS_IPSEC_DOSP_BLOCK:                       "IPSEC_DOSP_BLOCK",
	NT_STATUS_IPSEC_DOSP_RECEIVED_MULTICAST:          "IPSEC_DOSP_RECEIVED_MULTICAST",
	NT_STATUS_IPSEC_DOSP_INVALID_PACKET:              "IPSEC_DOSP_INVALID_PACKET",
	NT_STATUS_IPSEC_DOSP_STATE_LOOKUP_FAILED:         "IPSEC_DOSP_STATE_LOOKUP_FAILED",
	NT_STATUS_IPSEC_DOSP_MAX_ENTRIES:                 "IPSEC_DOSP_MAX_ENTRIES",
	NT_STATUS_IPSEC_DOSP_KEYMOD_NOT_ALLOWED:          "IPSEC_DOSP_KEYMOD_NOT_ALLOWED",
	NT_STATUS_IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES: "IPSEC_DOSP_MAX_PER_IP_RATELIMIT_QUEUES",
	NT_STATUS_VOLMGR_MIRROR_NOT_SUPPORTED:            "VOLMGR_MIRROR_NOT_SUPPORTED",
	NT_STATUS_VOLMGR_RAID5_NOT_SUPPORTED:             "VOLMGR_RAID5_NOT_SUPPORTED",
	NT_STATUS_VIRTDISK_PROVIDER_NOT_FOUND:            "VIRTDISK_PROVIDER_NOT_FOUND",
	NT_STATUS_VIRTDISK_NOT_VIRTUAL_DISK:              "VIRTDISK_NOT_VIRTUAL_DISK",
	NT_STATUS_VHD_PARENT_VHD_ACCESS_DENIED:           "VHD_PARENT_VHD_ACCESS_DENIED",
	NT_STATUS_VHD_CHILD_PARENT_SIZE_MISMATCH:         "VHD_CHILD_PARENT_SIZE_MISMATCH",
	NT_STATUS_VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED:  "VHD_DIFFERENCING_CHAIN_CYCLE_DETECTED",
	NT_STATUS_VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT: "VHD_DIFFERENCING_CHAIN_ERROR_IN_PARENT",
	NT_STATUS_SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP:  "SMB_NO_PREAUTH_INTEGRITY_HASH_OVERLAP",
	NT_STATUS_SMB_BAD_CLUSTER_DIALECT:                "SMB_BAD_CLUSTER_DIALECT",
}

func (s NT_STATUS) String() string {
	if str, exists := NTStatusToStringName[s]; exists {
		return str
	}
	return "UNKNOWN"
}

func (s NT_STATUS) Error() error {
	if s == NT_STATUS_SUCCESS {
		return nil
	}

	if str, exists := NTStatusToGoErrorMap[s]; exists {
		return fmt.Errorf("NT_STATUS(0x%08x): %s", uint32(s), str)
	}

	return nil
}
