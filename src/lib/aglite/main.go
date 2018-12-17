package aglite

import (
	"os"
	"runtime"

	"github.com/glog"
)

func init() {
	if cpu := runtime.NumCPU(); cpu == 1 {
		runtime.GOMAXPROCS(2)
	} else {
		runtime.GOMAXPROCS(cpu)
	}
}

func main() {
	if len(os.Args) < 2 {
		glog.Info("error: please input patten")
		os.Exit(1)
	}
	pt := Searcher{Out: os.Stdout, Err: os.Stderr}
	exitCode := pt.Run(os.Args[1:]) // ag keyword .
	// map[string]match :max size > LRU time limit
	os.Exit(exitCode)
}
