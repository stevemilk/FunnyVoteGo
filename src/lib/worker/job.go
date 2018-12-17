// Package worker exec job sync
package worker

// Job interface
type Job interface {
	Exec() error
}
