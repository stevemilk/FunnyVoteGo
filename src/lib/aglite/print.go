package aglite

import (
	"fmt"
	"io"
	"sync"
)

type printer struct {
	mu  *sync.Mutex
	out io.Writer
}

func newPrinter(out io.Writer) printer {
	return printer{
		mu:  new(sync.Mutex),
		out: out,
	}
}

func (p printer) print(match match) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Fprintln(p.out, match.Path)
	for _, line := range match.Lines {
		fmt.Fprintf(p.out, "%d:%s\n", line.Num, line.Text)
	}
}
