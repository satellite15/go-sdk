//
// Author:: Salim Afiune Maya (<afiune@lacework.net>)
// Copyright:: Copyright 2020, Lacework Inc.
// License:: Apache License, Version 2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

const (
	blobGo      string = "blob.go"
	staticFiles string = "static"
)

var (
	conv = map[string]interface{}{"conv": fmtByteSlice}
	tmpl = template.Must(template.New("").Funcs(conv).Parse(`
// Code generated by: internal/databox/generator/main.go
//
// <<< DO NOT EDIT >>>
//

package databox

func init() {
	{{- range $name, $file := . }}
    	box.Add("{{ $name }}", []byte{ {{ conv $file }} })
	{{- end }}
}`),
	)
)

func main() {
	if _, err := os.Stat(staticFiles); os.IsNotExist(err) {
		log.Fatal("Static directory does not exists!")
	}

	// map of data to be boxed
	content := make(map[string][]byte)

	// walking through the static/ directory
	err := filepath.Walk(staticFiles, func(path string, info os.FileInfo, err error) error {
		relativePath := filepath.ToSlash(strings.TrimPrefix(path, staticFiles))

		if info.IsDir() {
			// skip: do not box directories
			log.Println(path, "is a directory, skipping...")
			return nil
		} else {
			// only box files
			log.Println(path, "is a file, boxing it...")

			b, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("Error reading %s: %s", path, err)
				return err
			}

			// store file inside the map of data
			content[relativePath] = b
		}

		return nil
	})
	if err != nil {
		log.Fatal("Error walking through embed directory:", err)
	}

	// write blob.go file
	f, err := os.Create(blobGo)
	if err != nil {
		log.Fatal("Error creating blob file:", err)
	}
	defer f.Close()

	buff := &bytes.Buffer{}
	if err = tmpl.Execute(buff, content); err != nil {
		log.Fatal("Error executing template", err)
	}

	data, err := format.Source(buff.Bytes())
	if err != nil {
		log.Fatal("Error formatting generated code", err)
	}

	if err = ioutil.WriteFile(blobGo, data, os.ModePerm); err != nil {
		log.Fatal("Error writing blob file", err)
	}
}

func fmtByteSlice(s []byte) string {
	builder := strings.Builder{}

	for _, v := range s {
		builder.WriteString(fmt.Sprintf("%d,", int(v)))
	}

	return builder.String()
}