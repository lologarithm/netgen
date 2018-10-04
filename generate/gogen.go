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
	return fmt.Sprintf("// Code generated by netgen tool on %s. DO NOT EDIT\n", goTime.Format("Jan 2 2006 15:04 MST"))
}

// GoLibHeader will return all the bits needed to make the generated serializers/deserializers work
// Specifically that is package name, imports, an enum of all message types, and a generic parse message function.
func GoLibHeader(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("%spackage %s\n\nimport (\n\t\"github.com/lologarithm/netgen/lib/ngen\"", HeaderComment(), pkgname))
	gobuf.WriteString("\n)\n\n\n")
	// 1. List type values!
	gobuf.WriteString("const (\n\tUnknownMsgType ngen.MessageType = iota\n\tAckMsgType\n")
	for _, t := range messages {
		gobuf.WriteString("\t")
		gobuf.WriteString(t.Name)
		gobuf.WriteString("MsgType\n")
	}
	gobuf.WriteString(")\n\n")

	// 1.a. Parent parser function
	gobuf.WriteString("// ParseNetMessage accepts input of raw bytes from a NetMessage. Parses and returns a Net message.\n")
	gobuf.WriteString("func ParseNetMessage(packet ngen.Packet, content *ngen.Buffer) ngen.Net {\n")
	gobuf.WriteString("\tswitch packet.Header.MsgType {\n")
	for _, t := range messages {
		gobuf.WriteString("\tcase ")
		gobuf.WriteString(t.Name)
		gobuf.WriteString("MsgType:\n")
		gobuf.WriteString("\t\tmsg := ")
		gobuf.WriteString(t.Name)
		gobuf.WriteString("Deserialize(content)\n\t\treturn &msg\n")
	}
	gobuf.WriteString("\tdefault:\n\t\treturn nil\n\t}\n}\n\n")
	return gobuf.String()
}

// GoType will write a generated go struct that represents input msg
func GoType(msg Message) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString("type ")
	gobuf.WriteString(msg.Name)
	gobuf.WriteString(" struct {")
	for _, f := range msg.Fields {
		gobuf.WriteString("\n\t")
		gobuf.WriteString(f.Name)
		gobuf.WriteString(" ")
		if f.Array {
			gobuf.WriteString("[]")
		}
		if f.Pointer {
			gobuf.WriteString("*")
		}
		gobuf.WriteString(f.Type)
	}
	gobuf.WriteString("\n}")
	return gobuf.String()
}

// GoSerializers returns the generated code of Serialize, Len, and MessageType for the input msg
func GoSerializers(msg Message, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("\n\nfunc (m %s) Serialize(buffer []byte) {\n", msg.Name))
	if len(msg.Fields) > 0 {
		gobuf.WriteString("\tidx := 0\n")
	}
	for _, f := range msg.Fields {
		WriteGoSerializeField(f, 1, gobuf, messageMap, enumMap)
	}
	gobuf.WriteString("}\n")
	gobuf.WriteString(fmt.Sprintf("\nfunc (m %s) Len() int {\n\tmylen := 0\n", msg.Name))
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
	gobuf.WriteString(fmt.Sprintf("\nfunc %sDeserialize(buffer *ngen.Buffer) (m %s) {\n", msg.Name, msg.Name))
	for _, f := range msg.Fields {
		WriteGoDeserialField(f, 1, gobuf, messageMap, enumMap)
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
		if _, ok := messages[f.Type]; ok || f.Interface {
			if f.Pointer || f.Interface {
				buf.WriteString("\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("mylen++ // nil check \n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString(fmt.Sprintf("if %s != nil {\n", n))
				writeTabScope(buf, scopeDepth)
				if f.Interface {
					buf.WriteString("mylen+=2 // interface type value\n")
					writeTabScope(buf, scopeDepth)
				}
			}
			buf.WriteString(fmt.Sprintf("mylen += %s.Len()", n))
			if f.Pointer || f.Interface {
				buf.WriteString("\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("}")
			}
		} else if _, ok := enums[f.Type]; ok {
			buf.WriteString("mylen += 4 // enums are always int32... for now")
		} else {
			fmt.Printf("Can't write len for an unknown type... %#v", f)
		}
	}
	buf.WriteString("\n")
}

