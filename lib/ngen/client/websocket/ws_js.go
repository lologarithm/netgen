// +build js

package client

import (
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket/websocketjs"
	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngen/client"
)

func New(url, origin string) (client.Client, error) {
	conn, err := websocketjs.New(url)
	if err != nil {
		return nil, err
	}
	conn.BinaryType = "arraybuffer"

	ws := &wsjs{
		conn:     conn,
		buffer:   make([]byte, 4096),
		idx:      0,
		framebuf: make(chan []byte, 2), // can hold 2 frames
	}

	onMessage := func(ev *js.Object) {
		jsarr := js.Global.Get("Uint8Array").New(ev)
		slice := jsarr.Interface().([]byte)
		go func() {
			ws.framebuf <- slice
		}
	}

	conn.AddEventListener("message", false, onMessage)

	return &Client{
		Conn:     ws,
		Outgoing: make(chan *ngen.Packet, 20),
		Incoming: make(chan *ngen.Packet, 20),
	}, nil
}

type wsjs struct {
	conn     *websocketjs.WebSocket
	buffer   []byte
	idx      int
	framebuf chan []byte
}

func (ws *wsjs) Read(p []byte) (int, error) {
	if ws.idx == 0 {
		select {
		case 	slice := <- ws.framebuf:
			copy(ws.buffer[ws.idx:], slice)
			ws.idx += len(slice)
		case <-time.NewTimer(time.Second*60).C:
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
	return ws.conn.Close()
}

func (ws *wsjs) Write(p []byte) (int, error) {
	err = ws.conn.Send(p)
	// technically N is wrong here, but the err should make this ok...
	return len(p), err
}
