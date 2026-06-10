package structures

// Twr models twr_t ([C706] Appendix O):
//
//	typedef struct {
//	    unsigned32 tower_length;
//	    [size_is(tower_length)] byte tower_octet_string[];
//	} twr_t;
//
// It is a conformant structure: the NDR codec hoists the array's maximum_count to the
// front of the structure and derives TowerLength from the octet-string length, so
// callers set only TowerOctetString. The octet string itself is a protocol tower (see
// Tower); use NewTwr / DecodeTower to convert between the two.
type Twr struct {
	TowerLength      uint32
	TowerOctetString []byte `ndr:"conformant,size_is=TowerLength"`
}

// NewTwr wraps a protocol tower's octet-string encoding in a twr_t.
func NewTwr(tower Tower) Twr {
	return Twr{TowerOctetString: tower.Marshal()}
}

// DecodeTower parses the twr_t's octet string back into a protocol tower.
func (t Twr) DecodeTower() (Tower, error) {
	return UnmarshalTower(t.TowerOctetString)
}
