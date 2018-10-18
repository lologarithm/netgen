package client

import (
	"fmt"
	"io"

	"github.com/lologarithm/netgen/lib/ngen"
)

type Client struct {
	ID       int32
	Name     string
	Conn     io.ReadWriteCloser
	Outgoing chan *ngen.Packet
	Incoming chan *ngen.Packet
}

// Reader spawns a block for loop reading off the conn on Client
// it will put all read packets onto the incoming channel.
// This code requires the conn to not shard packets.
func Reader(c *Client, parser ngen.NetParser) {
	idx := 0
	buffer := make([]byte, 4096)
	// Cached versioning info.
	// This means we don't have to send it on every request, only on each connection.
	var remoteSettings *ngen.Settings

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

		p, ok := ngen.NextPacket(buffer[:idx+n], parser, remoteSettings)
		if !ok {
			// increment idx by how much we wrote.
			idx += n
			// fmt.Printf("No Packet... Buffer: %v\n", buffer[:idx])
			continue
		}

		if p.Header.MsgType == 0 {
			fmt.Printf("Got remote settings: %#v\n", p.NetMsg)
			remoteSettings = p.NetMsg.(*ngen.Settings)
		} else {
			// Successful packet read
			c.Incoming <- &p
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

func Sender(c *Client, ver *ngen.Settings) {
	// First message out is the settings (versioning info) for this instance.
	// This will allow the other side to read our versioned structs.
	c.Outgoing <- ngen.NewPacket(ver)
	fmt.Printf("ver data queued, now reading outgoing packets.\n")
	for m := range c.Outgoing {
		if m == nil {
			return // Empty message means die
		}
		n, err := c.Conn.Write(m.Pack())
		if err != nil {
			fmt.Printf("Writing failed: %s\n", err.Error())
			break
		} else if n == 0 {
			fmt.Printf("outgoing write failed, shutting down writer.\n")
			break
		}
	}
}
