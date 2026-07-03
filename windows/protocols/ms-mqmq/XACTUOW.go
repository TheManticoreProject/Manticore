package msmqmq

// XACTUOW is a transaction unit-of-work identifier ([MS-MQMQ] 2.2.18.1.8): a fixed
// 16-octet value. It identifies the transaction of a Send or Receive operation in
// [MS-MQMP].
type XACTUOW struct {
	Rgb [16]uint8
}
