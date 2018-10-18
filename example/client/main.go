package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/lologarithm/netgen/example/newmodels"
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

// Events exposes a set of callbacks that the controller logic can
// register for.
// func (c *ClientJS) Events() *js.Object {
// 	return js.MakeWrapper(c.events)
// }

func (c *ClientJS) SendMessage(jso *js.Object) {
	c.Outgoing <- ngen.NewPacket(newmodels.MessageFromJS(jso))
}

func (c *ClientJS) Dial(url string) {
	go func() {
		if url == "" {
			url = "ws://127.0.0.1:4567/ws"
		}
		var err error
		c.Client, err = ngwebsocket.New(url, "", func() {
			print("Connection active. starting client now.")
			go runClient(c)
		})
		if err != nil {
			print("Failed to connect:", err)
			return
		}
	}()
}

// JS Client closure for processing network messages
func runClient(c *ClientJS) {
	newmodels.ManageClient(c.Client)
	for packet := range c.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case newmodels.MessageMsgType:
			print("Got message: ", packet.NetMsg.(*newmodels.Message).Message)
		}
	}
	print("Got nil packet, shutting down client reader.")
}
