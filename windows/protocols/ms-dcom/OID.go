package msdcom

// OID is an object identifier ([MS-DCOM] 2.2.5): a 64-bit value that identifies an
// object within an object exporter's ping set. Carried as unsigned hyper.
type OID uint64
