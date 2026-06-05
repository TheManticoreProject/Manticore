package structures

// EFS_EXIM_PIPE is the EFSRPC raw import/export pipe ([MS-EFSR] 2.2.7): an NDR pipe
// of bytes (IDL `typedef pipe unsigned char`). On the wire it is a sequence of chunks
// — each a uint32 element count followed by that many bytes — terminated by a 0-count
// chunk ([C706] section 14.7). Marshal/unmarshal it with the `ndr:"pipe"` tag.
type EFS_EXIM_PIPE []byte
