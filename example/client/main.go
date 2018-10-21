package main

import (
	"fmt"

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
	c := &Client{
		CEvents: &Events{},
	}
	return js.MakeWrapper(c)
}

func runClient(c *Client) {
	for packet := range c.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case newmodels.MessageMsgType:
			print("Got message: ", packet.NetMsg.(*newmodels.Message).Message, "\n")
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

// Events exposes a set of callbacks that the controller logic can
// register for.
func (c *Client) Events() *js.Object {
	return js.MakeWrapper(c.CEvents)
}

func (c *Client) SendMessage(jso *js.Object) {
	c.Outgoing <- ngen.NewPacket(newmodels.MessageFromJS(jso))
}

func (c *Client) SendVerMessage(jso *js.Object) {
	c.Outgoing <- ngen.NewPacket(newmodels.VersionedMessageFromJS(jso))
}

func (c *Client) Dial(url string) {
	go func() {
		origin := ""
		if url == "" {
			url = "ws://127.0.0.1:4567/ws"
			origin = "http://127.0.0.1/"
		}
		var err error
		c.Client, err = ngwebsocket.New(url, origin, func() {
			print("Connection active. starting client now.\n")
			if c.CEvents != nil && c.CEvents.connected != nil {
				c.CEvents.connected(true)
			}
			newmodels.ManageClient(c.Client)
			go runClient(c)
		})
		if err != nil {
			fmt.Printf("Failed to connect: %s\n", err.Error())
			c.CEvents.connected(false)
			return
		}
	}()
}
