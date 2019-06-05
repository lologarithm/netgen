package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/build"
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
	"golang.org/x/tools/go/buildutil"
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

	// 1. search given package for all public types
	fset := token.NewFileSet()
	wd, _ := os.Getwd()
	pkgpath := filepath.Join(wd, *dir)

	bc := &build.Context{
		GOROOT:      build.Default.GOROOT,
		GOPATH:      build.Default.GOPATH,
		GOOS:        build.Default.GOOS,
		GOARCH:      build.Default.GOARCH,
		Compiler:    "gc",
		BuildTags:   []string{"purego"},
		ReleaseTags: build.Default.ReleaseTags,
		CgoEnabled:  true, // detect `import "C"` to throw proper error
	}
	pkg, err := bc.ImportDir(pkgpath, 0)
	if err != nil {
		panic(err)
	}

	type parsedPkg struct {
		name       string
		pkg        *build.Package
		files      []*ast.File
		messages   []generate.Message
		enums      []generate.Enum
		messageMap map[string]generate.Message
		enumMap    map[string]generate.Enum
	}
	pkgs := map[string]*parsedPkg{}

	var parseFile func(f *ast.File, pkg *parsedPkg)
	parseFile = func(f *ast.File, pkg *parsedPkg) {
		log.Printf("Parsing file: %s", f.Name.Name)
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
							msg := generate.Message{
								Name: ts.Name.Name,
							}
							log.Printf("Looking at struct: %#v", ts.Name)
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
								pkgSel, identType, isArray, isPointer := getidenttype(tfi.Type, false, false)
								log.Printf("Looking at field: %s, %s", pkgSel, identType)
								if identType == nil {
									// this means we don't handle this field type
									continue
								}
								typeval := identType.Name
								if pkgSel != nil {
									typeval = pkgSel.Name + "." + typeval
								}
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
								} else {
									msg.Versioned = true
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
							pkg.messages = append(pkg.messages, msg)
							if msg.Package != pkg.name {
								pkg.messageMap[msg.Package+"."+msg.Name] = msg
							} else {
								pkg.messageMap[msg.Name] = msg
							}
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
							pkg.enums = append(pkg.enums, enum)
							fmt.Printf("Added enum type %s\n", ts.Name.Name)
							pkg.enumMap[ts.Name.Name] = enum
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
	}

	var parsePkg func(pkg *build.Package)
	parsePkg = func(pkg *build.Package) {
		log.Printf("Parsing Package: %s", pkg.Name)
		for _, impt := range pkg.Imports {
			if _, ok := pkgs[impt]; ok {
				continue
			}
			importedPkg, err := bc.Import(impt, pkgpath, 0)
			if err != nil {
				log.Fatalf("Failed to import: %s", err)
			}
			parsePkg(importedPkg)

			pkgs[impt] = &parsedPkg{
				pkg:        importedPkg,
				messages:   []generate.Message{},
				enums:      []generate.Enum{},
				messageMap: map[string]generate.Message{},
				enumMap:    map[string]generate.Enum{},
			}

			for _, fname := range importedPkg.GoFiles {
				if !filepath.IsAbs(fname) { // name might be absolute if specified directly. E.g., `gopherjs build /abs/file.go`.
					fname = filepath.Join(importedPkg.Dir, fname)
				}
				r, err := buildutil.OpenFile(bc, fname)
				if err != nil {
					panic(err)
				}
				file, err := parser.ParseFile(fset, fname, r, parser.ParseComments)
				if err != nil {
					panic(err)
				}
				r.Close()
				pkgs[impt].files = append(pkgs[impt].files, file)
				parseFile(file, pkgs[impt])
			}
		}
	}

	parsePkg(pkg)

	// for _, fname := range pkg.GoFiles {
	// 	if !filepath.IsAbs(fname) { // name might be absolute if specified directly. E.g., `gopherjs build /abs/file.go`.
	// 		fname = filepath.Join(pkg.Dir, fname)
	// 	}
	// 	r, err := buildutil.OpenFile(bc, fname)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	file, err := parser.ParseFile(fset, fname, r, parser.ParseComments)
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// 	r.Close()
	// 	parseFile(file, pkgs[])
	// }

	// TODO: Validate that we are correctly using versioning
	// for _, msg := range messages {
	// 	if !msg.Versioned {
	// 		continue
	// 	}
	// 	sort.Slice(msg.Fields, func(i int, j int) bool {
	// 		return msg.Fields[i].Order < msg.Fields[j].Order
	// 	})
	// 	seen := map[int]bool{}
	// 	for _, f := range msg.Fields {
	// 		if ok := seen[f.Order]; ok {
	// 			log.Fatalf("Duplicate Field IDs on versioned struct: %s", msg.Name)
	// 		}
	// 		seen[f.Order] = true
	// 	}
	// }

	if outdir == nil || *outdir == "" {
		outdir = dir
	}

	msgs := []generate.Message{}
	msgMap := map[string]generate.Message{}
	enums := []generate.Enum{}
	enumMap := map[string]generate.Enum{}

	for _, l := range strings.Split(*genlist, ",") {
		switch l {
		case "go":
			outpkg := filepath.Base(*outdir)
			buf := &bytes.Buffer{}
			buf.WriteString(generate.GoLibHeader(outpkg, msgs, msgMap, enums, enumMap))

			// for _, msg := range messages {
			// 	buf.WriteString(generate.GoDeserializers(msg, pkg.messages, pkg.messageMap, pkg.enums, pkg.enumMap))
			// }

			ioutil.WriteFile(filepath.Join(filepath.Join(wd, *outdir), "ngenDeserial.go"), buf.Bytes(), 0644)

			buf.Reset()
			buf.WriteString(fmt.Sprintf("%spackage %s\n\nimport \"github.com/lologarithm/netgen/lib/ngen\"", generate.HeaderComment(), pkg.Name))
			// for _, msg := range messages {
			// 	buf.WriteString(generate.GoSerializers(msg, pkg.messages, pkg.messageMap, pkg.enums, pkg.enumMap))
			// }
			ioutil.WriteFile(filepath.Join(pkgpath, "ngenSerial.go"), buf.Bytes(), 0644)
		case "js":
			jsfile := generate.WriteJSConverter(pkg.Name, msgs, msgMap, enums, enumMap)
			rootpkg := filepath.Join(wd, *outdir)
			ioutil.WriteFile(path.Join(rootpkg, "ngenjs.go"), jsfile, 0666)

		case "cs":
			// generate.WriteCS(messages, messageMap)
		}
	}
}

// return is:
//  identifier type
//  isArray
//  isPointer
func getidenttype(e ast.Expr, isArray bool, isPointer bool) (*ast.Ident, *ast.Ident, bool, bool) {
	log.Printf("GetIdent: %#v", e)
	switch itf := e.(type) {
	case *ast.Ident:
		return nil, itf, isArray, isPointer
	case *ast.ArrayType:
		return getidenttype(itf.Elt, true, false)
	case *ast.StarExpr:
		return getidenttype(itf.X, isArray, true)
	case *ast.SelectorExpr:
		_, xv, _, _ := getidenttype(itf.X, false, false)
		return xv, itf.Sel, isArray, isPointer
	default:
		fmt.Printf("failed to handle a field type! %T, %#v\n", itf, itf)
	}
	return nil, nil, false, false
}
