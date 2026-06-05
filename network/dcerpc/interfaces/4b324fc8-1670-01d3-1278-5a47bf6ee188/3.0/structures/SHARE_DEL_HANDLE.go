package structures

// SHARE_DEL_HANDLE is a [context_handle] returned by NetrShareDelStart and
// consumed by NetrShareDelCommit ([MS-SRVS] 2.2.1.1 / 3.1.4.13). On the wire a
// context handle is an attributes DWORD followed by a 16-byte GUID, i.e. a
// 20-byte opaque blob.
type SHARE_DEL_HANDLE [20]byte
