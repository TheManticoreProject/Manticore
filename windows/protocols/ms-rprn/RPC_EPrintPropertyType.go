package msrprn

// RPC_EPrintPropertyType is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-RPRN]).
type RPC_EPrintPropertyType uint16

const (
	kRpcPropertyTypeString RPC_EPrintPropertyType = 1
	kRpcPropertyTypeInt32  RPC_EPrintPropertyType = 2
	kRpcPropertyTypeInt64  RPC_EPrintPropertyType = 3
	kRpcPropertyTypeByte   RPC_EPrintPropertyType = 4
	kRpcPropertyTypeBuffer RPC_EPrintPropertyType = 5
)
