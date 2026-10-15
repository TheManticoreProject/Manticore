// Package dataqueue is a client for the IBM i Data Queue host server
// (as-dtaq / as-dtaq-s, server ID 0xE007): query a queue's attributes,
// write, read and peek entries (sequential or by key), and create, clear
// and delete queues.
//
// A Session is one authenticated server job, started by Start on a
// connection dialed to the Data Queue endpoint with the same credential
// and password family as Sign-on (see network/ibmi/client.OpenDataQueue).
// Exchanges are serialized on the connection; a Session is safe for
// concurrent use, but a read that waits blocks every other operation until
// it returns, so give long or indefinite waits their own Session.
//
// Entry and key bytes are opaque: they are sent and returned exactly, never
// transcoded. Queue and library names, search operators and descriptions
// travel in the system CCSID learned from Sign-on.
//
// Read, write, clear, create and delete change the queue. When the reply
// to one of them is lost the outcome is unknown (ErrUncertainOutcome): the
// entry may already be gone or added. Nothing is ever retried; the
// connection is poisoned and a new Session is needed. Registered exit
// programs may reject an otherwise valid operation, so authentication does
// not imply authority to a queue.
//
// Errors report return codes and IBM message IDs, never entry, key or
// message text; Error.Text holds the decoded message for callers that
// want it.
//
// Layouts follow iSeriesNavigatorReverse docs/16-data-queue-host-server.md
// (the JTOpen 4.8 generation that the V5R4 baseline targets).
package dataqueue
