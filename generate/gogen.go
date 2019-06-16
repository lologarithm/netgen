package generate

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

var goTime = time.Now()

func HeaderComment() string {
	return fmt.Sprintf("// Code generated by netgen tool on %s. DO NOT EDIT", goTime.Format("Jan 2 2006 15:04 MST"))
}

func goName(m Message) string {
	if len(m.Package) > 0 {
		return m.Package + "." + m.Name
	}
	return m.Name
}

func goFieldName(m MessageField) string {
	if len(m.RemotePackage) > 0 {
		return m.RemotePackage + "." + m.Type
	}
	return m.Type
}

// GoLibHeader will return all the bits needed to make the generated serializers/deserializers work
// Specifically that is package name, imports, an enum of all message types, and a generic parse message function.
func GoLibHeader(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("%s\npackage %s\n\nimport (\n\t\"github.com/lologarithm/netgen/lib/ngen\"", HeaderComment(), pkgname))
	gobuf.WriteString("\n)\n\n\n")

	fldbuf := &bytes.Buffer{}
	for _, msg := range messages {
		if msg.Versioned {
			fldbuf.WriteString(fmt.Sprintf("%d: []byte{", MessageID(msg)))
			for _, f := range msg.Fields {
				fldbuf.WriteString(strconv.Itoa(f.Order))
				fldbuf.WriteString(",")
			}
			fldbuf.WriteString("},")
		}
	}

	// TODO: Add the Read/Write/Length functions attached to the settings
	gobuf.WriteString(fmt.Sprintf(`var Context = &ngen.Context {
		FieldVersions: map[ngen.MessageType][]byte{
			%s
		},
	}
`, fldbuf.String()))

	// TODO: Move this to ngservice package
	// 	gobuf.WriteString(`
	// func ManageClient(c *client.Client) {
	// 	settingsSync := make(chan *ngen.Context)
	// 	go client.Sender(c, Settings, settingsSync)
	// 	go client.Reader(c, ParseNetMessage, settingsSync)
	// }
	// `)

	// 1. List type values!
	gobuf.WriteString("const (\n")
	for _, t := range messages {
		gobuf.WriteString("\t")
		gobuf.WriteString(t.Name)
		gobuf.WriteString(fmt.Sprintf("MsgType = %d\n", MessageID(t)))
	}
	gobuf.WriteString(")\n\n")

	// type Reader func(*Context, MessageType, *Buffer) Message
	// 1.a. Parent parser function
	readFunc := `// Read accepts input of raw bytes and a type. Parses and returns a message.
func Read(ctx *ngen.Context, msgType ngen.MessageType, content *ngen.Buffer) ngen.Message {
	switch msgType {
		case ngen.MessageTypeContext:
			return ngen.DeserializeContext(Context, content)
%s
		default:
			return nil
	}
}
`
	caseTemplate := `	case %sMsgType:
			msg := Deserialize%s(ctx, content)
			return &msg
`
	caseBuffer := bytes.Buffer{}
	for _, t := range messages {
		caseBuffer.WriteString(fmt.Sprintf(caseTemplate, t.Name, t.Name))
	}
	gobuf.WriteString(fmt.Sprintf(readFunc, caseBuffer.String()))

	return gobuf.String()
}

// GoSerializers returns the generated code of Serialize, Len, and MessageType for the input msg
func GoSerializers(msg Message, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("\n\nfunc (m %s) Serialize(ctx *ngen.Context, buffer *ngen.Buffer) error {\n", msg.Name))
	if msg.Versioned {
		// If versioned we need to switch on the field indexes
		fldSwitch := &bytes.Buffer{}
		for _, f := range msg.Fields {
			fldSwitch.WriteString(fmt.Sprintf("\t\t\tcase %d:\n", f.Order))
			WriteGoSerializeField(f, 1, gobuf, messageMap, enumMap)
		}
		gobuf.WriteString(fmt.Sprintf(
			`	for _, fld := range settings.FieldVersions[%d] {
		switch fld {
%s		}
		}
`, MessageID(msg), fldSwitch.String()))
	} else {
		for _, f := range msg.Fields {
			WriteGoSerializeField(f, 1, gobuf, messageMap, enumMap)
		}
	}

	gobuf.WriteString("\n\treturn buffer.Err\n}\n")
	// TODO: Write the router Len function.
	gobuf.WriteString(fmt.Sprintf("\nfunc (m %s) Length(ctx *ngen.Context) int {\n\tmylen := 0\n", msg.Name))
	for _, f := range msg.Fields {
		WriteGoLen(f, 1, gobuf, messageMap, enumMap)
	}
	gobuf.WriteString("\treturn mylen\n}\n\n")

	gobuf.WriteString("func (m ")
	gobuf.WriteString(msg.Name)
	gobuf.WriteString(") MsgType() ngen.MessageType {\n\treturn ")
	gobuf.WriteString(msg.Name)
	gobuf.WriteString("MsgType\n}\n\n")
	return gobuf.String()
}

