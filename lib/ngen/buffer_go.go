// +build !js

package ngen

import "io"

type Buffer struct {
	Buf []byte
	Loc uint32
}

func NewBuffer(b []byte) *Buffer {
	return &Buffer{Buf: b}
}

func (b *Buffer) Reset() {
	b.Loc = 0
}

// ReadByte will read next byte from buffer and increment read location
func (b *Buffer) ReadByte() (byte, error) {
	if len(b.Buf) < int(b.Loc+1) {
		return 0, io.EOF
	}
	v := b.Buf[b.Loc]
	b.Loc++
	return v, nil
}

func (b *Buffer) ReadUint16() (uint16, error) {
	if len(b.Buf) < int(b.Loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.Buf[b.Loc:])
	b.Loc += 2
	return v, nil
}

func (b *Buffer) ReadInt16() (int16, error) {
	if len(b.Buf) < int(b.Loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.Buf[b.Loc:])
	b.Loc += 2
	return int16(v), nil
}

func (b *Buffer) ReadUint32() (uint32, error) {
	if len(b.Buf) < int(b.Loc+4) {
		return 0, io.EOF
	}
	v := Uint32(b.Buf[b.Loc:])
	b.Loc += 4
	return v, nil
}

func (b *Buffer) ReadInt32() (int32, error) {
	if len(b.Buf) < int(b.Loc+4) {
		return 0, io.EOF
	}
	v := Uint32(b.Buf[b.Loc:])
	b.Loc += 4
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
	if len(b.Buf) < int(b.Loc+8) {
		return 0, io.EOF
	}
	v := Uint64(b.Buf[b.Loc:])
	b.Loc += 8
	return v, nil
}

func (b *Buffer) ReadInt64() (int64, error) {
	if len(b.Buf) < int(b.Loc+8) {
		return 0, io.EOF
	}
	v := Uint64(b.Buf[b.Loc:])
	b.Loc += 8
	return int64(v), nil
}

func (b *Buffer) ReadFloat64() (float64, error) {
	if len(b.Buf) < int(b.Loc+8) {
		return 0, io.EOF
	}
	v := Float64(b.Buf[b.Loc:])
	b.Loc += 8
	return v, nil
}

func (b *Buffer) ReadString() (string, error) {
	l, err := b.ReadUint32()
	if err != nil {
		return "", err
	}
	if len(b.Buf) < int(b.Loc+l) {
		return "", io.EOF
	}
	v := string(b.Buf[b.Loc : b.Loc+l])
	b.Loc += l
	return v, err
}

func (b *Buffer) ReadByteSlice() ([]byte, error) {
	l, err := b.ReadUint32()
	if err != nil {
		return nil, err
	}
	if len(b.Buf) < int(b.Loc+l) {
		return nil, io.EOF
	}
	v := make([]byte, l)
	copy(v, b.Buf[b.Loc:b.Loc+l])
	b.Loc += l
	return v, nil
}

func (b *Buffer) ReadInt32Slice() ([]int32, error) {
	return nil, io.EOF
}
