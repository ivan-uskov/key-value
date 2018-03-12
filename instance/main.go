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
	"os"
	"os/signal"
	"syscall"
	"regexp"
	"key-value/instance/storages"
	"key-value/instance/replication"
)

func createSetter(s storages.Storage) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		s.Set(r.Option1, r.Option2)
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

var addr = flag.String("addr", ":8080", "http service address")

const dataPath = "storage.data"
const persistenceDelay = 2 * time.Second

func onShutDown(h func()) {
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)

	go func() {
		<-gracefulStop
		h()
		os.Exit(0)
	}()
}

func createRouter(storage storages.Storage) routers.Router {
	r := routers.NewRouter()
	r.AddRoute(routers.GET, createGetter(storage))
	r.AddRoute(routers.SET, createSetter(storage))
	r.AddRoute(routers.LIST, createLister(storage))
	r.AddRoute(routers.REMOVE, createRemover(storage))
	r.AddRoute(routers.PING, func(r routers.Request) (string, error) {
		return ``, nil
	})

	return r
}

func initializePersistence(storage storages.Storage) {
	portStr := regexp.MustCompile("[0-9]+$").FindString(*addr)
	p := NewPersister(dataPath+"."+portStr, storage.List)
	p.Load(storage.Set)
	p.RunSaveLoop(persistenceDelay)
	onShutDown(p.Persists)
}

func initializeReplication(s storages.Storage, router routers.Router, selfAddress string) {
	c := replication.NewClient(selfAddress)
	router.AddRoute(`NODES`, c.HandleNewNodesRequest)
	s.AddRemoveHandler(c.HandleRemoved)
	s.AddSetHandler(c.HandleUpdated)
	replication.NewServer(s, c).Bind()
}

func main() {
	flag.Parse()

	storage := storages.New()
	storage.AddSetHandler(func(key string, val string, ver int64) {
		log.Printf("Set %s : %s : %d", key, val, ver)
	})
	storage.AddRemoveHandler(func(key string, ver int64) {
		log.Printf("Remove %s : %d", key, ver)
	})

	initializePersistence(storage)

	router := createRouter(storage)
	initializeReplication(storage, router, *addr)

	server := ws.NewServer()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(w, r, router.CreateWebSocketHandler())
	})

	log.Fatal(http.ListenAndServe(*addr, nil))
}
