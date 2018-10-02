package client

import (
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
	for {
		n, err := c.Conn.Read(buffer[idx:])
		if err != nil {
			c.Incoming <- nil
			c.Outgoing <- nil
			return
			// ded?
		} else if n == 0 {
			c.Incoming <- nil
			c.Outgoing <- nil
			// ded?
			return
		} else if n+idx >= len(buffer) {
			// Expand buffer to hold the message!
			newbuff := make([]byte, len(buffer)*2)
			copy(newbuff, buffer)
			buffer = newbuff
			idx += n
			continue
		}

		p, ok := ngen.NextPacket(buffer, parser)
		if !ok {
			// increment idx by how much we wrote.
			idx += n
			// fmt.Printf("No Packet... Buffer: %v\n", buffer[:idx])
			continue
		}

		// Successful packet read
		c.Incoming <- &p
		// copy back in case we have multiple packets in the buffer
		l := p.Len()
		if l != idx {
			copy(buffer, buffer[l:])
		}
		idx = 0
	}
}

func Sender(c *Client) {
	for m := range c.Outgoing {
		if m == nil {
			return // Empty message means die
		}
		msg := m.Pack()
		n, err := c.Conn.Write(msg)
		if err != nil {
			// print("Failure writing to socket: ", err.Error())
			c.Outgoing <- nil
			return
			// ded?
		} else if n == 0 {
			// print("Wrote 0 bytes!?\n")
			c.Outgoing <- nil
			return
			// ded?
		}
	}
	close(c.Incoming)
	close(c.Outgoing)
}
