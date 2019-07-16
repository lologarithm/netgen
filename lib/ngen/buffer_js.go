// +build js

package ngen

import (
	"io"

	"github.com/gopherjs/gopherjs/js"
)

type Buffer struct {
	buf  []byte
	ab   *js.Object
	view *js.Object
	loc  uint32
	Err  error
}

func NewBuffer(b []byte) *Buffer {
	bytejs := js.InternalObject(b)
	arr := bytejs.Get("$array")
	ib := arr.Get("buffer")
	if bytejs.Get("$offset") != js.InternalObject(0) || bytejs.Get("$length") != bytejs.Get("byteLength") {
		ib = ib.Call("slice", bytejs.Get("$offset"), bytejs.Get("$offset").Int()+bytejs.Get("$length").Int())
	}
	dv := js.Global.Get("DataView").New(ib)
	return &Buffer{buf: b, ab: ib, view: dv}
}

// Bytes returns buffer up to the current write location.
// This is not useful for reading buffers.
func (b *Buffer) Bytes() []byte {
	return b.buf[:b.loc]
}

func (b *Buffer) Reset() {
	b.loc = 0
}

// ReadByte will read next byte from buffer and increment read location
func (b *Buffer) ReadByte() byte {
	if b.Err != nil {
		return 0
	}
	if len(b.buf) < int(b.loc+1) {
		b.Err = io.EOF
		return 0
	}
	v := b.buf[b.loc]
	b.loc++
	return v
}

func (b *Buffer) ReadUint16() uint16 {
	if b.Err != nil {
		return 0
	}
	if len(b.buf) < int(b.loc+2) {
		b.Err = io.EOF
		return 0
	}
	v := Uint16(b.buf[b.loc:])
	b.loc += 2
	return v
}

func (b *Buffer) ReadInt16() int16 {
	return int16(b.ReadUint16())
}

func (b *Buffer) ReadUint32() uint32 {
	if b.Err != nil {
		return 0
	}
	if len(b.buf) < int(b.loc+4) {
		b.Err = io.EOF
		return 0
	}
	v := uint32(b.buf[b.loc]) | uint32(b.buf[b.loc+1])<<8 | uint32(b.buf[b.loc+2])<<16 | uint32(b.buf[b.loc+3])<<24
	// v := b.view.Call("getUint32", b.loc, js.InternalObject(true)).Int()
	b.loc += 4
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

	if len(b.buf) < int(b.loc+8) {
		b.Err = io.EOF
		return 0
	}
	new64 := uint64(0)
	js.InternalObject(new64).Set("$low", b.view.Call("getUint32", b.loc, true).Int())
	js.InternalObject(new64).Set("$high", b.view.Call("getUint32", b.loc+4, true).Int())
	b.loc += 8
	return new64
}

func (b *Buffer) ReadInt64() int64 {
	return int64(b.ReadUint64())
}

func (b *Buffer) ReadFloat32() float32 {
	if b.Err != nil {
		return 0
	}
	if len(b.buf) < int(b.loc+4) {
		b.Err = io.EOF
		return 0
	}
	v := b.view.Call("getFloat32", b.loc, true).Float()
	b.loc += 4
	return float32(v)
}

func (b *Buffer) ReadFloat64() float64 {
	if b.Err != nil {
		return 0
	}
	if len(b.buf) < int(b.loc+8) {
		b.Err = io.EOF
		return 0
	}
	v := b.view.Call("getFloat64", b.loc, true).Float()
	b.loc += 8
	return v
}

func (b *Buffer) ReadString() string {
	return string(b.readByteSlice(b.ReadUint32()))
}

func (b *Buffer) ReadByteSlice() []byte {
	return b.readByteSlice(b.ReadUint32())
}

func (b *Buffer) readByteSlice(length uint32) []byte {
	if b.Err != nil {
		return nil
	}
	if len(b.buf) < int(b.loc+length) {
		b.Err = io.EOF
		return nil
	}
	v := make([]byte, length)
	copy(v, b.buf[b.loc:b.loc+length])
	b.loc += length
	return v
}

// func (b *Buffer) ReadInt32Slice() ([]int32, error) {
// 	return nil, io.EOF
// }

func (b *Buffer) WriteByte(v byte) {
	if b.Err != nil {
		return
	}
	if len(b.buf) < int(b.loc+1) {
		b.Err = io.EOF
		return
	}
	b.buf[b.loc] = v
	b.loc++
	return
}

func (b *Buffer) WriteUint16(v uint16) {
	if b.Err != nil {
		return
	}
	if len(b.buf) < int(b.loc+2) {
		b.Err = io.EOF
		return
	}
	_ = b.buf[b.loc+1] // early bounds check to guarantee safety of writes below
	b.buf[b.loc] = byte(v)
	b.buf[b.loc+1] = byte(v >> 8)
	b.loc += 2
}

func (b *Buffer) WriteInt16(v int16) {
	b.WriteUint16(uint16(v))
}

func (b *Buffer) WriteUint32(v uint32) {
	if b.Err != nil {
		return
	}
	if len(b.buf) < int(b.loc+4) {
		b.Err = io.EOF
		return
	}
	_ = b.buf[b.loc+3] // early bounds check to guarantee safety of writes below
	b.buf[b.loc] = byte(v)
	b.buf[b.loc+1] = byte(v >> 8)
	b.buf[b.loc+2] = byte(v >> 16)
	b.buf[b.loc+3] = byte(v >> 24)
	b.loc += 4
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
	if len(b.buf) < int(b.loc+8) {
		b.Err = io.EOF
		return
	}
	iv := js.InternalObject(v)
	// TODO: what if this doesnt align with the 4 bytes?
	b.view.Call("setUint32", b.loc, iv.Get("$low").Int(), true)
	b.view.Call("setUint32", b.loc+4, iv.Get("$high").Int(), true)
	b.loc += 8
	return
}

func (b *Buffer) WriteInt64(v int64) {
	b.WriteUint64(uint64(v))
}

func (b *Buffer) WriteFloat32(v float32) {
	if b.Err != nil {
		return
	}
	if len(b.buf) < int(b.loc+4) {
		b.Err = io.EOF
		return
	}
	b.view.Call("setFloat32", b.loc, v, true)
	b.loc += 4
}

func (b *Buffer) WriteFloat64(v float64) {
	if b.Err != nil {
		return
	}
	if len(b.buf) < int(b.loc+8) {
		b.Err = io.EOF
		return
	}
	b.view.Call("setFloat64", b.loc, v, true)
	b.loc += 8
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
	if len(b.buf) < int(b.loc)+l {
		b.Err = io.EOF
		return
	}
	copy(b.buf[b.loc:], v)
	b.loc += uint32(l)
}