// GoDeserializers returns the generated code of Deserialize
func GoDeserializers(msg Message, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("\nfunc Deserialize%s(ctx *ngen.Context, buffer *ngen.Buffer) (m %s) {\n", msg.Name, msg.Name))
	if msg.Versioned {
		// If versioned we need to switch on the field indexes
		fldSwitch := &bytes.Buffer{}
		for _, f := range msg.Fields {
			fldSwitch.WriteString(fmt.Sprintf("\t\t\tcase %d:\n", f.Order))
			WriteGoDeserialField(f, true, 4, fldSwitch, messageMap, enumMap)
		}
		gobuf.WriteString(fmt.Sprintf(
			`	for _, fld := range settings.FieldVersions[%d] {
		switch fld {
%s		}
		}
`, MessageID(msg), fldSwitch.String()))
	} else {
		for _, f := range msg.Fields {
			WriteGoDeserialField(f, true, 1, gobuf, messageMap, enumMap)
		}
	}
	gobuf.WriteString("\treturn m\n}\n")
	return gobuf.String()
}

func WriteGoLen(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	n := ""
	if scopeDepth == 1 {
		n = "m."
	}
	n += f.Name

	writeTabScope(buf, scopeDepth)
	if f.Array && f.Type != ByteType { // array handling for non-byte type
		buf.WriteString("mylen += 4\n\t")
		fn := "v" + strconv.Itoa(scopeDepth+1)
		buf.WriteString(fmt.Sprintf("for _, %s := range %s {\n", fn, n))
		WriteGoLen(MessageField{Name: fn, Type: f.Type, Order: f.Order, Pointer: f.Pointer}, scopeDepth+1, buf, messages, enums)
		writeTabScope(buf, scopeDepth)
		buf.WriteString("}\n")
		return
	}
	switch f.Type {
	case ByteType, BoolType:
		if f.Array {
			buf.WriteString(fmt.Sprintf("mylen += 4 + len(%s)", n))
		} else {
			buf.WriteString("mylen += 1")
		}
	case Uint16Type, Int16Type:
		buf.WriteString("mylen += 2")
	case Uint32Type, Int32Type, RuneType, IntType:
		buf.WriteString("mylen += 4")
	case Uint64Type, Int64Type, Float64Type:
		buf.WriteString("mylen += 8")
	case StringType:
		buf.WriteString(fmt.Sprintf("mylen += 4 + len(%s)", n))
	default:
		if _, ok := messages[goFieldName(f)]; ok || f.Interface {
			if f.Pointer || f.Interface {
				buf.WriteString("\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("mylen++ // nil check \n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString(fmt.Sprintf("if %s != nil {\n", n))
				writeTabScope(buf, scopeDepth)
				if f.Interface {
					buf.WriteString("mylen+=4 // interface type value\n")
					writeTabScope(buf, scopeDepth)
				}
			}
			buf.WriteString(fmt.Sprintf("mylen += %s.Length(ctx)", n))
			if f.Pointer || f.Interface {
				buf.WriteString("\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("}")
			}
		} else if _, ok := enums[goFieldName(f)]; ok {
			buf.WriteString("mylen += 4 // enums are always int32... for now")
		} else {
			fmt.Printf("Can't write len for an unknown type... %#v\nMessage Type Map:\n%#v\n", f, messages)
		}
	}
	buf.WriteString("\n")
}

func writeArrayLen(f MessageField, scopeDepth int, buf *bytes.Buffer) {
	name := f.Name
	if scopeDepth == 1 {
		name = "m." + name
	}
	buf.WriteString(fmt.Sprintf("buffer.WriteUint32(uint32(len(%s)))\n", name))
	writeTabScope(buf, scopeDepth)
}

func WriteGoSerializeField(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	tabString := strings.Repeat("\t", scopeDepth)
	n := f.Name
	if scopeDepth == 1 {
		n = "m." + n
	}

	buf.WriteString(tabString)
	if f.Array && f.Type != ByteType { // Specially handle byte/bool array type.
		// Array!
		writeArrayLen(f, scopeDepth, buf)
		fn := "v" + strconv.Itoa(scopeDepth+1)
		buf.WriteString(fmt.Sprintf("for _, %s := range %s {\n", fn, n))
		WriteGoSerializeField(MessageField{Name: fn, Type: f.Type, Order: f.Order, Pointer: f.Pointer}, scopeDepth+1, buf, messages, enums)
		writeTabScope(buf, scopeDepth)
		buf.WriteString("}\n")
		return
	}

	switch f.Type {
	case ByteType:
		if f.Array {
			// Faster handler for byte arrays
			buf.WriteString(fmt.Sprintf("buffer.WriteByteSlice(%s)\n", n))
		} else {
			// Single byte
			buf.WriteString(fmt.Sprintf("buffer.WriteByte(%s)\n", n))
		}
	case BoolType:
		buf.WriteString(fmt.Sprintf("buffer.WriteBool(%s)\n", n))
	case Int16Type, Uint16Type:
		buf.WriteString(fmt.Sprintf("buffer.WriteUint16(uint16(%s))\n", n))
	case Int32Type, Uint32Type, RuneType, IntType:
		buf.WriteString(fmt.Sprintf("buffer.WriteUint32(uint32(%s))\n", n))
	case Int64Type, Uint64Type:
		buf.WriteString(fmt.Sprintf("buffer.WriteUint64(uint64(%s))\n", n))
	case Float64Type:
		buf.WriteString(fmt.Sprintf("buffer.WriteFloat64(%s)\n", n))
	case StringType:
		buf.WriteString(fmt.Sprintf("buffer.WriteString(%s)\n", n))
	default:
		if _, ok := messages[goFieldName(f)]; ok || f.Interface {
			varname := f.Name
			// Custom message deserial here.
			if scopeDepth == 1 {
				varname = "m." + varname
			}
			if f.Pointer || f.Interface {
				buf.WriteString(fmt.Sprintf(`if %s != nil {
					buffer.WriteBool(true)
`, varname))
				writeTabScope(buf, scopeDepth)
				if f.Interface {
					buf.WriteString(fmt.Sprintf("buffer.WriteUint32(uint32(%s.MsgType()))\n", varname))
					writeTabScope(buf, scopeDepth)
				}
				buf.WriteString(fmt.Sprintf("%s.Serialize(ctx, buffer)\n", varname))
				buf.WriteString(`} else {
	buffer.WriteBool(false)
}
`)
			} else {
				buf.WriteString(fmt.Sprintf("%s.Serialize(ctx, buffer)\n%s", varname, tabString))
			}
		} else if _, ok := enums[f.Type]; ok {
			buf.WriteString(fmt.Sprintf("buffer.WriteUint32(uint32(%s))\n", n))
			writeTabScope(buf, scopeDepth)
		} else {
			buf.WriteString(fmt.Sprintf("%s.Serialize(ctx, buffer)\n%s", n, tabString))
		}
	}
}

func WriteGoDeserialField(f MessageField, includeM bool, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	n := ""
	if includeM {
		n = "m."
	}
	n += f.Name

	writeTabScope(buf, scopeDepth)
	if f.Array && f.Type != ByteType { // handle byte array specially
		// Get len of array
		lname := "l" + strconv.Itoa(f.Order) + "_" + strconv.Itoa(scopeDepth)
		buf.WriteString(lname)
		buf.WriteString(" := buffer.ReadUint32()\n")
		writeTabScope(buf, scopeDepth)

		// 	// Create array variable
		buf.WriteString(n)
		buf.WriteString(" = make([]")
		if f.Pointer {
			buf.WriteString("*")
		}
		buf.WriteString(f.Type)
		buf.WriteString(", ")
		buf.WriteString(lname)
		buf.WriteString(")\n")
		//
		// Read each var into the array in loop
		writeTabScope(buf, scopeDepth)
		buf.WriteString("for i := uint32(0); i < ")
		buf.WriteString(lname)
		buf.WriteString("; i++ {\n")
		fn := ""
		if scopeDepth == 1 {
			fn += "m."
		}
		fn += f.Name + "[i]"
		WriteGoDeserialField(MessageField{Name: fn, Type: f.Type, Pointer: f.Pointer}, false, scopeDepth+1, buf, messages, enums)
		writeTabScope(buf, scopeDepth)
		buf.WriteString("}\n")
		return
	}
	switch f.Type {
	case ByteType:
		buf.WriteString(n)
		if f.Array {
			buf.WriteString(" = buffer.ReadByteSlice()\n")
		} else {
			buf.WriteString(" = buffer.ReadByte()\n")
		}
	case Int16Type, Int32Type, Int64Type, Uint16Type, Uint32Type, Uint64Type, Float64Type, RuneType, IntType:
		buf.WriteString(n)
		buf.WriteString(" = buffer.Read")
		buf.WriteString(strings.Title(f.Type))
		buf.WriteString("()\n")
	case StringType:
		buf.WriteString(n)
		buf.WriteString(" = buffer.ReadString()\n")
	default:
		if f.Interface {
			writeInterDeserial(buf, f, scopeDepth)
			return
		}

		if _, ok := messages[goFieldName(f)]; ok {
			// Custom message deserial here.
			if f.Pointer {
				buf.WriteString("if v := buffer.ReadByte(); v == 1 {\n")
				writeTabScope(buf, scopeDepth)
				subName := "sub" + f.Name
				if strings.Contains(f.Name, "[") {
					subName = "subi"
				}
				pkg := ""
				if f.RemotePackage != "" {
					pkg = f.RemotePackage + "."
				}
				buf.WriteString(fmt.Sprintf("\tvar %s = %sDeserialize%s(ctx, buffer)\n", subName, pkg, f.Type))
				writeTabScope(buf, scopeDepth)
				buf.WriteString(fmt.Sprintf("\t%s = &%s\n", n, subName))
				writeTabScope(buf, scopeDepth)
				buf.WriteString("}\n")
			} else {
				buf.WriteString(n)
				buf.WriteString(" = ")
				buf.WriteString(f.Type[0:])
				buf.WriteString("Deserialize(ctx, buffer)\n")
			}
		} else if _, ok := enums[goFieldName(f)]; ok {
			name := "tmp" + f.Name
			buf.WriteString(name)
			buf.WriteString(" := buffer.ReadUint32()\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString(n)
			buf.WriteString(" = ")
			buf.WriteString(f.Type)
			buf.WriteString("(")
			buf.WriteString(name)
			buf.WriteString(")\n")
		}
	}
}

// writeInterDeserial is just like write dynamic deserial except its for when the underlying type
// is an interface instead of a struct.
func writeInterDeserial(buf *bytes.Buffer, f MessageField, scopeDepth int) {
	buf.WriteString("if v := buffer.ReadByte(); v == 1 {\n")
	writeTabScope(buf, scopeDepth)
	mt := fmt.Sprintf("\tiType%d", f.Order)
	buf.WriteString(mt)
	buf.WriteString(" := buffer.ReadUint32()\n\t")

	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(fmt.Sprintf(" = Read(ctx, ngen.MessageType(%s), buffer).(%s)\n", mt, f.Type))
	writeTabScope(buf, scopeDepth)
	buf.WriteString("}\n")
}

func lowerFirst(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToLower(r)) + s[n:]
}

func writeTabScope(buf *bytes.Buffer, scopeDepth int) {
	for i := 0; i < scopeDepth; i++ {
		buf.WriteString("\t")
	}
}
