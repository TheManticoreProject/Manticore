# Kerberos live integration tests

The native (stdlib-only) Kerberos paths have an env-gated live test harness that
runs against a real KDC / domain controller. Every test is behind the
`//go:build integration` tag, so it is excluded from the default build and from
`go test ./...`; it also **skips cleanly** when its configuration is not present
in the environment. Running the suite without configuration is a no-op.

## Configuration

Baseline (drives the always-on paths across every subsystem):

| Variable           | Meaning                                                        |
| ------------------ | ------------------------------------------------------------- |
| `KRB5_TEST_KDC`    | KDC / domain-controller host or IP (obtains the TGT)          |
| `KRB5_TEST_REALM`  | Kerberos realm / AD domain (e.g. `EXAMPLE.LOCAL`)             |
| `KRB5_TEST_USER`   | account sAMAccountName (e.g. `Administrator`)                  |
| `KRB5_TEST_PASS`   | account password                                              |
| `KRB5_TEST_SPN`    | an existing SPN the account may request (e.g. `ldap/dc.example.local`) — used by GetTGS / Kerberoast / pass-the-ticket |
| `KRB5_TEST_TARGET` | SMB / LDAP / RPC server FQDN (defaults to `KRB5_TEST_KDC`); the SMB/LDAP/RPC SPNs are built from it, so a fully-qualified name is required |

Optional (each unlocks a path a vanilla single-domain lab cannot exercise; the
test skips when its variable is unset):

| Variable                | Unlocks                                                  |
| ----------------------- | -------------------------------------------------------- |
| `KRB5_TEST_KEYTAB`      | keytab-based `GetTGT`                                     |
| `KRB5_TEST_ASREP_USER`  | `ASREPRoast` (needs a pre-auth-disabled account)         |
| `KRB5_TEST_FAST`        | FAST-armored AS-REQ (needs KDC Kerberos-armoring support) |
| `KRB5_TEST_S4U_SPN`     | `S4U2Self` → `S4U2Proxy` (needs constrained delegation)  |
| `KRB5_TEST_S4U_USER`    | user to impersonate for S4U (defaults to `KRB5_TEST_USER`) |
| `KRB5_TEST_KRBTGT_KEY` + `KRB5_TEST_DOMAIN_SID` | golden-ticket forge → use       |
| `KRB5_TEST_SMB21`       | SMB 2.1 signing (most Windows DCs reject 2.x — target a member server) |

No secrets are ever committed; the harness reads everything from the environment.

## Running

```sh
KRB5_TEST_KDC=10.0.0.10 \
KRB5_TEST_REALM=EXAMPLE.LOCAL \
KRB5_TEST_USER=Administrator \
KRB5_TEST_PASS='…' \
KRB5_TEST_SPN=ldap/dc.example.local \
KRB5_TEST_TARGET=dc.example.local \
  go test -tags integration -v \
    ./network/kerberos/v5/... \
    ./network/smb/smb_v20/client/ \
    ./network/ldap/ \
    ./network/dcerpc/v5/client/
```

## Coverage

- **network/kerberos/v5** — `GetTGT` (password / NT-hash / AES-key / keytab),
  `GetTGS`, Kerberoast, ASREPRoast, U2U, `Renew`, kirbi + ccache export→import
  pass-the-ticket, golden forge→use, S4U2Self→S4U2Proxy, FAST-armored AS-REQ, and
  a live PAC decode (U2U-to-self ticket decrypted with the client's TGT session
  key, PAC parsed).
- **network/smb/smb_v20/client** — `SessionSetupKerberos` with an SMB 3.1.1 signed
  (AES-128-CMAC) tree connect and an AES-GCM-encrypted tree connect, plus a gated
  SMB 2.1 (HMAC-SHA256) signing test.
- **network/ldap** — GSSAPI/Kerberos bind with the integrity (sign) and
  confidentiality (seal) security layers, each followed by a real search.
- **network/dcerpc/v5/client** — `SetAuthKerberos` at CONNECT / PKT /
  PKT_INTEGRITY / PKT_PRIVACY, each with a protected `ept_lookup` against the
  endpoint mapper.

Cross-realm and PKINIT paths are intentionally out of scope for a single-domain
lab and are not exercised here.
