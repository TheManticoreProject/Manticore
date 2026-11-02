"""Drive srvsvc and wkssvc over SMB1 named pipes with impacket.

impacket writes each RPC PDU with SMB_COM_WRITE_ANDX and reads the answer with
SMB_COM_READ_ANDX instead of using TRANS_TRANSACT_NMPIPE, so this exercises the
server's pipe write and read path. An optional fragment size makes impacket
split each request into several PDUs.

An optional RPC auth level (2 connect, 5 packet integrity, 6 packet privacy)
makes each bind NTLM-authenticated (RPC_C_AUTHN_WINNT) with the same
credentials, and every request and response signed or sealed at that level.

An optional RPC password, used for the RPC binds only, lets a caller check that
a wrong one is refused while the SMB login still succeeds.

usage: impacket_pipe.py <port> <domain> <user> <password> [max_fragment] [auth_level] [rpc_password]
"""
import sys

from impacket.smb import SMB_DIALECT
from impacket.smbconnection import SMBConnection
from impacket.dcerpc.v5 import transport, srvs, wkst

port, domain, user, password = int(sys.argv[1]), sys.argv[2], sys.argv[3], sys.argv[4]
fragment = int(sys.argv[5]) if len(sys.argv) > 5 else 0
auth_level = int(sys.argv[6]) if len(sys.argv) > 6 else 0
rpc_password = sys.argv[7] if len(sys.argv) > 7 else password

smb = SMBConnection("127.0.0.1", "127.0.0.1", sess_port=port, preferredDialect=SMB_DIALECT)
smb.login(user, password, domain)

def bound(pipe, uuid):
    rpc = transport.SMBTransport("127.0.0.1", port, pipe, smb_connection=smb)
    dce = rpc.get_dce_rpc()
    if auth_level:
        rpc.set_credentials(user, rpc_password, domain)
        dce.set_auth_type(10)  # RPC_C_AUTHN_WINNT
        dce.set_auth_level(auth_level)
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
