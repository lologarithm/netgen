package main

import "github.com/lologarithm/netgen/lib/ngen"

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

