package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
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

const persistenceDelay = 2 * time.Second
const tmpDir = `tmp`

func getDataPath(port string) string {
	return tmpDir+"/storage."+port+".data"
}

func getLogPath(port string) string {
	return tmpDir+"/storage."+port+".log"
}

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
	p := NewPersister(getDataPath(getPort()), storage.List)
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
	os.Mkdir(tmpDir, os.ModePerm)
	initLogger()

	storage := storages.New()
	initializePersistence(storage)

	router := createRouter(storage)
	initializeReplication(storage, router, *addr)

	server := ws.NewServer()
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(w, r, router.CreateWebSocketHandler())
	})

	log.WithFields(log.Fields{`addr`: *addr}).Info(`Listen`)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func initLogger() {
	log.SetFormatter(&log.JSONFormatter{})
	file, err := os.OpenFile(getLogPath(getPort()), os.O_CREATE | os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
}

func getPort() string {
	return regexp.MustCompile("[0-9]+$").FindString(*addr)
}
