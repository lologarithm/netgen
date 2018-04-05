package main

import (
	"github.com/lologarithm/netgen/lib/ngen"
)


const (
	UnknownMsgType ngen.MessageType = iota
	AckMsgType
	BenchyMsgType
)

// ParseNetMessage accepts input of raw bytes from a NetMessage. Parses and returns a Net message.
func ParseNetMessage(packet ngen.Packet, content *ngen.Buffer) ngen.Net {
	switch packet.Header.MsgType {
	case BenchyMsgType:
		msg := BenchyDeserialize(content)
		return &msg
	default:
		return nil
	}
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
