package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/lologarithm/netgen/generate"
)

var genlist = flag.String("gen", "go", "list of languages to generate bindings for, separated by commas")
var dir = flag.String("dir", "", "Input directory to transpile")
var outdir = flag.String("out", "", "Output directory for deserializer package")
var version = flag.Bool("version", false, "Prints the version")

var verNum = "0.0.1"

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s", verNum)
		os.Exit(0)
	}

	messages := []generate.Message{}
	enums := []generate.Enum{}
	messageMap := map[string]generate.Message{}
	enumMap := map[string]generate.Enum{}

	// 1. search given package for all public types
	count := 0
	fset := token.NewFileSet()
	wd, _ := os.Getwd()
	pkgpath := filepath.Join(wd, *dir)
	files, err := ioutil.ReadDir(pkgpath)
	if err != nil {
		panic(err)
	}
	parsed := make([]*ast.File, len(files))
	for _, fi := range files {
		fn := fi.Name()
		if strings.HasSuffix(fn, ".go") && !strings.HasSuffix(fn, "_test.go") {
			f, err := parser.ParseFile(fset, filepath.Join(pkgpath, fn), nil, 0)
			if err == nil {
				parsed[count] = f
				count++
			} else {
				fmt.Fprintf(os.Stderr, "Exception: %v\n", err)
				os.Exit(1)
			}
		}
	}
	parsed = parsed[:count]
	if count == 0 {
		fmt.Printf("No go files found to parse.\n")
		os.Exit(1)
	}
	pkgname := parsed[0].Name.Name

	for _, f := range parsed {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				switch d.Tok {
				case token.TYPE:
					for _, s := range d.Specs {
						ts := s.(*ast.TypeSpec)
						if !ts.Name.IsExported() {
							continue
						}
						switch tsType := ts.Type.(type) {
						case *ast.StructType:
							msg := generate.Message{}
							msg.Name = ts.Name.Name
							var fields []generate.MessageField
							for _, tfi := range tsType.Fields.List {
								emb := false
								name := ""
								if len(tfi.Names) == 0 {
									emb = true
								} else if !tfi.Names[0].IsExported() {
									continue
								} else {
									name = tfi.Names[0].Name
								}

								customOrder := -1
								if tfi.Tag != nil && len(tfi.Tag.Value) > 0 {
									doSkip := false
									tag := reflect.StructTag(tfi.Tag.Value[1 : len(tfi.Tag.Value)-1])
									tagv, ok := tag.Lookup("ngen")
									if ok {
										tags := strings.Split(tagv, ",")
										for _, t := range tags {
											if t == "-" {
												doSkip = true
												break
											} else {
												// This is therefore a verioning tag
												customOrder, err = strconv.Atoi(t)
												if err != nil {
													log.Fatalf("Invalid ngen field tag (%s) found at: %s", t, fset.Position(tfi.Pos()).String())
												}
											}
										}
									}
									if doSkip {
										continue
									}
								}
								size := 0
								identType, isArray, isPointer := getidenttype(tfi.Type, false, false)
								if identType == nil {
									// this means we don't handle this field type
									continue
								}
								typeval := identType.Name
								if emb {
									name = typeval
								}
								isInterface := false
								if identType.Obj != nil {
									if dec, ok := identType.Obj.Decl.(*ast.TypeSpec); ok {
										if _, ok := dec.Type.(*ast.InterfaceType); ok {
											isInterface = true
										}
									}
								}
								if customOrder == -1 {
									customOrder = len(fields)
								}
								fields = append(fields, generate.MessageField{
									Name:      name,
									Type:      typeval,
									Array:     isArray,
									Pointer:   isPointer,
									Order:     customOrder,
									Size:      size,
									Embedded:  emb,
									Interface: isInterface,
								})
							}
							msg.Fields = fields
							messages = append(messages, msg)
							messageMap[msg.Name] = msg
							fmt.Printf("Added message type %s\n", msg.Name)
						case *ast.InterfaceType:
							// skip - no need to handle this i think
						case *ast.Ident:
							// this is a const type
							if tsType.Name == "string" {
								fmt.Printf("can't use const of type %s\n", tsType.Name)
								break
							}
							enum := generate.Enum{Name: ts.Name.Name}
							enums = append(enums, enum)
							fmt.Printf("Added enum type %s\n", ts.Name.Name)
							enumMap[ts.Name.Name] = enum
						default:
							fmt.Printf("Unknown type lib declaration: %s, %v\n", reflect.TypeOf(ts.Type), ts.Type)
						}
					}
				case token.CONST:
					// fmt.Printf("found const: %#v\n", d.Specs)
					// for _, s := range d.Specs {
					// 	switch ts := s.(type) {
					// 	case *ast.TypeSpec:
					// 	case *ast.ValueSpec:
					// 		for ni, val := range ts.Names {
					// 			fmt.Printf("  value spec name (%T): %#v\n", val, val)
					// 			enum, ok := enumMap[val.Name]
					// 			if !ok {
					// 				continue // no enum of this type
					// 			}
					// 			bl := ts.Values[ni].(*ast.BasicLit)
					// 			intval, _ := strconv.Atoi(bl.Value)
					// 			enum.Values = append(enum.Values, generate.EnumValue{
					// 				Name:  val.Name,
					// 				Value: intval,
					// 			})
					//
					// 		}
					// 		for _, val := range ts.Values {
					// 			fmt.Printf("  value spec value (%T): %#v\n", val, val)
					// 			// switch vt := val.(type) {
					// 			// case *ast.BasicLit:
					// 			// 	vt.Value
					// 			// }
					// 		}
					// 	}
					// }
				}
			case *ast.FuncDecl:
				// skip, we don't care about functions
			default:
				fmt.Printf("Other declaration in file? %T, %#v\n", d, d)
			}
		}
		// for _, imp := range f.Imports {
		// 	fmt.Printf("import %#v\n", imp.Path.Value)
		// 	// TODO: also create the imports serializers?
		// }
	}
	for _, msg := range messages {
		// 1. figure out which fields are versioned.
		for _, f := range msg.Fields {
			if f.Versioned {
				f.Order += len(msg.Fields)
			}
		}
	}

	if outdir == nil || *outdir == "" {
		outdir = dir
	}

	for _, l := range strings.Split(*genlist, ",") {
		switch l {
		case "go":
			// outpkg := filepath.Base(*outdir)
			buf := &bytes.Buffer{}
			buf.WriteString(generate.GoLibHeader(pkgname, messages, messageMap, enums, enumMap))

			for _, msg := range messages {
				buf.WriteString(generate.GoDeserializers(msg, messages, messageMap, enums, enumMap))
			}

			ioutil.WriteFile(filepath.Join(filepath.Join(wd, *outdir), "deserial.go"), buf.Bytes(), 0644)

			buf.Reset()
			buf.WriteString(fmt.Sprintf("%spackage %s\n\nimport \"github.com/lologarithm/netgen/lib/ngen\"", generate.HeaderComment(), pkgname))
			for _, msg := range messages {
				buf.WriteString(generate.GoSerializers(msg, messages, messageMap, enums, enumMap))
			}
			ioutil.WriteFile(filepath.Join(pkgpath, "gongen.go"), buf.Bytes(), 0644)
		case "js":
			jsfile := generate.WriteJSConverter(pkgname, messages, messageMap, enums, enumMap)
			rootpkg := filepath.Join(wd, *outdir)
			ioutil.WriteFile(path.Join(rootpkg, "jsSerial.go"), jsfile, 0666)

		case "cs":
			// generate.WriteCS(messages, messageMap)
		}
	}
}

// return is:
//  identifier type
//  isArray
//  isPointer
func getidenttype(e ast.Expr, isArray bool, isPointer bool) (*ast.Ident, bool, bool) {
	switch itf := e.(type) {
	case *ast.Ident:
		return itf, isArray, isPointer
	case *ast.ArrayType:
		return getidenttype(itf.Elt, true, false)
	case *ast.StarExpr:
		return getidenttype(itf.X, isArray, true)
	case *ast.SelectorExpr:
		return itf.Sel, isArray, isPointer
	default:
		fmt.Printf("failed to handle a field type! %T, %#v\n", itf, itf)
	}
	return nil, false, false
}
