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
func ParseNetMessage(packet ngen.Packet, content []byte) ngen.Net {
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


func MultipartDeserialize(buffer []byte) (m Multipart) {
	idx := 0
	m.ID = ngen.Uint16(buffer[idx:])
	idx+=2
	m.GroupID = ngen.Uint32(buffer[idx:])
	idx+=4
	m.NumParts = ngen.Uint16(buffer[idx:])
	idx+=2
	l3_1 := int(ngen.Uint32(buffer[idx:]))
	idx += 4
	m.Content = make([]byte, l3_1)
	copy(m.Content, buffer[idx:idx+l3_1])
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


func HeartbeatDeserialize(buffer []byte) (m Heartbeat) {
	idx := 0
	m.Time = int64(ngen.Uint64(buffer[idx:]))
	idx+=8
	m.Latency = int64(ngen.Uint64(buffer[idx:]))
	idx+=8
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


func BenchyDeserialize(buffer []byte) (m Benchy) {
	idx := 0
	l0_1 := int(ngen.Uint32(buffer[idx:]))
	idx += 4
	m.Name = string(buffer[idx:idx+l0_1])
	idx+=len(m.Name)
	m.BirthDay = int64(ngen.Uint64(buffer[idx:]))
	idx+=8
	l2_1 := int(ngen.Uint32(buffer[idx:]))
	idx += 4
	m.Phone = string(buffer[idx:idx+l2_1])
	idx+=len(m.Phone)
	m.Siblings = int32(ngen.Uint32(buffer[idx:]))
	idx+=4
	m.Spouse = buffer[idx]

	idx+=1
	m.Money = ngen.Float64(buffer[idx:])
	idx+=8
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


func NestedDeserialize(buffer []byte) (m Nested) {
	idx := 0
	m.A = SubNestDeserialize(buffer[idx:])
	idx+=m.A.Len()
	var subB = SubNest2Deserialize(buffer[idx:])
	m.B = &subB
	idx+=m.B.Len()
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


func ExampleDynDeserialize(buffer []byte) (m ExampleDyn) {
	idx := 0
	m.dynFieldType = ngen.MessageType(ngen.Uint16(buffer[idx:]))
	idx+=2
	p := ngen.Packet{Header: ngen.Header{MsgType: m.dynFieldType}}
	m.DynField = ParseNetMessage(p, buffer[idx:])
		idx+=m.DynField.(ngen.Net).Len()
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


func SubNestDeserialize(buffer []byte) (m SubNest) {
	idx := 0
	m.B = int32(ngen.Uint32(buffer[idx:]))
	idx+=4
	m.C = ngen.Float64(buffer[idx:])
	idx+=8
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


func SubNest2Deserialize(buffer []byte) (m SubNest2) {
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


func ConnectedDeserialize(buffer []byte) (m Connected) {
	idx := 0
	m.Awesomeness = Level(ngen.Uint32(buffer[idx:]))
	idx+=4
	return m
}
