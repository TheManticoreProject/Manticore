package msfrs2

// FRS_RDC_PARAMETERS ([MS-FRS2] 2.2.1.4.10) holds the RDC filter configuration selected
// by rdcChunkerAlgorithm ([MS-FRS2] enumeration RDC_CHUNKER_ALGORITHM).
type FRS_RDC_PARAMETERS struct {
	RdcChunkerAlgorithm uint16
	U                   FRS_RDC_PARAMETERS_U
}

// FRS_RDC_PARAMETERS_U is a discriminated union ([MS-FRS2] 2.2.1.4.10); the discriminant
// precedes the selected arm ([C706] 14.3.8). The IDL switches on rdcChunkerAlgorithm, an
// unsigned short, so the inline discriminant is 16-bit (not a DWORD).
type FRS_RDC_PARAMETERS_U struct {
	Tag           uint16                         `ndr:"switch"`
	FilterGeneric FRS_RDC_PARAMETERS_GENERIC     `ndr:"case=0"`
	FilterMax     FRS_RDC_PARAMETERS_FILTERMAX   `ndr:"case=1"`
	FilterPoint   FRS_RDC_PARAMETERS_FILTERPOINT `ndr:"case=2"`
}
