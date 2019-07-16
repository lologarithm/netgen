package generate

import (
	"bytes"
	"fmt"
	"strings"
)

func WriteJSConverter(pkg *ParsedPkg) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString(fmt.Sprintf("%s\npackage %s\n\nimport (\n\t\"github.com/gopherjs/gopherjs/js\"\n\t\"github.com/lologarithm/netgen/lib/ngen\"\n)\n\n", HeaderComment(), pkg.Name))

	// 1.a. Parent parser function
	buf.WriteString("// ParseNetMessageJS accepts input of js.Object, parses it and returns a Net message.\n")
	buf.WriteString("func ParseNetMessageJS(jso *js.Object, t ngen.MessageType) ngen.Message {\n")
	buf.WriteString("\tswitch t {\n")
	for _, t := range pkg.Messages {
		buf.WriteString(fmt.Sprintf("\tcase %sMsgType:\n", t.Name))
		buf.WriteString("\t\tmsg := ")
		buf.WriteString(t.Name)
		buf.WriteString("FromJS(jso)\n\t\treturn &msg\n")
	}
	buf.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n\n")

	for _, msg := range pkg.Messages {
		WriteJSConvertFunc(buf, msg, pkg)
	}
	return buf.Bytes()
}

func WriteJSConvertFunc(buf *bytes.Buffer, msg Message, pkg *ParsedPkg) {
	buf.WriteString(fmt.Sprintf("func %sFromJS(jso *js.Object) (m %s) {", msg.Name, msg.Name))
	for _, f := range msg.Fields {
		WriteJSConvertField(buf, f, "", pkg, 1)
	}
	buf.WriteString("\n\treturn m\n}\n")
}

func WriteJSConvertField(buf *bytes.Buffer, f MessageField, subindex string, pkg *ParsedPkg, scopeDepth int) {
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

	if f.Interface {
		getnametype := fmt.Sprintf("Get(\"%sType\")", f.Name)
		buf.WriteString(fmt.Sprintf("%s = ParseNetMessageJS(jso.%s, MessageType(jso.%s.Int())).(%s)", setname, getname, getnametype, f.Type))
	} else if f.Array {
		// We have an array
		buf.WriteString(setname)
		buf.WriteString(" = make([]")
		if f.Pointer {
			buf.WriteString("*")
		}
		buf.WriteString(f.Type)
		buf.WriteString(", ")
		sublen := fmt.Sprintf("jso.Get(\"%s\").Length()", f.Name)
		buf.WriteString(sublen)
		buf.WriteString(")\n")
		buf.WriteString(strings.Repeat("\t", scopeDepth))
		buf.WriteString(fmt.Sprintf("for i := 0; i < %s; i++ {", sublen))
		WriteJSConvertField(buf, MessageField{Name: f.Name, Type: f.Type, Pointer: f.Pointer}, "i", pkg, scopeDepth+1)
		buf.WriteString("\n\t}")
	} else if _, ok := pkg.MessageMap[f.Type]; ok {
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
	} else if _, ok := pkg.EnumMap[f.Type]; ok {
		buf.WriteString(fmt.Sprintf("%s = %s(jso.%s.Int64())", setname, f.Type, getname))
	} else {
		buf.WriteString(fmt.Sprintf("%s = ", setname))
		switch f.Type {
		case Float64Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Float())", f.Type, getname))
		case ByteType, IntType, RuneType, Int16Type, Int32Type, Uint16Type, Uint32Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Int())", f.Type, getname))
		case Int64Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Int64())", f.Type, getname))
		case Uint64Type:
			buf.WriteString(fmt.Sprintf("%s(jso.%s.Uint64())", f.Type, getname))
		case StringType:
			buf.WriteString(fmt.Sprintf("jso.%s.String()", getname))
		default:
			panic("Unknown type: " + f.Type)
		}
	}
}
