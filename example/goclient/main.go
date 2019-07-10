package main

import (
	"fmt"
	"os"

	"github.com/lologarithm/netgen/example/newmodels"
	"github.com/lologarithm/netgen/lib/ngservice/client"
	"github.com/lologarithm/netgen/lib/ngservice/client/ngwebsocket"
)

func main() {
	url := "ws://127.0.0.1:4567/ws"
	origin := "http://127.0.0.1/"
	ngclient, err := ngwebsocket.New(url, origin, func() {})
	if err != nil {
		fmt.Printf("Failed to connect: %s\n", err.Error())
		os.Exit(1)
	}

	client.ManageClient(newmodels.Context, ngclient)
	ngclient.Outgoing <- newmodels.Message{
		Message: "HELLOOOO",
	}

	for msg := range ngclient.Incoming {
		if msg == nil {
			break
		}
		switch msg.MsgType() {
		case newmodels.MessageMsgType:
			fmt.Printf("Got message: %s\n", msg.(*newmodels.Message).Message)
			os.Exit(0)
		}
	}
}
