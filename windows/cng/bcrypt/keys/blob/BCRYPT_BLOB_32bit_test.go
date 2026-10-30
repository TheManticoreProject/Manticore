package blob

import (
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys/headers"
)

// TestUnmarshalRejectsSizesAbove2GiB checks that every blob parser rejects a
// header size >= 2^31. On 32-bit builds int() of such a size is negative, which
// used to pass the "size > remaining" bounds check and panic on the slice.
func TestUnmarshalRejectsSizesAbove2GiB(t *testing.T) {
	const big = 0xFFFFFFF0
	value := make([]byte, 32)

	rsaHeaders := map[string]headers.BCRYPT_RSA_KEY_BLOB{
		"CbPublicExp": {CbPublicExp: big},
		"CbModulus":   {CbModulus: big},
		"CbPrime1":    {CbPrime1: big},
		"CbPrime2":    {CbPrime2: big},
	}

	cases := map[string]func() (int, error){
		"ECC_PUBLIC": func() (int, error) {
			return (&BCRYPT_ECC_PUBLIC_BLOB{}).Unmarshal(headers.BCRYPT_ECC_KEY_BLOB{KeySize: big}, value)
		},
		"ECC_PRIVATE": func() (int, error) {
			return (&BCRYPT_ECC_PRIVATE_BLOB{}).Unmarshal(headers.BCRYPT_ECC_KEY_BLOB{KeySize: big}, value)
		},
		"DSA_PUBLIC": func() (int, error) {
			return (&BCRYPT_DSA_PUBLIC_BLOB{}).Unmarshal(headers.BCRYPT_DSA_KEY_BLOB{CbKey: big}, value)
		},
		"DSA_PRIVATE": func() (int, error) {
			return (&BCRYPT_DSA_PRIVATE_BLOB{}).Unmarshal(headers.BCRYPT_DSA_KEY_BLOB{CbKey: big}, value)
		},
	}
	for field, h := range rsaHeaders {
		h := h
		if field == "CbPublicExp" || field == "CbModulus" {
			// The public blob carries no primes.
			cases["RSA_PUBLIC/"+field] = func() (int, error) { return (&BCRYPT_RSA_PUBLIC_BLOB{}).Unmarshal(h, value) }
		}
		cases["RSA_PRIVATE/"+field] = func() (int, error) { return (&BCRYPT_RSA_PRIVATE_BLOB{}).Unmarshal(h, value) }
		cases["RSA_FULLPRIVATE/"+field] = func() (int, error) { return (&BCRYPT_RSA_FULLPRIVATE_BLOB{}).Unmarshal(h, value) }
	}

	for name, unmarshal := range cases {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Unmarshal panicked: %v", r)
				}
			}()
			if _, err := unmarshal(); err == nil {
				t.Fatal("expected error for a size >= 2^31, got nil")
			}
		})
	}
}
