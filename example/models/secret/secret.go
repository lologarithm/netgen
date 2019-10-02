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

// Unused isn't used in the package that imports this one (example/models).
// If serializers are generated from example/models then this struct will not be included.
type Unused struct {
	Something int
}
