package ngen

import (
	"io"
	"math"
)

type Buffer struct {
	Buf []byte
	Loc uint32
	Err error
}

func NewBuffer(b []byte) *Buffer {
	return &Buffer{Buf: b}
}

// Bytes returns buffer up to the current write location.
// This is not useful for reading buffers.
func (b *Buffer) Bytes() []byte {
	return b.Buf[:b.Loc]
}

func (b *Buffer) Reset() {
	b.Loc = 0
	b.Err = nil
}

// READ FUNCS

// ReadBool will read next byte from buffer and increment read location
func (b *Buffer) ReadBool() bool {
	if b.Err != nil {
		return false
	}
	if len(b.Buf) < int(b.Loc+1) {
		b.Err = io.EOF
		return false
	}
	v := b.Buf[b.Loc] == 1
	b.Loc++
	return v
}

// ReadByte will read next byte from buffer and increment read location
func (b *Buffer) ReadByte() byte {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+1) {
		b.Err = io.EOF
		return 0
	}
	v := b.Buf[b.Loc]
	b.Loc++
	return v
}

func (b *Buffer) ReadUint16() uint16 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+2) {
		b.Err = io.EOF
		return 0
	}
	v := Uint16(b.Buf[b.Loc:])
	b.Loc += 2
	return v
}

func (b *Buffer) ReadInt16() int16 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+2) {
		b.Err = io.EOF
		return 0
	}
	_ = b.Buf[b.Loc+1] // bounds check hint to compiler; see golang.org/issue/14808
	v := uint16(b.Buf[b.Loc]) | uint16(b.Buf[b.Loc+1])<<8
	b.Loc += 2
	return int16(v)
}

func (b *Buffer) ReadUint32() uint32 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+4) {
		b.Err = io.EOF
		return 0
	}
	_ = b.Buf[b.Loc+3] // bounds check hint to compiler; see golang.org/issue/14808
	v := uint32(b.Buf[b.Loc]) | uint32(b.Buf[b.Loc+1])<<8 | uint32(b.Buf[b.Loc+2])<<16 | uint32(b.Buf[b.Loc+3])<<24
	b.Loc += 4
	return v
}

func (b *Buffer) ReadInt32() int32 {
	return int32(b.ReadUint32())
}

// ReadRune returns a single rune from the buffer
func (b *Buffer) ReadRune() rune {
	return b.ReadInt32()
}

func (b *Buffer) ReadInt() int {
	return int(b.ReadInt32())
}

func (b *Buffer) ReadUint64() uint64 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return 0
	}
	v := Uint64(b.Buf[b.Loc:])
	b.Loc += 8
	return v
}

func (b *Buffer) ReadInt64() int64 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return 0
	}
	v := Uint64(b.Buf[b.Loc:])
	b.Loc += 8
	return int64(v)
}

func (b *Buffer) ReadFloat32() float32 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return 0
	}
	v := Float32(b.Buf[b.Loc:])
	b.Loc += 8
	return v
}

func (b *Buffer) ReadFloat64() float64 {
	if b.Err != nil {
		return 0
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return 0
	}
	v := Float64(b.Buf[b.Loc:])
	b.Loc += 8
	return v
}

func (b *Buffer) ReadString() string {
	if b.Err != nil {
		return ""
	}
	l := b.ReadUint32()
	if len(b.Buf) < int(b.Loc+l) {
		b.Err = io.EOF
	}
	v := string(b.Buf[b.Loc : b.Loc+l])
	b.Loc += l
	return v
}

func (b *Buffer) ReadByteSlice() []byte {
	if b.Err != nil {
		return nil
	}
	l := b.ReadUint32()
	if len(b.Buf) < int(b.Loc+l) {
		b.Err = io.EOF
	}
	return b.readByteSlice(l)
}

func (b *Buffer) readByteSlice(length uint32) []byte {
	if b.Err != nil {
		return nil
	}
	if length == 0 {
		return nil
	}
	v := make([]byte, length)
	copy(v, b.Buf[b.Loc:b.Loc+length])
	b.Loc += length
	return v
}

