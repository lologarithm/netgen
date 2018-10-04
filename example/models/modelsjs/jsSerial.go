package modelsjs

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/lologarithm/netgen/example/models"
	"github.com/lologarithm/netgen/lib/ngen"
)

// ParseNetMessageJS accepts input of js.Object, parses it and returns a Net message.
func ParseNetMessageJS(jso *js.Object, t ngen.MessageType) ngen.Net {
	switch t {
	case models.MessageMsgType:
		msg := MessageFromJS(jso)
		return &msg
	default:
		return nil
	}
}

func MessageFromJS(jso *js.Object) (m models.Message) {
	m.Message = jso.Get("Message").String()
	return m
}
