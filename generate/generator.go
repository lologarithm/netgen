package generate

// Message is a message that can be serialized across network.
type Message struct {
	Name     string         // name of message
	Fields   []MessageField // list of fields on the message
	SelfSize int            // size of message not counting sub objects
}

// Enum represents a list of values with a shared type
type Enum struct {
	Name   string      // name of enum
	Values []EnumValue // list of enum values
}

// EnumValue is a single value from an enum
type EnumValue struct {
	Name  string
	Value int
}

// MessageField is a single field of a message.
type MessageField struct {
	Name      string
	Type      string
	Array     bool
	Pointer   bool
	Order     int
	Size      int
	Embedded  bool
	Interface bool // used only for generating from existing interfaces
}

// Allowed types to generate from
const (
	IntType     string = "int"
	RuneType    string = "rune"
	BoolType    string = "bool"
	StringType  string = "string"
	ByteType    string = "byte"
	Int16Type   string = "int16"
	Uint16Type  string = "uint16"
	Int32Type   string = "int32"
	Uint32Type  string = "uint32"
	Int64Type   string = "int64"
	Uint64Type  string = "uint64"
	Float64Type string = "float64"
	DynamicType string = "dynamic"
)
