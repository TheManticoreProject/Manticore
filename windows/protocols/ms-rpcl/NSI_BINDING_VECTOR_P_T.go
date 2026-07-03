package msrpcl

// NSI_BINDING_VECTOR_P_T is a [unique] pointer to an NSI_BINDING_VECTOR_T ([MS-RPCL] 2.2).
// The pointer framing (referent id) comes from the ndr:"unique" tag at each use site.
type NSI_BINDING_VECTOR_P_T = *NSI_BINDING_VECTOR_T
