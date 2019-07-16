package ngen

// MessageType is the hash of a message to uniquely identify the type.
type MessageType uint32

// Message is a single message that can be serialized and setn
type Message interface {
	MsgType() MessageType
	Serialize(*Context, *Buffer) error // Writes the serialized message to the given buffer
	// Deserialize(*Context, *Buffer) error // Deserializes the given buffer into this object instance. Requires that the message type matches the underlying struct
	Length(*Context) int // Returns the size of the object. Used by serialize to pre-allocate a buffer.
}

// Reader takes a context and message type and deserializes a message from the given buffer.
type Reader func(*Context, MessageType, *Buffer) Message

// Context represents serialization settings
// The current main use for this is to exchange versions of objects.
type Context struct {
	Read Reader

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
func (v Context) Serialize(_ *Context, buf *Buffer) error {
	buf.WriteUint32(uint32(len(v.FieldVersions)))
	for k, fv := range v.FieldVersions {
		buf.WriteUint32(uint32(k))
		l := len(fv)
		buf.WriteByte(byte(l)) // write len as byte to keep msg small
		buf.writeByteSlice(fv)
	}
	return buf.Err
}

// Length returns length of this message
func (v Context) Length(_ *Context) int {
	total := 4
	for _, fv := range v.FieldVersions {
		total += 5 + len(fv) // Field key (4) + field length (1) + field values (len of array)
	}
	return total
}

func (c *Context) Deserialize(ctx *Context, buf *Buffer) error {
	*c = *DeserializeContext(ctx, buf)
	return buf.Err
}

// DeserializeContext constructs a new Context from the binary data.
// Requires the existing context to clone reader/writer functions.
func DeserializeContext(ctx *Context, b *Buffer) *Context {
	s := &Context{
		Read: ctx.Read,
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
