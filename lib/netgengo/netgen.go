package netgengo

var LittleEndian littleEndian

type littleEndian struct{}

func (littleEndian) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

func (littleEndian) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func (littleEndian) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func (littleEndian) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func (littleEndian) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func (littleEndian) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

type Net interface {
	Serialize([]byte)
	Len() int
	MsgType() MessageType
}

const FrameLen int = 6

func NewPacket(msg Net) *Packet {
	return &Packet{
		Frame: Frame{
			MsgType:       msg.MsgType(),
			ContentLength: uint16(msg.Len()),
		},
		NetMsg: msg,
	}
}

type Packet struct {
	Frame  Frame
	NetMsg Net
}

// Pack serializes the content into RawBytes.
func (p *Packet) Pack() []byte {
	buf := make([]byte, p.Len())
	LittleEndian.PutUint16(buf, uint16(p.Frame.MsgType))
	LittleEndian.PutUint16(buf[2:], p.Frame.Seq)
	LittleEndian.PutUint16(buf[4:], p.Frame.ContentLength)
	p.NetMsg.Serialize(buf[6:])
	return buf
}

// Len returns the total length of the message including the frame
func (p *Packet) Len() int {
	return int(p.Frame.ContentLength) + FrameLen
}

type Frame struct {
	MsgType       MessageType // byte 0-1, type
	Seq           uint16      // byte 2-3, order of message
	ContentLength uint16      // byte 4-5, content length
}

func ParseFrame(rawBytes []byte) (mf Frame, ok bool) {
	if len(rawBytes) < FrameLen {
		return
	}
	mf.MsgType = MessageType(LittleEndian.Uint16(rawBytes[0:2]))
	mf.Seq = LittleEndian.Uint16(rawBytes[2:4])
	mf.ContentLength = LittleEndian.Uint16(rawBytes[4:6])
	return mf, true
}

type NetParser func(Packet, []byte) Net

func NextPacket(rawBytes []byte, parser NetParser) (packet Packet, ok bool) {
	packet.Frame, ok = ParseFrame(rawBytes)
	if !ok {
		return
	}

	ok = false
	if packet.Len() <= len(rawBytes) {
		packet.NetMsg = parser(packet, rawBytes[FrameLen:packet.Len()])
		if packet.NetMsg != nil {
			ok = true
		}
	}
	return
}

type MessageType uint16
