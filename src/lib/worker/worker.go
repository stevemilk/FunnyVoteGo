package worker

// Worker request the worker that executes the job.
type worker struct {
	workerPool chan chan execJob
	jobChanel  chan execJob
	quit       chan bool
}

// NewWorker return a worker.
func newWorker(workerPool chan chan execJob) worker {
	return worker{
		workerPool: workerPool,
		jobChanel:  make(chan execJob),
		quit:       make(chan bool),
	}
}

// Start method starts the run loop for worker,
// listen for a quit channel case we need to stop.
func (w worker) start() {
	go func() {
		for {
			// release a jobChanel before exec it
			w.workerPool <- w.jobChanel
			select {
			// received a wor request.

			case job := <-w.jobChanel:
				job.exec()
			case <-w.quit:
				// have received a signal to stop.

				return
			}
		}
	}()
}

// Stop signals the worker to stop listening for work requests.
func (w worker) stop() {
	go func() {
		w.quit <- true
	}()
}
