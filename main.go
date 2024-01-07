package main

import (
	"bytes"
	"embed"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/samber/lo"
)

const (
	FileSuffix = "_tabl.templ"
)

var (
	//go:embed templates
	embeddedFS embed.FS

	file = flag.String("file", "", "File to parse")
)

func main() {
	flag.Parse()
	data, err := ParseFile(*file, flag.Args()...)
	if err != nil {
		panic(err)
	}
	outputPath := filepath.Dir(*file)
	templateData := TemplateData{
		Name:    strings.TrimSuffix(filepath.Base(*file), filepath.Ext(*file)),
		Package: data.Package,
		Types:   data.Structs,
	}
	renderTemplates(templateData, outputPath)
}

// Struct contains the property of parsed struct
type Struct struct {
	Name       string
	Properties []Property
}

// Property contains the necessary data about a struct property for tabl component generation
type Property struct {
	Name      string
	FieldName string
}

// TemplateData contains all of the data necessary for the code generation templates to be generated
type TemplateData struct {
	Name    string
	Package string
	Types   []Struct
}

func loadTemplates(data *TemplateData) (*template.Template, error) {
	tmpl := template.New("")
	tmpl.Funcs(template.FuncMap{
		"templateToString": func(name string) (string, error) {
			return renderTemplate(tmpl, name, data)
		},
	})
	return tmpl.ParseFS(embeddedFS, "templates/*.tmpl")
}

// renderTemplate executes a template by its name with provided data.
func renderTemplate(t *template.Template, name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	err := t.ExecuteTemplate(&buf, name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func renderTemplates(data TemplateData, target string) error {
	tpl, err := loadTemplates(&data)
	if err != nil {
		panic(err)
	}
	os.MkdirAll(target, 0755)
	file, err := os.OpenFile(filepath.Join(target, data.Name+FileSuffix), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	tpl.ExecuteTemplate(file, "base.tmpl", data)
	return nil
}

type FileData struct {
	FileName string
	Package  string
	Structs  []Struct
}

// parseStructHeaders parses a Go source file and extracts headers from a struct.
func ParseFile(filename string, typeList ...string) (*FileData, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	fileData := FileData{
		FileName: filename,
		Package:  node.Name.Name,
	}
	ast.Inspect(node, func(n ast.Node) bool {
		// Check if it's a type declaration
		t, ok := n.(*ast.TypeSpec)
		if !ok {
			return true // continue searching
		}

		// Check if the type is the struct we are looking for, if no types are provider
		// then any type as accepted
		if len(typeList) > 0 && !lo.Contains(typeList, t.Name.Name) {
			return true
		}
		structType, ok := t.Type.(*ast.StructType)
		if !ok {
			return false // not a struct, stop searching
		}
		data := Struct{
			Name: t.Name.Name,
		}
		for _, field := range structType.Fields.List {
			property := Property{
				FieldName: field.Names[0].Name,
				Name:      field.Names[0].Name,
			}
			if field.Tag != nil {
				tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
				if tag.Get("compl") == "-" {
					continue
				}
				switch name := tag.Get("name"); name {
				case "-":
					property.Name = ""
				default:
					property.Name = name
				}
			}
			data.Properties = append(data.Properties, property)
		}
		fileData.Structs = append(fileData.Structs, data)
		return true // continue searching
	})

	return &fileData, nil
}
