package aglite

import (
	"os"
)

type find struct {
	out chan string
}

func (f find) start(root string) {
	f.findFile(root)
}

func (f find) findFile(root string) {
	concurrentWalk(root, func(path string, info os.FileInfo) error {
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}
		f.out <- path
		return nil
	})
	close(f.out)
}
