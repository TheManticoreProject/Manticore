# Reference corpora

Each file records one SMB exchange with a real peer: the messages in order, with
the direction each travelled, as hexadecimal. The Direct TCP framing is stripped
([MS-SMB2] 2.1), so a corpus holds SMB messages and is unaffected by how they
were split across TCP segments.

| Corpus | Peer | Dialect |
|---|---|---|
| `smb311-server2025.json` | Windows Server 2025 Standard | SMB 3.1.1, signing required |
| `smb302-server2012r2.json` | Windows Server 2012 R2 Standard 9600 | SMB 3.0.2, signing required |
| `smb1-server2012r2.json` | Windows Server 2012 R2 Standard 9600 | SMB 1.0 (NT LM 0.12), Unicode |

## What they are for

Pairing this repository's client with its own server proves very little: a wire
detail both halves get wrong agrees with itself, and every round trip passes.
Real defects have hidden in that gap — entry names carrying their terminator, a
Unicode path a byte out of phase, a credit charge no server would accept — each
found only once a third party was on the other end.

Two checks run against a corpus offline, in CI, with no network and no lab:

- **What the client emits**, byte for byte, against what it emitted when the
  exchange was recorded, ignoring the fields that legitimately differ between
  runs. This is a regression guard on this implementation's own output.
- **That the parsers accept every message the peer sent.** Those bytes were
  produced by Windows rather than by this code, so this is the check with
  independent authority.

A corpus is *not* a parity reference. The client messages in it were emitted by
this implementation, not by Windows, so they say nothing about whether this
client resembles a Windows client. Establishing that needs a capture of a
Windows client talking to a server, which these are not.

## Re-recording

    tcpdump -i <iface> -s 0 -w session.pcap "host <peer> and tcp port 445"
    # drive the exchange, then stop the capture
    ./extract.py session.pcap <peer-ip> <name> "<peer description>" > <name>.json

`tcpdump` is used rather than `dumpcap` because it is commonly installed setuid
root and world-executable, where `dumpcap` is usually restricted to a group.

Corpora are cheap to regenerate and should be re-recorded whenever the peer is
upgraded; the peer description records what produced them.
