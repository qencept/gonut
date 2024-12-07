package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/qencept/gonut/pkg/peerlinker"
)

var id = flag.String("id", "", "link id")
var stun1 = flag.String("stun1", "52.24.174.49:3478", "stun server")
var stun2 = flag.String("stun2", "52.26.251.34:3478", "stun server")
var rendezvous = flag.String("rendezvous", "", "rendezvous server")

func main() {
	flag.Parse()
	if *id == "" {
		log.Fatal("must specify link id")
	}

	linker := peerlinker.New([2]string{*stun1, *stun2}, *rendezvous)
	conn, addr, err := linker.Link(*id)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	go func() {
		for i := 0; ; i++ {
			msg := fmt.Sprintf("k/a %d", i)
			_, err = conn.WriteToUDP([]byte(msg), addr)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("TX %v\n", msg)
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {
		for {
			buf := make([]byte, 1536)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("RX %v\n", string(buf[:n]))
		}
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
