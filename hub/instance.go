package main

import (
	"key-value/lib/processes"
	"os"
	"path/filepath"
	"runtime"
	"fmt"
	"key-value/lib/routers"
)

type Instance interface {
	Ping() bool
	Kill()
}

type instance struct {
	address string
	worker processes.Worker
	ws routers.Client
}

func (i *instance) Ping() bool {
	resp, err := i.ws.SendSync(routers.Request{Action:routers.PING})
	return err != nil || resp != `1`
}

func (i *instance) Kill() {
	i.worker.Kill()
}

func NewInstance(address string) (Instance, error) {
	instancePath, err := getInstanceExecutablePath()
	if err != nil {
		return nil, err
	}

	args := []string{fmt.Sprintf(`--addr=%s`, address)}
	worker, err := processes.Run(instancePath, args)
	if err != nil {
		return nil, err
	}

	c , err := routers.NewClient(address, `ws`)
	if err != nil {
		worker.Kill()
		return nil, err
	}

	resp, err := c.SendSync(routers.Request{Action:routers.PING})
	if err != nil || resp != `1`{
		worker.Kill()
		return nil, err
	}

	return &instance{address, worker, c}, nil
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