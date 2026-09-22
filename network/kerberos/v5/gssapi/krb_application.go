package gssapi

import (
	"bytes"
	"fmt"
	"time"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// KRBApplicationOptions supplies the RFC-native application-message metadata.
// SenderAddress is required. Timestamp defaults to now; SequenceNumber and
// RecipientAddress are optional.
type KRBApplicationOptions struct {
	Timestamp        time.Time
	SequenceNumber   *int
	SenderAddress    messages.HostAddress
	RecipientAddress *messages.HostAddress
}

// KRBApplicationReceiveOptions controls freshness, sequence, and address
// validation while receiving KRB-SAFE or KRB-PRIV.
type KRBApplicationReceiveOptions struct {
	Now              time.Time
	ClockSkew        time.Duration
	SequenceNumber   *int
	SenderAddress    *messages.HostAddress
	RecipientAddress *messages.HostAddress
}

func applicationBody(data []byte, opts KRBApplicationOptions) (messages.KRBApplicationBody, error) {
	if len(opts.SenderAddress.Address) == 0 {
		return messages.KRBApplicationBody{}, fmt.Errorf("gssapi: KRB application sender address is required")
	}
	now := opts.Timestamp
	if now.IsZero() {
		now = time.Now().UTC()
	}
	body := messages.KRBApplicationBody{
		UserData: append([]byte(nil), data...), Timestamp: now,
		Usec: now.Nanosecond() / 1000, SenderAddress: opts.SenderAddress,
	}
	if opts.SequenceNumber != nil {
		body.SequenceNumber = *opts.SequenceNumber
	}
	if opts.RecipientAddress != nil {
		body.RecipientAddress = *opts.RecipientAddress
	}
	return body, nil
}

// MakeKRBSafe creates an integrity-protected RFC 4120 KRB-SAFE message.
func (ctx *SecContext) MakeKRBSafe(data []byte, opts KRBApplicationOptions) ([]byte, error) {
	body, err := applicationBody(data, opts)
	if err != nil {
		return nil, err
	}
	bodyDER, err := body.Marshal()
	if err != nil {
		return nil, err
	}
	key, etype := ctx.baseKey()
	cksumType, ok := kerbcrypto.ChecksumTypeForEType(etype)
	if !ok {
		return nil, fmt.Errorf("gssapi: no checksum for KRB-SAFE etype %d", etype)
	}
	checksum, err := kerbcrypto.GetChecksum(cksumType, key, iana.KeyUsageKRBSafeCksum, bodyDER)
	if err != nil {
		return nil, err
	}
	return (&messages.KRBSafe{SafeBody: body, Checksum: messages.Checksum{CKSumType: cksumType, Checksum: checksum}}).Marshal()
}

// ReadKRBSafe authenticates and decodes an RFC 4120 KRB-SAFE message.
func (ctx *SecContext) ReadKRBSafe(wire []byte, opts KRBApplicationReceiveOptions) ([]byte, error) {
	var msg messages.KRBSafe
	if _, err := msg.Unmarshal(wire); err != nil {
		return nil, err
	}
	bodyDER, err := msg.SafeBody.Marshal()
	if err != nil {
		return nil, err
	}
	key, _ := ctx.baseKey()
	if !kerbcrypto.VerifyChecksum(msg.Checksum.CKSumType, key, iana.KeyUsageKRBSafeCksum, bodyDER, msg.Checksum.Checksum) {
		return nil, fmt.Errorf("gssapi: invalid KRB-SAFE checksum")
	}
	if err := validateApplicationBody(msg.SafeBody, opts); err != nil {
		return nil, err
	}
	return append([]byte(nil), msg.SafeBody.UserData...), nil
}

// MakeKRBPriv creates a confidentiality-protected RFC 4120 KRB-PRIV message.
func (ctx *SecContext) MakeKRBPriv(data []byte, opts KRBApplicationOptions) ([]byte, error) {
	body, err := applicationBody(data, opts)
	if err != nil {
		return nil, err
	}
	plain, err := (*messages.EncKRBPrivPart)(&body).Marshal()
	if err != nil {
		return nil, err
	}
	key, etype := ctx.baseKey()
	cipher, err := kerbcrypto.Encrypt(etype, key, iana.KeyUsageKRBPrivEncPart, plain)
	if err != nil {
		return nil, err
	}
	return (&messages.KRBPriv{EncPart: messages.EncryptedData{EType: etype, Cipher: cipher}}).Marshal()
}

// ReadKRBPriv decrypts, validates, and returns an RFC 4120 KRB-PRIV payload.
func (ctx *SecContext) ReadKRBPriv(wire []byte, opts KRBApplicationReceiveOptions) ([]byte, error) {
	var msg messages.KRBPriv
	if _, err := msg.Unmarshal(wire); err != nil {
		return nil, err
	}
	key, etype := ctx.baseKey()
	if msg.EncPart.EType != etype {
		return nil, fmt.Errorf("gssapi: KRB-PRIV enctype %d does not match context enctype %d", msg.EncPart.EType, etype)
	}
	plain, err := kerbcrypto.Decrypt(etype, key, iana.KeyUsageKRBPrivEncPart, msg.EncPart.Cipher)
	if err != nil {
		return nil, fmt.Errorf("gssapi: decrypt KRB-PRIV: %w", err)
	}
	var part messages.EncKRBPrivPart
	if _, err := part.Unmarshal(plain); err != nil {
		return nil, err
	}
	body := messages.KRBApplicationBody(part)
	if err := validateApplicationBody(body, opts); err != nil {
		return nil, err
	}
	return append([]byte(nil), body.UserData...), nil
}

func validateApplicationBody(body messages.KRBApplicationBody, opts KRBApplicationReceiveOptions) error {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	skew := opts.ClockSkew
	if skew <= 0 {
		skew = DefaultClockSkew
	}
	if body.Timestamp.IsZero() && opts.SequenceNumber == nil {
		return fmt.Errorf("gssapi: KRB application message has neither timestamp nor expected sequence")
	}
	if !body.Timestamp.IsZero() {
		delta := now.Sub(body.Timestamp.UTC())
		if delta < 0 {
			delta = -delta
		}
		if delta > skew {
			return fmt.Errorf("gssapi: KRB application message clock skew too large")
		}
	}
	if opts.SequenceNumber != nil && body.SequenceNumber != *opts.SequenceNumber {
		return fmt.Errorf("gssapi: KRB application sequence mismatch: got %d, want %d", body.SequenceNumber, *opts.SequenceNumber)
	}
	if opts.SenderAddress != nil && !sameHostAddress(body.SenderAddress, *opts.SenderAddress) {
		return fmt.Errorf("gssapi: KRB application sender address mismatch")
	}
	if opts.RecipientAddress != nil && !sameHostAddress(body.RecipientAddress, *opts.RecipientAddress) {
		return fmt.Errorf("gssapi: KRB application recipient address mismatch")
	}
	return nil
}

func sameHostAddress(a, b messages.HostAddress) bool {
	return a.AddrType == b.AddrType && bytes.Equal(a.Address, b.Address)
}
