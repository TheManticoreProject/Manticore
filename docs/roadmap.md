# Manticore module roadmap

This graph mirrors the package tree on disk (top-level modules → sub-packages) and
color-codes each module by its advancement state.

**Legend**

| Color | Meaning |
| :--- | :--- |
| 🟢 green | Implemented and verified — real code exercised by real tests. |
| 🟠 orange | Partial / unverified — code present, but minimal (stub) or without meaningful test coverage. |
| 🔴 red | Not implemented — directory reserved but empty, or explicitly planned-only. |

**Propagation rule:** a parent takes the *worst* color of any of its descendants. If any
submodule is orange or red, the parent is at best that color. This is why the top-level
modules and the `Manticore` root read red/orange even though most leaves are green: the
tree still contains planned-but-empty subsystems (e.g. SMB 2.1+, raw DNS/IP/TCP, NBF).

**How colors are assigned (per directory, from its own `.go` files):**
- 🔴 red — no production Go code at all (empty leaf / planned).
- 🟢 green — ≥ 30 lines of production code **and** ≥ 20 lines of test code.
- 🟠 orange — otherwise (stub-sized code, or code without a real test).

A small set of headline modules carry explicit states from the human-maintained roadmap
(e.g. SMB versions, LDAP, Kerberos, raw DNS/IP/TCP) that override the metric; the
`docs/gen_roadmap_graph.py` generator is the source of truth and can regenerate this file.
The DCE/RPC UUID-keyed interface registry and the mechanical NDR `structures/`/`functions/`
partitions are collapsed for readability (they still count toward their parent's color).

```mermaid
graph LR
  root["Manticore 🔴"]:::c0
  root --> n_crypto
  root --> n_encoding
  root --> n_logger
  root --> n_network
  root --> n_utils
  root --> n_windows
  n_crypto["crypto 🟠"]:::c1
  n_crypto_aes["aes 🟢"]:::c2
  n_crypto_aes_cfb8["cfb8 🟢"]:::c2
  n_crypto_aes_cts["cts 🟢"]:::c2
  n_crypto_cmac["cmac 🟢"]:::c2
  n_crypto_dcc["dcc 🟢"]:::c2
  n_crypto_dcc2["dcc2 🟢"]:::c2
  n_crypto_gppp["gppp 🟢"]:::c2
  n_crypto_lm["lm 🟢"]:::c2
  n_crypto_md4["md4 🟢"]:::c2
  n_crypto_nfold["nfold 🟢"]:::c2
  n_crypto_nt["nt 🟢"]:::c2
  n_crypto_ntlmv1["ntlmv1 🟢"]:::c2
  n_crypto_ntlmv2["ntlmv2 🟢"]:::c2
  n_crypto_pkcs7["pkcs7 🟢"]:::c2
  n_crypto_rc4["rc4 🟢"]:::c2
  n_crypto_spnego["spnego 🟠"]:::c1
  n_crypto_spnego_ntlm["ntlm 🟠"]:::c1
  n_crypto_spnego_ntlm_avpair["avpair 🟢"]:::c2
  n_crypto_spnego_ntlm_datafields["datafields 🟢"]:::c2
  n_crypto_spnego_ntlm_message["message 🟠"]:::c1
  n_crypto_spnego_ntlm_message_authenticate["authenticate 🟢"]:::c2
  n_crypto_spnego_ntlm_message_challenge["challenge 🟢"]:::c2
  n_crypto_spnego_ntlm_message_header["header 🟢"]:::c2
  n_crypto_spnego_ntlm_message_negotiate["negotiate 🟢"]:::c2
  n_crypto_spnego_ntlm_message_negotiate_flags["flags 🟢"]:::c2
  n_crypto_spnego_ntlm_message_types["types 🟢"]:::c2
  n_crypto_spnego_ntlm_security["security 🟢"]:::c2
  n_crypto_spnego_ntlm_targetinfo["targetinfo 🟢"]:::c2
  n_crypto_spnego_ntlm_version["version 🟢"]:::c2
  n_crypto_uuid["uuid 🟢"]:::c2
  n_crypto_uuid_uuid_v1["uuid_v1 🟢"]:::c2
  n_crypto_uuid_uuid_v2["uuid_v2 🟢"]:::c2
  n_crypto_uuid_uuid_v3["uuid_v3 🟢"]:::c2
  n_crypto_uuid_uuid_v4["uuid_v4 🟢"]:::c2
  n_crypto_uuid_uuid_v5["uuid_v5 🟢"]:::c2
  n_crypto_uuid_uuid_v6["uuid_v6 🟢"]:::c2
  n_crypto_uuid_uuid_v7["uuid_v7 🟢"]:::c2
  n_crypto_uuid_uuid_v8["uuid_v8 🟢"]:::c2
  n_encoding["encoding 🟠"]:::c1
  n_encoding_ascii["ascii 🟠"]:::c1
  n_encoding_ebcdic["ebcdic 🟢"]:::c2
  n_encoding_ebcdic_cp037["cp037 🟢"]:::c2
  n_encoding_ebcdic_cp500["cp500 🟢"]:::c2
  n_encoding_utf16["utf16 🟢"]:::c2
  n_logger["logger 🟢"]:::c2
  n_network["network 🔴"]:::c0
  n_network_dcerpc["dcerpc 🟠"]:::c1
  n_network_dcerpc_interfaces["interfaces 🟠"]:::c1
  n_network_dcerpc_ms_protocols["ms-protocols 🟢"]:::c2
  n_network_dcerpc_ms_protocols_ms_drsr["ms-drsr 🟢"]:::c2
  n_network_dcerpc_ms_protocols_ms_rrp["ms-rrp 🟢"]:::c2
  n_network_dcerpc_ms_protocols_ms_srvs["ms-srvs 🟢"]:::c2
  n_network_dcerpc_ms_protocols_msproto["msproto 🟢"]:::c2
  n_network_dcerpc_ndr["ndr 🟢"]:::c2
  n_network_dcerpc_syntax["syntax 🟢"]:::c2
  n_network_dcerpc_v4["v4 🟠"]:::c1
  n_network_dcerpc_v4_client["client 🟢"]:::c2
  n_network_dcerpc_v4_epm["epm 🟢"]:::c2
  n_network_dcerpc_v4_interfaces["interfaces 🟢"]:::c2
  n_network_dcerpc_v4_interfaces_mgmt["mgmt 🟢"]:::c2
  n_network_dcerpc_v4_internal["internal 🟢"]:::c2
  n_network_dcerpc_v4_internal_ndr["ndr 🟢"]:::c2
  n_network_dcerpc_v4_pdu["pdu 🟢"]:::c2
  n_network_dcerpc_v4_transport["transport 🟠"]:::c1
  n_network_dcerpc_v4_transport_udp["udp 🟢"]:::c2
  n_network_dcerpc_v5["v5 🟠"]:::c1
  n_network_dcerpc_v5_client["client 🟢"]:::c2
  n_network_dcerpc_v5_pdu["pdu 🟢"]:::c2
  n_network_dcerpc_v5_transport["transport 🟠"]:::c1
  n_network_dcerpc_v5_transport_smb["smb 🟢"]:::c2
  n_network_dcerpc_v5_transport_smb2["smb2 🟢"]:::c2
  n_network_dcerpc_v5_transport_tcp["tcp 🟢"]:::c2
  n_network_dns["dns 🔴"]:::c0
  n_network_gssapi["gssapi 🟠"]:::c1
  n_network_gssapi_context["context 🟠"]:::c1
  n_network_gssapi_status["status 🟢"]:::c2
  n_network_ip["ip 🔴"]:::c0
  n_network_kerberos["kerberos 🟠"]:::c1
  n_network_kerberos_v5["v5 🟠"]:::c1
  n_network_kerberos_v5_attacks["attacks 🟢"]:::c2
  n_network_kerberos_v5_credcache["credcache 🟢"]:::c2
  n_network_kerberos_v5_credcache_ccache["ccache 🟢"]:::c2
  n_network_kerberos_v5_credcache_keytab["keytab 🟢"]:::c2
  n_network_kerberos_v5_credcache_kirbi["kirbi 🟢"]:::c2
  n_network_kerberos_v5_credentials["credentials 🟢"]:::c2
  n_network_kerberos_v5_crypto["crypto 🟢"]:::c2
  n_network_kerberos_v5_gssapi["gssapi 🟢"]:::c2
  n_network_kerberos_v5_iana["iana 🟠"]:::c1
  n_network_kerberos_v5_messages["messages 🟢"]:::c2
  n_network_kerberos_v5_mskile["mskile 🟢"]:::c2
  n_network_kerberos_v5_pac["pac 🟢"]:::c2
  n_network_kerberos_v5_pkinit["pkinit 🟢"]:::c2
  n_network_kerberos_v5_sfu["sfu 🟢"]:::c2
  n_network_ldap["ldap 🟠"]:::c1
  n_network_ldap_ldap_attributes["ldap_attributes 🟢"]:::c2
  n_network_ldap_objects["objects 🟠"]:::c1
  n_network_ldap_schema["schema 🟢"]:::c2
  n_network_llmnr["llmnr 🟠"]:::c1
  n_network_llmnr_class["class 🟢"]:::c2
  n_network_llmnr_constants["constants 🟠"]:::c1
  n_network_llmnr_domain_name["domain_name 🟢"]:::c2
  n_network_llmnr_errors["errors 🟠"]:::c1
  n_network_llmnr_llmnr_type["llmnr_type 🟢"]:::c2
  n_network_llmnr_message["message 🟢"]:::c2
  n_network_llmnr_message_header["header 🟢"]:::c2
  n_network_llmnr_question["question 🟢"]:::c2
  n_network_llmnr_resourcerecord["resourcerecord 🟢"]:::c2
  n_network_llmnr_server["server 🟢"]:::c2
  n_network_netbios["netbios 🔴"]:::c0
  n_network_netbios_nbdgm["nbdgm 🟢"]:::c2
  n_network_netbios_nbf["nbf 🔴"]:::c0
  n_network_netbios_nbns["nbns 🟢"]:::c2
  n_network_netbios_nbt["nbt 🟠"]:::c1
  n_network_smb["smb 🔴"]:::c0
  n_network_smb_client["client 🟢"]:::c2
  n_network_smb_common["common 🟢"]:::c2
  n_network_smb_common_transport["transport 🟢"]:::c2
  n_network_smb_smb_v10["smb_v10 🟠"]:::c1
  n_network_smb_smb_v10_capabilities["capabilities 🟢"]:::c2
  n_network_smb_smb_v10_client["client 🟢"]:::c2
  n_network_smb_smb_v10_dialects["dialects 🟢"]:::c2
  n_network_smb_smb_v10_errors["errors 🟠"]:::c1
  n_network_smb_smb_v10_informationlevels["informationlevels 🟢"]:::c2
  n_network_smb_smb_v10_message["message 🟠"]:::c1
  n_network_smb_smb_v10_message_commands["commands 🟠"]:::c1
  n_network_smb_smb_v10_message_commands_andx["andx 🟢"]:::c2
  n_network_smb_smb_v10_message_commands_codes["codes 🟠"]:::c1
  n_network_smb_smb_v10_message_commands_command_interface["command_interface 🟢"]:::c2
  n_network_smb_smb_v10_message_commands_utils["utils 🟢"]:::c2
  n_network_smb_smb_v10_message_data["data 🟢"]:::c2
  n_network_smb_smb_v10_message_header["header 🟢"]:::c2
  n_network_smb_smb_v10_message_header_flags["flags 🟢"]:::c2
  n_network_smb_smb_v10_message_header_flags2["flags2 🟢"]:::c2
  n_network_smb_smb_v10_message_parameters["parameters 🟢"]:::c2
  n_network_smb_smb_v10_message_securityfeatures["securityfeatures 🟢"]:::c2
  n_network_smb_smb_v10_securitymode["securitymode 🟢"]:::c2
  n_network_smb_smb_v10_subcommands["subcommands 🟢"]:::c2
  n_network_smb_smb_v10_types["types 🟢"]:::c2
  n_network_smb_smb_v20["smb_v20 🟠"]:::c1
  n_network_smb_smb_v20_capabilities["capabilities 🟢"]:::c2
  n_network_smb_smb_v20_client["client 🟢"]:::c2
  n_network_smb_smb_v20_createcontext["createcontext 🟢"]:::c2
  n_network_smb_smb_v20_dialects["dialects 🟠"]:::c1
  n_network_smb_smb_v20_message["message 🟠"]:::c1
  n_network_smb_smb_v20_message_commands["commands 🟠"]:::c1
  n_network_smb_smb_v20_message_commands_codes["codes 🟢"]:::c2
  n_network_smb_smb_v20_message_commands_command_interface["command_interface 🟠"]:::c1
  n_network_smb_smb_v20_message_header["header 🟢"]:::c2
  n_network_smb_smb_v20_message_header_flags["flags 🟢"]:::c2
  n_network_smb_smb_v20_securitymode["securitymode 🟢"]:::c2
  n_network_smb_smb_v20_types["types 🟢"]:::c2
  n_network_smb_smb_v21["smb_v21 🔴"]:::c0
  n_network_smb_smb_v30["smb_v30 🔴"]:::c0
  n_network_smb_smb_v302["smb_v302 🔴"]:::c0
  n_network_smb_smb_v311["smb_v311 🔴"]:::c0
  n_network_tcp["tcp 🔴"]:::c0
  n_utils["utils 🟢"]:::c2
  n_windows["windows 🟠"]:::c1
  n_windows_activedirectory["activedirectory 🟢"]:::c2
  n_windows_activedirectory_replication["replication 🟢"]:::c2
  n_windows_activedirectory_replication_dsrepl["dsrepl 🟢"]:::c2
  n_windows_cng["cng 🟠"]:::c1
  n_windows_cng_bcrypt["bcrypt 🟠"]:::c1
  n_windows_cng_bcrypt_keys["keys 🟠"]:::c1
  n_windows_cng_bcrypt_keys_blob["blob 🟢"]:::c2
  n_windows_cng_bcrypt_keys_headers["headers 🟢"]:::c2
  n_windows_cng_bcrypt_keys_magic["magic 🟢"]:::c2
  n_windows_cng_bcrypt_keys_types["types 🟠"]:::c1
  n_windows_credentials["credentials 🟢"]:::c2
  n_windows_database["database 🟢"]:::c2
  n_windows_database_ese["ese 🟢"]:::c2
  n_windows_database_ntds["ntds 🟢"]:::c2
  n_windows_fileflags["fileflags 🟠"]:::c1
  n_windows_filesystem["filesystem 🟠"]:::c1
  n_windows_filesystem_infoclass["infoclass 🟠"]:::c1
  n_windows_guid["guid 🟢"]:::c2
  n_windows_kerberos["kerberos 🟢"]:::c2
  n_windows_kerberos_serviceprincipalname["serviceprincipalname 🟢"]:::c2
  n_windows_keycredentiallink["keycredentiallink 🟢"]:::c2
  n_windows_keycredentiallink_crypto["crypto 🟢"]:::c2
  n_windows_keycredentiallink_key["key 🟢"]:::c2
  n_windows_keycredentiallink_key_customkeyinformation["customkeyinformation 🟢"]:::c2
  n_windows_keycredentiallink_key_material["material 🟢"]:::c2
  n_windows_keycredentiallink_key_material_fek["fek 🟢"]:::c2
  n_windows_keycredentiallink_key_material_fek_blob["blob 🟢"]:::c2
  n_windows_keycredentiallink_key_material_fek_headers["headers 🟢"]:::c2
  n_windows_keycredentiallink_key_material_fek_magic["magic 🟢"]:::c2
  n_windows_keycredentiallink_key_source["source 🟢"]:::c2
  n_windows_keycredentiallink_key_strength["strength 🟢"]:::c2
  n_windows_keycredentiallink_key_usage["usage 🟢"]:::c2
  n_windows_keycredentiallink_utils["utils 🟢"]:::c2
  n_windows_keycredentiallink_version["version 🟢"]:::c2
  n_windows_ms_dtyp["ms-dtyp 🟠"]:::c1
  n_windows_nt_status["nt_status 🟢"]:::c2
  n_windows_protocols["protocols 🟠"]:::c1
  n_windows_protocols_ms_bpau["ms-bpau 🟠"]:::c1
  n_windows_protocols_ms_brwsa["ms-brwsa 🟢"]:::c2
  n_windows_protocols_ms_capr["ms-capr 🟠"]:::c1
  n_windows_protocols_ms_cmpo["ms-cmpo 🟢"]:::c2
  n_windows_protocols_ms_cmrp["ms-cmrp 🟢"]:::c2
  n_windows_protocols_ms_dcom["ms-dcom 🟢"]:::c2
  n_windows_protocols_ms_dfsnm["ms-dfsnm 🟢"]:::c2
  n_windows_protocols_ms_dhcpm["ms-dhcpm 🟢"]:::c2
  n_windows_protocols_ms_dltw["ms-dltw 🟢"]:::c2
  n_windows_protocols_ms_dnsp["ms-dnsp 🟢"]:::c2
  n_windows_protocols_ms_drsr["ms-drsr 🟢"]:::c2
  n_windows_protocols_ms_dssp["ms-dssp 🟢"]:::c2
  n_windows_protocols_ms_eerr["ms-eerr 🟢"]:::c2
  n_windows_protocols_ms_efsr["ms-efsr 🟢"]:::c2
  n_windows_protocols_ms_even["ms-even 🟢"]:::c2
  n_windows_protocols_ms_even6["ms-even6 🟢"]:::c2
  n_windows_protocols_ms_fasp["ms-fasp 🟢"]:::c2
  n_windows_protocols_ms_fax["ms-fax 🟢"]:::c2
  n_windows_protocols_ms_frs1["ms-frs1 🟠"]:::c1
  n_windows_protocols_ms_frs2["ms-frs2 🟢"]:::c2
  n_windows_protocols_ms_fsrvp["ms-fsrvp 🟢"]:::c2
  n_windows_protocols_ms_irp["ms-irp 🟢"]:::c2
  n_windows_protocols_ms_lrec["ms-lrec 🟢"]:::c2
  n_windows_protocols_ms_lsad["ms-lsad 🟢"]:::c2
  n_windows_protocols_ms_lsat["ms-lsat 🟢"]:::c2
  n_windows_protocols_ms_mqds["ms-mqds 🟢"]:::c2
  n_windows_protocols_ms_mqmp["ms-mqmp 🟢"]:::c2
  n_windows_protocols_ms_mqmq["ms-mqmq 🟢"]:::c2
  n_windows_protocols_ms_mqmr["ms-mqmr 🟢"]:::c2
  n_windows_protocols_ms_mqqp["ms-mqqp 🟢"]:::c2
  n_windows_protocols_ms_mqrr["ms-mqrr 🟢"]:::c2
  n_windows_protocols_ms_msrp["ms-msrp 🟢"]:::c2
  n_windows_protocols_ms_nrpc["ms-nrpc 🟢"]:::c2
  n_windows_protocols_ms_nrpc_crypto["crypto 🟢"]:::c2
  n_windows_protocols_ms_nrpc_securechannel["securechannel 🟢"]:::c2
  n_windows_protocols_ms_nspi["ms-nspi 🟢"]:::c2
  n_windows_protocols_ms_pan["ms-pan 🟢"]:::c2
  n_windows_protocols_ms_par["ms-par 🟠"]:::c1
  n_windows_protocols_ms_pcq["ms-pcq 🟠"]:::c1
  n_windows_protocols_ms_raa["ms-raa 🟢"]:::c2
  n_windows_protocols_ms_raiw["ms-raiw 🟢"]:::c2
  n_windows_protocols_ms_rpce["ms-rpce 🟢"]:::c2
  n_windows_protocols_ms_rpcl["ms-rpcl 🟢"]:::c2
  n_windows_protocols_ms_rprn["ms-rprn 🟢"]:::c2
  n_windows_protocols_ms_rrasm["ms-rrasm 🟢"]:::c2
  n_windows_protocols_ms_rrp["ms-rrp 🟢"]:::c2
  n_windows_protocols_ms_rsp["ms-rsp 🟠"]:::c1
  n_windows_protocols_ms_samr["ms-samr 🟢"]:::c2
  n_windows_protocols_ms_scmr["ms-scmr 🟢"]:::c2
  n_windows_protocols_ms_srvs["ms-srvs 🟢"]:::c2
  n_windows_protocols_ms_swn["ms-swn 🟢"]:::c2
  n_windows_protocols_ms_trp["ms-trp 🟢"]:::c2
  n_windows_protocols_ms_tsch["ms-tsch 🟢"]:::c2
  n_windows_protocols_ms_tsgu["ms-tsgu 🟢"]:::c2
  n_windows_protocols_ms_tsts["ms-tsts 🟢"]:::c2
  n_windows_protocols_ms_w32t["ms-w32t 🟢"]:::c2
  n_windows_protocols_ms_wcce["ms-wcce 🟠"]:::c1
  n_windows_protocols_ms_wkst["ms-wkst 🟢"]:::c2
  n_windows_registry["registry 🟢"]:::c2
  n_windows_registry_regf["regf 🟢"]:::c2
  n_windows_registry_regfile["regfile 🟢"]:::c2
  n_crypto --> n_crypto_aes
  n_crypto_aes --> n_crypto_aes_cfb8
  n_crypto_aes --> n_crypto_aes_cts
  n_crypto --> n_crypto_cmac
  n_crypto --> n_crypto_dcc
  n_crypto --> n_crypto_dcc2
  n_crypto --> n_crypto_gppp
  n_crypto --> n_crypto_lm
  n_crypto --> n_crypto_md4
  n_crypto --> n_crypto_nfold
  n_crypto --> n_crypto_nt
  n_crypto --> n_crypto_ntlmv1
  n_crypto --> n_crypto_ntlmv2
  n_crypto --> n_crypto_pkcs7
  n_crypto --> n_crypto_rc4
  n_crypto --> n_crypto_spnego
  n_crypto_spnego --> n_crypto_spnego_ntlm
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_avpair
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_datafields
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_message
  n_crypto_spnego_ntlm_message --> n_crypto_spnego_ntlm_message_authenticate
  n_crypto_spnego_ntlm_message --> n_crypto_spnego_ntlm_message_challenge
  n_crypto_spnego_ntlm_message --> n_crypto_spnego_ntlm_message_header
  n_crypto_spnego_ntlm_message --> n_crypto_spnego_ntlm_message_negotiate
  n_crypto_spnego_ntlm_message_negotiate --> n_crypto_spnego_ntlm_message_negotiate_flags
  n_crypto_spnego_ntlm_message --> n_crypto_spnego_ntlm_message_types
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_security
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_targetinfo
  n_crypto_spnego_ntlm --> n_crypto_spnego_ntlm_version
  n_crypto --> n_crypto_uuid
  n_crypto_uuid --> n_crypto_uuid_uuid_v1
  n_crypto_uuid --> n_crypto_uuid_uuid_v2
  n_crypto_uuid --> n_crypto_uuid_uuid_v3
  n_crypto_uuid --> n_crypto_uuid_uuid_v4
  n_crypto_uuid --> n_crypto_uuid_uuid_v5
  n_crypto_uuid --> n_crypto_uuid_uuid_v6
  n_crypto_uuid --> n_crypto_uuid_uuid_v7
  n_crypto_uuid --> n_crypto_uuid_uuid_v8
  n_encoding --> n_encoding_ascii
  n_encoding --> n_encoding_ebcdic
  n_encoding_ebcdic --> n_encoding_ebcdic_cp037
  n_encoding_ebcdic --> n_encoding_ebcdic_cp500
  n_encoding --> n_encoding_utf16
  n_network --> n_network_dcerpc
  n_network_dcerpc --> n_network_dcerpc_interfaces
  n_network_dcerpc --> n_network_dcerpc_ms_protocols
  n_network_dcerpc_ms_protocols --> n_network_dcerpc_ms_protocols_ms_drsr
  n_network_dcerpc_ms_protocols --> n_network_dcerpc_ms_protocols_ms_rrp
  n_network_dcerpc_ms_protocols --> n_network_dcerpc_ms_protocols_ms_srvs
  n_network_dcerpc_ms_protocols --> n_network_dcerpc_ms_protocols_msproto
  n_network_dcerpc --> n_network_dcerpc_ndr
  n_network_dcerpc --> n_network_dcerpc_syntax
  n_network_dcerpc --> n_network_dcerpc_v4
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_client
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_epm
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_interfaces
  n_network_dcerpc_v4_interfaces --> n_network_dcerpc_v4_interfaces_mgmt
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_internal
  n_network_dcerpc_v4_internal --> n_network_dcerpc_v4_internal_ndr
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_pdu
  n_network_dcerpc_v4 --> n_network_dcerpc_v4_transport
  n_network_dcerpc_v4_transport --> n_network_dcerpc_v4_transport_udp
  n_network_dcerpc --> n_network_dcerpc_v5
  n_network_dcerpc_v5 --> n_network_dcerpc_v5_client
  n_network_dcerpc_v5 --> n_network_dcerpc_v5_pdu
  n_network_dcerpc_v5 --> n_network_dcerpc_v5_transport
  n_network_dcerpc_v5_transport --> n_network_dcerpc_v5_transport_smb
  n_network_dcerpc_v5_transport --> n_network_dcerpc_v5_transport_smb2
  n_network_dcerpc_v5_transport --> n_network_dcerpc_v5_transport_tcp
  n_network --> n_network_dns
  n_network --> n_network_gssapi
  n_network_gssapi --> n_network_gssapi_context
  n_network_gssapi --> n_network_gssapi_status
  n_network --> n_network_ip
  n_network --> n_network_kerberos
  n_network_kerberos --> n_network_kerberos_v5
  n_network_kerberos_v5 --> n_network_kerberos_v5_attacks
  n_network_kerberos_v5 --> n_network_kerberos_v5_credcache
  n_network_kerberos_v5_credcache --> n_network_kerberos_v5_credcache_ccache
  n_network_kerberos_v5_credcache --> n_network_kerberos_v5_credcache_keytab
  n_network_kerberos_v5_credcache --> n_network_kerberos_v5_credcache_kirbi
  n_network_kerberos_v5 --> n_network_kerberos_v5_credentials
  n_network_kerberos_v5 --> n_network_kerberos_v5_crypto
  n_network_kerberos_v5 --> n_network_kerberos_v5_gssapi
  n_network_kerberos_v5 --> n_network_kerberos_v5_iana
  n_network_kerberos_v5 --> n_network_kerberos_v5_messages
  n_network_kerberos_v5 --> n_network_kerberos_v5_mskile
  n_network_kerberos_v5 --> n_network_kerberos_v5_pac
  n_network_kerberos_v5 --> n_network_kerberos_v5_pkinit
  n_network_kerberos_v5 --> n_network_kerberos_v5_sfu
  n_network --> n_network_ldap
  n_network_ldap --> n_network_ldap_ldap_attributes
  n_network_ldap --> n_network_ldap_objects
  n_network_ldap --> n_network_ldap_schema
  n_network --> n_network_llmnr
  n_network_llmnr --> n_network_llmnr_class
  n_network_llmnr --> n_network_llmnr_constants
  n_network_llmnr --> n_network_llmnr_domain_name
  n_network_llmnr --> n_network_llmnr_errors
  n_network_llmnr --> n_network_llmnr_llmnr_type
  n_network_llmnr --> n_network_llmnr_message
  n_network_llmnr_message --> n_network_llmnr_message_header
  n_network_llmnr --> n_network_llmnr_question
  n_network_llmnr --> n_network_llmnr_resourcerecord
  n_network_llmnr --> n_network_llmnr_server
  n_network --> n_network_netbios
  n_network_netbios --> n_network_netbios_nbdgm
  n_network_netbios --> n_network_netbios_nbf
  n_network_netbios --> n_network_netbios_nbns
  n_network_netbios --> n_network_netbios_nbt
  n_network --> n_network_smb
  n_network_smb --> n_network_smb_client
  n_network_smb --> n_network_smb_common
  n_network_smb_common --> n_network_smb_common_transport
  n_network_smb --> n_network_smb_smb_v10
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_capabilities
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_client
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_dialects
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_errors
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_informationlevels
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_message
  n_network_smb_smb_v10_message --> n_network_smb_smb_v10_message_commands
  n_network_smb_smb_v10_message_commands --> n_network_smb_smb_v10_message_commands_andx
  n_network_smb_smb_v10_message_commands --> n_network_smb_smb_v10_message_commands_codes
  n_network_smb_smb_v10_message_commands --> n_network_smb_smb_v10_message_commands_command_interface
  n_network_smb_smb_v10_message_commands --> n_network_smb_smb_v10_message_commands_utils
  n_network_smb_smb_v10_message --> n_network_smb_smb_v10_message_data
  n_network_smb_smb_v10_message --> n_network_smb_smb_v10_message_header
  n_network_smb_smb_v10_message_header --> n_network_smb_smb_v10_message_header_flags
  n_network_smb_smb_v10_message_header --> n_network_smb_smb_v10_message_header_flags2
  n_network_smb_smb_v10_message --> n_network_smb_smb_v10_message_parameters
  n_network_smb_smb_v10_message --> n_network_smb_smb_v10_message_securityfeatures
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_securitymode
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_subcommands
  n_network_smb_smb_v10 --> n_network_smb_smb_v10_types
  n_network_smb --> n_network_smb_smb_v20
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_capabilities
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_client
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_createcontext
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_dialects
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_message
  n_network_smb_smb_v20_message --> n_network_smb_smb_v20_message_commands
  n_network_smb_smb_v20_message_commands --> n_network_smb_smb_v20_message_commands_codes
  n_network_smb_smb_v20_message_commands --> n_network_smb_smb_v20_message_commands_command_interface
  n_network_smb_smb_v20_message --> n_network_smb_smb_v20_message_header
  n_network_smb_smb_v20_message_header --> n_network_smb_smb_v20_message_header_flags
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_securitymode
  n_network_smb_smb_v20 --> n_network_smb_smb_v20_types
  n_network_smb --> n_network_smb_smb_v21
  n_network_smb --> n_network_smb_smb_v30
  n_network_smb --> n_network_smb_smb_v302
  n_network_smb --> n_network_smb_smb_v311
  n_network --> n_network_tcp
  n_windows --> n_windows_activedirectory
  n_windows_activedirectory --> n_windows_activedirectory_replication
  n_windows_activedirectory_replication --> n_windows_activedirectory_replication_dsrepl
  n_windows --> n_windows_cng
  n_windows_cng --> n_windows_cng_bcrypt
  n_windows_cng_bcrypt --> n_windows_cng_bcrypt_keys
  n_windows_cng_bcrypt_keys --> n_windows_cng_bcrypt_keys_blob
  n_windows_cng_bcrypt_keys --> n_windows_cng_bcrypt_keys_headers
  n_windows_cng_bcrypt_keys --> n_windows_cng_bcrypt_keys_magic
  n_windows_cng_bcrypt_keys --> n_windows_cng_bcrypt_keys_types
  n_windows --> n_windows_credentials
  n_windows --> n_windows_database
  n_windows_database --> n_windows_database_ese
  n_windows_database --> n_windows_database_ntds
  n_windows --> n_windows_fileflags
  n_windows --> n_windows_filesystem
  n_windows_filesystem --> n_windows_filesystem_infoclass
  n_windows --> n_windows_guid
  n_windows --> n_windows_kerberos
  n_windows_kerberos --> n_windows_kerberos_serviceprincipalname
  n_windows --> n_windows_keycredentiallink
  n_windows_keycredentiallink --> n_windows_keycredentiallink_crypto
  n_windows_keycredentiallink --> n_windows_keycredentiallink_key
  n_windows_keycredentiallink_key --> n_windows_keycredentiallink_key_customkeyinformation
  n_windows_keycredentiallink_key --> n_windows_keycredentiallink_key_material
  n_windows_keycredentiallink_key_material --> n_windows_keycredentiallink_key_material_fek
  n_windows_keycredentiallink_key_material_fek --> n_windows_keycredentiallink_key_material_fek_blob
  n_windows_keycredentiallink_key_material_fek --> n_windows_keycredentiallink_key_material_fek_headers
  n_windows_keycredentiallink_key_material_fek --> n_windows_keycredentiallink_key_material_fek_magic
  n_windows_keycredentiallink_key --> n_windows_keycredentiallink_key_source
  n_windows_keycredentiallink_key --> n_windows_keycredentiallink_key_strength
  n_windows_keycredentiallink_key --> n_windows_keycredentiallink_key_usage
  n_windows_keycredentiallink --> n_windows_keycredentiallink_utils
  n_windows_keycredentiallink --> n_windows_keycredentiallink_version
  n_windows --> n_windows_ms_dtyp
  n_windows --> n_windows_nt_status
  n_windows --> n_windows_protocols
  n_windows_protocols --> n_windows_protocols_ms_bpau
  n_windows_protocols --> n_windows_protocols_ms_brwsa
  n_windows_protocols --> n_windows_protocols_ms_capr
  n_windows_protocols --> n_windows_protocols_ms_cmpo
  n_windows_protocols --> n_windows_protocols_ms_cmrp
  n_windows_protocols --> n_windows_protocols_ms_dcom
  n_windows_protocols --> n_windows_protocols_ms_dfsnm
  n_windows_protocols --> n_windows_protocols_ms_dhcpm
  n_windows_protocols --> n_windows_protocols_ms_dltw
  n_windows_protocols --> n_windows_protocols_ms_dnsp
  n_windows_protocols --> n_windows_protocols_ms_drsr
  n_windows_protocols --> n_windows_protocols_ms_dssp
  n_windows_protocols --> n_windows_protocols_ms_eerr
  n_windows_protocols --> n_windows_protocols_ms_efsr
  n_windows_protocols --> n_windows_protocols_ms_even
  n_windows_protocols --> n_windows_protocols_ms_even6
  n_windows_protocols --> n_windows_protocols_ms_fasp
  n_windows_protocols --> n_windows_protocols_ms_fax
  n_windows_protocols --> n_windows_protocols_ms_frs1
  n_windows_protocols --> n_windows_protocols_ms_frs2
  n_windows_protocols --> n_windows_protocols_ms_fsrvp
  n_windows_protocols --> n_windows_protocols_ms_irp
  n_windows_protocols --> n_windows_protocols_ms_lrec
  n_windows_protocols --> n_windows_protocols_ms_lsad
  n_windows_protocols --> n_windows_protocols_ms_lsat
  n_windows_protocols --> n_windows_protocols_ms_mqds
  n_windows_protocols --> n_windows_protocols_ms_mqmp
  n_windows_protocols --> n_windows_protocols_ms_mqmq
  n_windows_protocols --> n_windows_protocols_ms_mqmr
  n_windows_protocols --> n_windows_protocols_ms_mqqp
  n_windows_protocols --> n_windows_protocols_ms_mqrr
  n_windows_protocols --> n_windows_protocols_ms_msrp
  n_windows_protocols --> n_windows_protocols_ms_nrpc
  n_windows_protocols_ms_nrpc --> n_windows_protocols_ms_nrpc_crypto
  n_windows_protocols_ms_nrpc --> n_windows_protocols_ms_nrpc_securechannel
  n_windows_protocols --> n_windows_protocols_ms_nspi
  n_windows_protocols --> n_windows_protocols_ms_pan
  n_windows_protocols --> n_windows_protocols_ms_par
  n_windows_protocols --> n_windows_protocols_ms_pcq
  n_windows_protocols --> n_windows_protocols_ms_raa
  n_windows_protocols --> n_windows_protocols_ms_raiw
  n_windows_protocols --> n_windows_protocols_ms_rpce
  n_windows_protocols --> n_windows_protocols_ms_rpcl
  n_windows_protocols --> n_windows_protocols_ms_rprn
  n_windows_protocols --> n_windows_protocols_ms_rrasm
  n_windows_protocols --> n_windows_protocols_ms_rrp
  n_windows_protocols --> n_windows_protocols_ms_rsp
  n_windows_protocols --> n_windows_protocols_ms_samr
  n_windows_protocols --> n_windows_protocols_ms_scmr
  n_windows_protocols --> n_windows_protocols_ms_srvs
  n_windows_protocols --> n_windows_protocols_ms_swn
  n_windows_protocols --> n_windows_protocols_ms_trp
  n_windows_protocols --> n_windows_protocols_ms_tsch
  n_windows_protocols --> n_windows_protocols_ms_tsgu
  n_windows_protocols --> n_windows_protocols_ms_tsts
  n_windows_protocols --> n_windows_protocols_ms_w32t
  n_windows_protocols --> n_windows_protocols_ms_wcce
  n_windows_protocols --> n_windows_protocols_ms_wkst
  n_windows --> n_windows_registry
  n_windows_registry --> n_windows_registry_regf
  n_windows_registry --> n_windows_registry_regfile
  classDef c2 fill:#1f8b4c,stroke:#155d33,color:#fff;
  classDef c1 fill:#d9822b,stroke:#a35d17,color:#fff;
  classDef c0 fill:#c0392b,stroke:#7d2419,color:#fff;
```
