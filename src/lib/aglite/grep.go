package aglite

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"sync"

	"github.com/glog"
)

//var TASK = make(chan match, 1000)

var newLine = []byte("\n")

// grep model
type grep struct {
	in      chan string
	done    chan struct{}
	printer printer
}

// start to start many goroutines to search file according to pattern
func (g grep) start(pattern string, task chan match) {
	sem := make(chan struct{}, 208)
	wg := &sync.WaitGroup{}

	p := []byte(pattern)
	for path := range g.in {
		sem <- struct{}{}
		wg.Add(1)
		go g.grep(path, p, sem, wg, task)
	}
	wg.Wait()
	//println("dddddddddddd")
	g.done <- struct{}{}
}

// grep is one search in a goroutine
func (g grep) grep(path string, pattern []byte, sem chan struct{}, wg *sync.WaitGroup, task chan match) {

	f, err := os.Open(path)
	if err != nil {
		glog.Info("open: ", err)
		return
	}

	buf := make([]byte, 8196)
	var stash []byte
	identified := false
	var encoding int

	for {
		c, err := f.Read(buf)
		if err != nil && err != io.EOF {
			glog.Info("read: ", err)
			break
		}

		if err == io.EOF {
			break
		}

		if !identified {
			limit := c
			if limit > 512 {
				limit = 512
			}

			encoding = detectEncoding(buf[:limit])
			if encoding == ERROR || encoding == BINARY {
				// ignore unknown file and binary file
				break
			}
			identified = true
		}

		// repair first line from previous last line.
		if len(stash) > 0 {
			var repaired []byte
			index := bytes.Index(buf[:c], newLine)
			if index == -1 {
				repaired = append(stash, buf[:c]...)
			} else {
				repaired = append(stash, buf[:index]...)
			}
			// grep from repaied line.
			if bytes.Contains(bytes.ToLower(repaired), bytes.ToLower(pattern)) {
				m := g.grepEachLines(f, pattern)
				task <- m
				//matches = append(matches, m)
				break
			}
		}

		// grep from buffer.
		if bytes.Contains(bytes.ToLower(buf[:c]), bytes.ToLower(pattern)) {
			m := g.grepEachLines(f, pattern)
			task <- m
			//matches = append(matches, m)
			break
		}

		// stash last line.
		index := bytes.LastIndex(buf[:c], newLine)
		if index == -1 {
			stash = append(stash, buf[:c]...)
		} else {
			stash = make([]byte, c-index)
			copy(stash, buf[index:c])
		}
	}

	f.Close()
	<-sem
	wg.Done()
}

// grepEachLines grep each line
func (g grep) grepEachLines(f *os.File, pattern []byte) match {
	f.Seek(0, 0)
	match := match{Path: f.Name()}
	scanner := bufio.NewScanner(f)
	line := 1
	num := 1
	for scanner.Scan() && num <= 10 {
		if bytes.Contains(bytes.ToLower(scanner.Bytes()), bytes.ToLower(pattern)) {
			match.add(line, scanner.Text())
			num++
		}
		line++
	}
	g.printer.print(match)
	return match

}
