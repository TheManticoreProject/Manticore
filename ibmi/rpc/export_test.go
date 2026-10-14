package rpc

// CheckInputLength exposes the 32-bit input cap, which cannot be reached
// through Input without allocating 4 GiB.
var CheckInputLength = checkInputLength
