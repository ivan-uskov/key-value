package main

import (
	"flag"
	"log"
	"key-value/lib/ws"
	"time"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	client := ws.NewClient(*addr, `ws`)

	for {
		select {
			case <-time.After(5 * time.Second):
				mc, err := client.Send(`{"Action":"LIST"}`)
				if err != nil {
					log.Println(err)
					return
				}
				log.Println(<-mc)
		}
	}
}