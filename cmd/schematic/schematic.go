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
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/interagent/schematic"
)

var output = flag.String("o", "", "Output file")

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Fatal(r)
		}
	}()

	log.SetFlags(0)
	log.SetPrefix("schematic: ")

	flag.Parse()

	if flag.NArg() != 1 {
		log.Fatal("missing schema file")
	}

	var i io.Reader
	var err error
	if flag.Arg(0) == "-" {
		i = os.Stdin
	} else {
		if i, err = os.Open(flag.Arg(0)); err != nil {
			log.Fatal(err)
		}
	}

	var o io.Writer
	if *output == "" {
		o = os.Stdout
	} else {
		if o, err = os.Create(*output); err != nil {
			log.Fatal(err)
		}
	}

	var s schematic.Schema
	d := json.NewDecoder(i)
	if err := d.Decode(&s); err != nil {
		log.Fatal(err)
	}

	code, err := s.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", code)
		log.Fatal(err)
	}

	fmt.Fprintln(o, string(code))
}
