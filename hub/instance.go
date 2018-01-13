package main

import (
	"key-value/lib/processes"
	"os"
	"path/filepath"
	"runtime"
	"fmt"
)

type Instance interface {
	Ping() bool
	Kill()
}

type instance struct {
	address string
	worker processes.Worker
}

func (i *instance) Ping() bool {
	return true
}

func (i *instance) Kill() {
	i.worker.Kill()
}

func NewInstance(address string) (Instance, error) {
	instancePath, err := getInstanceExecutablePath()
	if err != nil {
		return nil, err
	}

	args := []string{fmt.Sprintf(`--addr=%s`, r.Option1)}
	worker, err := processes.Run(instancePath, args)
	if err != nil {
		return nil, err
	}

	return &instance{address, worker}, nil
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