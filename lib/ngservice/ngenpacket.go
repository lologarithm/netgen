package ngservice

import "github.com/lologarithm/netgen/lib/ngen"

const headerLen int = 6

// Packet is a single network message.
type Packet struct {
	Header  Header
	RawData []byte
	NetMsg  ngen.Message
}

// Len returns the total length of the message including the frame
func (p *Packet) Len() int {
	return int(p.Header.ContentLength) + headerLen
}

// Header is the first bytes of a packet
type Header struct {
	MsgType       ngen.MessageType // byte 0-3, type
	ContentLength uint16           // byte 4-6, content length
}

// writeHeader will write header bytes to the given byte slice.
func writeHeader(h Header, buf []byte) {
	ngen.PutUint32(buf, uint32(h.MsgType))
	ngen.PutUint16(buf[4:], h.ContentLength)
}

// parseHeader will parse the header off a byte array.
func parseHeader(rawBytes []byte) (mf Header, ok bool) {
	if len(rawBytes) < headerLen {
		return
	}
	mf.MsgType = ngen.MessageType(ngen.Uint32(rawBytes[0:4]))
	mf.ContentLength = ngen.Uint16(rawBytes[4:6])
	return mf, true
}

// ReadPacket takes a context and a byte slice and tries to read a packet from it.
func ReadPacket(ctx *ngen.Context, rawBytes []byte) (packet Packet, ok bool) {
	if packet.Header, ok = parseHeader(rawBytes); !ok {
		return packet, ok
	}

	if packet.Len() <= len(rawBytes) {
		packet.NetMsg = ctx.Read(ctx, packet.Header.MsgType, ngen.NewBuffer(rawBytes[headerLen:packet.Len()]))
	}
	return packet, packet.NetMsg != nil
}

// WriteMessage turns a message into byte slice for writing to network
func WriteMessage(ctx *ngen.Context, msg ngen.Message) []byte {
	bytes := make([]byte, ctx.Length(ctx, msg)+headerLen)
	writeHeader(Header{MsgType: msg.MsgType(), ContentLength: uint16(len(bytes))}, bytes)
	ctx.Write(ctx, msg, ngen.NewBuffer(bytes[headerLen:]))
	return bytes
}
