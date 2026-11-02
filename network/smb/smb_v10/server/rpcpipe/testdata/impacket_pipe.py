"""Drive srvsvc and wkssvc over SMB1 named pipes with impacket.

impacket writes each RPC PDU with SMB_COM_WRITE_ANDX and reads the answer with
SMB_COM_READ_ANDX instead of using TRANS_TRANSACT_NMPIPE, so this exercises the
server's pipe write and read path. An optional fragment size makes impacket
split each request into several PDUs.

usage: impacket_pipe.py <port> <domain> <user> <password> [max_fragment]
"""
import sys

from impacket.smb import SMB_DIALECT
from impacket.smbconnection import SMBConnection
from impacket.dcerpc.v5 import transport, srvs, wkst

port, domain, user, password = int(sys.argv[1]), sys.argv[2], sys.argv[3], sys.argv[4]
fragment = int(sys.argv[5]) if len(sys.argv) > 5 else 0

smb = SMBConnection("127.0.0.1", "127.0.0.1", sess_port=port, preferredDialect=SMB_DIALECT)
smb.login(user, password, domain)

def bound(pipe, uuid):
    rpc = transport.SMBTransport("127.0.0.1", port, pipe, smb_connection=smb)
    dce = rpc.get_dce_rpc()
    if fragment:
        dce.set_max_fragment_size(fragment)
    dce.connect()
    dce.bind(uuid)
    return dce

dce = bound(r"\srvsvc", srvs.MSRPC_UUID_SRVS)
resp = srvs.hNetrShareEnum(dce, 1)
for share in resp["InfoStruct"]["ShareInfo"]["Level1"]["Buffer"]:
    print("share", share["shi1_netname"][:-1])
dce.disconnect()

dce = bound(r"\wkssvc", wkst.MSRPC_UUID_WKST)
resp = wkst.hNetrWkstaGetInfo(dce, 100)
print("computer", resp["WkstaInfo"]["WkstaInfo100"]["wki100_computername"][:-1])
print("domain", resp["WkstaInfo"]["WkstaInfo100"]["wki100_langroup"][:-1])
dce.disconnect()
print("OK")
