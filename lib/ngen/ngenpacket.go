package ngen

type Net interface {
	Serialize([]byte)
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

type Packet struct {
	Header Header
	NetMsg Net
}

// Pack serializes the content into RawBytes.
func (p *Packet) Pack() []byte {
	buf := make([]byte, p.Len())
	PutUint32(buf, uint32(p.Header.MsgType))
	PutUint16(buf[4:], p.Header.Seq)
	PutUint16(buf[6:], p.Header.ContentLength)
	p.NetMsg.Serialize(buf[8:])
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

type NetParser func(Packet, *Buffer) Net

func NextPacket(rawBytes []byte, parser NetParser) (packet Packet, ok bool) {
	packet.Header, ok = ParseHeader(rawBytes)
	if !ok {
		return
	}

	ok = false
	if packet.Len() <= len(rawBytes) {
		packet.NetMsg = parser(packet, NewBuffer(rawBytes[HeaderLen:packet.Len()]))
		if packet.NetMsg != nil {
			ok = true
		}
	}
	return
}

type MessageType uint32
