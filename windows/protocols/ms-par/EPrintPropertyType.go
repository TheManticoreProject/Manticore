package mspar

// EPrintPropertyType is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-PAR]).
type EPrintPropertyType uint16

const (
	kPropertyTypeString              EPrintPropertyType = 1
	kPropertyTypeInt32               EPrintPropertyType = 2
	kPropertyTypeInt64               EPrintPropertyType = 3
	kPropertyTypeByte                EPrintPropertyType = 4
	kPropertyTypeTime                EPrintPropertyType = 5
	kPropertyTypeDevMode             EPrintPropertyType = 6
	kPropertyTypeSD                  EPrintPropertyType = 7
	kPropertyTypeNotificationReply   EPrintPropertyType = 8
	kPropertyTypeNotificationOptions EPrintPropertyType = 9
)
