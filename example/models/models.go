package models

import "github.com/lologarithm/netgen/example/models/secret"

type Message struct {
	Message string
}

type VersionedMessage struct {
	Message string `ngen:"1"`
	From    string `ngen:"2"`
}

// SuperMessage can have a secret.
type SuperMessage struct {
	Normal string
	Secret *secret.Msg
}
