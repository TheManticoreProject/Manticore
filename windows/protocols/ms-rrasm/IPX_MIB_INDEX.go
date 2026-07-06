package msrrasm

// IPX_MIB_INDEX ([MS-RRASM] 2.2.1.2.190) is a C union with no on-wire discriminant;
// the active arm is selected by the sibling TableId of the containing
// IPX_MIB_GET_INPUT_DATA. It is modeled by its largest arm,
// STATIC_SERVICES_TABLE_INDEX; the other arms (IF_TABLE_INDEX, ROUTING_TABLE_INDEX,
// STATIC_ROUTES_TABLE_INDEX, SERVICES_TABLE_INDEX) overlay the same bytes.
type IPX_MIB_INDEX struct {
	StaticServicesTableIndex STATIC_SERVICES_TABLE_INDEX
}
