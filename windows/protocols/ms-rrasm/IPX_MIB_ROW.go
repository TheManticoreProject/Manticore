package msrrasm

// IPX_MIB_ROW ([MS-RRASM] 2.2.1.2.195) is a C union with no on-wire discriminant;
// the active arm is selected by the sibling TableId of the containing
// IPX_MIB_SET_INPUT_DATA. It is modeled by its largest arm, IPX_INTERFACE; the
// other arms (IPX_ROUTE, IPX_SERVICE) overlay the same bytes.
type IPX_MIB_ROW struct {
	Interface IPX_INTERFACE
}
