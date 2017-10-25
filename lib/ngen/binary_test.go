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
		v64 := Uint64(b)
		t.Logf("v64 value: %d\n", v64)
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
