// +build !js

package client

import (
	"github.com/lologarithm/netgen/lib/ngen"
	"golang.org/x/net/websocket"
)

func AcceptConn(conn *websocket.Conn) Client {
	// fmt.Printf("Accepting Connection: %s\n", conn.RemoteAddr().String())
	return Client{
		Name:     conn.RemoteAddr().String(),
		Conn:     &BinaryWebsocket{socket: conn},
		Outgoing: make(chan *ngen.Packet, 100),
		Incoming: make(chan *ngen.Packet, 100),
	}
}

func New(url, origin string) (*Client, error) {
	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	ws := &BinaryWebsocket{
		socket: conn,
	}

	return &Client{
		Conn:     ws,
		Outgoing: make(chan *ngen.Packet, 100),
		Incoming: make(chan *ngen.Packet, 100),
	}, nil
}

type BinaryWebsocket struct {
	socket *websocket.Conn
}

func (ws *BinaryWebsocket) Read(p []byte) (n int, err error) {
	var inbuf []byte
	err = websocket.Message.Receive(ws.socket, &inbuf)
	if err != nil {
		return 0, err
	}
	copy(p, inbuf)
	return len(inbuf), err
}

func (ws *BinaryWebsocket) Close() error {
	return ws.socket.Close()
}

func (ws *BinaryWebsocket) Write(p []byte) (n int, err error) {
	err = websocket.Message.Send(ws.socket, p)
	return len(p), err
}
