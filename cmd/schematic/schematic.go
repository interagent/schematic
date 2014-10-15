// Schematic
//
// Generate Go client code for HTTP APIs
// described by JSON Hyper-Schemas: http://json-schema.org/latest/json-schema-hypermedia.html.
//
// Run it against your schema:
//
//     $ schematic platform-api.json > heroku/heroku.go
//
// This will generate a Go package named after your schema:
//
//     package heroku
//     ...
//
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/interagent/schematic"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	log.SetFlags(0)

	if len(os.Args) != 2 {
		log.Fatal("schematic: missing schema file")
	}

	var f *os.File
	var err error
	if os.Args[1] == "-" {
		f = os.Stdin
	} else {
		if f, err = os.Open(os.Args[1]); err != nil {
			log.Fatal(err)
		}
	}

	var s schematic.Schema
	d := json.NewDecoder(f)
	if err := d.Decode(&s); err != nil {
		log.Fatal(err)
	}

	code, err := s.Generate()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(code))
}
