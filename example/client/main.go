package main

import (
	"fmt"

	"github.com/lologarithm/netgen/example/newmodels"
	"github.com/lologarithm/netgen/lib/ngservice/client"
	"github.com/lologarithm/netgen/lib/ngservice/client/ngwebsocket"
)

func main() {
	print("Starting!\n")
	blocker := make(chan bool)
	c := &Client{
		CEvents: &Events{},
	} // ngex.newClient()
	c.CEvents.OnConnected(func(arg1 bool) {
		c.Outgoing <- newmodels.Message{Message: "Hello World."}
		c.Outgoing <- newmodels.VersionedMessage{Message: "Version Hello World!", From: "The Client", NewHotness: 42.0}
		print("Connected...\n")
	})
	c.Dial("ws://127.0.0.1:4567/ws") // example/server websocket endpoint
	<-blocker
}

func runClient(c *Client) {
	for msg := range c.Incoming {
		if msg == nil {
			break
		}
		switch msg.MsgType() {
		case newmodels.MessageMsgType:
			print("Got message: ", msg.(*newmodels.Message).Message, "\n")
		case newmodels.VersionedMessageMsgType:
			vm := msg.(*newmodels.VersionedMessage)
			print("Got versioned message: `" + vm.Message + "` From: `" + vm.From + "`\n")
		}
	}
	print("Got nil packet, shutting down client reader.\n")
}

type Client struct {
	*client.Client
	CEvents *Events
}

type Events struct {
	connected func(bool)
}

func (ce *Events) OnConnected(cb func(bool)) {
	ce.connected = cb
}

// If implementing a go network client but js controller, you would expose event objects.
// // Events exposes a set of callbacks that the controller logic can register for.
// func (c *Client) Events() js.Value {
// 	return js.MakeWrapper(c.CEvents)
// }
//
// func (c *Client) SendMessage(jso js.Value) {
// 	c.Outgoing <- newmodels.MessageFromJS(jso)
// }
//
// func (c *Client) SendVerMessage(jso js.Value) {
// 	c.Outgoing <- newmodels.VersionedMessageFromJS(jso)
// }

func (c *Client) Dial(url string) {
	var err error
	c.Client, err = ngwebsocket.New(url, "", func() {
		print("Connection active. starting client now.\n")
		if c.CEvents != nil && c.CEvents.connected != nil {
			c.CEvents.connected(true)
		}
		client.ManageClient(newmodels.Context, c.Client)
		go runClient(c)
	})
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err.Error())
		c.CEvents.connected(false)
		return
	}
}
