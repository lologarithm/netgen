// +build js

package ngen

import (
	"github.com/gopherjs/gopherjs/js"
)

// This is simply a clone of littleEndian conversion code from std lib.
// This is done to allow for custom implementations

func Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[0]) | uint16(b[1])<<8
}

func PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
}

func Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

func PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
}

func Uint64(b []byte) uint64 {
	iba := js.InternalObject(b)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	new64 := uint64(0)
	js.InternalObject(new64).Set("$low", view.Call("getUint32", 0, true).Int())
	js.InternalObject(new64).Set("$high", view.Call("getUint32", 4, true).Int())
	return new64
}

func PutUint64(b []byte, v uint64) {
	iba := js.InternalObject(b)
	iv := js.InternalObject(v)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	view.Call("setUint32", 0, iv.Get("$low").Int(), true)
	view.Call("setUint32", 4, iv.Get("$high").Int(), true)
}

func Float32(b []byte) float32 {
	iba := js.InternalObject(b)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	return float32(view.Call("getFloat32", 0, true).Float())
}

func Float64(b []byte) float64 {
	iba := js.InternalObject(b)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	return view.Call("getFloat64", 0, true).Float()
}

func PutFloat32(b []byte, v float32) {
	iba := js.InternalObject(b)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	view.Call("setFloat32", 0, v, true)
}

func PutFloat64(b []byte, v float64) {
	iba := js.InternalObject(b)
	buf := iba.Get("$array").Get("buffer")
	view := js.Global.Get("DataView").New(buf, iba.Get("$offset"), 8)
	view.Call("setFloat64", 0, v, true)
}
