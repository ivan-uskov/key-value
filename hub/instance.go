package main

import (
	"key-value/lib/processes"
	"os"
	"path/filepath"
	"runtime"
	"fmt"
	"key-value/lib/routers"
	"sync"
)

type Instance interface {
	Ping() bool
	Restart() error
	Kill()
}

type instance struct {
	address             string
	worker              processes.Worker
	ws                  routers.Client
	sync.RWMutex
	stopRestartLoopChan chan bool
}

func NewInstance(address string) (Instance, error) {
	i := &instance{address, nil, nil, sync.RWMutex{}, nil}
	err := i.start()

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

func (i *instance) Restart() error {
	i.Lock()
	defer i.Unlock()

	i.unsafeKill()
	return i.start()
}

func (i *instance) unsafeKill() {
	if i.launched() {
		i.stopRestartLoopChan <- true
		i.ws.Close()
		i.ws = nil
		i.worker.Kill()
		i.worker = nil
	}
}

func (i *instance) launched() bool {
	return i.ws != nil && i.worker != nil
}

func (i *instance) start() error {
	instancePath, err := getInstanceExecutablePath()
	if err != nil {
		return err
	}

	i.worker, err = processes.Run(instancePath, `-addr`, i.address)
	if err != nil {
		return err
	}

	i.ws, err = routers.NewClient(i.address, `ws`)
	if err != nil {
		i.worker.Kill()
		i.worker = nil
		return err
	}

	resp, err := i.ws.SendSync(routers.Request{Action: routers.PING})
	if err != nil {
		i.unsafeKill()
		return err
	} else if !resp.Success {
		return fmt.Errorf(resp.Error)
	}

	i.runHandler()

	return nil
}

func (i *instance) runHandler() {
	i.stopRestartLoopChan = make(chan bool, 1)
	go func() {
		for {
			select {
			case <-i.stopRestartLoopChan:
				fmt.Println(`stop restart`)
				return
			case <-i.worker.GetStopChan():
				if i.launched() {
					err := i.Restart()
					if err != nil {
						fmt.Printf("%s killed and got error on restart %v\n", i.address, err)
					} else {
						fmt.Printf("%s restarted\n", i.address)
					}
				}
				return
			}
		}
	}()
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
