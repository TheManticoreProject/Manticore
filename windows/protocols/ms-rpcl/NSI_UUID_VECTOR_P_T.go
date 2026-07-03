package msrpcl

// NSI_UUID_VECTOR_P_T is a [unique] pointer to an NSI_UUID_VECTOR_T ([MS-RPCL] 2.2).
// The pointer framing (referent id) comes from the ndr:"unique" tag at each use site.
type NSI_UUID_VECTOR_P_T = *NSI_UUID_VECTOR_T
