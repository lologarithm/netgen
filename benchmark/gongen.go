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



func (m Features) Serialize(buffer []byte) {
	idx := 0
	m.Dynd.Serialize(buffer[idx:])
	idx+=m.Dynd.Len()
	ngen.PutUint32(buffer[idx:], uint32(len(m.Bin)))
	idx += 4
	copy(buffer[idx:], m.Bin)
	idx+=len(m.Bin)
	ngen.PutUint32(buffer[idx:], uint32(len(m.OtherFeatures)))
	idx += 4
	for _, v2 := range m.OtherFeatures {
		if v2 != nil {
				buffer[idx] = 1
				idx++
				v2.Serialize(buffer[idx:])
		idx+=v2.Len()
		} else {
		buffer[idx] = 0
		idx++
		}	}
	m.DatBenchy.Serialize(buffer[idx:])
	idx+=m.DatBenchy.Len()
		ngen.PutUint32(buffer[idx:], uint32(m.EnumyV))
	idx+=4
}

func (m Features) Len() int {
	mylen := 0
	
	mylen += 4 + len(m.Bin)
	mylen += 4
	for _, v2 := range m.OtherFeatures {
	_ = v2
		mylen += v2.Len()		
mylen++
	}
	mylen += m.DatBenchy.Len()
	mylen += 4
	return mylen
}

func (m Features) MsgType() ngen.MessageType {
	return FeaturesMsgType
}

