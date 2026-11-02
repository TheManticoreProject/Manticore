package bcrypt

import (
	"bytes"
	"crypto/elliptic"
	"encoding/base64"
	"math/big"
	"testing"

	"github.com/TheManticoreProject/Manticore/windows/cng/bcrypt/keys"
)

// windowsBlobs are key blobs exported by Windows itself: BCryptGenerateKeyPair
// then BCryptExportKey (ECCPUBLICBLOB/ECCPRIVATEBLOB, DSAPUBLICBLOB/
// DSAPRIVATEBLOB) on Windows Server 2012 R2, for throwaway keys destroyed right
// after the export.
var windowsBlobs = []struct{ kind, alg, b64 string }{
	{"PUB", "ECDSA_P256", "RUNTMSAAAACaVTXbta3ZswKIWiYRA/vcgcpPoWQ3sVZqRyPmbj9+LChD6e535xMLDPicwl38moDQ41KkQosczpMofL5Mmfpv"},
	{"PRV", "ECDSA_P256", "RUNTMiAAAACaVTXbta3ZswKIWiYRA/vcgcpPoWQ3sVZqRyPmbj9+LChD6e535xMLDPicwl38moDQ41KkQosczpMofL5Mmfpv6KJN9JckLnPeLPyoQK1a9GlURdUbu9hw370U+ujMuv0="},
	{"PUB", "ECDH_P384", "RUNLMzAAAAB1vWVT0jYnajzgj06YluWO0SdFMjrugt/Hp1liwKLzcz8tYvKcLgsFxD1FGxJuNY/YnGGVftfXnqJb6fDim4/E1ImMVUCSuVp2Lwlfc3swQUb/xMkkVIDZOvSsP688rm8="},
	{"PRV", "ECDH_P384", "RUNLNDAAAAB1vWVT0jYnajzgj06YluWO0SdFMjrugt/Hp1liwKLzcz8tYvKcLgsFxD1FGxJuNY/YnGGVftfXnqJb6fDim4/E1ImMVUCSuVp2Lwlfc3swQUb/xMkkVIDZOvSsP688rm+AhU+E1SrPpPjKSjL/kkmeHUHD1Pq7g2/17TB6VoE0mdeCjq2ryrq+EKds3gJsVhk="},
	{"PUB", "ECDSA_P521", "RUNTNUIAAAABeQh+7Lr37pwEzibkDbcFNWypory39RRJab8OqTfyhFak1RuTCut/KxFyriGthRFr+2OPLDArg0mI2ad9JT997akAIRl0BLD5cunRcEekVXuMQ/rEFV7FYUZk/Tlw9cAeHnVWHo26Z04V+Cyg356AV6vLNUrUe2v/WqRIz6w45bvcnvc="},
	{"PRV", "ECDSA_P521", "RUNTNkIAAAABeQh+7Lr37pwEzibkDbcFNWypory39RRJab8OqTfyhFak1RuTCut/KxFyriGthRFr+2OPLDArg0mI2ad9JT997akAIRl0BLD5cunRcEekVXuMQ/rEFV7FYUZk/Tlw9cAeHnVWHo26Z04V+Cyg356AV6vLNUrUe2v/WqRIz6w45bvcnvcBOfHstT9S4RyL3cZ2lYWRAVr7YfXgqbmuUvspPn0gY/u6g6+ck1bxjNLCqWPGcLJIVWIDcg2XpKKOW1jtNxsazPM="},
	{"PUB", "DSA", "RFNQQoAAAAAAAADmy0MpsNAPy+LkG5I+G8Ia0xM9KdP9IqoLmLIK7n9tvL5LLaFjBgv9L5d8t4l30EZ0z3JnElsSm1B971VpF6/AQPJ9hDHYuinP5FexTipzGBpuh8QRS7HR3OLR0Rue4zrOQ4/Ux9khmF9nnhlk5VtFTQ9j/uB5ZLP6FtC4Jm5l7sQh8J09pPIrRiaMlIGqdceioCedytxf0bXpGDi4ECcbucCtQ8xxlVepPlhKk/ex16LinZremEClKM5cua2ly93AtvCVBMtt7PTV/e+sGL+8hXkRuzWaByWoO3sqzpIXvbBbA0v9VqUyr7cOiaBRRwdze9Lqi06VMALtOKNFMavzmJFWT4et2y8mcQ0WUCRQheKU9JraJbnHWtR5P0tXZ972mF5zf9OPY5RUvV83b0Br/Eeum/QTPvYdRGKtlAh2xw46+zCdZRgrswTN7rD9JM9mfXP987bhPZNgBj+niGpT9vB/sDMD55ZJH41BchbC+uW0jNazUzyZyWb0aE7ao3PImJd2DqdRcyZI9M+l5PBQtipi4t0JjBliVPalODdtqQ94vKntRwrqAw=="},
	{"PRV", "DSA", "RFNQVoAAAAAAAADmy0MpsNAPy+LkG5I+G8Ia0xM9KdP9IqoLmLIK7n9tvL5LLaFjBgv9L5d8t4l30EZ0z3JnElsSm1B971VpF6/AQPJ9hDHYuinP5FexTipzGBpuh8QRS7HR3OLR0Rue4zrOQ4/Ux9khmF9nnhlk5VtFTQ9j/uB5ZLP6FtC4Jm5l7sQh8J09pPIrRiaMlIGqdceioCedytxf0bXpGDi4ECcbucCtQ8xxlVepPlhKk/ex16LinZremEClKM5cua2ly93AtvCVBMtt7PTV/e+sGL+8hXkRuzWaByWoO3sqzpIXvbBbA0v9VqUyr7cOiaBRRwdze9Lqi06VMALtOKNFMavzmJFWT4et2y8mcQ0WUCRQheKU9JraJbnHWtR5P0tXZ972mF5zf9OPY5RUvV83b0Br/Eeum/QTPvYdRGKtlAh2xw46+zCdZRgrswTN7rD9JM9mfXP987bhPZNgBj+niGpT9vB/sDMD55ZJH41BchbC+uW0jNazUzyZyWb0aE7ao3PImJd2DqdRcyZI9M+l5PBQtipi4t0JjBliVPalODdtqQ94vKntRwrqA1PHCbAKFbIKHRZGAnel9OsH1WEz"},
}

