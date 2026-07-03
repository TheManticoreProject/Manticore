package msmqqp

// REMOTEREADDESC2 ([MS-MQQP] 2.2.2.3) extends the receive descriptor used by
// RemoteQMStartReceive2 and RemoteQMStartReceiveByLookupId: a unique pointer to a
// REMOTEREADDESC plus the 64-bit SequentialId of the message.
type REMOTEREADDESC2 struct {
	PRemoteReadDesc *REMOTEREADDESC `ndr:"unique"`
	SequentialId    uint64
}
