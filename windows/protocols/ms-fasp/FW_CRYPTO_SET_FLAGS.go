package msfasp

// FW_CRYPTO_SET_FLAGS is an NDR enum, transmitted as a 16-bit value ([C706] 14.3.6, [MS-FASP]).
type FW_CRYPTO_SET_FLAGS uint16

const (
	FW_CRYPTO_SET_FLAGS_NONE FW_CRYPTO_SET_FLAGS = 0x00
	FW_CRYPTO_SET_FLAGS_MAX  FW_CRYPTO_SET_FLAGS = 0x01
)
