// Package ibmi is the IBM i platform and ABI boundary of Manticore.
//
// It holds network-independent IBM i definitions: release identity, CCSIDs,
// program-call contracts (ibmi/rpc) and typed feature workflows (ibmi/api).
// The Client Access wire protocol that carries those calls lives separately in
// network/ibmi, mirroring the split between network/... transports and the
// root windows/... platform structures.
//
// Entry points: ibmi/api/users lists user and group profiles over any
// rpc.ExclusiveCaller; ibmi/rpc/qgy/... are the program-qualified QGY call
// contracts; network/ibmi/client provides the connection that implements
// the callers. The supported releases, password levels, CCSIDs and
// transports are listed in the network/ibmi package documentation.
//
// Dependency rules, enforced by architecture_test.go:
//
//   - no production package under ibmi/... imports network/ibmi/...;
//   - network/ibmi/... imports only ibmi, ibmi/ccsid and ibmi/rpc from here; and
//   - root ibmi and root ibmi/rpc never import their own children.
//
// Evidence and license boundary: Manticore is MIT-licensed. IBM's JTOpen
// (IBM Public License 1.0) may be used as behavioral evidence and to derive
// independently expressed tests, but its code must not be copied or
// mechanically translated here.
package ibmi
