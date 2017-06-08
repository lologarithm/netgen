package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/lologarithm/netgen/generate"
)

var genlist = flag.String("gen", "go", "list of languages to generate bindings for, separated by commas")
var input = flag.String("input", "", "Input defition file to generate from")

func main() {
	flag.Parse()

	messages := []generate.Message{}
	enums := []generate.Enum{}
	messageMap := map[string]generate.Message{}
	enumMap := map[string]generate.Enum{}

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
	// Parse types
	lines := strings.Split(string(data), "\n")
	pkgName := "netgen"
	message := generate.Message{}
	enum := generate.Enum{}
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) > 1 {
			if parts[0] == "enum" {
				enum.Name = parts[1]
				if parts[1] == generate.DynamicType {
					panic("dynamic is not valid type name")
				}
				continue
			} else if parts[0] == "struct" {
				message.Name = parts[1]
				if parts[1] == generate.DynamicType {
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
					message = generate.Message{}
				} else if enum.Name != "" {
					enums = append(enums, enum)
					enumMap[enum.Name] = enum
					enum = generate.Enum{}
				}
			} else if len(parts) > 1 && message.Name != "" {
				// probably a message field in format "<NAME> <TYPE>"
				field := generate.MessageField{
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
				ev := generate.EnumValue{
					Name:  parts[0],
					Value: val,
				}
				enum.Values = append(enum.Values, ev)
			}
		}
	}
	for _, l := range strings.Split(*genlist, ",") {
		switch l {
		case "go":
			generate.WriteGo(pkgName, messages, messageMap, enums, enumMap)
		case "dart":
			generate.WriteDartBindings(pkgName, messages, messageMap, enums, enumMap)
			generate.WriteJSConverter(pkgName, messages, messageMap, enums, enumMap)
		case "js":
			generate.WriteJSConverter(pkgName, messages, messageMap, enums, enumMap)
		case "cs":
			// generate.WriteCS(messages, messageMap)
		}
	}
}
