package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

var genlist = flag.String("gen", "go", "list of languages to generate bindings for, separated by commas")
var input = flag.String("input", "", "Input defition file to generate from")

func main() {
	flag.Parse()

	// 1. Read defs.ng
	inputFile := "defs.ng"
	if *input != "" {
		inputFile = *input
	}
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		log.Printf("Failed to read definition file: %s", err)
		return
	}

	pkgName, messages, enums, messageMap, enumMap := parseNG(data)
	for _, l := range strings.Split(*genlist, ",") {
		switch l {
		case "go":
			WriteGo(pkgName, messages, messageMap, enums, enumMap)
		case "dart":
			WriteDartBindings(pkgName, messages, messageMap, enums, enumMap)
			WriteJSConverter(pkgName, messages, messageMap, enums, enumMap)
		case "js":
			WriteJSConverter(pkgName, messages, messageMap, enums, enumMap)
		case "cs":
			// WriteCS(messages, messageMap)
		}
	}
}

func parseNG(data []byte) (string, []Message, []Enum, map[string]Message, map[string]Enum) {
	messages := []Message{}
	enums := []Enum{}
	messageMap := map[string]Message{}
	enumMap := map[string]Enum{}

	// Parse types
	lines := strings.Split(string(data), "\n")
	pkgName := "netgen"
	message := Message{}
	enum := Enum{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 1 {
			if parts[0] == "enum" {
				enum.Name = parts[1]
				if parts[1] == DynamicType {
					panic("dynamic is not valid type name")
				}
				continue
			} else if parts[0] == "struct" {
				message.Name = parts[1]
				if parts[1] == DynamicType {
					panic("dynamic is not valid type name")
				}
				continue
			} else if parts[0] == "package" {
				pkgName = parts[1]
				continue
			}
		}
		if len(parts) > 0 {
			if parts[0] == "}" {
				if message.Name != "" {
					messages = append(messages, message)
					messageMap[message.Name] = message
					message = Message{}
				} else if enum.Name != "" {
					enums = append(enums, enum)
					enumMap[enum.Name] = enum
					enum = Enum{}
				}
			} else if len(parts) > 1 && message.Name != "" {
				// probably a message field in format "<NAME> <TYPE>"
				field := MessageField{
					Name:  parts[0],
					Type:  parts[1],
					Order: len(message.Fields),
				}
				if field.Type[0] == '[' {
					field.Array = true
					field.Type = field.Type[2:]
				}
				if field.Type[0] == '*' {
					field.Type = field.Type[1:]
					field.Pointer = true
				}
				switch field.Type {
				case "byte":
					field.Size = 1
				case "uint16", "int16":
					field.Size = 2
				case "uint32", "int32":
					field.Size = 4
				case "uint64", "int64", "float64":
					field.Size = 8
				case "string":
					field.Size = 4
				}
				message.SelfSize += field.Size
				message.Fields = append(message.Fields, field)
			} else if len(parts) > 2 && enum.Name != "" {
				// enum field in format "<NAME> = <TYPE>"
				val, err := strconv.Atoi(parts[2])
				if err != nil {
					fmt.Printf("Trying to parse enum %s, field %s, value is not valid integer.\n", enum.Name, parts[0])
					panic("invalid formatted definition file.")
				}
				ev := EnumValue{
					Name:  parts[0],
					Value: val,
				}
				enum.Values = append(enum.Values, ev)
			}
		}
	}

	return pkgName, messages, enums, messageMap, enumMap
}

// Message is a message that can be serialized across network.
type Message struct {
	Name     string
	Fields   []MessageField
	SelfSize int
}

type Enum struct {
	Name   string
	Values []EnumValue
}

type EnumValue struct {
	Name  string
	Value int
}

// MessageField is a single field of a message.
type MessageField struct {
	Name    string
	Type    string
	Array   bool
	Pointer bool
	Order   int
	Size    int
}

const (
	StringType  string = "string"
	ByteType    string = "byte"
	Int16Type   string = "int16"
	Uint16Type  string = "uint16"
	Int32Type   string = "int32"
	Uint32Type  string = "uint32"
	Int64Type   string = "int64"
	Uint64Type  string = "uint64"
	Float64Type string = "float64"
	DynamicType string = "dynamic"
)
