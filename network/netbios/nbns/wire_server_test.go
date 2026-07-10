package nbns

import (
	"encoding/binary"
	"net"
	"testing"
	"time"
)

// TestNodeStatusResponderEndToEnd drives a NODE STATUS REQUEST from the package's
// own NodeStatus client through a running UDP server with node status enabled,
// and confirms the server answers from its local name table so client and server
// agree on the RFC 1002 4.2.18 wire format: the registered names (and their
// group flag) are enumerated and the configured adapter MAC is recovered from
// the STATISTICS UNIT_ID.
func TestNodeStatusResponderEndToEnd(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	if err := nbns.RegisterName("MANTICORE-01", "", Unique, net.ParseIP("10.0.0.5"), time.Hour); err != nil {
		t.Fatalf("RegisterName (unique): %v", err)
	}
	if err := nbns.RegisterName("WORKGROUP", "", Group, net.ParseIP("10.0.0.6"), time.Hour); err != nil {
		t.Fatalf("RegisterName (group): %v", err)
	}

	udp, err := NewUDPServer("127.0.0.1:0", nbns)
	if err != nil {
		t.Fatalf("NewUDPServer: %v", err)
	}
	wantMAC, err := net.ParseMAC("de:ad:be:ef:00:01")
	if err != nil {
		t.Fatalf("ParseMAC: %v", err)
	}
	udp.EnableNodeStatus(wantMAC)
	if err := udp.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer udp.Stop()

	target := udp.conn.LocalAddr().String()
	client := &Client{Timeout: 2 * time.Second, Retransmit: 1}
	result, err := client.NodeStatus(target)
	if err != nil {
		t.Fatalf("NodeStatus: %v", err)
	}

	names := map[string]NodeName{}
	for _, n := range result.Names {
		names[n.Name] = n
	}
	if _, ok := names["MANTICORE-01"]; !ok {
		t.Errorf("node status response missing MANTICORE-01: got %v", result.Names)
	}
	wg, ok := names["WORKGROUP"]
	if !ok {
		t.Errorf("node status response missing WORKGROUP: got %v", result.Names)
	} else if !wg.IsGroup() {
		t.Errorf("WORKGROUP entry not marked as a group name (flags 0x%04x)", wg.Flags)
	}
	for _, n := range result.Names {
		if !n.IsActive() {
			t.Errorf("entry %q not marked active (flags 0x%04x)", n.Name, n.Flags)
		}
	}

	if result.MAC == nil || result.MAC.String() != wantMAC.String() {
		t.Errorf("STATISTICS UNIT_ID MAC = %v, want %v", result.MAC, wantMAC)
	}
}

// TestHandleNodeStatusRDATA exercises the handler directly and re-parses the
// NBSTAT answer's RDATA to confirm the NUM_NAMES count, per-entry NAME_FLAGS
// (group/active) and the UNIT_ID are all built correctly.
func TestHandleNodeStatusRDATA(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	_ = nbns.RegisterName("HOST", "", Unique, net.ParseIP("10.0.0.1"), time.Hour)
	_ = nbns.RegisterName("GRP", "", Group, net.ParseIP("10.0.0.2"), time.Hour)

	h := NewPacketHandler(nbns)
	mac, _ := net.ParseMAC("01:02:03:04:05:06")
	h.EnableNodeStatus(mac)

	req := &NBNSPacket{
		Header:    NBNSHeader{Flags: OpNameQuery, Questions: 1},
		Questions: []NBNSQuestion{{Name: &NetBIOSName{Name: "*"}, Type: QuestionTypeNBSTAT, Class: QuestionClassIn}},
	}
	if !h.isNodeStatusQuery(req) {
		t.Fatal("isNodeStatusQuery = false for an NBSTAT question")
	}

	resp := &NBNSPacket{}
	h.handleNodeStatus(req, resp)

	if len(resp.Answers) != 1 || resp.Answers[0].Type != QuestionTypeNBSTAT {
		t.Fatalf("expected 1 NBSTAT answer, got %+v", resp.Answers)
	}

	result, err := parseNodeStatusRData(resp.Answers[0].RData)
	if err != nil {
		t.Fatalf("parseNodeStatusRData: %v", err)
	}
	if len(result.Names) != 2 {
		t.Fatalf("NUM_NAMES = %d, want 2", len(result.Names))
	}
	byName := map[string]NodeName{}
	for _, n := range result.Names {
		byName[n.Name] = n
	}
	if n, ok := byName["GRP"]; !ok || !n.IsGroup() || !n.IsActive() {
		t.Errorf("GRP entry = %+v, want group+active", n)
	}
	if n, ok := byName["HOST"]; !ok || n.IsGroup() || !n.IsActive() {
		t.Errorf("HOST entry = %+v, want unique+active", n)
	}
	if result.MAC == nil || result.MAC.String() != mac.String() {
		t.Errorf("UNIT_ID MAC = %v, want %v", result.MAC, mac)
	}
}

