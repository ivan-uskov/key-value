package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"key-value/lib/processes"
	"key-value/lib/routers"
	"key-value/lib/ws"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var addr = flag.String("addr", ":8372", "http service address")

func getInstanceExecutablePath() (string, error) {
	exePath, err := os.Executable()
	exeDir := filepath.Dir(exePath)

	filename := "instance"
	if runtime.GOOS == "windows" {
		filename = "instance.exe"
	}
	candidates := []string{filename, filepath.Join("..", "instance", filename)}
	for _, candidate := range candidates {
		instancePath := filepath.Join(exeDir, candidate)
		_, err = os.Stat(instancePath)
		if err != nil {
			continue
		}
		return instancePath, nil
	}
	return "", err
}

func createSetter(reg Register) routers.RequestStrategy {
	return func(r routers.Request) (string, error) {
		if reg.Exists(r.Option1) {
			return ``, errors.New(`Instance already started`)
		}

		args := []string{}
		args = append(args, fmt.Sprintf(`--addr=%s`, r.Option1))
		instancePath, err := getInstanceExecutablePath()
		if err != nil {
			return "", err
		}

		worker := processes.Run(instancePath, args)
		reg.Add(r.Option1, worker)

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
	server := ws.NewServer()
	router := routers.New()
	router.SetSetter(createSetter(register))
	router.SetLister(createLister(register))
	router.SetRemover(createRemover(register))

	http.HandleFunc("/ctl", func(w http.ResponseWriter, r *http.Request) {
		server.Serve(w, r, router.CreateWebSocketHandler())
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
