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
