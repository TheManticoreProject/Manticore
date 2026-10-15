package message

import (
	"github.com/TheManticoreProject/Manticore/network/ibmi/hostserver"
	"github.com/TheManticoreProject/Manticore/network/ibmi/internal/startup"
)

// Header byte 4 of 0x7001 advertises the client's password capabilities:
// the JTOpen 4.8 client sends SeedAttributes (SHA-1 capable); current JTOpen
// sends SeedAttributesLevel4, which also advertises the password level 4
// algorithm (iSeriesNavigatorReverse docs/04).
const (
	SeedAttributes       = startup.SeedAttributes
	SeedAttributesLevel4 = startup.SeedAttributesLevel4
)

// ExchangeSeedsFrame encodes request 0x7001 with SeedAttributes: the 8-byte
// client seed is the whole template.
func ExchangeSeedsFrame(clientSeed [8]byte) hostserver.Frame {
	return ExchangeSeedsFrameWithAttributes(clientSeed, SeedAttributes)
}

// ExchangeSeedsFrameWithAttributes encodes request 0x7001 with the given
// client attributes byte.
func ExchangeSeedsFrameWithAttributes(clientSeed [8]byte, attributes byte) hostserver.Frame {
	return startup.ExchangeSeedsFrame(hostserver.ServerIDRemoteCommand, clientSeed, attributes)
}

// ExchangeSeedsReply is reply 0xF001.
type ExchangeSeedsReply = startup.ExchangeSeedsReply

// ParseExchangeSeedsReply decodes reply 0xF001: a 4-byte return code, then
// with a zero return code the 8-byte server seed and optional LL/CP fields.
func ParseExchangeSeedsReply(f hostserver.Frame) (*ExchangeSeedsReply, error) {
	return startup.ParseExchangeSeedsReply(f)
}
