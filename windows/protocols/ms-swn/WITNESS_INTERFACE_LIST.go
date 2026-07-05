package msswn

// WITNESS_INTERFACE_LIST is the list of witness-capable interfaces returned by
// WitnessrGetInterfaceList ([MS-SWN] 2.2.2.6). NumberOfInterfaces is the element count;
// InterfaceInfo is a [size_is(NumberOfInterfaces)] [unique] pointer to a conformant array
// of WITNESS_INTERFACE_INFO (the codec emits a referent id, then defers the array body).
type WITNESS_INTERFACE_LIST struct {
	NumberOfInterfaces uint32
	InterfaceInfo      []WITNESS_INTERFACE_INFO `ndr:"unique,size_is=NumberOfInterfaces"`
}
