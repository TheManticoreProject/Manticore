package structures

// LSA_FOREST_TRUST_COLLISION_RECORD_TYPE enumerates the kinds of forest-trust collision
// record ([MS-LSAD] 2.2.7.25). As an NDR enum it is transmitted as a 16-bit unsigned
// value ([C706] section 14.3.6).
type LSA_FOREST_TRUST_COLLISION_RECORD_TYPE uint16

const (
	CollisionTdo   LSA_FOREST_TRUST_COLLISION_RECORD_TYPE = 0
	CollisionXref  LSA_FOREST_TRUST_COLLISION_RECORD_TYPE = 1
	CollisionOther LSA_FOREST_TRUST_COLLISION_RECORD_TYPE = 2
)
