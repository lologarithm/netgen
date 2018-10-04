package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/lologarithm/netgen/example/models"
	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngen/client"
	"github.com/lologarithm/netgen/lib/ngen/client/ngwebsocket"
)

func main() {
	print("loaded")
	js.Global.Set("ngex", map[string]interface{}{
		"newClient": newClient,
	})
}

func newClient() *js.Object {
	c := &ClientJS{
		// events: &ClientEvents{},
	}
	return js.MakeWrapper(c)
}

type ClientJS struct {
	*client.Client
	// events *ClientEvents
}

// func (c *ClientJS) Events() *js.Object {
// 	return js.MakeWrapper(c.events)
// }

func (c *ClientJS) SendMessage(jso *js.Object) {
	c.Outgoing <- ngen.NewPacket(models.MessageFromJS(jso))
}

func (c *ClientJS) Dial(url string) {
	go func() {
		if url == "" {
			url = "ws://127.0.0.1:4567/ws"
		}
		var err error
		c.Client, err = ngwebsocket.New(url, "")
		if err != nil {
			print("Failed to connect:", err)
			return
		}
		go runClient(c)
		// , func(o *js.Object) {
		// 	c.Outgoing <- mafianet.NewPacket(&mafianet.Heartbeat{})
		// 	go runClient(c)
		// }, func(o *js.Object) {
		// }, func(o *js.Object) {
		// })
	}()
}

// JS Client closure for processing network messages
func runClient(c *ClientJS) {
	go client.Sender(c.Client)
	go client.Reader(c.Client, models.ParseNetMessage)

	for packet := range c.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case models.MessageMsgType:
			print("Got message: ", packet.NetMsg.(*models.Message).Message)
		}
	}
	print("Got nil packet, shutting down client reader.")
}
