package secret

import "fmt"

// Msg is a secret message for someone.
type Msg struct {
	Message string
	To      string
}

func (m Msg) String() string {
	return fmt.Sprintf("M: %s", m.Message)
}

// Unused isn't used in example app. This will not be included in the serializers
type Unused struct {
	Something int
}
