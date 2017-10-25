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
	buf := js.InternalObject(b).Get("$array").Get("buffer")
	view := js.Global.Get("Uint32Array").New(buf, 0, 2)
	new64 := uint64(0)
	js.InternalObject(new64).Set("$high", view.Index(1).Int())
	js.InternalObject(new64).Set("$low", view.Index(0).Int())
	return new64
}

func PutUint64(b []byte, v uint64) {
	iv := js.InternalObject(v)
	buf := js.InternalObject(b).Get("$array").Get("buffer")
	view := js.Global.Get("Uint32Array").New(buf, 0, 2)
	view.SetIndex(0, iv.Get("$low").Int())
	view.SetIndex(1, iv.Get("$high").Int())
}

func Float64(b []byte) float64 {
	buf := js.InternalObject(b).Get("$array").Get("buffer")
	view := js.Global.Get("Float64Array").New(buf, 0, 1)
	return view.Index(0).Float()
}

func PutFloat64(b []byte, v float64) {
	buf := js.InternalObject(b).Get("$array").Get("buffer")
	view := js.Global.Get("Float64Array").New(buf, 0, 1)
	view.SetIndex(0, v)
}
