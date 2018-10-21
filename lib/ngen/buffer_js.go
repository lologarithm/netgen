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

func (b *Buffer) Reset() {
	b.loc = 0
}

// ReadByte will read next byte from buffer and increment read location
func (b *Buffer) ReadByte() (byte, error) {
	if len(b.buf) < int(b.loc+1) {
		return 0, io.EOF
	}
	v := b.buf[b.loc]
	b.loc++
	return v, nil
}

func (b *Buffer) ReadUint16() (uint16, error) {
	if len(b.buf) < int(b.loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.buf[b.loc:])
	b.loc += 2
	return v, nil
}

func (b *Buffer) ReadInt16() (int16, error) {
	if len(b.buf) < int(b.loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.buf[b.loc:])
	b.loc += 2
	return int16(v), nil
}

func (b *Buffer) ReadUint32() (uint32, error) {
	if len(b.buf) < int(b.loc+4) {
		return 0, io.EOF
	}
	v := b.view.Call("getUint32", b.loc, js.InternalObject(true)).Int()
	b.loc += 4
	return uint32(v), nil
}

func (b *Buffer) ReadInt32() (int32, error) {
	if len(b.buf) < int(b.loc+4) {
		return 0, io.EOF
	}
	v := b.view.Call("getUint32", b.loc, js.InternalObject(true)).Int()
	b.loc += 4
	return int32(v), nil
}

// ReadRune returns a single rune from the buffer
func (b *Buffer) ReadRune() (rune, error) {
	return b.ReadInt32()
}

func (b *Buffer) ReadInt() (int, error) {
	v, err := b.ReadInt32()
	return int(v), err
}

func (b *Buffer) ReadUint64() (uint64, error) {
	if len(b.buf) < int(b.loc+8) {
		return 0, io.EOF
	}
	new64 := uint64(0)
	js.InternalObject(new64).Set("$low", b.view.Call("getUint32", b.loc, true).Int())
	js.InternalObject(new64).Set("$high", b.view.Call("getUint32", b.loc+4, true).Int())
	b.loc += 8
	return new64, nil
}

func (b *Buffer) ReadInt64() (int64, error) {
	v, err := b.ReadUint64()
	return int64(v), err
}

func (b *Buffer) ReadFloat64() (float64, error) {
	if len(b.buf) < int(b.loc+8) {
		return 0, io.EOF
	}
	v := b.view.Call("getFloat64", b.loc, true).Float()
	b.loc += 8
	return v, nil
}

func (b *Buffer) ReadString() (string, error) {
	l, err := b.ReadUint32()
	if err != nil {
		return "", err
	}
	if len(b.buf) < int(b.loc+l) {
		return "", io.EOF
	}
	v := string(b.buf[b.loc : b.loc+l])
	b.loc += l
	return v, err
}

func (b *Buffer) ReadByteSlice() ([]byte, error) {
	l, err := b.ReadUint32()
	if err != nil {
		return nil, err
	}
	return b.readByteSlice(l)
}

func (b *Buffer) readByteSlice(length uint32) ([]byte, error) {
	if len(b.buf) < int(b.loc+length) {
		return nil, io.EOF
	}
	v := make([]byte, length)
	copy(v, b.buf[b.loc:b.loc+length])
	b.loc += length
	return v, nil
}

func (b *Buffer) ReadInt32Slice() ([]int32, error) {
	return nil, io.EOF
}
