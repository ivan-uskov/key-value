package main

import (
	"flag"
	"key-value/lib/ws"
	"net/http"
	"log"
	"key-value/lib/routers"
	"key-value/lib/processes"
	"fmt"
	"encoding/json"
	"errors"
)

var addr = flag.String("addr", ":8372", "http service address")

func createSetter(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		if reg.Exists(r.Option1) {
			return ``, errors.New(`Instance already started`)
		}

		args := []string{}
		args = append(args, fmt.Sprintf(`--addr=%s`, r.Option1))
		reg.Add(r.Option1, processes.Run(`../instance/instance.exe`, args))

		return ``, nil
	}
}

func createLister(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		items := reg.List()
		result := []string{}
		for address := range items {
			result = append(result, address)
		}

		res, err := json.Marshal(result)
		if err != nil {
			return fmt.Sprintf(`Cant marshall %s`, err.Error()), err
		}

		return string(res), nil
	}
}

func createRemover(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		worker, ok := reg.Get(r.Option1)
		if !ok {
			return ``, errors.New(`Item not exists`)
		}

		worker.Kill()
		if !reg.Remove(r.Option1) {
			return ``, errors.New(`Remove error`)
		}

		return ``, nil
	}
}

func main() {
	flag.Parse()

	register := NewRegister()
	server := ws.NewWebSocketServer()
	router := routers.New()
	router.SetSetter(createSetter(register))
	router.SetLister(createLister(register))
	router.SetRemover(createRemover(register))

	http.HandleFunc("/ctl", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWebSocket(w, r, router.CreateWebSocketHandler())
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
