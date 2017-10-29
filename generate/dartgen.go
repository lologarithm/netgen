package generate

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
)

func WriteDartBindings(pkgname string, messages []Message, messageMap map[string]Message, enums []Enum, enumMap map[string]Enum) {
	gobuf := &bytes.Buffer{}
	gobuf.WriteString(fmt.Sprintf("part of %s;\n", pkgname))

	for _, en := range enums {
		gobuf.WriteString("class ")
		gobuf.WriteString(en.Name)
		gobuf.WriteString(" {\n")
		for _, ev := range en.Values {
			gobuf.WriteString(fmt.Sprintf("static const int %s = %d;\n", ev.Name, ev.Value))
		}
		gobuf.WriteString("}\n\n")
	}

	for _, m := range messages {
		gobuf.WriteString("@JS()\n@anonymous\nclass ")
		gobuf.WriteString(m.Name)
		gobuf.WriteString(" {\n")
		gobuf.WriteString("external factory ")
		gobuf.WriteString(m.Name)
		gobuf.WriteString("({")
		for i, f := range m.Fields {
			gobuf.WriteString(fmt.Sprintf("%s %s", dartType(f, enumMap), f.Name))
			if i != len(m.Fields)-1 {
				gobuf.WriteString(",")
			}
		}
		gobuf.WriteString("});\n")

		for _, f := range m.Fields {
			dartT := dartType(f, enumMap)
			gobuf.WriteString(fmt.Sprintf("\nexternal %s get %s;\n", dartT, f.Name))
			gobuf.WriteString(fmt.Sprintf("external void set %s(%s val);\n", f.Name, dartT))
		}
		gobuf.WriteString("}\n\n")
	}

	ioutil.WriteFile(path.Join(pkgname, pkgname+".dart"), gobuf.Bytes(), 0666)
}

func dartType(f MessageField, enums map[string]Enum) string {
	if f.Array {
		return "List<" + dartType(MessageField{Type: f.Type}, enums) + ">"
	}
	switch f.Type {
	case ByteType, Int16Type, Int32Type, Int64Type, Uint16Type, Uint32Type, Uint64Type:
		return "int"
	case StringType:
		return "String"
	default:
		if _, ok := enums[f.Type]; ok {
			return "int"
		}
		return f.Type
	}
}
