package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func WriteJSConverter(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) {
	buf := &bytes.Buffer{}
	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic("Unable to find current working directory")
	}
	gopath := path.Join(os.Getenv("GOPATH"), "src")
	rel, err := filepath.Rel(gopath, pwd)
	if err != nil {
		panic("Current directory not in gopath!")
	}
	rootpkg := path.Join(rel, pkgname)
	buf.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"github.com/gopherjs/gopherjs/js\"\n\t\"%s\"\n)\n\n", pkgname+"js", rootpkg))

	// 1.a. Parent parser function
	buf.WriteString("// ParseNetMessageJS accepts input of js.Object, parses it and returns a Net message.\n")
	buf.WriteString(fmt.Sprintf("func ParseNetMessageJS(jso *js.Object, t %s.MessageType) %s.Net {\n", pkgname, pkgname))
	buf.WriteString("\tswitch t {\n")
	for _, t := range messages {
		buf.WriteString(fmt.Sprintf("\tcase %s.", pkgname))
		buf.WriteString(t.Name)
		buf.WriteString("MsgType:\n")
		buf.WriteString("\t\tmsg := ")
		buf.WriteString(t.Name)
		buf.WriteString("FromJS(jso)\n\t\treturn &msg\n")
	}
	buf.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n\n")

	for _, msg := range messages {
		WriteJSConvertFunc(buf, pkgname, msg, messageMap, enumMap)
	}
	jsdir := path.Join(pkgname, pkgname+"js")
	os.MkdirAll(jsdir, 0777)
	ioutil.WriteFile(path.Join(jsdir, "jsSerial.go"), buf.Bytes(), 0666)
}

func WriteJSConvertFunc(buf *bytes.Buffer, pkgname string, msg Message, msgMap map[string]Message, enumMap map[string]Enum) {
	buf.WriteString(fmt.Sprintf("func %sFromJS(jso *js.Object) (m %s.%s) {", msg.Name, pkgname, msg.Name))
	for _, f := range msg.Fields {
		WriteJSConvertField(buf, pkgname, f, "", msgMap, enumMap, 1)
	}
	buf.WriteString("\n\treturn m\n}\n")
}

func WriteJSConvertField(buf *bytes.Buffer, pkgname string, f MessageField, subindex string, msgMap map[string]Message, enumMap map[string]Enum, scopeDepth int) {
	buf.WriteString("\n")
	buf.WriteString(strings.Repeat("\t", scopeDepth))

	getname := fmt.Sprintf("Get(\"%s\")", f.Name)
	setname := f.Name
	if scopeDepth <= 2 {
		setname = "m." + setname
	}
	if subindex != "" {
		getname = fmt.Sprintf("Get(\"%s\").Index(%s)", f.Name, subindex)
		setname += "[" + subindex + "]"
	}

	if f.Array {
		// We have an array
		buf.WriteString(setname)
		buf.WriteString(" = make([]")
		if f.Pointer {
			buf.WriteString("*")
		}
		switch f.Type {
		case ByteType, Int16Type, Int32Type, Int64Type:
		case Uint16Type, Uint32Type, Uint64Type:
		case StringType:
		default:
			buf.WriteString(pkgname)
			buf.WriteString(".")
		}
		buf.WriteString(f.Type)
		buf.WriteString(", ")
		sublen := fmt.Sprintf("jso.Get(\"%s\").Length()", f.Name)
		buf.WriteString(sublen)
		buf.WriteString(")\n")
		buf.WriteString(strings.Repeat("\t", scopeDepth))
		buf.WriteString(fmt.Sprintf("for i := 0; i < %s; i++ {", sublen))
		WriteJSConvertField(buf, pkgname, MessageField{Name: f.Name, Type: f.Type, Pointer: f.Pointer}, "i", msgMap, enumMap, scopeDepth+1)
		buf.WriteString("\n\t}")
	} else if _, ok := msgMap[f.Type]; ok {
		// We have another message type
		if f.Pointer {
			subName := "sub" + f.Name
			if subindex != "" {
				subName = "subi"
			}
			buf.WriteString(fmt.Sprintf("var %s = %sFromJS(jso.%s)\n", subName, f.Type, getname))
			buf.WriteString(strings.Repeat("\t", scopeDepth))
			buf.WriteString(fmt.Sprintf("%s = &%s", setname, subName))
		} else {
			buf.WriteString(fmt.Sprintf("%s = %sFromJS(jso.%s)", setname, f.Type, getname))
		}
	} else if _, ok := enumMap[f.Type]; ok {
		buf.WriteString(fmt.Sprintf("%s = %s(jso.%s.Int64())", setname, pkgname+"."+f.Type, getname))
	} else {
		buf.WriteString(fmt.Sprintf("%s = ", setname))
		switch f.Type {
		case ByteType, Int16Type, Int32Type, Int64Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Int64())", f.Type, getname))
		case Uint16Type, Uint32Type, Uint64Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Uint64())", f.Type, getname))
		case StringType:
			buf.WriteString(fmt.Sprintf("jso.%s.String()", getname))
		case DynamicType:
			getnametype := fmt.Sprintf("Get(\"%sType\")", f.Name)
			buf.WriteString(fmt.Sprintf("ParseNetMessageJS(jso.%s, %s.MessageType(jso.%s.Int()))", getname, pkgname, getnametype))
		default:
			panic("Unknown type: " + f.Type)
		}
	}
}