// TestNodeStatusDisabledByDefault confirms the responder is opt-in: without
// EnableNodeStatus an NBSTAT query is not answered from the name table.
func TestNodeStatusDisabledByDefault(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	h := NewPacketHandler(nbns)
	if h.nodeStatusEnabled {
		t.Fatal("node status enabled by default")
	}
}

// TestBuildWACKWire asserts the WAIT FOR ACKNOWLEDGEMENT (WACK) RESPONSE layout
// (RFC 1002 4.2.16): the transaction ID is echoed, the flags are a WACK response
// with AA set, the TTL is the wait time in seconds, and the 2-byte RDATA carries
// the request's OPCODE and NM_FLAGS (its header flags with the response bit and
// RCODE cleared).
func TestBuildWACKWire(t *testing.T) {
	const trn uint16 = 0xbeef
	reqFlags := OpRegistration | FlagRecursion | FlagBroadcast
	req := &NBNSPacket{
		Header: NBNSHeader{TransactionID: trn, Flags: reqFlags, Answers: 1},
		Answers: []NBNSResourceRecord{{
			Name: &NetBIOSName{Name: "WAITNAME"}, Type: QuestionTypeNB, Class: QuestionClassIn,
		}},
	}

	wack := buildWACK(req, DefaultWACKWait)

	if wack.Header.TransactionID != trn {
		t.Errorf("WACK TRN_ID = 0x%04x, want 0x%04x", wack.Header.TransactionID, trn)
	}
	wantFlags := FlagResponse | OpWACK | FlagAuthoritative
	if wack.Header.Flags != wantFlags {
		t.Errorf("WACK flags = 0x%04x, want 0x%04x", wack.Header.Flags, wantFlags)
	}
	if len(wack.Answers) != 1 {
		t.Fatalf("WACK answers = %d, want 1", len(wack.Answers))
	}
	rr := wack.Answers[0]
	if rr.TTL != uint32(DefaultWACKWait.Seconds()) {
		t.Errorf("WACK TTL = %d, want %d", rr.TTL, uint32(DefaultWACKWait.Seconds()))
	}
	if rr.Class != QuestionClassIn {
		t.Errorf("WACK RR_CLASS = 0x%04x, want IN", rr.Class)
	}
	if rr.RDLength != 2 || len(rr.RData) != 2 {
		t.Fatalf("WACK RDLENGTH = %d / len(RData) = %d, want 2/2", rr.RDLength, len(rr.RData))
	}
	if got := binary.BigEndian.Uint16(rr.RData); got != reqFlags {
		t.Errorf("WACK RDATA OPCODE/NM_FLAGS = 0x%04x, want 0x%04x", got, reqFlags)
	}

	// The WACK must marshal and round-trip through the packet codec.
	data, err := wack.Marshal()
	if err != nil {
		t.Fatalf("WACK Marshal: %v", err)
	}
	var back NBNSPacket
	if _, err := back.Unmarshal(data); err != nil {
		t.Fatalf("WACK Unmarshal: %v", err)
	}
	if back.Header.Flags&OpcodeMask != OpWACK {
		t.Errorf("round-tripped WACK opcode = 0x%04x, want OpWACK", back.Header.Flags&OpcodeMask)
	}
}

