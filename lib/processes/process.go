package processes

import (
	"os/exec"
	"sync"
	"fmt"
)

type Worker interface {
	Kill()
}

type worker struct {
	endWorkersWaitGroup  *sync.WaitGroup
	killWorkersChan      chan bool
	terminateWorkersChan chan bool
}

func (w *worker) Kill() {
	w.killWorkersChan <- true
}

func createWorker() * worker {
	w := worker{}
	w.endWorkersWaitGroup = &sync.WaitGroup{}
	w.killWorkersChan = make(chan bool)
	w.terminateWorkersChan = make(chan bool)

	return &w
}

func Run(cmd string, args []string) (Worker, error) {
	w := createWorker()

	errChan := make(chan error)
	go func() {
		command := exec.Command(cmd, args...)
		successChan := make(chan bool)
		w.endWorkersWaitGroup.Add(1)
		fmt.Println("Start command", cmd, args)
		err := command.Start()
		errChan <- err
		if err != nil {
			return
		}

		go func() {
			select {
			case <-w.killWorkersChan:
				fmt.Println("Kill worker catched", args)
				command.Process.Kill()
			case <-w.terminateWorkersChan:
				fmt.Println("Terminate worker catched", args)
			case <-successChan:
			}
		}()

		err = command.Wait()
		w.endWorkersWaitGroup.Done()

		if err != nil {
			successChan <- false
		}

		successChan <- true
	} ()

	return w, <-errChan
}
