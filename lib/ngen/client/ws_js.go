// +build js

package client

import (
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
	"github.com/lologarithm/netgen/lib/ngen"
)

func New(url, origin string) (*Client, error) {
	conn, err := websocketjs.New(url)
	if err != nil {
		return nil, err
	}
	conn.BinaryType = "arraybuffer"

	ws := &wsjs{
		conn:   conn,
		buffer: make([]byte, 4096),
		idx:    0,
	}

	onMessage := func(ev *js.Object) {
		jsarr := js.Global.Get("Uint8Array").New(ev)
		slice := jsarr.Interface().([]byte)
		copy(ws.buffer[ws.idx:], slice)
		ws.idx += len(slice)
	}

	conn.AddEventListener("message", false, onMessage)

	return &Client{
		Conn:     ws,
		Outgoing: make(chan *ngen.Packet, 100),
		Incoming: make(chan *ngen.Packet, 100),
	}, nil
}

type wsjs struct {
	conn   *websocketjs.WebSocket
	buffer []byte
	idx    int
}

func (ws *wsjs) Read(p []byte) (int, error) {
	for ws.idx == 0 {
		// So this only works because JS is single threaded.... this is probably bad.
		// Just waiting for an event to push data onto the struct.
		time.Sleep(time.Millisecond * 100)
		// TODO: implement a read timeout here.
	}
	num := ws.idx
	copy(p, ws.buffer[:num])
	ws.idx = 0
	return num, nil
}

func (ws *wsjs) Close() error {
	return ws.conn.Close()
}

func (ws *wsjs) Write(p []byte) (int, error) {
	err = ws.conn.Send(p)
	// technically N is wrong here, but the err should make this ok...
	return len(p), err
}
