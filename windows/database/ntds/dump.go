package ntds

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"github.com/TheManticoreProject/Manticore/windows/database/ese"
)

// NTDS datatable column names. In NTDS.dit each attribute is an ESE column named
// "ATT<type><id>"; these are the columns needed to recover account secrets.
const (
	AttSAMAccountName          = "ATTm590045"
	AttSAMAccountType          = "ATTj590126"
	AttObjectSid               = "ATTr589970"
	AttUserAccountControl      = "ATTj589832"
	AttUnicodePwd              = "ATTk589914" // NT hash
	AttDBCSPwd                 = "ATTk589879" // LM hash
	AttNTPwdHistory            = "ATTk589918"
	AttLMPwdHistory            = "ATTk589984"
	AttPEKList                 = "ATTk590689"
	AttUserPrincipalName       = "ATTm590480"
	AttSupplementalCredentials = "ATTk589949"
)

// uacAccountDisable is the ADS_UF_ACCOUNTDISABLE bit of userAccountControl.
const uacAccountDisable = 0x0002

// sAMAccountType values that identify a security principal carrying password material
// (user / machine / trust). Group and alias objects have other values and are skipped.
const (
	samNormalUserAccount = 0x30000000
	samMachineAccount    = 0x30000001
	samTrustAccount      = 0x30000002
)

func isAccountType(t uint32) bool {
	return t == samNormalUserAccount || t == samMachineAccount || t == samTrustAccount
}

func hexBytes(s string) []byte { b, _ := hex.DecodeString(s); return b }

// Empty-password hashes, used when an account has no stored NT/LM hash (matching
// impacket's NTOWFv1(”,”) / LMOWFv1(”,”)).
var (
	emptyNTHash = hexBytes("31d6cfe0d16ae931b73c59d7e0c089c0")
	emptyLMHash = hexBytes("aad3b435b51404eeaad3b435b51404ee")
)

// Account is one decrypted NTDS account.
type Account struct {
	SAMAccountName     string
	RID                uint32
	LMHash             []byte // 16 bytes
	NTHash             []byte // 16 bytes
	LMHistory          [][]byte
	NTHistory          [][]byte
	UserAccountControl uint32
	HasUAC             bool
}

// Disabled reports whether the account's userAccountControl marks it disabled.
func (a *Account) Disabled() bool {
	return a.HasUAC && a.UserAccountControl&uacAccountDisable != 0
}

// SecretsdumpLine formats the account as the secretsdump line "user:rid:lm:nt:::".
func (a *Account) SecretsdumpLine() string {
	return fmt.Sprintf("%s:%d:%s:%s:::", a.SAMAccountName, a.RID,
		hex.EncodeToString(a.LMHash), hex.EncodeToString(a.NTHash))
}

// HistoryLines formats the account's password-history entries as secretsdump
// "user_historyN:rid:lm:nt:::" lines, skipping the current password (index 0).
func (a *Account) HistoryLines() []string {
	var lines []string
	n := len(a.NTHistory)
	if len(a.LMHistory) > n {
		n = len(a.LMHistory)
	}
	for i := 1; i < n; i++ {
		lm, nt := emptyLMHash, emptyNTHash
		if i < len(a.LMHistory) {
			lm = a.LMHistory[i]
		}
		if i < len(a.NTHistory) {
			nt = a.NTHistory[i]
		}
		lines = append(lines, fmt.Sprintf("%s_history%d:%d:%s:%s:::", a.SAMAccountName, i-1, a.RID,
			hex.EncodeToString(lm), hex.EncodeToString(nt)))
	}
	return lines
}

// ridFromSID extracts the RID (last sub-authority) from a binary objectSid. NTDS stores
// the objectSid with its sub-authorities big-endian, so the RID is the big-endian uint32
// of the final four bytes.
func ridFromSID(sid []byte) uint32 {
	if len(sid) < 4 {
		return 0
	}
	return binary.BigEndian.Uint32(sid[len(sid)-4:])
}

// FindPEKList scans the datatable for the pekList attribute and decrypts it with bootKey.
func FindPEKList(table *ese.Table, bootKey []byte) ([]PEK, error) {
	cur, err := table.Rows()
	if err != nil {
		return nil, err
	}
	for cur.Next() {
		if blob, ok := cur.Row().Raw(AttPEKList); ok {
			return DecryptPEKList(bootKey, blob)
		}
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("ntds: pekList attribute not found in datatable")
}

// Dump iterates the datatable of an opened NTDS database, decrypts each user account's
// secrets with the PEK list (recovered from bootKey), and invokes fn for every account
// that has a sAMAccountName and objectSid. bootKey is the 16-byte SYSTEM boot key (the
// caller derives it).
func Dump(db *ese.Database, bootKey []byte, fn func(Account) error) error {
	table, err := db.Table("datatable")
	if err != nil {
		return err
	}
	peks, err := FindPEKList(table, bootKey)
	if err != nil {
		return err
	}

	cur, err := table.Rows()
	if err != nil {
		return err
	}
	for cur.Next() {
		row := cur.Row()
		// Only objects that are security principals with password material (user /
		// machine / trust accounts); skip groups, aliases, and other objects.
		accountType, ok := row.Uint32(AttSAMAccountType)
		if !ok || !isAccountType(accountType) {
			continue
		}
		sid, ok := row.Raw(AttObjectSid)
		if !ok {
			continue
		}
		sam, ok := row.String(AttSAMAccountName)
		if !ok || sam == "" {
			continue
		}
		acct := decryptAccount(peks, row, sid, sam)
		if err := fn(acct); err != nil {
			return err
		}
	}
	return cur.Err()
}

// decryptAccount builds an Account from a datatable row, decrypting its hash attributes.
// Missing hashes default to the empty-password hash.
func decryptAccount(peks []PEK, row *ese.Row, sid []byte, sam string) Account {
	rid := ridFromSID(sid)
	acct := Account{SAMAccountName: sam, RID: rid, NTHash: emptyNTHash, LMHash: emptyLMHash}

	if blob, ok := row.Raw(AttUnicodePwd); ok {
		if h, err := DecryptHash(peks, rid, blob); err == nil {
			acct.NTHash = h
		}
	}
	if blob, ok := row.Raw(AttDBCSPwd); ok {
		if h, err := DecryptHash(peks, rid, blob); err == nil {
			acct.LMHash = h
		}
	}
	if uac, ok := row.Uint32(AttUserAccountControl); ok {
		acct.UserAccountControl = uac
		acct.HasUAC = true
	}
	if blob, ok := row.Raw(AttNTPwdHistory); ok {
		if h, err := DecryptHashHistory(peks, rid, blob); err == nil {
			acct.NTHistory = h
		}
	}
	if blob, ok := row.Raw(AttLMPwdHistory); ok {
		if h, err := DecryptHashHistory(peks, rid, blob); err == nil {
			acct.LMHistory = h
		}
	}
	return acct
}
