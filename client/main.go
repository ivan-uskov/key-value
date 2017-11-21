package main

import (
	"flag"
	"key-value/lib/ws"
	"fmt"
	"os"
	"log"
	"bufio"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func main() {
	flag.Parse()

	client := ws.NewClient(*addr, `ws`)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		res, err := client.SendSync(scanner.Text())
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}