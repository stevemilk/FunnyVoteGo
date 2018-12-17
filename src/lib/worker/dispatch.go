package worker

import (
	"github.com/spf13/viper"
)

var (
	// nodeActionDispatcher global dispatcher
	nodeActionDispatcher Dispatcher
	// jobQueue is the global job queue
	// a buffered channel that we can send work requests on.
	jobQueue    chan execJob
	errQueue    chan *ExecJobError
	onJobErrors func(err *ExecJobError)
)

// ExecJobError def job error
type ExecJobError struct {
	Job Job
	Err error
}

// Dispatcher is the worker Dispatcher
type Dispatcher struct {
	// A pool of workers channels that are registered with the Dispatcher

	workerPool chan chan execJob
	maxWorkers int
}

type execJob struct {
	Ejob Job
	Done chan bool
}

// InitWorker make init
func InitWorker() {
	// init max queue size
	jobQueue = make(chan execJob, viper.GetInt("worker.queue_size"))
	errQueue = make(chan *ExecJobError, viper.GetInt("worker.max_worker_num"))

	nodeActionDispatcher := newDispatcher(viper.GetInt("worker.max_worker_num"))
	nodeActionDispatcher.run()
}

// SubmitFunc push func to worker
func SubmitFunc(f func() (err error)) {
	Submit(newFuncJob(f))
}

// Submit a job to job queue
func Submit(job Job) chan bool {
	wg := make(chan bool, 1)
	exej := execJob{Ejob: job, Done: wg}
	jobQueue <- exej

	return wg
}

// SetErrorHandle handle job exec error
func SetErrorHandle(handler func(err *ExecJobError)) {
	onJobErrors = handler
}

// NewDispatcher return a Dispatcher
func newDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan execJob, maxWorkers)
	return &Dispatcher{workerPool: pool, maxWorkers: maxWorkers}
}

// Run start worker loop
func (d *Dispatcher) run() {
	// starting n number of workers

	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.workerPool)
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-jobQueue:
			// a job request has been received

			go func(job execJob) {
				// try to obtain a worker job channel that is available.

				// this will block until a worker is idle

				jobChannel := <-d.workerPool

				// dispatch the job to the worker job channel

				jobChannel <- job
			}(job)
		case err := <-errQueue:
			// an error occurred

			// dispatch error to error handler

			go func(err *ExecJobError) {
				if onJobErrors != nil {
					// error callback
					onJobErrors(err)
				}
			}(err)
		}
	}
}

// Exec do exact  job
func (j execJob) exec() {
	if err := j.Ejob.Exec(); err != nil {
		errQueue <- &ExecJobError{Job: j.Ejob, Err: err}
	}
	// set job done
	j.Done <- true
}
