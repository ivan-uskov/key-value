package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"key-value/lib/routers"
	"key-value/lib/ws"
	"log"
	"net/http"
)

var addr = flag.String("addr", ":8372", "http service address")

func createRunner(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		i, ok := reg.Get(r.Option1)
		if ok {
			if i.Ping() {
				return ``, nil
			} else {
				i.Kill()
				reg.Remove(r.Option1)
			}
		}

		i, err := NewInstance(r.Option1)
		if err != nil {
			return ``, err
		}
		reg.Add(r.Option1, i)

		return ``, nil
	}
}

func createLister(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		items := reg.List()
		result := make([]string, 0, len(items))
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
		i, ok := reg.Get(r.Option1)
		if !ok {
			return ``, errors.New(`item not exists`)
		}

		i.Kill()
		if !reg.Remove(r.Option1) {
			return ``, errors.New(`remove error`)
		}

		return ``, nil
	}
}

func main() {
	flag.Parse()

	register := NewRegister()
	server := ws.NewServer()
	router := routers.NewRouter()
	router.AddRoute(routers.RUN, createRunner(register))
	router.AddRoute(routers.LIST, createLister(register))
	router.AddRoute(routers.REMOVE, createRemover(register))

	http.HandleFunc("/ctl", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(w, r, router.CreateWebSocketHandler())
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
