package aglite

import "io"

type search struct {
	Task    chan match
	root    string
	pattern string
	out     io.Writer
	Done    chan struct{}
}

func (s search) start() {
	grepChan := make(chan string, 5000)

	go find{out: grepChan}.start(s.root)
	go grep{in: grepChan, done: s.Done, printer: newPrinter(s.out)}.start(s.pattern, s.Task)
}