// TODO: pointer hacks are fun
// func (b *Buffer) ReadInt32Slice() []int32 {
// 	return nil, io.EOF
// }

// WRITE FUNCS

func (b *Buffer) WriteBool(v bool) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+1) {
		b.Err = io.EOF
		return
	}
	if v {
		b.Buf[b.Loc] = 1
	} else {
		b.Buf[b.Loc] = 0
	}
	b.Loc++
	return
}

func (b *Buffer) WriteByte(v byte) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+1) {
		b.Err = io.EOF
		return
	}
	b.Buf[b.Loc] = v
	b.Loc++
	return
}

func (b *Buffer) WriteUint16(v uint16) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+2) {
		b.Err = io.EOF
		return
	}
	_ = b.Buf[b.Loc+1] // early bounds check to guarantee safety of writes below
	b.Buf[b.Loc] = byte(v)
	b.Buf[b.Loc+1] = byte(v >> 8)
	b.Loc += 2
}

func (b *Buffer) WriteInt16(v int16) {
	b.WriteUint16(uint16(v))
}

func (b *Buffer) WriteUint32(v uint32) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+4) {
		b.Err = io.EOF
		return
	}
	_ = b.Buf[b.Loc+3] // early bounds check to guarantee safety of writes below
	b.Buf[b.Loc] = byte(v)
	b.Buf[b.Loc+1] = byte(v >> 8)
	b.Buf[b.Loc+2] = byte(v >> 16)
	b.Buf[b.Loc+3] = byte(v >> 24)
	b.Loc += 4
	return
}

func (b *Buffer) WriteInt32(v int32) {
	b.WriteUint32(uint32(v))
}

func (b *Buffer) WriteRune(v rune) {
	b.WriteUint32(uint32(v))
}

// WriteInt writes an int as int32
func (b *Buffer) WriteInt(v int) {
	b.WriteInt32(int32(v))
}

func (b *Buffer) WriteUint64(v uint64) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return
	}
	_ = b.Buf[b.Loc+7] // early bounds check to guarantee safety of writes below
	b.Buf[b.Loc] = byte(v)
	b.Buf[b.Loc+1] = byte(v >> 8)
	b.Buf[b.Loc+2] = byte(v >> 16)
	b.Buf[b.Loc+3] = byte(v >> 24)
	b.Buf[b.Loc+4] = byte(v >> 32)
	b.Buf[b.Loc+5] = byte(v >> 40)
	b.Buf[b.Loc+6] = byte(v >> 48)
	b.Buf[b.Loc+7] = byte(v >> 56)
	b.Loc += 8
	return
}

func (b *Buffer) WriteInt64(v int64) {
	b.WriteUint64(uint64(v))
}

func (b *Buffer) WriteFloat32(v float32) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return
	}
	b.WriteUint32(math.Float32bits(v))
}

func (b *Buffer) WriteFloat64(v float64) {
	if b.Err != nil {
		return
	}
	if len(b.Buf) < int(b.Loc+8) {
		b.Err = io.EOF
		return
	}
	b.WriteUint64(math.Float64bits(v))
}

func (b *Buffer) WriteString(v string) {
	b.WriteByteSlice([]byte(v))
}

func (b *Buffer) WriteByteSlice(v []byte) {
	b.WriteUint32(uint32(len(v)))
	b.writeByteSlice(v)
}

func (b *Buffer) writeByteSlice(v []byte) {
	if b.Err != nil {
		return
	}
	l := len(v)
	if l == 0 {
		return
	}
	if len(b.Buf) < int(b.Loc)+l {
		b.Err = io.EOF
		return
	}
	copy(b.Buf[b.Loc:], v)
	b.Loc += uint32(l)
}

//
// func (b *Buffer) ReadInt32Slice() []int32 {
// 	return nil, io.EOF
// }
