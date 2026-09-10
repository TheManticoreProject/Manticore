#!/usr/bin/env python3
"""Turn a packet capture of an SMB session into a reference corpus.

The corpus records the SMB messages of one exchange, in order, with the
direction each travelled. It is the input to the wirediff harness, which
replays the peer's half back to this implementation's client and compares what
the client emits against what was recorded here.

Usage:
    tcpdump -i <iface> -s 0 -w session.pcap "host <peer> and tcp port 445"
    ./extract.py session.pcap <peer-ip> <name> <peer-description> > corpus.json

The Direct TCP framing (a zero byte and a 24-bit length, [MS-SMB2] 2.1) is
stripped: a corpus holds SMB messages, not transport frames, so a change in how
messages are split across TCP segments does not alter it.
"""
import json
import subprocess
import sys


def payloads(pcap, peer):
    """Yield (from_client, payload_bytes) for every TCP segment carrying data."""
    out = subprocess.run(
        ["tshark", "-r", pcap, "-Y", "tcp.len>0", "-T", "fields",
         "-e", "ip.src", "-e", "tcp.payload"],
        capture_output=True, text=True, check=True,
    ).stdout
    for line in out.splitlines():
        parts = line.split("\t")
        if len(parts) != 2 or not parts[1]:
            continue
        src, hexbytes = parts
        # tshark emits a comma-separated list when a frame carries several PDUs.
        raw = bytes.fromhex(hexbytes.replace(",", "").replace(":", ""))
        yield src != peer, raw


def messages(stream):
    """Split a reassembled direction into SMB messages on the Direct TCP header."""
    offset = 0
    while offset + 4 <= len(stream):
        if stream[offset] != 0x00:
            raise SystemExit(f"unexpected Direct TCP header byte {stream[offset]:#04x} at {offset}")
        length = int.from_bytes(stream[offset + 1:offset + 4], "big")
        if offset + 4 + length > len(stream):
            break  # a trailing partial message, e.g. the capture was cut short
        yield stream[offset + 4:offset + 4 + length]
        offset += 4 + length


def main():
    if len(sys.argv) != 5:
        raise SystemExit(__doc__)
    pcap, peer, name, description = sys.argv[1:5]

    # Reassemble each direction, then split; interleave by preserving arrival order.
    order, streams = [], {True: bytearray(), False: bytearray()}
    for from_client, raw in payloads(pcap, peer):
        streams[from_client] += raw
        order.append(from_client)

    split = {d: list(messages(bytes(s))) for d, s in streams.items()}
    cursor = {True: 0, False: 0}

    frames, seen = [], set()
    for from_client in order:
        if cursor[from_client] >= len(split[from_client]):
            continue
        msg = split[from_client][cursor[from_client]]
        cursor[from_client] += 1
        key = (from_client, cursor[from_client])
        if key in seen:
            continue
        seen.add(key)
        frames.append({
            "direction": "client" if from_client else "server",
            "message": msg.hex(),
        })

    json.dump({
        "name": name,
        "peer": description,
        "frames": frames,
    }, sys.stdout, indent=2)
    print()


if __name__ == "__main__":
    main()
