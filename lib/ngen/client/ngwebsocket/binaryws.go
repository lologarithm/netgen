// +build !js

package ngwebsocket

import (
	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngen/client"
	"golang.org/x/net/websocket"
)

// AcceptConn is used by a server to accept a new connection
// from a client.
func AcceptConn(conn *websocket.Conn) *client.Client {
	// fmt.Printf("Accepting Connection: %s\n", conn.RemoteAddr().String())
	return &client.Client{
		Name:     conn.RemoteAddr().String(),
		Conn:     &BinaryWebsocket{socket: conn},
		Outgoing: make(chan *ngen.Packet, 10),
		Incoming: make(chan *ngen.Packet, 10),
	}
}

// New is used by a go client to open a websocket to a server.
func New(url, origin string, onOpen func()) (*client.Client, error) {
	conn, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}
	ws := &BinaryWebsocket{
		socket: conn,
	}
	if onOpen != nil {
		go func() {
			onOpen()
		}()
	}

	return &client.Client{
		Conn:     ws,
		Outgoing: make(chan *ngen.Packet, 10),
		Incoming: make(chan *ngen.Packet, 10),
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
