package models

import "github.com/lologarithm/netgen/lib/ngen"

func (m Message) Serialize(buffer []byte) {
	idx := 0
	ngen.PutUint32(buffer[idx:], uint32(len(m.Message)))
	idx += 4
	copy(buffer[idx:], []byte(m.Message))
	idx += len(m.Message)
}

func (m Message) Len() int {
	mylen := 0
	mylen += 4 + len(m.Message)
	return mylen
}

func (m Message) MsgType() ngen.MessageType {
	return MessageMsgType
}

