package client

// TreeConnect represents an established tree connect between the client and share on the server
type TreeConnect struct {
	Connection *Connection // The SMB connection associated with this tree connect
	ShareName  string      // The share name corresponding to this tree connect
	TreeID     uint16      // The TreeID (TID) that identifies this tree connect
	Session    *Session    // A reference to the session on which this tree connect was established
	IsDfsShare bool        // A Boolean that, if set, indicates that the tree connect was established to a DFS share
}