func writeArrayLen(f MessageField, scopeDepth int, buf *bytes.Buffer) {
	buf.WriteString("ngen.PutUint32(buffer[idx:], uint32(len(")
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(")))\n")
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx += 4\n")
	writeTabScope(buf, scopeDepth)
}

func writeIdxInc(f MessageField, scopeDepth int, buf *bytes.Buffer) {
	n := ""
	if scopeDepth == 1 {
		n = "m."
	}
	n += f.Name

	buf.WriteString("\n")
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx += ")
	switch f.Type {
	case ByteType, BoolType:
		if f.Array {
			buf.WriteString("len(")
			buf.WriteString(n)
			buf.WriteString(")")
		} else {
			buf.WriteString("1")
		}
	case Int16Type, Uint16Type:
		buf.WriteString("2")
	case Int32Type, Uint32Type, RuneType, IntType:
		buf.WriteString("4")
	case Int64Type, Uint64Type, Float64Type:
		buf.WriteString("8")
	default:
		// Array probably
		if f.Type == StringType || f.Array {
			buf.WriteString("len(")
			buf.WriteString(n)
			buf.WriteString(")")
		} else {
			fmt.Printf("Failed to find type for '%s' to write idx incrementer for.", f.Type)
			panic("unknown type")
		}
	}
	buf.WriteString("\n")
}

func WriteGoSerializeField(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	tabString := strings.Repeat("\t", scopeDepth)
	n := ""
	if scopeDepth == 1 {
		n = "m."
	}
	n += f.Name

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
			writeArrayLen(f, scopeDepth, buf)
			buf.WriteString("copy(buffer[idx:], ")
			buf.WriteString(n)
			buf.WriteString(")")
			writeIdxInc(f, scopeDepth, buf)
		} else {
			// Single byte
			buf.WriteString("buffer[idx] = ")
			buf.WriteString(n)
			writeIdxInc(f, scopeDepth, buf)
		}
	case BoolType:
		typename := f.Name
		if scopeDepth == 1 {
			typename = "m." + typename
		}
		buf.WriteString(fmt.Sprintf("if %s { buffer[idx] = 1 }", typename))
		writeIdxInc(f, scopeDepth, buf)
	case Int16Type, Uint16Type:
		buf.WriteString("ngen.PutUint16(buffer[idx:], uint16(")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	case Int32Type, Uint32Type, RuneType, IntType:
		buf.WriteString("ngen.PutUint32(buffer[idx:], uint32(")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	case Int64Type, Uint64Type:
		buf.WriteString("ngen.PutUint64(buffer[idx:], uint64(")
		buf.WriteString(n)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	case Float64Type:
		buf.WriteString("ngen.PutFloat64(buffer[idx:], ")
		buf.WriteString(n)
		buf.WriteString(")")
		writeIdxInc(f, scopeDepth, buf)
	case StringType:
		writeArrayLen(f, scopeDepth, buf)
		buf.WriteString("copy(buffer[idx:], []byte(")
		buf.WriteString(n)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	default:
		if _, ok := messages[f.Type]; ok || f.Interface {
			varname := f.Name
			// Custom message deserial here.
			if scopeDepth == 1 {
				varname = "m." + varname
			}
			if f.Pointer || f.Interface {
				buf.WriteString(fmt.Sprintf("if %s != nil {\n", varname))
				buf.WriteString(fmt.Sprintf("%s%sbuffer[idx] = 1\n%s%sidx++\n%s", tabString, tabString, tabString, tabString, tabString))
				if f.Interface {
					buf.WriteString(fmt.Sprintf("%s\tngen.PutUint16(buffer[idx:], uint16(%s.MsgType()))\n", tabString, varname))
					writeTabScope(buf, scopeDepth)
					buf.WriteString("idx += 2\n")
					writeTabScope(buf, scopeDepth)
				}
				buf.WriteString(fmt.Sprintf("%s%s.Serialize(buffer[idx:])\n%sidx += %s.Len()\n%s", tabString, varname, tabString, varname, tabString))
				buf.WriteString("} else {\n")
				buf.WriteString(fmt.Sprintf("%sbuffer[idx] = 0\n%sidx++\n%s", tabString, tabString, tabString))
				buf.WriteString("}\n")
			} else {
				buf.WriteString(fmt.Sprintf("%s.Serialize(buffer[idx:])\n%sidx += %s.Len()\n%s", varname, tabString, varname, tabString))
			}
		} else if _, ok := enums[f.Type]; ok {
			buf.WriteString("ngen.PutUint32(buffer[idx:], uint32(")
			buf.WriteString(n)
			buf.WriteString("))")
			buf.WriteString("\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx += 4\n")
		} else {
			buf.WriteString(n)
			buf.WriteString(".Serialize(buffer[idx:])\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx += ")
			buf.WriteString(n)
			buf.WriteString(".Len()\n")
		}
	}
}

