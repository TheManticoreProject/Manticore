package commands

// ReadRawResponse
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/3f3914f6-d251-48ab-89c0-35349cb6d1c8

// The server's "raw" response
// - No SMB Header at all. Instead, the server uses the underlying SMB transport's session-service framing
// (4-byte NetBIOS header: 1-byte MessageType + 3-byte Length) and then emits only the file-data bytes. There is:
// [NetBIOS Session Service header:
//     - MessageType = 0x00 (session message)
//     - Length = N (the number of data bytes)
// ]
// [Raw data payload:  N bytes of file or pipe data]

// There is no SMB Status/Flags fields, no WordCount, no DataOffset-just a single contiguous byte-stream whose length
// is inferred from the transport header.
// Source: https://learn.microsoft.com/en-us/openspecs/windows_protocols/ms-cifs/07e4a838-0137-4405-b534-1e5eeaf16812
// - If the server can't fulfill the read (error or EOF), it simply returns a zero-length payload (Length=0)

// Why "no SMB Header" on the response?

// Because SMB_COM_READ_RAW is optimized for bulk reads, the designers stripped away the per-message header to avoid
// extra framing overhead and allow messages larger than the negotiated buffer size. The only framing is done by the
// NetBIOS session layer, so the client must quiesce all other requests until it has received exactly the number of
// bytes (or fewer, on EOF) that it asked for
