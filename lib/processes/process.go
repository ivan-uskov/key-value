package processes

import (
	"os/exec"
)

type Worker interface {
	Kill()
}

type worker struct {
	killWorkersChan      chan bool
}

func (w *worker) Kill() {
	w.killWorkersChan <- true
}

func Run(cmd string, args []string) (Worker, error) {
	w := &worker{make(chan bool, 1)}

	errChan := make(chan error)
	go func() {
		command := exec.Command(cmd, args...)
		successChan := make(chan bool)
		err := command.Start()
		errChan <- err
		if err != nil {
			return
		}

		go func() {
			select {
			case <-w.killWorkersChan:
				command.Process.Kill()
			case <-successChan:
			}
		}()

		err = command.Wait()

		if err != nil {
			successChan <- false
		}

		successChan <- true
	} ()

	return w, <-errChan
}
