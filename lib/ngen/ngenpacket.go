package ngen

type Net interface {
	Serialize([]byte, *Settings)
	Len() int
	MsgType() MessageType
}

const HeaderLen int = 8

func NewPacket(msg Net) *Packet {
	return &Packet{
		Header: Header{
			MsgType:       msg.MsgType(),
			ContentLength: uint16(msg.Len()),
		},
		NetMsg: msg,
	}
}

type Settings struct {
	FieldVersions map[MessageType][]byte

	// FUTURE IDEA: negociate messages that don't need variable length
	// then we can remove that value from every message.
	FixedSizeMessages map[MessageType]int
}

// Serialize for Settings doesn't need a settings because it is the settings.
func (v Settings) Serialize(buffer []byte, _ *Settings) {
	i := 4
	PutUint32(buffer, uint32(len(v.FieldVersions)))
	for k, fv := range v.FieldVersions {
		PutUint32(buffer[i:], uint32(k))
		i += 4
		buffer[i] = byte(len(fv))
		i += 1
		copy(buffer[i:], fv)
		i += len(fv)
	}
}

func (v Settings) Len() int {
	total := 4
	for _, fv := range v.FieldVersions {
		total += 5 + len(fv) // Field key (4) + field length (1) + field values (len of array)
	}
	return total
}

func (v Settings) MsgType() MessageType {
	return 0
}

func DeserializeSettings(b *Buffer) *Settings {
	s := &Settings{}
	num, _ := b.ReadInt()
	s.FieldVersions = make(map[MessageType][]byte, num)
	for i := 0; i < num; i++ {
		// First read the type
		k, _ := b.ReadUint32()

		// Tomfoolery to have byte length array instead of uint32
		v, _ := b.ReadByte()
		buf, _ := b.readByteSlice(uint32(v))
		s.FieldVersions[MessageType(k)] = buf
	}
	return s
}

type Packet struct {
	Header Header
	NetMsg Net
}

// Pack serializes the content into RawBytes.
func (p *Packet) Pack(settings *Settings) []byte {
	buf := make([]byte, p.Len())
	PutUint32(buf, uint32(p.Header.MsgType))
	PutUint16(buf[4:], p.Header.Seq)
	PutUint16(buf[6:], p.Header.ContentLength)
	p.NetMsg.Serialize(buf[8:], settings)
	return buf
}

// Len returns the total length of the message including the frame
func (p *Packet) Len() int {
	return int(p.Header.ContentLength) + HeaderLen
}

type Header struct {
	MsgType       MessageType // byte 0-3, type
	Seq           uint16      // byte 4-5, order of message
	ContentLength uint16      // byte 6-7, content length
}

func ParseHeader(rawBytes []byte) (mf Header, ok bool) {
	if len(rawBytes) < HeaderLen {
		return
	}
	mf.MsgType = MessageType(Uint32(rawBytes[0:4]))
	mf.Seq = Uint16(rawBytes[4:6])
	mf.ContentLength = Uint16(rawBytes[6:8])
	return mf, true
}

type NetParser func(Packet, *Buffer, *Settings) Net

func NextPacket(rawBytes []byte, parser NetParser, ver *Settings) (packet Packet, ok bool) {
	packet.Header, ok = ParseHeader(rawBytes)
	if !ok {
		return
	}

	ok = false
	if packet.Len() <= len(rawBytes) {
		packet.NetMsg = parser(packet, NewBuffer(rawBytes[HeaderLen:packet.Len()]), ver)
		if packet.NetMsg != nil {
			ok = true
		}
	}
	return
}

type MessageType uint32
