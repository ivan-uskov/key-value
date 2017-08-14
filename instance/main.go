package main

import (
	"flag"
	"log"
	"net/http"
	"key-value/lib/ws"
	"key-value/lib/storages"
	"key-value/lib/routers"
	"encoding/json"
	"errors"
)

var addr = flag.String("addr", ":8080", "http service address")

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "index.html")
}

func createSetter(storage storages.Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		storage.Set(r.Option1, r.Option2)
		return ``, nil
	}
}

func createGetter(storage storages.Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		v, ok := storage.Get(r.Option1)
		if !ok {
			return ``, errors.New(`Item not exists`)
		}

		return v, nil
	}
}


func createLister(reg storages.Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		res, err := json.Marshal(reg.List())
		if err != nil {
			return ``, err
		}

		return string(res), nil
	}
}

func createRemover(reg storages.Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		if !reg.Remove(r.Option1) {
			return ``, errors.New(`Not exists`)
		}

		return ``, nil
	}
}

func main() {
	flag.Parse()

	storage := storages.New()
	server := ws.NewWebSocketServer()

	router := routers.New()
	router.SetSetter(createSetter(storage))
	router.SetLister(createLister(storage))
	router.SetRemover(createRemover(storage))
	router.SetGetter(createGetter(storage))

	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.ServeWebSocket(w, r, router.CreateWebSocketHandler())
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}