// TestMarkNameConflictEmitsDemand confirms MarkNameConflict now flips local state
// AND emits a NAME CONFLICT DEMAND (RFC 1002 4.2.15) to the affected owner when a
// conflict-demand sender is installed.
func TestMarkNameConflictEmitsDemand(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	owner := net.ParseIP("192.168.1.50")
	if err := nbns.RegisterName("DUPE", "", Unique, owner, time.Hour); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}

	var gotPkt *NBNSPacket
	var gotOwner net.IP
	nbns.SetConflictDemandSender(func(p *NBNSPacket, o net.IP) error {
		gotPkt, gotOwner = p, o
		return nil
	})

	if err := nbns.MarkNameConflict("DUPE", ""); err != nil {
		t.Fatalf("MarkNameConflict: %v", err)
	}

	if gotPkt == nil {
		t.Fatal("MarkNameConflict did not emit a NAME CONFLICT DEMAND")
	}
	if !gotOwner.Equal(owner) {
		t.Errorf("demand sent to %v, want %v", gotOwner, owner)
	}
	if gotPkt.Header.Flags&OpcodeMask != OpConflict {
		t.Errorf("demand opcode = 0x%04x, want OpConflict", gotPkt.Header.Flags&OpcodeMask)
	}
	if gotPkt.Header.Flags&RcodeMask != RcodeConflict {
		t.Errorf("demand RCODE = %d, want CFT_ERR (%d)", gotPkt.Header.Flags&RcodeMask, RcodeConflict)
	}
	if len(gotPkt.Answers) != 1 {
		t.Fatalf("demand answers = %d, want 1", len(gotPkt.Answers))
	}
	ip, err := ParseIPFromRData(gotPkt.Answers[0].RData)
	if err != nil {
		t.Fatalf("ParseIPFromRData: %v", err)
	}
	if !ip.Equal(owner) {
		t.Errorf("demand ADDR_ENTRY = %v, want %v", ip, owner)
	}

	// The record must also be flagged in conflict locally.
	if _, _, _, err := nbns.QueryName("DUPE", ""); err == nil {
		t.Error("QueryName succeeded for a name marked in conflict")
	}
}

// TestMarkNameConflictNoSender confirms that, absent a sender, MarkNameConflict
// keeps its historical local-only behaviour (no panic, status flipped).
func TestMarkNameConflictNoSender(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	_ = nbns.RegisterName("LOCAL", "", Unique, net.ParseIP("10.1.1.1"), time.Hour)
	if err := nbns.MarkNameConflict("LOCAL", ""); err != nil {
		t.Fatalf("MarkNameConflict: %v", err)
	}
	if _, _, _, err := nbns.QueryName("LOCAL", ""); err == nil {
		t.Error("QueryName succeeded for a name marked in conflict")
	}
}

// TestRedirectWiredPath confirms that with a redirect manager installed, a name
// query for a configured scope is turned into a REDIRECT NAME QUERY RESPONSE
// (OpRedirect) rather than an authoritative name-table answer.
func TestRedirectWiredPath(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	h := NewPacketHandler(nbns)

	rm := NewRedirectManager()
	rm.AddRedirect("", net.ParseIP("10.9.9.9").To4(), 137)
	h.SetRedirectManager(rm)

	req := &NBNSPacket{
		Header:    NBNSHeader{Flags: OpNameQuery, Questions: 1},
		Questions: []NBNSQuestion{{Name: &NetBIOSName{Name: "ANY"}, Type: QuestionTypeNB, Class: QuestionClassIn}},
	}
	resp := &NBNSPacket{}
	h.handleNameQueryWithRedirect(req, resp)

	if resp.Header.Flags&OpcodeMask != OpRedirect {
		t.Errorf("query flags = 0x%04x, want OpRedirect", resp.Header.Flags&OpcodeMask)
	}
	if resp.Header.Additional != 1 || len(resp.Additional) != 1 {
		t.Errorf("redirect additional count = %d/%d, want 1/1", resp.Header.Additional, len(resp.Additional))
	}
}

// TestRedirectDisabledByDefault confirms an ordinary query with no redirect
// manager falls through to the authoritative name-table lookup unchanged.
func TestRedirectDisabledByDefault(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	_ = nbns.RegisterName("REAL", "", Unique, net.ParseIP("10.0.0.42"), time.Hour)
	h := NewPacketHandler(nbns)

	req := &NBNSPacket{
		Header:    NBNSHeader{Flags: OpNameQuery, Questions: 1},
		Questions: []NBNSQuestion{{Name: &NetBIOSName{Name: "REAL"}, Type: QuestionTypeNB, Class: QuestionClassIn}},
	}
	resp := &NBNSPacket{}
	h.handleNameQueryWithRedirect(req, resp)

	if resp.Header.Flags&OpcodeMask == OpRedirect {
		t.Error("query was redirected with no redirect manager installed")
	}
	if len(resp.Answers) != 1 {
		t.Fatalf("expected 1 authoritative answer, got %d", len(resp.Answers))
	}
}

