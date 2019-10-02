package ngwebsocket

import (
	"errors"
	"syscall/js"
	"time"

	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngservice/client"
)

// New has _ to follow the pattern from the Go client.
func New(url, _ string, onConnect func()) (*client.Client, error) {
	conn := js.Global().Get("WebSocket").New(url)
	conn.Set("binaryType", "arraybuffer")
	conn.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} { onConnect(); return nil }), false)

	ws := &wsjs{
		conn:     conn,
		buffer:   make([]byte, 4096),
		idx:      0,
		framebuf: make(chan []byte, 2), // can hold 2 frames
	}

	onMessage := func(this js.Value, args []js.Value) interface{} {
		data := args[0].Get("data")
		slice := make([]byte, data.Get("byteLength").Int())
		view := js.Global().Get("Uint8Array").New(data)
		js.CopyBytesToGo(slice, view)
		go func() {
			ws.framebuf <- slice
		}()
		return nil
	}

	conn.Call("addEventListener", "message", js.FuncOf(onMessage), false)

	return &client.Client{
		Conn:     ws,
		Outgoing: make(chan ngen.Message, 10),
		Incoming: make(chan ngen.Message, 10),
	}, nil
}

type wsjs struct {
	conn     js.Value
	buffer   []byte
	idx      int
	framebuf chan []byte
}

func (ws *wsjs) Read(p []byte) (int, error) {
	if ws.idx == 0 {
		select {
		case slice := <-ws.framebuf:
			copy(ws.buffer[ws.idx:], slice)
			ws.idx += len(slice)
		case <-time.NewTimer(time.Second * 60).C:
			// No message for 60 seconds.. seems like its dead?
			return 0, errors.New("failed to read")
		}
	}

	num := ws.idx
	if len(p) < num {
		num = len(p) // can't read more than will fit.
	}
	copy(p, ws.buffer[:num])
	ws.idx -= num
	return num, nil
}

func (ws *wsjs) Close() error {
	ws.conn.Call("close")
	return nil
}

func (ws *wsjs) Write(p []byte) (int, error) {
	// technically N is wrong here, but the err should make this ok...
	var err error
	defer func() {
		e := recover()
		if e == nil {
			return
		}
		if jsErr, ok := e.(*js.Error); ok && jsErr != nil {
			err = jsErr
		} else {
			panic(e)
		}
	}()

	view := js.Global().Get("Uint8Array").New(len(p))
	num := js.CopyBytesToJS(view, p)
	if num != len(p) {
		err = errors.New("failed to copy bytes all bytes to js for network transmission")
	}
	ws.conn.Call("send", view)
	return num, err
}