func writeArrayLenRead(lname string, scopeDepth int, buf *bytes.Buffer) {
	buf.WriteString(lname)
	buf.WriteString(" := int(ngen.Uint32(buffer[idx:]))\n")
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx += 4\n")
	writeTabScope(buf, scopeDepth)
}

func WriteGoDeserialField(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	n := ""
	if scopeDepth == 1 {
		n = "m."
	}
	n += f.Name

	writeTabScope(buf, scopeDepth)
	if f.Array && f.Type != ByteType { // handle byte array specially
		// Get len of array
		lname := "l" + strconv.Itoa(f.Order) + "_" + strconv.Itoa(scopeDepth)
		buf.WriteString(lname)
		buf.WriteString(", _ := buffer.ReadUint32()\n")
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
		WriteGoDeserialField(MessageField{Name: fn, Type: f.Type, Pointer: f.Pointer}, scopeDepth+1, buf, messages, enums)
		writeTabScope(buf, scopeDepth)
		buf.WriteString("}\n")
		return
	}
	switch f.Type {
	case ByteType:
		if f.Array {
			buf.WriteString(n)
			buf.WriteString(", _ = buffer.ReadByteSlice()\n")
		} else {
			buf.WriteString(n)
			buf.WriteString(", _ = buffer.ReadByte()\n")
		}
	case Int16Type, Int32Type, Int64Type, Uint16Type, Uint32Type, Uint64Type, Float64Type, RuneType, IntType:
		buf.WriteString(n)
		buf.WriteString(", _ = buffer.Read")
		buf.WriteString(strings.Title(f.Type))
		buf.WriteString("()\n")
	case StringType:
		buf.WriteString(n)
		buf.WriteString(", _ = buffer.ReadString()\n")
	default:
		if f.Interface {
			writeInterDeserial(buf, f, scopeDepth)
			return
		}

		if _, ok := messages[f.Type]; ok {
			// Custom message deserial here.
			if f.Pointer {
				buf.WriteString("if v, _ := buffer.ReadByte(); v == 1 {\n")
				writeTabScope(buf, scopeDepth)
				subName := "sub" + f.Name
				if strings.Contains(f.Name, "[") {
					subName = "subi"
				}
				buf.WriteString("\tvar ")
				buf.WriteString(subName)
				buf.WriteString(" = ")
				buf.WriteString(f.Type)
				buf.WriteString("Deserialize(buffer)\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("\t")
				buf.WriteString(n)
				buf.WriteString(" = &")
				buf.WriteString(subName)
				buf.WriteString("\n")
				writeTabScope(buf, scopeDepth)
				buf.WriteString("}\n")
			} else {
				buf.WriteString(n)
				buf.WriteString(" = ")
				buf.WriteString(f.Type[0:])
				buf.WriteString("Deserialize(buffer)\n")
			}
		} else if _, ok := enums[f.Type]; ok {
			name := "tmp" + f.Name
			buf.WriteString(name)
			buf.WriteString(", _ := buffer.ReadUint32()\n")
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
	buf.WriteString("if v, _ := buffer.ReadByte(); v == 1 {\n")
	writeTabScope(buf, scopeDepth)
	mt := fmt.Sprintf("\tiType%d", f.Order)
	buf.WriteString(mt)
	buf.WriteString(", _ := buffer.ReadUint16()\n\t")

	//ParseNetMessage
	buf.WriteString(fmt.Sprintf("\tp := ngen.Packet{Header: ngen.Header{MsgType: ngen.MessageType(%s)}}\n", mt))
	writeTabScope(buf, scopeDepth)
	buf.WriteString("\t")
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(fmt.Sprintf(" = ParseNetMessage(p, buffer).(%s)\n", f.Type))
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
