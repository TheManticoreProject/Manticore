package rmtcmd

import "fmt"

// Default operation limits.
const (
	DefaultMaxReceiver     uint32 = 16 << 20
	DefaultMaxMessageText  uint32 = 1 << 20
	DefaultMaxRLEExpansion uint32 = 16 << 20
	DefaultMaxMessages            = 256

	// MaxLimit is the largest receiver or RLE expansion accepted.
	MaxLimit uint32 = 64 << 20
)

// Limits bound every peer-controlled allocation of Call and RunCommand.
// Zero fields select the defaults.
type Limits struct {
	// MaxReceiver caps each output parameter, requested and returned.
	MaxReceiver uint32
	// MaxMessageText caps each message's rendered text.
	MaxMessageText uint32
	// MaxRLEExpansion caps the decompressed size of one RLE output.
	MaxRLEExpansion uint32
	// MaxMessages caps the message records in one reply.
	MaxMessages int
}

// Resolve applies defaults and validates the limits.
func (l Limits) Resolve() (Limits, error) {
	if l.MaxReceiver == 0 {
		l.MaxReceiver = DefaultMaxReceiver
	}
	if l.MaxMessageText == 0 {
		l.MaxMessageText = DefaultMaxMessageText
	}
	if l.MaxRLEExpansion == 0 {
		l.MaxRLEExpansion = DefaultMaxRLEExpansion
	}
	if l.MaxMessages == 0 {
		l.MaxMessages = DefaultMaxMessages
	}
	switch {
	case l.MaxReceiver > MaxLimit:
		return Limits{}, fmt.Errorf("rmtcmd: MaxReceiver %d exceeds %d", l.MaxReceiver, MaxLimit)
	case l.MaxRLEExpansion > MaxLimit:
		return Limits{}, fmt.Errorf("rmtcmd: MaxRLEExpansion %d exceeds %d", l.MaxRLEExpansion, MaxLimit)
	case l.MaxMessageText > MaxLimit:
		return Limits{}, fmt.Errorf("rmtcmd: MaxMessageText %d exceeds %d", l.MaxMessageText, MaxLimit)
	case l.MaxMessages < 0 || l.MaxMessages > 0xFFFF:
		return Limits{}, fmt.Errorf("rmtcmd: MaxMessages %d outside 1..65535", l.MaxMessages)
	}
	return l, nil
}