func curveFor(t *testing.T, alg string) elliptic.Curve {
	t.Helper()
	switch alg {
	case "ECDSA_P256", "ECDH_P256":
		return elliptic.P256()
	case "ECDSA_P384", "ECDH_P384":
		return elliptic.P384()
	case "ECDSA_P521", "ECDH_P521":
		return elliptic.P521()
	}
	t.Fatalf("no curve for %s", alg)
	return nil
}

// TestUnmarshalKeyMaterialWindowsBlobs checks every ECC and DSA blob Windows
// produced decodes through UnmarshalKeyMaterial into the right key type,
// consumes the whole blob, re-marshals byte for byte, and holds a consistent
// key: an ECC public point on its curve, an ECC private scalar d with d*G equal
// to the public point, and a DSA public value y = g^x mod p.
func TestUnmarshalKeyMaterialWindowsBlobs(t *testing.T) {
	for _, tc := range windowsBlobs {
		raw, err := base64.StdEncoding.DecodeString(tc.b64)
		if err != nil {
			t.Fatal(err)
		}
		km, n, err := UnmarshalKeyMaterial(raw)
		if err != nil {
			t.Errorf("%s %s: UnmarshalKeyMaterial: %v", tc.kind, tc.alg, err)
			continue
		}
		if n != len(raw) {
			t.Errorf("%s %s: consumed %d of %d bytes", tc.kind, tc.alg, n, len(raw))
		}
		out, err := km.Marshal()
		if err != nil || !bytes.Equal(out, raw) {
			t.Errorf("%s %s: re-marshal differs (err %v)", tc.kind, tc.alg, err)
		}

		switch k := km.(type) {
		case *keys.BCRYPT_ECC_PUBLIC_KEY:
			c := curveFor(t, tc.alg)
			x, y := new(big.Int).SetBytes(k.Content.X), new(big.Int).SetBytes(k.Content.Y)
			if tc.kind != "PUB" || !c.IsOnCurve(x, y) {
				t.Errorf("%s %s: public point not on %s", tc.kind, tc.alg, c.Params().Name)
			}
		case *keys.BCRYPT_ECC_PRIVATE_KEY:
			c := curveFor(t, tc.alg)
			gx, gy := c.ScalarBaseMult(k.Content.D)
			if tc.kind != "PRV" || gx.Cmp(new(big.Int).SetBytes(k.Content.X)) != 0 || gy.Cmp(new(big.Int).SetBytes(k.Content.Y)) != 0 {
				t.Errorf("%s %s: d*G is not the public point", tc.kind, tc.alg)
			}
		case *keys.BCRYPT_DSA_PUBLIC_KEY:
			if tc.kind != "PUB" || len(k.Content.Modulus) != 128 {
				t.Errorf("%s %s: DSA public key with %d-byte p", tc.kind, tc.alg, len(k.Content.Modulus))
			}
		case *keys.BCRYPT_DSA_PRIVATE_KEY:
			p := new(big.Int).SetBytes(k.Content.Modulus)
			g := new(big.Int).SetBytes(k.Content.Generator)
			x := new(big.Int).SetBytes(k.Content.PrivateExponent[:])
			if tc.kind != "PRV" || new(big.Int).Exp(g, x, p).Cmp(new(big.Int).SetBytes(k.Content.Public)) != 0 {
				t.Errorf("%s %s: g^x mod p is not y", tc.kind, tc.alg)
			}
		default:
			t.Errorf("%s %s: decoded as %T", tc.kind, tc.alg, km)
		}
	}
}
