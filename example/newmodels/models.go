package newmodels

type Message struct {
	Message string
}

type VersionedMessage struct {
	Message string `ngen:"1"`
	// From    string `ngen:"2"`
	Blah string `ngen:"3"`
}
