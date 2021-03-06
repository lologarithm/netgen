package client

import (
	"fmt"
	"io"

	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngservice"
)

type Client struct {
	ID       int32
	Name     string
	Conn     io.ReadWriteCloser
	Outgoing chan ngen.Message
	Incoming chan ngen.Message
}

func ManageClient(ctx *ngen.Context, c *Client) {
	settingsSync := make(chan *ngen.Context)
	go Sender(c, ctx, settingsSync)
	go Reader(c, ctx, settingsSync)
}

// Reader spawns a block for loop reading off the conn on Client
// it will put all read packets onto the incoming channel.
// This code requires the conn to not shard packets.
func Reader(c *Client, local *ngen.Context, remote chan *ngen.Context) {
	idx := 0
	buffer := make([]byte, 4096)
	// Cached versioning info.
	// This means we don't have to send it on every request, only on each connection.

	remoteSettings := local // Use local settings until we have a remote.

	for {
		n, err := c.Conn.Read(buffer[idx:])
		if err != nil {
			// ded?
			break
		} else if n == 0 {
			break
		} else if n+idx >= len(buffer) {
			// Expand buffer to hold the message!
			newbuff := make([]byte, len(buffer)*2)
			copy(newbuff, buffer)
			buffer = newbuff
			idx += n
			continue
		}

		p, ok := ngservice.ReadPacket(remoteSettings, buffer[:idx+n])
		if !ok {
			// increment idx by how much we wrote.
			idx += n
			// fmt.Printf("No Packet... Buffer: %v\n", buffer[:idx])
			continue
		}

		if p.Header.MsgType == ngen.MessageTypeContext {
			fmt.Printf("Got remote settings: %#v\n", p.NetMsg)
			remoteSettings = p.NetMsg.(*ngen.Context)
			remote <- remoteSettings // send to 'sender' channel now
		} else {
			// Successful packet read
			c.Incoming <- p.NetMsg
		}

		// copy back in case we have multiple packets in the buffer
		l := p.Len()
		if l != idx {
			copy(buffer, buffer[l:])
		}
		idx = 0
	}
	close(c.Incoming)
}

func Sender(c *Client, local *ngen.Context, remote chan *ngen.Context) {
	remoteSettings := local // start with local settings by default

	// Only send and wait for versioning message if we have versioned messages
	if len(local.FieldVersions) > 0 {
		// First message out is the settings (versioning info) for this instance.
		// This will allow the other side to read our versioned structs.
		n, err := c.Conn.Write(ngservice.WriteMessage(nil, local))
		if err != nil || n == 0 {
			fmt.Printf("Failed to write handshake settings with remote: %s", err.Error())
			return
		}

		remoteSettings = <-remote
	}

	for m := range c.Outgoing {
		if m == nil {
			return // Empty message means die
		}
		n, err := c.Conn.Write(ngservice.WriteMessage(remoteSettings, m))
		if err != nil {
			fmt.Printf("Writing failed: %s\n", err.Error())
			break
		} else if n == 0 {
			fmt.Printf("outgoing write failed, shutting down writer.\n")
			break
		}
	}
}
