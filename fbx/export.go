package fbx

import (
	"archive/zip"
	"bufio"
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

type exporter struct {
	tabs   int
	w      *bufio.Writer
	params int
}

func (e *exporter) fillTabs(diff int) {
	for i := 0; i < e.tabs+diff; i++ {
		e.w.WriteRune('\t')
	}
}

func (e *exporter) tabsInc() { e.tabs++ }
func (e *exporter) tabsDec() { e.tabs-- }

func (e *exporter) printf(format string, args ...interface{}) {
	e.w.WriteString(fmt.Sprintf(format, args...))
}
func (e *exporter) print(s string) {
	e.w.WriteString(s)
}

func (e *exporter) simpleValueString(nodeValue reflect.Value) string {
	switch nodeValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%d", nodeValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", nodeValue.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", nodeValue.Float())
	case reflect.String:
		return fmt.Sprintf("\"%s\"", nodeValue.String())
	case reflect.Bool:
		if nodeValue.Bool() {
			return "T"
		} else {
			return ""
		}
	default:
		return ""
	}
}

func (e *exporter) exportParameters(paramValue reflect.Value) {
	switch paramValue.Type().Kind() {
	case reflect.Interface, reflect.Ptr:
		e.exportParameters(paramValue.Elem())
		return
	case reflect.Slice, reflect.Array:
		for i := 0; i < paramValue.Len(); i++ {
			e.exportParameters(paramValue.Index(i))
		}
		return
	}

	if e.params != 0 {
		e.print(", ")
	}
	switch paramValue.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		e.print(e.simpleValueString(paramValue))
		e.params++
	default:
		panic(paramValue.Type().Kind())
	}
}

func (e *exporter) exportStruct(name string, nodeValue reflect.Value) {
	nodeType := nodeValue.Type()

	e.params = 0
	e.fillTabs(0)
	if e.tabs >= 0 {
		e.printf("%s: ", name)
	}

	// Print parameters
	for i := 0; i < nodeValue.NumField(); i++ {
		fieldStruct := nodeType.Field(i)

		tag := fieldStruct.Tag.Get("fbx")
		if tag == "" {
			continue
		}
		tags := strings.Split(tag, ",")
		if tags[0] != "p" {
			continue
		}

		e.exportParameters(nodeValue.Field(i))
	}

	openedBracket := false

	openBrackets := func() {
		if !openedBracket {
			if e.tabs >= 0 {
				if e.params != 0 {
					e.print(" ")
				}
				e.printf("{\n")
			}
			openedBracket = true
		}
	}
	if e.tabs == 0 {
		openBrackets()
	}

	// Print subnodes
	for i := 0; i < nodeValue.NumField(); i++ {
		fieldStruct := nodeType.Field(i)
		if fieldStruct.PkgPath != "" {
			continue
		}

		subNodeName := fieldStruct.Name
		isArrayNode := false
		inclusiveArrayNode := false
		tag := fieldStruct.Tag.Get("fbx")
		if tag != "" {
			tags := strings.Split(tag, ",")
			switch tags[0] {
			case "a":
				isArrayNode = true
			case "i":
				inclusiveArrayNode = true
			case "p":
				continue
			default:
				log.Panicf("Unknown fbx tag type: '%v'", tags[0])
			}
		}

		openBrackets()

		subValue := nodeValue.Field(i)
		e.tabsInc()
		if isArrayNode {
			e.exportArray(subNodeName, subValue)
		} else if inclusiveArrayNode {
			if !subValue.IsNil() {
				e.fillTabs(0)
				e.print(subNodeName)
				e.print(": ")
				for i := 0; i < subValue.Len(); i++ {
					if i != 0 {
						e.print(", ")
					}
					e.print(e.simpleValueString(subValue.Index(i)))
				}
				e.print("\n")
			}
		} else {
			e.exportNode(subNodeName, subValue)
		}
		e.tabsDec()
	}

	if e.tabs >= 0 {
		if openedBracket {
			e.fillTabs(0)
			e.print("}")
		}
		e.print("\n")
		if e.tabs == 0 {
			e.print("\n")
		}
	}
}

func (e *exporter) exportArray(name string, nodeValue reflect.Value) {
	if nodeValue.IsNil() {
		return
	}
	switch nodeValue.Type().Kind() {
	case reflect.Interface, reflect.Ptr:
		e.exportArray(name, nodeValue.Elem())
	case reflect.Array, reflect.Slice:
		l := nodeValue.Len()
		e.fillTabs(0)
		e.printf("%s: *%d {\n", name, l)
		e.fillTabs(1)
		e.print("a: ")
		for i := 0; i < l; i++ {
			if i != 0 {
				e.print(",")
			}
			e.print(e.simpleValueString(nodeValue.Index(i)))
		}
		e.print("\n")
		e.fillTabs(0)
		e.print("}\n")
	default:
		panic(nodeValue.Type().Kind())
	}
}

func (e *exporter) exportNode(name string, nodeValue reflect.Value) {
	switch nodeValue.Kind() {
	case reflect.Interface, reflect.Ptr:
		if nodeValue.IsNil() {
			return
		}
		e.exportNode(name, nodeValue.Elem())
		return
	}

	switch nodeValue.Kind() {
	case reflect.Struct:
		e.exportStruct(name, nodeValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		e.fillTabs(0)
		e.printf("%s: %s\n", name, e.simpleValueString(nodeValue))
	case reflect.Array, reflect.Slice:
		for i := 0; i < nodeValue.Len(); i++ {
			e.exportNode(name, nodeValue.Index(i))
		}
	}
}

func (f *FBX) Export(w io.Writer) error {
	return Export(f, w)
}

func (f *FBX) AddExportFile(name string, data []byte) {
	f.files[name] = data
}

func (f *FBX) ExportZip(w io.Writer, name string) error {
	zw := zip.NewWriter(w)

	fbxW, err := zw.Create(name)
	if err != nil {
		return errors.Wrapf(err, "Can't create zip fbx for %q", name)
	}
	if err := f.Export(fbxW); err != nil {
		return errors.Wrapf(err, "Fbx exporting failed")
	}

	for name, file := range f.files {
		fw, err := zw.Create(name)
		if err != nil {
			return errors.Wrapf(err, "Can't create zip for %q", name)
		}
		if _, err := fw.Write(file); err != nil {
			return errors.Wrapf(err, "Can't write zip for %q", name)
		}
	}

	return zw.Close()
}

func Export(f *FBX, originalWriter io.Writer) error {
	w := bufio.NewWriter(originalWriter)

	exporter := exporter{
		w:    w,
		tabs: -1,
	}

	exporter.exportNode("", reflect.ValueOf(f))

	return w.Flush()
}
