package msdltw

// CDomainRelativeObjId is a domain-relative object identifier ([MS-DLTW] 2.2.5): the
// pairing of a VolumeID and an ObjectID that together locate a tracked file. In the
// protocol it names both a FileID (the birth droid) and a FileLocation (the current
// droid). Both members are transmitted inline in declaration order (Volume then Object),
// for a total of 32 octets on the wire.
type CDomainRelativeObjId struct {
	Volume CVolumeId
	Object CObjId
}
