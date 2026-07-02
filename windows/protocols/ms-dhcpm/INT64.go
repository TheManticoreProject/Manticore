package msdhcpm

// INT64 is a 64-bit signed integer. The MS-DHCPM IDL uses the bare MIDL base
// type INT64 (e.g. the reserved fields of DHCP_SUBNET_INFO_VQ) without a
// typedef, so it is modeled here as a named 8-byte signed integer.
type INT64 int64
