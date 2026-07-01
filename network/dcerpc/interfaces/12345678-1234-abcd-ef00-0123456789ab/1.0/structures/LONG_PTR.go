package structures

// LONG_PTR is the [MS-DTYP] pointer-sized signed integer. It appears only in the reserved
// SPLCLIENT_INFO_2.notUsed field ([MS-RPRN] 2.2.1.3.2), which the spooler marshals in a
// 32-bit context, so it is transmitted as a 4-byte value.
type LONG_PTR int32
