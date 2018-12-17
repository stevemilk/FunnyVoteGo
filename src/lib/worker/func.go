package worker

// FuncJob for func submit
type FuncJob struct {
	f func() error
}

func newFuncJob(fb func() error) FuncJob {
	return FuncJob{
		f: fb,
	}
}

// Exec impl
func (f FuncJob) Exec() error {
	return f.f()
}
