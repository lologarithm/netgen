package main

import (
	"fmt"
	"os"

	"github.com/lologarithm/netgen/example/newmodels"
	"github.com/lologarithm/netgen/lib/ngen"
	"github.com/lologarithm/netgen/lib/ngen/service/client/ngwebsocket"
)

func main() {
	url := "ws://127.0.0.1:4567/ws"
	origin := "http://127.0.0.1/"
	client, err := ngwebsocket.New(url, origin, func() {})
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err.Error())
		os.Exit(1)
	}

	newmodels.ManageClient(client)
	client.Outgoing <- ngen.NewPacket(newmodels.Message{
		Message: "HELLOOOO",
	})

	for packet := range client.Incoming {
		if packet == nil {
			break
		}
		switch packet.Header.MsgType {
		case newmodels.MessageMsgType:
			fmt.Printf("Got message: %s\n", packet.NetMsg.(*newmodels.Message).Message)
			os.Exit(0)
		}
	}
}
