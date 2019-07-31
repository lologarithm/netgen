package ngen

import (
	"io"
	"testing"
)

func TestBasicBuffer(t *testing.T) {
	buf := NewBuffer(make([]byte, 0))

	buf.WriteByte(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteByte to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteUint16(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteUint16 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteInt16(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteInt16 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteUint32(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteUint32 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteInt32(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteInt32 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteRune(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteRune to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteInt(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteInt to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteUint64(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteUint64 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteInt64(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteInt64 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteFloat32(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteFloat32 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteFloat64(0)
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteFloat64 to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteString("asdf")
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteString to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf.WriteByteSlice([]byte{1})
	if buf.Err != io.EOF {
		t.Fatalf("Didn't fail when expected when WriteByteSlice to buffer: %s", buf.Err)
	}
	buf.Reset()

	buf = NewBuffer(make([]byte, 36))
	buf.WriteByteSlice([]byte{1, 2, 3})
	buf.WriteUint64(12)
	buf.WriteFloat64(123.456)
	buf.WriteFloat32(654.321)
	buf.WriteString("hello")
	if buf.Err != nil {
		t.Fatalf("Failed to serialize buffer: %s", buf.Err)
	}
	buf.Reset()

	slice := buf.ReadByteSlice()
	if slice[1] != 2 {
		t.Fatalf("Failed to read byte slice back out of buffer: %#v, %s", buf, buf.Err)
	}
	ui64 := buf.ReadUint64()
	if ui64 != 12 {
		t.Fatalf("Failed to read uint64 back out of buffer(%d): %#v, %s", ui64, buf, buf.Err)
	}
	if buf.ReadFloat64() != 123.456 {
		t.Fatalf("Failed to float64 back out of buffer: %#v, %s", buf, buf.Err)
	}
	if buf.ReadFloat32() != 654.321 {
		t.Fatalf("Failed to float32 back out of buffer: %#v, %s", buf, buf.Err)
	}
	if buf.ReadString() != "hello" {
		t.Fatalf("Failed to read string back out of buffer: %#v, %s", buf, buf.Err)
	}
}
