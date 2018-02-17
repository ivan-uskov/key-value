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
	"os"
	"os/signal"
	"syscall"
	"context"
)

func createRunner(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		reg.Lock()
		defer reg.Unlock()
		i, ok := reg.Get(r.Option1)
		if ok {
			if !i.Ping() {
				err := i.Restart()
				if err != nil {
					i.Kill()
					reg.Remove(r.Option1)
					return ``, err
				}

				fmt.Printf("%s restarted by request \n", r.Option1)
			}
		} else {
			i, err := NewInstance(r.Option1)
			if err != nil {
				return ``, err
			}

			reg.Add(r.Option1, i)
			fmt.Printf("Run new instance on %s\n", r.Option1)
		}

		return ``, nil
	}
}

func createLister(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		reg.RLock()
		defer reg.RUnlock()

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
		reg.Lock()
		defer reg.Unlock()
		i, ok := reg.Get(r.Option1)
		if !ok {
			return ``, errors.New(`item not exists`)
		}

		i.Kill()
		reg.Remove(r.Option1)

		return ``, nil
	}
}

var addr = flag.String("addr", ":8372", "http service address")

func getKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Kill, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChan
}

func waitForKillSignal(killSignalChan chan os.Signal) {
	killSignal := <-killSignalChan
	switch killSignal {
	case os.Interrupt:
		log.Println("got SIGINT...")
	case syscall.SIGTERM:
		log.Println("got SIGTERM...")
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

	killChan := getKillSignalChan()
	srv := &http.Server{Addr:    *addr}
	go func() {
		log.Fatal(srv.ListenAndServe())
	}()

	waitForKillSignal(killChan)

	srv.Shutdown(context.Background())
	for _, i := range register.List() {
		i.Kill()
	}
}
