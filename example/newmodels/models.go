package newmodels

type Message struct {
	Message string
}

type VersionedMessage struct {
	Message string `ngen:"1"`
	From    string `ngen:"2"`
	// UselessData int    `ngen:"3"` Don't need useless data anymore
	NewHotness float64 `ngen:"4"`
}
