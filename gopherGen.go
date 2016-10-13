package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
)

func WriteJSConverter(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("// +build js\n\npackage %s\n\nimport (\n\t\"github.com/gopherjs/gopherjs/js\"\n)\n\n", pkgname))
	for _, msg := range messages {
		WriteJSConvertFunc(buf, msg, messageMap, enumMap, 1)
	}
	ioutil.WriteFile(path.Join(pkgname, pkgname+"js.go"), buf.Bytes(), 0666)
}

func WriteJSConvertFunc(buf *bytes.Buffer, msg Message, msgMap map[string]Message, enumMap map[string]Enum, scopeDepth int) {
	buf.WriteString(fmt.Sprintf("func %sFromJS(jso *js.Object) (m %s) {", msg.Name, msg.Name))
	for _, f := range msg.Fields {
		if f.Array {
			// We have an array
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(" = make([]")
			if f.Pointer {
				buf.WriteString("*")
			}
			buf.WriteString(f.Type)
			buf.WriteString(", jso.Length())")
			buf.WriteString("\n\tfor i := 0; i < jso.Length(); i++ {")

			buf.WriteString("\n\t}")
		} else if _, ok := msgMap[f.Type]; ok {
			// We have another message type
			if f.Pointer {
				buf.WriteString(fmt.Sprintf("\n\tvar sub%s = %sFromJS(jso.Get(\"%s\"))", f.Name, f.Type, f.Name))
				buf.WriteString(fmt.Sprintf("\n\tm.%s = &sub%s", f.Name, f.Name))
			} else {
				buf.WriteString(fmt.Sprintf("\n\tm.%s = %sFromJS(jso.Get(\"%s\"))", f.Name, f.Type, f.Name))
			}

		} else if _, ok := enumMap[f.Type]; ok {
			buf.WriteString(fmt.Sprintf("\n\tm.%s = %s(jso.Get(\"%s\").Int64())", f.Name, f.Type, f.Name))
		} else {
			buf.WriteString(fmt.Sprintf("\n\tm.%s = ", f.Name))
			switch f.Type {
			case "int8", "int16", "int32", "int64":
				buf.WriteString(fmt.Sprintf("%s(jso.Get(\"%s\").Int64())", f.Type, f.Name))
			case "uint8", "uint16", "uint32", "uint64":
				buf.WriteString(fmt.Sprintf("%s(jso.Get(\"%s\").Uint64())", f.Type, f.Name))
			case "string":
				buf.WriteString(fmt.Sprintf("jso.Get(\"%s\").String()", f.Name))
			default:
				panic("Unknown type: " + f.Type)
			}
		}
	}
	buf.WriteString("\n\treturn m\n}\n")
}
