package processes

import (
	"os/exec"
)

type Worker interface {
	Kill()
	GetStopChan() chan bool
}

type worker struct {
	killWorkersChan chan bool
	killWaitChan    chan bool
	stopChan        chan bool
}

func (w *worker) Kill() {
	w.killWorkersChan <- true
	<-w.killWaitChan
}

func (w *worker) GetStopChan() chan bool {
	return w.stopChan
}

func Run(cmd string, args []string) (Worker, error) {
	w := &worker{make(chan bool, 1), make(chan bool, 1), make(chan bool, 1)}

	errChan := make(chan error, 1)
	go func() {
		command := exec.Command(cmd, args...)
		waitChan := make(chan error)
		err := command.Start()
		errChan <- err
		if err != nil {
			return
		}

		go func() {
			for {
				select {
				case <-w.killWorkersChan:
					command.Process.Kill()
				case err := <-waitChan:
					w.stopChan <- err == nil
					w.killWaitChan <- true
					break
				}
			}
		}()

		waitChan <- command.Wait()
	}()

	return w, <-errChan
}
