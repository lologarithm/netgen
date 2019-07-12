// +build js

package ngwebsocket

import (
	"errors"
	"log"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngservice/client"
)

// New has _ to follow the pattern from the Go client.
func New(url, _ string, onConnect func()) (*client.Client, error) {
	conn := js.Global.Get("WebSocket").New(url)
	conn.Set("binaryType", "arraybuffer")
	conn.Call("addEventListener", "open", func(ev *js.Object) { onConnect() }, false)

	ws := &wsjs{
		conn:     conn,
		buffer:   make([]byte, 4096),
		idx:      0,
		framebuf: make(chan []byte, 2), // can hold 2 frames
	}

	onMessage := func(ev *js.Object) {
		jsarr := js.Global.Get("Uint8Array").New(ev.Get("data"))
		slice := jsarr.Interface().([]byte)
		go func() {
			ws.framebuf <- slice
		}()
	}

	conn.Call("addEventListener", "message", onMessage, false)

	return &client.Client{
		Conn:     ws,
		Outgoing: make(chan ngen.Message, 10),
		Incoming: make(chan ngen.Message, 10),
	}, nil
}

type wsjs struct {
	conn     *js.Object
	buffer   []byte
	idx      int
	framebuf chan []byte
}

func (ws *wsjs) Read(p []byte) (int, error) {
	if ws.idx == 0 {
		select {
		case slice := <-ws.framebuf:
			log.Printf("Got message from network: %v", slice)
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
	print("Writing bytes: ", p)
	ws.conn.Call("send", p)
	return len(p), err
}
