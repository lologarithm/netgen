package netmsg

import (
	"github.com/lologarithm/netgen/lib/ngen"
)


const (
	UnknownMsgType ngen.MessageType = iota
	AckMsgType
	MultipartMsgType
	HeartbeatMsgType
	BenchyMsgType
	NestedMsgType
	ExampleDynMsgType
	SubNestMsgType
	SubNest2MsgType
	ConnectedMsgType
)

// ParseNetMessage accepts input of raw bytes from a NetMessage. Parses and returns a Net message.
func ParseNetMessage(packet ngen.Packet, content *ngen.Buffer) ngen.Net {
	switch packet.Header.MsgType {
	case MultipartMsgType:
		msg := MultipartDeserialize(content)
		return &msg
	case HeartbeatMsgType:
		msg := HeartbeatDeserialize(content)
		return &msg
	case BenchyMsgType:
		msg := BenchyDeserialize(content)
		return &msg
	case NestedMsgType:
		msg := NestedDeserialize(content)
		return &msg
	case ExampleDynMsgType:
		msg := ExampleDynDeserialize(content)
		return &msg
	case SubNestMsgType:
		msg := SubNestDeserialize(content)
		return &msg
	case SubNest2MsgType:
		msg := SubNest2Deserialize(content)
		return &msg
	case ConnectedMsgType:
		msg := ConnectedDeserialize(content)
		return &msg
	default:
		return nil
	}
}

type Level int

const(
	PrettyLow	 Level = 0
	PrettyOk	 Level = 1
	PrettyAwesome	 Level = 2
)

type Multipart struct {
	ID uint16
	GroupID uint32
	NumParts uint16
	Content []byte
}

func (m Multipart) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint16(buffer[idx:], uint16(m.ID))
	idx+=2
	ngen.PutUint32(buffer[idx:], uint32(m.GroupID))
	idx+=4
	ngen.PutUint16(buffer[idx:], uint16(m.NumParts))
	idx+=2
	ngen.PutUint32(buffer[idx:], uint32(len(m.Content)))
	idx += 4
	copy(buffer[idx:], m.Content)
	idx+=len(m.Content)
}

func (m Multipart) Len() int {
	mylen := 0
	mylen += 2
	mylen += 4
	mylen += 2
	mylen += 4 + len(m.Content)
	return mylen
}

func (m Multipart) MsgType() ngen.MessageType {
	return MultipartMsgType
}


func MultipartDeserialize(buffer *ngen.Buffer) (m Multipart) {
	m.ID, _ = buffer.ReadUint16()
	m.GroupID, _ = buffer.ReadUint32()
	m.NumParts, _ = buffer.ReadUint16()
	m.Content, _ = buffer.ReadByteSlice()
	return m
}
type Heartbeat struct {
	Time int64
	Latency int64
}

func (m Heartbeat) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint64(buffer[idx:], uint64(m.Time))
	idx+=8
	ngen.PutUint64(buffer[idx:], uint64(m.Latency))
	idx+=8
}

func (m Heartbeat) Len() int {
	mylen := 0
	mylen += 8
	mylen += 8
	return mylen
}

func (m Heartbeat) MsgType() ngen.MessageType {
	return HeartbeatMsgType
}


func HeartbeatDeserialize(buffer *ngen.Buffer) (m Heartbeat) {
	m.Time, _ = buffer.ReadInt64()
	m.Latency, _ = buffer.ReadInt64()
	return m
}
type Benchy struct {
	Name string
	BirthDay int64
	Phone string
	Siblings int32
	Spouse byte
	Money float64
}

func (m Benchy) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint32(buffer[idx:], uint32(len(m.Name)))
	idx += 4
	copy(buffer[idx:], []byte(m.Name))
	idx+=len(m.Name)
	ngen.PutUint64(buffer[idx:], uint64(m.BirthDay))
	idx+=8
	ngen.PutUint32(buffer[idx:], uint32(len(m.Phone)))
	idx += 4
	copy(buffer[idx:], []byte(m.Phone))
	idx+=len(m.Phone)
	ngen.PutUint32(buffer[idx:], uint32(m.Siblings))
	idx+=4
	buffer[idx] = m.Spouse
	idx+=1
	ngen.PutFloat64(buffer[idx:], m.Money)
	idx+=8
}

