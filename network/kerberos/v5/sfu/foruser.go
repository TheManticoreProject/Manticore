// Package sfu implements the Microsoft Service for User and Constrained
// Delegation extensions ([MS-SFU]) that layer onto RFC 4120's TGS exchange:
// PA-FOR-USER (S4U2Self) and the padata used for S4U2Proxy. Like the rest of
// the MS extensions, these ride the RFC's padata / additional-tickets extension
// points — the TGS-REQ/REP messages are unchanged.
package sfu

import (
	"encoding/asn1"
	"encoding/binary"
	"fmt"

	kerbcrypto "github.com/TheManticoreProject/Manticore/network/kerberos/v5/crypto"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/iana"
	"github.com/TheManticoreProject/Manticore/network/kerberos/v5/messages"
)

// authPackageKerberos is the fixed auth-package value in PA-FOR-USER ([MS-SFU]
// 2.2.1): the authentication mechanism name, which MUST be "Kerberos".
const authPackageKerberos = "Kerberos"

// s4uByteArray builds the S4UByteArray checksummed by PA-FOR-USER ([MS-SFU]
// 2.2.1): the userName.name-type as a 4-byte little-endian integer, followed by
// each userName.name-string component, then userRealm, then auth-package — with
// no null terminators.
func s4uByteArray(userName messages.PrincipalName, userRealm, authPackage string) []byte {
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(userName.NameType))
	for _, s := range userName.NameString {
		buf = append(buf, []byte(s)...)
	}
	buf = append(buf, []byte(userRealm)...)
	buf = append(buf, []byte(authPackage)...)
	return buf
}

// BuildPAForUser builds the PA-FOR-USER pre-authentication data element used in
// an S4U2Self TGS-REQ, on behalf of the user (userName, userRealm). The keyed
// checksum is computed with the requesting service's TGT session key at key
// usage KERB_NON_KERB_CKSUM_SALT (17).
//
// [MS-SFU] 2.2.1 specifies KERB_CHECKSUM_HMAC_MD5, which is correct when the TGT
// session key is RC4. For AES TGT session keys, modern KDCs expect the checksum
// type paired with the session key's enctype; this function selects that type
// from sessionKeyEType so it interoperates with current Active Directory.
func BuildPAForUser(userName messages.PrincipalName, userRealm string, sessionKey []byte, sessionKeyEType int) (messages.PAData, error) {
	cksumType, ok := kerbcrypto.ChecksumTypeForEType(sessionKeyEType)
	if !ok {
		return messages.PAData{}, fmt.Errorf("sfu: unsupported session-key etype %d", sessionKeyEType)
	}
	buf := s4uByteArray(userName, userRealm, authPackageKerberos)
	sum, err := kerbcrypto.GetChecksum(cksumType, sessionKey, iana.KeyUsageKerbNonKerbCksumSalt, buf)
	if err != nil {
		return messages.PAData{}, fmt.Errorf("sfu: PA-FOR-USER checksum: %w", err)
	}

	userNameElem, err := asn1.MarshalWithParams(messages.MarshalPrincipalName(userName), "explicit,tag:0")
	if err != nil {
		return messages.PAData{}, err
	}
	userRealmElem, err := asn1.Marshal(messages.ExplicitGeneralString(1, userRealm))
	if err != nil {
		return messages.PAData{}, err
	}
	cksumElem, err := asn1.MarshalWithParams(messages.Checksum{CKSumType: cksumType, Checksum: sum}, "explicit,tag:2")
	if err != nil {
		return messages.PAData{}, err
	}
	authPkgElem, err := asn1.Marshal(messages.ExplicitGeneralString(3, authPackageKerberos))
	if err != nil {
		return messages.PAData{}, err
	}

	body := make([]byte, 0, len(userNameElem)+len(userRealmElem)+len(cksumElem)+len(authPkgElem))
	body = append(body, userNameElem...)
	body = append(body, userRealmElem...)
	body = append(body, cksumElem...)
	body = append(body, authPkgElem...)

	seq, err := asn1.Marshal(asn1.RawValue{Class: asn1.ClassUniversal, Tag: asn1.TagSequence, IsCompound: true, Bytes: body})
	if err != nil {
		return messages.PAData{}, err
	}
	return messages.PAData{PADataType: iana.PAForUser, PADataValue: seq}, nil
}

// PAForUser is the decoded content of a PA-FOR-USER element.
type PAForUser struct {
	UserName    messages.PrincipalName
	UserRealm   string
	Cksum       messages.Checksum
	AuthPackage string
}

type paForUserInner struct {
	UserName    messages.PrincipalName `asn1:"explicit,tag:0"`
	UserRealm   string                 `asn1:"explicit,tag:1,generalstring"`
	Cksum       messages.Checksum      `asn1:"explicit,tag:2"`
	AuthPackage string                 `asn1:"explicit,tag:3,generalstring"`
}

// ParsePAForUser decodes a PA-FOR-USER padata-value.
func ParsePAForUser(b []byte) (*PAForUser, error) {
	var inner paForUserInner
	if _, err := asn1.Unmarshal(b, &inner); err != nil {
		return nil, fmt.Errorf("sfu: parse PA-FOR-USER: %w", err)
	}
	return &PAForUser{
		UserName:    inner.UserName,
		UserRealm:   inner.UserRealm,
		Cksum:       inner.Cksum,
		AuthPackage: inner.AuthPackage,
	}, nil
}

// VerifyPAForUser recomputes the PA-FOR-USER checksum with the given TGT session
// key and reports whether it matches, defending against tampering with the
// impersonated identity.
func VerifyPAForUser(p *PAForUser, sessionKey []byte, sessionKeyEType int) bool {
	cksumType, ok := kerbcrypto.ChecksumTypeForEType(sessionKeyEType)
	if !ok || cksumType != p.Cksum.CKSumType {
		return false
	}
	buf := s4uByteArray(p.UserName, p.UserRealm, p.AuthPackage)
	return kerbcrypto.VerifyChecksum(cksumType, sessionKey, iana.KeyUsageKerbNonKerbCksumSalt, buf, p.Cksum.Checksum)
}
