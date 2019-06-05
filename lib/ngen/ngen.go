package ngen

// MessageType is the hash of a message to uniquely identify the type.
type MessageType uint32

// Message is a single message that can be serialized and setn
type Message interface{}

// Writer is a function that accepts a context and a message and then serializes to the given buffer.
type Writer func(*Context, Message, *Buffer) error

// Reader takes a context and message type and deserializes a message from the given buffer.
type Reader func(*Context, MessageType, *Buffer) Message

// Lengther returns the length of the given message for allocating sizes.
type Lengther func(*Context, Message) int

// Context represents serialization settings
// The current main use for this is to exchange versions of objects.
type Context struct {
	Write  Writer
	Read   Reader
	Length Lengther

	FieldVersions map[MessageType][]byte

	// FUTURE IDEA: negociate messages that don't need variable length
	// then we can remove that value from every message.
	FixedSizeMessages map[MessageType]int
}

// MessageTypeContext is the message type of the context object itself.
const MessageTypeContext MessageType = 1

// MsgType is to implement the Message interface
func (v Context) MsgType() MessageType {
	return MessageTypeContext // Context gets a special message type. It is number one!
}

// Serialize will convert the settings to a byte slice
func (v Context) Serialize() []byte {
	buffer := make([]byte, v.length())
	i := 4
	PutUint32(buffer, uint32(len(v.FieldVersions)))
	for k, fv := range v.FieldVersions {
		PutUint32(buffer[i:], uint32(k))
		i += 4
		buffer[i] = byte(len(fv))
		i++
		copy(buffer[i:], fv)
		i += len(fv)
	}
	return buffer
}

// Len returns
func (v Context) length() int {
	total := 4
	for _, fv := range v.FieldVersions {
		total += 5 + len(fv) // Field key (4) + field length (1) + field values (len of array)
	}
	return total
}

// DeserializeContext constructs a new Context from the binary data.
// Requires the existing context to clone reader/writer functions.
func DeserializeContext(ctx *Context, b *Buffer) *Context {
	s := &Context{
		Read:   ctx.Read,
		Write:  ctx.Write,
		Length: ctx.Length,
	}
	num := b.ReadInt()
	s.FieldVersions = make(map[MessageType][]byte, num)
	for i := 0; i < num; i++ {
		// First read the type
		k := b.ReadUint32()

		// Tomfoolery to have byte length array instead of uint32
		v := b.ReadByte()
		buf := b.readByteSlice(uint32(v))
		s.FieldVersions[MessageType(k)] = buf
	}
	return s
}