func (m Benchy) Len() int {
	mylen := 0
	mylen += 4 + len(m.Name)
	mylen += 8
	mylen += 4 + len(m.Phone)
	mylen += 4
	mylen += 1
	mylen += 8
	return mylen
}

func (m Benchy) MsgType() ngen.MessageType {
	return BenchyMsgType
}


func BenchyDeserialize(buffer *ngen.Buffer) (m Benchy) {
	m.Name, _ = buffer.ReadString()
	m.BirthDay, _ = buffer.ReadInt64()
	m.Phone, _ = buffer.ReadString()
	m.Siblings, _ = buffer.ReadInt32()
	m.Spouse, _ = buffer.ReadByte()
	m.Money, _ = buffer.ReadFloat64()
	return m
}
type Nested struct {
	A SubNest
	B *SubNest2
}

func (m Nested) Serialize(buffer []byte) {
	idx := 0
	m.A.Serialize(buffer[idx:])
	idx+=m.A.Len()
		if m.B != nil {
		buffer[idx] = 1
		idx++
		m.B.Serialize(buffer[idx:])
	idx+=m.B.Len()
	} else {
	buffer[idx] = 0
	idx++
	}}

func (m Nested) Len() int {
	mylen := 0
	mylen += m.A.Len()
	mylen += m.B.Len()	
mylen++
	return mylen
}

func (m Nested) MsgType() ngen.MessageType {
	return NestedMsgType
}


func NestedDeserialize(buffer *ngen.Buffer) (m Nested) {
	m.A = SubNestDeserialize(buffer)
	var subB = SubNest2Deserialize(buffer)
	m.B = &subB
	return m
}
type ExampleDyn struct {
	DynField ngen.Net
	dynFieldType ngen.MessageType
}

func (m ExampleDyn) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint16(buffer[idx:], uint16(m.DynField.MsgType()))
	idx+=2
	m.DynField.(ngen.Net).Serialize(buffer[idx:])
	idx+=m.DynField.(ngen.Net).Len()
}

func (m ExampleDyn) Len() int {
	mylen := 0
	mylen += 2
	mylen += m.DynField.(ngen.Net).Len()
	return mylen
}

func (m ExampleDyn) MsgType() ngen.MessageType {
	return ExampleDynMsgType
}


func ExampleDynDeserialize(buffer *ngen.Buffer) (m ExampleDyn) {
	ttDynField, _ := buffer.ReadUint16()
m.dynFieldType = ngen.MessageType(ttDynField)
	p := ngen.Packet{Header: ngen.Header{MsgType: ngen.MessageType(m.dynFieldType)}}
	m.DynField = ParseNetMessage(p, buffer)
	return m
}
type SubNest struct {
	B int32
	C float64
}

func (m SubNest) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint32(buffer[idx:], uint32(m.B))
	idx+=4
	ngen.PutFloat64(buffer[idx:], m.C)
	idx+=8
}

func (m SubNest) Len() int {
	mylen := 0
	mylen += 4
	mylen += 8
	return mylen
}

func (m SubNest) MsgType() ngen.MessageType {
	return SubNestMsgType
}


func SubNestDeserialize(buffer *ngen.Buffer) (m SubNest) {
	m.B, _ = buffer.ReadInt32()
	m.C, _ = buffer.ReadFloat64()
	return m
}
type SubNest2 struct {
}

func (m SubNest2) Serialize(buffer []byte) {
}

func (m SubNest2) Len() int {
	mylen := 0
	return mylen
}

func (m SubNest2) MsgType() ngen.MessageType {
	return SubNest2MsgType
}


func SubNest2Deserialize(buffer *ngen.Buffer) (m SubNest2) {
	return m
}
type Connected struct {
	Awesomeness Level
}

func (m Connected) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint32(buffer[idx:], uint32(m.Awesomeness))
	idx+=4
}

func (m Connected) Len() int {
	mylen := 0
	mylen += 4
	return mylen
}

func (m Connected) MsgType() ngen.MessageType {
	return ConnectedMsgType
}


func ConnectedDeserialize(buffer *ngen.Buffer) (m Connected) {
	tmpAwesomeness, _ := buffer.ReadUint32()
	m.Awesomeness = Level(tmpAwesomeness)
	return m
}
