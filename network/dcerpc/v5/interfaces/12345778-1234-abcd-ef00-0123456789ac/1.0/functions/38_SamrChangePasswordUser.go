package functions

import (
	"fmt"

	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/client"
	samr "github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0"
	"github.com/TheManticoreProject/Manticore/network/dcerpc/v5/interfaces/12345778-1234-abcd-ef00-0123456789ac/1.0/structures"
)

// samrChangePasswordUserRequest carries the [in] user handle and the (mutually
// cross-encrypted) LM and NT password OWFs of SamrChangePasswordUser, in exact IDL field
// order. Each [in,unique] OWF pointer is optional; the *Present flags indicate which arms
// are supplied.
type samrChangePasswordUserRequest struct {
	UserHandle               structures.SAMPR_HANDLE
	LmPresent                uint8
	OldLmEncryptedWithNewLm  *structures.ENCRYPTED_LM_OWF_PASSWORD `ndr:"unique"`
	NewLmEncryptedWithOldLm  *structures.ENCRYPTED_LM_OWF_PASSWORD `ndr:"unique"`
	NtPresent                uint8
	OldNtEncryptedWithNewNt  *structures.ENCRYPTED_NT_OWF_PASSWORD `ndr:"unique"`
	NewNtEncryptedWithOldNt  *structures.ENCRYPTED_NT_OWF_PASSWORD `ndr:"unique"`
	NtCrossEncryptionPresent uint8
	NewNtEncryptedWithNewLm  *structures.ENCRYPTED_NT_OWF_PASSWORD `ndr:"unique"`
	LmCrossEncryptionPresent uint8
	NewLmEncryptedWithNewNt  *structures.ENCRYPTED_LM_OWF_PASSWORD `ndr:"unique"`
}

func (*samrChangePasswordUserRequest) Opnum() uint16 { return samr.OpnumSamrChangePasswordUser }

// SamrChangePasswordUser calls SamrChangePasswordUser (opnum 38), changing a user's password
// given the old and new LM/NT OWFs cross-encrypted with one another ([MS-SAMR] 3.1.5.10.2).
func SamrChangePasswordUser(rpc *client.Client, userHandle structures.SAMPR_HANDLE, lmPresent uint8, oldLmEncryptedWithNewLm *structures.ENCRYPTED_LM_OWF_PASSWORD, newLmEncryptedWithOldLm *structures.ENCRYPTED_LM_OWF_PASSWORD, ntPresent uint8, oldNtEncryptedWithNewNt *structures.ENCRYPTED_NT_OWF_PASSWORD, newNtEncryptedWithOldNt *structures.ENCRYPTED_NT_OWF_PASSWORD, ntCrossEncryptionPresent uint8, newNtEncryptedWithNewLm *structures.ENCRYPTED_NT_OWF_PASSWORD, lmCrossEncryptionPresent uint8, newLmEncryptedWithNewNt *structures.ENCRYPTED_LM_OWF_PASSWORD) error {
	req := &samrChangePasswordUserRequest{
		UserHandle:               userHandle,
		LmPresent:                lmPresent,
		OldLmEncryptedWithNewLm:  oldLmEncryptedWithNewLm,
		NewLmEncryptedWithOldLm:  newLmEncryptedWithOldLm,
		NtPresent:                ntPresent,
		OldNtEncryptedWithNewNt:  oldNtEncryptedWithNewNt,
		NewNtEncryptedWithOldNt:  newNtEncryptedWithOldNt,
		NtCrossEncryptionPresent: ntCrossEncryptionPresent,
		NewNtEncryptedWithNewLm:  newNtEncryptedWithNewLm,
		LmCrossEncryptionPresent: lmCrossEncryptionPresent,
		NewLmEncryptedWithNewNt:  newLmEncryptedWithNewNt,
	}
	var resp statusResponse
	if err := rpc.Invoke(req, &resp); err != nil {
		return fmt.Errorf("SamrChangePasswordUser: %w", err)
	}
	if uint32(resp.Status) != samr.StatusSuccess {
		return fmt.Errorf("SamrChangePasswordUser failed: %s", samr.StatusString(uint32(resp.Status)))
	}
	return nil
}
