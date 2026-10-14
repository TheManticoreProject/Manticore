package rmtcmd

import (
	"strings"

	"github.com/TheManticoreProject/Manticore/ibmi/ccsid"
	"github.com/TheManticoreProject/Manticore/ibmi/rpc"
	"github.com/TheManticoreProject/Manticore/network/ibmi/rmtcmd/message"
)

// decodeMessages parses count records from b and converts their character
// fields with the job-CCSID codec. The message ID and the message-file
// names are trimmed of EBCDIC blanks; the rendered text is kept verbatim.
func (s *Session) decodeMessages(count uint16, b []byte) ([]rpc.IBMMessage, error) {
	records, err := message.ParseMessages(count, b, message.MessageLimits{MaxMessages: s.limits.MaxMessages, MaxText: s.limits.MaxMessageText})
	if err != nil {
		return nil, err
	}
	out := make([]rpc.IBMMessage, 0, len(records))
	for _, r := range records {
		id, err := ccsid.DecodeFixed(s.codec, r.ID[:])
		if err != nil {
			return nil, err
		}
		file, err := ccsid.DecodeFixed(s.codec, r.File[:])
		if err != nil {
			return nil, err
		}
		lib, err := ccsid.DecodeFixed(s.codec, r.Library[:])
		if err != nil {
			return nil, err
		}
		text, err := s.codec.Decode(r.Text)
		if err != nil {
			return nil, err
		}
		out = append(out, rpc.IBMMessage{
			ID:               strings.TrimRight(id, " "),
			Type:             uint16(r.Type),
			Severity:         r.Severity,
			FileName:         file,
			LibraryName:      lib,
			Text:             text,
			SubstitutionData: r.Substitution,
		})
	}
	return out, nil
}

// wireMessageOption maps a semantic option to the request byte for the
// negotiated level and reports the option the server will honor:
//
//   - below level 7, "all" is downgraded to "up to ten";
//   - at level 10, "up to ten" is sent as 3 and "all" as 4;
//   - at level 11 and later, as 5 and 6.
//
// The level-10/11 values come from current JTOpen and do not apply to
// V5R4-era servers.
func wireMessageOption(o rpc.MessageOption, caps Capabilities) (byte, rpc.MessageOption) {
	effective := o
	if o == rpc.MessagesAll && !caps.AllCommandMessages {
		effective = rpc.MessagesUpToTen
	}
	level := caps.RemoteCommandLevel
	switch effective {
	case rpc.MessagesNone:
		return 1, effective
	case rpc.MessagesAll:
		switch {
		case level >= 11:
			return 6, effective
		case level == 10:
			return 4, effective
		}
		return 2, effective
	}
	switch {
	case level >= 11:
		return 5, effective
	case level == 10:
		return 3, effective
	}
	return 0, effective
}
