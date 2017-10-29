package ngen

import "io"

type Buffer struct {
	buf []byte
	loc uint32
}

func NewBuffer(b []byte) *Buffer {
	return &Buffer{buf: b}
}

// ReadByte will read next byte from buffer and increment read location
func (b *Buffer) ReadByte() (byte, error) {
	if len(b.buf) <= int(b.loc) {
		return 0, io.EOF
	}
	v := b.buf[b.loc]
	b.loc++
	return v, nil
}

func (b *Buffer) ReadUint16() (uint16, error) {
	if len(b.buf) <= int(b.loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.buf[b.loc:])
	b.loc += 2
	return v, nil
}

func (b *Buffer) ReadInt16() (int16, error) {
	if len(b.buf) <= int(b.loc+2) {
		return 0, io.EOF
	}
	v := Uint16(b.buf[b.loc:])
	b.loc += 2
	return int16(v), nil
}

func (b *Buffer) ReadUint32() (uint32, error) {
	if len(b.buf) <= int(b.loc+4) {
		return 0, io.EOF
	}
	v := Uint32(b.buf[b.loc:])
	b.loc += 4
	return v, nil
}

func (b *Buffer) ReadInt32() (int32, error) {
	if len(b.buf) <= int(b.loc+4) {
		return 0, io.EOF
	}
	v := Uint32(b.buf[b.loc:])
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
	if len(b.buf) <= int(b.loc+8) {
		return 0, io.EOF
	}
	v := Uint64(b.buf[b.loc:])
	b.loc += 8
	return v, nil
}

func (b *Buffer) ReadInt64() (int64, error) {
	if len(b.buf) <= int(b.loc+8) {
		return 0, io.EOF
	}
	v := Uint64(b.buf[b.loc:])
	b.loc += 8
	return int64(v), nil
}

func (b *Buffer) ReadFloat64() (float64, error) {
	if len(b.buf) <= int(b.loc+8) {
		return 0, io.EOF
	}
	v := Float64(b.buf[b.loc:])
	b.loc += 8
	return v, nil
}

func (b *Buffer) ReadString() (string, error) {
	bs, err := b.ReadByteSlice()
	return string(bs), err
}

func (b *Buffer) ReadByteSlice() ([]byte, error) {
	l, err := b.ReadUint32()
	if err != nil {
		return nil, err
	}
	if len(b.buf) < int(b.loc+l) {
		return nil, io.EOF
	}
	v := make([]byte, l)
	copy(v, b.buf[b.loc:b.loc+l])
	return v, nil
}

func (b *Buffer) ReadInt32Slice() ([]int32, error) {
	return nil, io.EOF
}
