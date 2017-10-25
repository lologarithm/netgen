package generate

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// search messages to see if we have a float64 field.
// this will help us decide if we need to import math package.
func hasfloat64(messages []Message) bool {
	for _, m := range messages {
		for _, f := range m.Fields {
			if f.Type == Float64Type {
				return true
			}
		}
	}
	return false
}

// GoLibHeader will return all the bits needed to make the generated serializers/deserializers work
// Specifically that is package name, imports, an enum of all message types, and a generic parse message function.
func GoLibHeader(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) string {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("package %s\n\nimport (\n\t\"github.com/lologarithm/netgen/lib/ngen\"", pkgname))
	if hasfloat64(messages) {
		// only import math if we need to serial/deserial float64s
		gobuf.WriteString("\n\t\"math\"")
	}
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
	gobuf.WriteString("func ParseNetMessage(packet ngen.Packet, content []byte) ngen.Net {\n")
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
		if f.Type == DynamicType {
			gobuf.WriteString("ngen.Net")
			gobuf.WriteString("\n\t")
			gobuf.WriteString(lowerFirst(f.Name))
			gobuf.WriteString("Type ")
			gobuf.WriteString("ngen.MessageType")
		} else {
			gobuf.WriteString(f.Type)
		}
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
	gobuf.WriteString(fmt.Sprintf("\nfunc %sDeserialize(buffer []byte) (m %s) {\n", msg.Name, msg.Name))
	if len(msg.Fields) > 0 {
		gobuf.WriteString("\tidx := 0\n")
	}
	for _, f := range msg.Fields {
		WriteGoDeserialField(f, 1, gobuf, messageMap, enumMap)
	}
	gobuf.WriteString("\treturn m\n}\n")
	return gobuf.String()
}