// TestChallengePathEmitsWACK drives a conflicting registration through a UDP
// server with name defence enabled and confirms the first datagram the requestor
// receives is a WAIT FOR ACKNOWLEDGEMENT (WACK) emitted before the challenge
// runs (RFC 1002 4.2.10 / 4.2.16). The challenge itself targets a non-listening
// address, but the WACK precedes it and arrives promptly.
func TestChallengePathEmitsWACK(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	if err := nbns.RegisterName("OWNED", "", Unique, net.ParseIP("127.0.0.2"), time.Hour); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}

	udp, err := NewUDPServer("127.0.0.1:0", nbns)
	if err != nil {
		t.Fatalf("NewUDPServer: %v", err)
	}
	udp.EnableNameDefense()
	if err := udp.Start(); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer udp.Stop()

	srvAddr := udp.conn.LocalAddr().(*net.UDPAddr)
	client, err := net.DialUDP("udp", nil, srvAddr)
	if err != nil {
		t.Fatalf("DialUDP: %v", err)
	}
	defer client.Close()

	entry := ADDR_ENTRY{Address: binary.BigEndian.Uint32(net.ParseIP("127.0.0.3").To4())}
	reg := &NBNSPacket{
		Header: NBNSHeader{TransactionID: 0x1234, Flags: OpRegistration, Answers: 1},
		Answers: []NBNSResourceRecord{{
			Name: &NetBIOSName{Name: "OWNED"}, Type: QuestionTypeNB, Class: QuestionClassIn,
			TTL: 3600, RDLength: uint16(entry.Length()), RData: entry.Marshal(),
		}},
	}
	data, err := reg.Marshal()
	if err != nil {
		t.Fatalf("Marshal registration: %v", err)
	}
	if _, err := client.Write(data); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := client.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatalf("SetReadDeadline: %v", err)
	}
	buf := make([]byte, MaxUDPSize)
	n, err := client.Read(buf)
	if err != nil {
		t.Fatalf("Read (expected WACK): %v", err)
	}

	var resp NBNSPacket
	if _, err := resp.Unmarshal(buf[:n]); err != nil {
		t.Fatalf("Unmarshal WACK: %v", err)
	}
	if resp.Header.Flags&OpcodeMask != OpWACK {
		t.Errorf("first datagram opcode = 0x%04x, want OpWACK", resp.Header.Flags&OpcodeMask)
	}
	if resp.Header.TransactionID != 0x1234 {
		t.Errorf("WACK TRN_ID = 0x%04x, want 0x1234", resp.Header.TransactionID)
	}
}

// TestNameDefenseNoConflictRegisters confirms name defence is transparent when
// there is no conflict: NewNameChallenger is wired in, but a fresh registration
// still succeeds without any challenge.
func TestNameDefenseNoConflictRegisters(t *testing.T) {
	nbns := NewNetBIOSNameServer(false)
	udp, err := NewUDPServer("127.0.0.1:0", nbns)
	if err != nil {
		t.Fatalf("NewUDPServer: %v", err)
	}
	udp.EnableNameDefense()
	if udp.handlers.challenger == nil {
		t.Fatal("EnableNameDefense did not install a challenger")
	}

	entry := ADDR_ENTRY{Address: binary.BigEndian.Uint32(net.ParseIP("10.5.5.5").To4())}
	reg := &NBNSPacket{
		Header: NBNSHeader{Flags: OpRegistration, Answers: 1},
		Answers: []NBNSResourceRecord{{
			Name: &NetBIOSName{Name: "FRESH"}, Type: QuestionTypeNB, Class: QuestionClassIn,
			TTL: 3600, RDLength: uint16(entry.Length()), RData: entry.Marshal(),
		}},
	}
	resp := &NBNSPacket{}
	udp.handlers.handleRegistrationWithChallenge(reg, resp, nil)

	if resp.Header.Flags&RcodeMask != RcodeSuccess {
		t.Errorf("fresh registration RCODE = %d, want success", resp.Header.Flags&RcodeMask)
	}
	if _, _, _, err := nbns.QueryName("FRESH", ""); err != nil {
		t.Errorf("QueryName after registration: %v", err)
	}
}
