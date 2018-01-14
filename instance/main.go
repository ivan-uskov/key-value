package main

import (
	"flag"
	"log"
	"net/http"
	"key-value/lib/ws"
	"key-value/lib/routers"
	"encoding/json"
	"errors"
	"time"
)

func createSetter(s Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		s.Set(r.Option1, r.Option2)
		return ``, nil
	}
}

func createGetter(storage Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		v, ok := storage.Get(r.Option1)
		if !ok {
			return ``, errors.New(`Item not exists`)
		}

		return v, nil
	}
}

func createLister(reg Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		res, err := json.Marshal(reg.List())
		if err != nil {
			return ``, err
		}

		return string(res), nil
	}
}

func createRemover(reg Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		if !reg.Remove(r.Option1) {
			return ``, errors.New(`Not exists`)
		}

		return ``, nil
	}
}


var addr = flag.String("addr", ":8080", "http service address")

const dataPath = "storage.data"
const persistenceDelay = 2 * time.Second

func main() {
	flag.Parse()

	storage := New()
	storage.AddSetHandler(func(key string, val string, ver int64) {
		log.Printf("Set %s : %s : %d", key, val, ver)
	})
	storage.AddRemoveHandler(func(key string, ver int64) {
		log.Printf("Remove %s : %d", key, ver)
	})

	p := NewPersister(dataPath, storage.List)
	p.Load(storage.Set)
	p.Run(persistenceDelay)

	router := routers.NewRouter()
	router.AddRoute(routers.GET, createGetter(storage))
	router.AddRoute(routers.SET, createSetter(storage))
	router.AddRoute(routers.LIST, createLister(storage))
	router.AddRoute(routers.REMOVE, createRemover(storage))
	router.AddRoute(routers.PING, func(r routers.Request) (string, error) {
		return ``, nil
	})

	server := ws.NewServer()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(w, r, router.CreateWebSocketHandler())
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}