package main

import (
	"flag"
	"log"

	"github.com/qencept/gonut/pkg/rendezvousserver"
)

var addr = flag.String("addr", ":3999", "listen addr")
var cert = flag.String("cert", "", "cert")
var key = flag.String("key", "", "key")

func main() {
	flag.Parse()

	server := rendezvousserver.New(*addr, *cert, *key)
	if err := server.Serve(); err != nil {
		log.Fatal(err)
	}
}
