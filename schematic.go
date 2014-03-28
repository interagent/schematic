package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/heroku/schematic/schema"
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

	var s schema.Schema
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
