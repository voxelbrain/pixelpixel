package main

import (
	"bytes"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

func format() {
	filepath.Walk(".", func(path string, info os.FileInfo, _ error) error {
		if strings.HasPrefix(path, ".") {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || info.IsDir() {
			return nil
		}
		code, err := ioutil.ReadFile(path)
		if err != nil {
			log.Fatalf("Could not read file %s: %s", path, err)
		}

		fmtCode, err := gofmt(code)
		if err != nil {
			log.Fatalf("File invalid %s: %s", path, err)
		}
		if !reflect.DeepEqual(code, fmtCode) {
			log.Printf("%s has been gofmt’d", path)
		}

		err = ioutil.WriteFile(path, fmtCode, os.FileMode(0644))
		if err != nil {
			log.Fatalf("Could not write file %s: %s", path, err)
		}
		return nil
	})
}

func gofmt(code []byte) ([]byte, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "someFile.go", code, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	err = (&printer.Config{Mode: printer.TabIndent | printer.UseSpaces, Tabwidth: 8}).Fprint(buf, fset, f)
	return buf.Bytes(), err
}
