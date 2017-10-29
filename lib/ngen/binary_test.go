package ngen

import "testing"

func TestEncode(t *testing.T) {
	b := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	PutUint16(b, 1)
	v16 := Uint16(b)
	if v16 != 1 {
		t.FailNow()
	}

	PutUint32(b, 65537)
	v32 := Uint32(b)
	if v32 != 65537 {
		t.Logf("v32 value: %d\n", v32)
		t.FailNow()
	}

	PutUint64(b, 4295032833)
	v64 := Uint64(b)
	if v64 != 4295032833 {
		t.Logf("v64 value: %d\n", v64)
		t.FailNow()
	}
}

func TestEncodeMulti(t *testing.T) {
	b := make([]byte, 2+4+8+4+8)
	PutUint16(b, 1)
	PutUint32(b[2:], 65537)
	PutUint64(b[6:], 4295032833)
	PutFloat64(b[14:], 123.456)
	PutUint32(b[22:], 65536)
	if Uint16(b[0:]) != 1 {
		t.Logf("uint16 failed...")
		t.FailNow()
	}
	if Uint32(b[2:]) != 65537 {
		t.Logf("uint32 wrong")
		t.FailNow()
	}
	if Uint64(b[6:]) != 4295032833 {
		v64 := Uint64(b[6:])
		t.Logf("v64 value: %d\n", v64)
		t.FailNow()
	}
	if Float64(b[14:]) != 123.456 {
		f64 := Float64(b[14:])
		t.Logf("f64 value: %f\n", f64)
		t.FailNow()
	}
	if Uint32(b[22:]) != 65536 {
		u32 := Uint32(b[22:])
		t.Logf("u32 value: %d\n", u32)
		t.FailNow()
	}
}

func TestDecode(t *testing.T) {
	b := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	v16 := Uint16(b)
	if v16 != 1 {
		t.FailNow()
	}
	v32 := Uint32(b)
	if v32 != 65537 {
		t.Logf("value: %d\n", v32)
		t.FailNow()
	}
	v64 := Uint64(b)
	if v64 != 4295032833 {
		t.Logf("value: %d\n", v64)
		t.FailNow()
	}
}

func BenchmarkDecode16(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		Uint16(bytes)
	}
}

func BenchmarkDecode32(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		Uint32(bytes)
	}
}

func BenchmarkDecode64(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		Uint64(bytes)
	}
}

func BenchmarkDecodeF64(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		Float64(bytes)
	}
}

func BenchmarkEncode16(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		PutUint16(bytes, 1)
	}
}

func BenchmarkEncode32(b *testing.B) {
	bytes := []byte{1, 0, 1, 0, 1, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		PutUint32(bytes, 65337)
	}
}

func BenchmarkEncode64(b *testing.B) {
	bytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		PutUint64(bytes, 4295032833)
	}
}

func BenchmarkEncodeF64(b *testing.B) {
	bytes := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < b.N; i++ {
		PutFloat64(bytes, 4295032833)
	}
}