func WriteGoLen(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	writeTabScope(buf, scopeDepth)
	if f.Array && f.Type != ByteType { // array handling for non-byte type
		buf.WriteString("mylen += 4\n\t")
		fn := "v" + strconv.Itoa(scopeDepth+1)
		buf.WriteString("for _, ")
		buf.WriteString(fn)
		buf.WriteString(" := range ")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(" {\n")
		buf.WriteString("\t_ = ")
		buf.WriteString(fn)
		buf.WriteString("\n")
		WriteGoLen(MessageField{Name: fn, Type: f.Type, Order: f.Order, Pointer: f.Pointer}, scopeDepth+1, buf, messages, enums)
		writeTabScope(buf, scopeDepth)
		buf.WriteString("}\n")
		return
	}
	switch f.Type {
	case ByteType, BoolType:
		if f.Array {
			buf.WriteString("mylen += 4 + len(")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(")")
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
		buf.WriteString("mylen += 4 + len(")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(")")
	case DynamicType:
		buf.WriteString("mylen += 2\n\t")
		buf.WriteString("mylen += ")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(".(ngen.Net).Len()")
	default:
		if _, ok := messages[f.Type]; ok {
			buf.WriteString("mylen += ")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(".Len()")
			if f.Pointer {
				writeTabScope(buf, scopeDepth)
				buf.WriteString("\n")
				buf.WriteString("mylen++")
			}
		} else if _, ok := enums[f.Type]; ok {
			buf.WriteString("mylen += 4")
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
	buf.WriteString("\n")
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx+=")
	switch f.Type {
	case ByteType, BoolType:
		if f.Array {
			buf.WriteString("len(")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
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
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
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
	buf.WriteString(tabString)
	if f.Array && f.Type != ByteType { // Specially handle byte/bool array type.
		// Array!
		writeArrayLen(f, scopeDepth, buf)
		fn := "v" + strconv.Itoa(scopeDepth+1)
		buf.WriteString("for _, ")
		buf.WriteString(fn)
		buf.WriteString(" := range ")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(" {\n")
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
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(")")
			writeIdxInc(f, scopeDepth, buf)
		} else {
			// Single byte
			buf.WriteString("buffer[idx] = ")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
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
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	case Float64Type:
		buf.WriteString("ngen.PutFloat64(buffer[idx:], ")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(")")
		writeIdxInc(f, scopeDepth, buf)
	case StringType:
		writeArrayLen(f, scopeDepth, buf)
		buf.WriteString("copy(buffer[idx:], []byte(")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString("))")
		writeIdxInc(f, scopeDepth, buf)
	case DynamicType:
		// Custom message deserial here.
		buf.WriteString("ngen.PutUint16(buffer[idx:], uint16(")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(".MsgType()))")
		buf.WriteString("\n")
		writeTabScope(buf, scopeDepth)
		buf.WriteString("idx+=2")
		buf.WriteString("\n")
		writeTabScope(buf, scopeDepth)

		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(".(ngen.Net).Serialize(buffer[idx:])\n")
		writeTabScope(buf, scopeDepth)
		buf.WriteString("idx+=")
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(".(ngen.Net).Len()\n")
	default:
		if _, ok := messages[f.Type]; ok {
			varname := f.Name
			// Custom message deserial here.
			if scopeDepth == 1 {
				varname = "m." + varname
			}
			if f.Pointer {
				buf.WriteString(fmt.Sprintf("if %s != nil {\n", varname))
				buf.WriteString(fmt.Sprintf("%s%sbuffer[idx] = 1\n%s%sidx++\n%s", tabString, tabString, tabString, tabString, tabString))
				buf.WriteString(fmt.Sprintf("%s%s.Serialize(buffer[idx:])\n%sidx+=%s.Len()\n%s", tabString, varname, tabString, varname, tabString))
				buf.WriteString("} else {\n")
				buf.WriteString(fmt.Sprintf("%sbuffer[idx] = 0\n%sidx++\n%s", tabString, tabString, tabString))
				buf.WriteString("}")
			} else {
				buf.WriteString(fmt.Sprintf("%s.Serialize(buffer[idx:])\n%sidx+=%s.Len()\n%s", varname, tabString, varname, tabString))
			}
		} else if _, ok := enums[f.Type]; ok {
			buf.WriteString("ngen.PutUint32(buffer[idx:], uint32(")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString("))")
			buf.WriteString("\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx+=4\n")
		} else {
			fmt.Printf("can't serialize %s!??!\n", f.Type)
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(".Serialize(buffer[idx:])\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx+=")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(".Len()\n")
		}
	}
}

func writeNumericDeserialFunc(f MessageField, scopeDepth int, buf *bytes.Buffer) {
	buf.WriteString("ngen.")
	switch f.Type {
	case Int16Type, Uint16Type:
		buf.WriteString("Uint16(")
	case Int32Type, Uint32Type, RuneType, IntType:
		buf.WriteString("Uint32(")
	case Int64Type, Uint64Type:
		buf.WriteString("Uint64(")
	case Float64Type:
		buf.WriteString("Float64(")
	}
	buf.WriteString("buffer[idx:]")
	buf.WriteString(")")
}

func writeArrayLenRead(lname string, scopeDepth int, buf *bytes.Buffer) {
	buf.WriteString(lname)
	buf.WriteString(" := int(ngen.Uint32(buffer[idx:]))\n")
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx += 4\n")
	writeTabScope(buf, scopeDepth)
}

func WriteGoDeserialField(f MessageField, scopeDepth int, buf *bytes.Buffer, messages map[string]Message, enums map[string]Enum) {
	writeTabScope(buf, scopeDepth)
	if f.Array && f.Type != ByteType { // handle byte array specially
		// Get len of array
		lname := "l" + strconv.Itoa(f.Order) + "_" + strconv.Itoa(scopeDepth)
		writeArrayLenRead(lname, scopeDepth, buf)

		// 	// Create array variable
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
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
		buf.WriteString("for i := 0; i < int(")
		buf.WriteString(lname)
		buf.WriteString("); i++ {\n")
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
			lname := "l" + strconv.Itoa(f.Order) + "_" + strconv.Itoa(scopeDepth)
			writeArrayLenRead(lname, scopeDepth, buf)

			dest := ""
			if scopeDepth == 1 {
				dest = "m."
			}
			dest += f.Name
			buf.WriteString(dest)
			buf.WriteString(" = make([]byte, ")
			buf.WriteString(lname)
			buf.WriteString(")\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString(fmt.Sprintf("copy(%s, buffer[idx:idx+%s])\n", dest, lname))
		} else {
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(" = buffer[idx]\n")
			writeIdxInc(f, scopeDepth, buf)
		}
	case Int16Type, Int32Type, Int64Type, Uint16Type, Uint32Type, Uint64Type, Float64Type, RuneType, IntType:
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(" = ")
		switch f.Type {
		case Int16Type, Int32Type, Int64Type, RuneType, IntType:
			buf.WriteString(f.Type)
			buf.WriteString("(")
		}
		writeNumericDeserialFunc(f, scopeDepth, buf)
		if f.Type[0] == 'i' || f.Type[0] == 'r' {
			buf.WriteString(")")
		}
		writeIdxInc(f, scopeDepth, buf)
	case StringType:
		// Get length of string first
		lname := "l" + strconv.Itoa(f.Order) + "_" + strconv.Itoa(scopeDepth)
		writeArrayLenRead(lname, scopeDepth, buf)
		if scopeDepth == 1 {
			buf.WriteString("m.")
		}
		buf.WriteString(f.Name)
		buf.WriteString(" = string(buffer[idx:idx+")
		buf.WriteString(lname)
		buf.WriteString("])")
		writeIdxInc(f, scopeDepth, buf)
	case DynamicType:
		writeDynDeserial(buf, f, scopeDepth)
	default:
		if f.Interface {
			writeInterDeserial(buf, f, scopeDepth)
			return
		}

		if _, ok := messages[f.Type]; ok {
			// Custom message deserial here.
			if f.Pointer {
				subName := "sub" + f.Name
				if strings.Contains(f.Name, "[") {
					subName = "subi"
				}
				buf.WriteString("var ")
				buf.WriteString(subName)
				buf.WriteString(" = ")
				buf.WriteString(f.Type)
				buf.WriteString("Deserialize(buffer[idx:])\n")
				writeTabScope(buf, scopeDepth)
				if scopeDepth == 1 {
					buf.WriteString("m.")
				}
				buf.WriteString(f.Name)
				buf.WriteString(" = &")
				buf.WriteString(subName)
				buf.WriteString("\n")
			} else {
				if scopeDepth == 1 {
					buf.WriteString("m.")
				}
				buf.WriteString(f.Name)
				buf.WriteString(" = ")
				buf.WriteString(f.Type[0:])
				buf.WriteString("Deserialize(buffer[idx:])\n")
			}
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx+=")
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(".Len()\n")
		} else if _, ok := enums[f.Type]; ok {
			if scopeDepth == 1 {
				buf.WriteString("m.")
			}
			buf.WriteString(f.Name)
			buf.WriteString(" = ")
			buf.WriteString(f.Type)
			buf.WriteString("(")
			buf.WriteString("ngen.")
			buf.WriteString("Uint32(")
			buf.WriteString("buffer[idx:]")
			buf.WriteString("))")
			buf.WriteString("\n")
			writeTabScope(buf, scopeDepth)
			buf.WriteString("idx+=4\n")
		}
	}
}

// writeInterDeserial is just like write dynamic deserial except its for when the underlying type
// is an interface instead of a struct.
func writeInterDeserial(buf *bytes.Buffer, f MessageField, scopeDepth int) {
	mt := fmt.Sprintf("iType%d", f.Order)
	buf.WriteString(mt)
	buf.WriteString(" := ngen.MessageType(ngen.Uint16(buffer[idx:]))\n\t")
	buf.WriteString("idx+=2\n\t")

	//ParseNetMessage
	buf.WriteString(fmt.Sprintf("p := ngen.Packet{Header: ngen.Header{MsgType: %s}}\n", mt))
	writeTabScope(buf, scopeDepth)
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(fmt.Sprintf(" = ParseNetMessage(p, buffer[idx:]).(%s)\n", f.Type))
	writeTabScope(buf, scopeDepth)

	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx+=")
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(".(ngen.Net).Len()\n")
}

func writeDynDeserial(buf *bytes.Buffer, f MessageField, scopeDepth int) {
	pre := ""
	if scopeDepth == 1 {
		pre = "m."
	}
	mt := fmt.Sprintf("%s%sType", pre, lowerFirst(f.Name))

	buf.WriteString(mt)
	buf.WriteString(" = ngen.MessageType(ngen.Uint16(buffer[idx:]))\n\t")
	buf.WriteString("idx+=2\n\t")

	//ParseNetMessage
	buf.WriteString(fmt.Sprintf("p := ngen.Packet{Header: ngen.Header{MsgType: %s}}\n", mt))
	writeTabScope(buf, scopeDepth)
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(" = ParseNetMessage(p, buffer[idx:])\n")
	writeTabScope(buf, scopeDepth)
	writeTabScope(buf, scopeDepth)
	buf.WriteString("idx+=")
	if scopeDepth == 1 {
		buf.WriteString("m.")
	}
	buf.WriteString(f.Name)
	buf.WriteString(".(ngen.Net).Len()\n")
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
