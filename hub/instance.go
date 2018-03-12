package main

import (
	"key-value/lib/processes"
	"os"
	"path/filepath"
	"runtime"
	"fmt"
	"key-value/lib/routers"
	"sync"
	"time"
	"encoding/json"
)

type Instance interface {
	Ping() bool
	Restart(others []string) error
	Kill()
}

type instance struct {
	address             string
	worker              processes.Worker
	ws                  routers.Client
	rws                 routers.Client
	sync.RWMutex
}

func NewInstance(address string, others []string) (Instance, error) {
	i := &instance{address: address, RWMutex: sync.RWMutex{}}
	err := i.start(others)

	if err != nil {
		i = nil
	}

	return i, err
}

func (i *instance) Ping() bool {
	i.RLock()
	defer i.RUnlock()
	resp, err := i.ws.SendSync(routers.Request{Action: routers.PING})
	return err == nil && resp.Success
}

func (i *instance) Kill() {
	i.Lock()
	i.unsafeKill()
	i.Unlock()
}

func (i *instance) Restart(others []string) error {
	i.Lock()
	defer i.Unlock()

	i.unsafeKill()
	return i.start(others)
}

func (i *instance) unsafeKill() {
	if i.launched() {
		i.ws.Close()
		i.ws = nil
		i.worker.Kill()
		i.worker = nil
	}
}

func (i *instance) launched() bool {
	return i.ws != nil && i.worker != nil
}

func (i *instance) start(others []string) error {
	err := i.runWorker()
	if err != nil {
		return err
	}

	err = i.createWS()
	if err != nil {
		i.worker.Kill()
		i.worker = nil
		return err
	}

	err = i.sendOther(others)
	if err != nil {
		i.unsafeKill()
		return err
	}

	return nil
}

func (i *instance) runWorker() error {
	instancePath, err := getInstanceExecutablePath()
	if err != nil {
		return err
	}

	i.worker, err = processes.Run(instancePath, `-addr`, i.address)
	if err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)
	return nil
}

func (i *instance) createWS() error {
	var err error
	i.ws, err = routers.NewClient(i.address, `ws`)
	if err != nil {
		return err
	}

	resp, err := i.ws.SendSync(routers.Request{Action: routers.PING})
	if err != nil {
		i.ws.Close()
		return err
	} else if !resp.Success {
		i.ws.Close()
		return fmt.Errorf(resp.Error)
	}

	return nil
}

func (i *instance) sendOther(others []string) error {
	data, err := json.Marshal(others)
	if err != nil {
		return err
	}

	resp, err := i.ws.SendSync(routers.Request{Action: `NODES`, Option1: string(data)})
	if err != nil {
		return err
	} else if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	return nil
}

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
