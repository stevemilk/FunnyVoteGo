package aglite

import (
	"hyperbaas/src/api/vm"
	"hyperbaas/src/model"
	"hyperbaas/src/util"
	"io"
	"os"
	"strings"
	"time"

	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const (
	// ExitCodeOK enum
	ExitCodeOK = iota
)

// Searcher def
type Searcher struct {
	Out, Err io.Writer
}

// Run to start search
func (p Searcher) Run(args []string) int {
	search := search{
		pattern: p.patternFrom(args),
		root:    p.rootFrom(args),
		out:     p.Out,
	}
	search.start()
	return ExitCodeOK
}

func (p Searcher) patternFrom(args []string) string {
	return args[0]
}

func (p Searcher) rootFrom(args []string) string {
	var root string
	if len(args) > 1 {
		root = args[1]
	} else {
		root = "."
	}
	return root
}

// AgSearch for entry
// TODO : entry
func AgSearch(keyword, path string, index, size int) vm.Page {
	//Mcache := mcache.StartInstance()
	//if pointer, ok := Mcache.GetPointer(keyword); ok {
	//	return pointer.(vm.Page)
	//}
	var ms []match
	search := search{
		Task:    make(chan match, 1000),
		pattern: keyword,
		root:    path,
		out:     os.Stdout,
		Done:    make(chan struct{}),
	}
	search.start()

	done := false
	for !done {
		select {
		case m := <-search.Task:
			ms = append(ms, m)
		case <-search.Done:
			done = true
			break
		case <-time.After(time.Second * 2):
			done = true
			break
		}
	}

	// deal with match
	var rm []returnmatch
	for _, m := range ms {
		//pattern := "^conf/doc/[0-9.]+(.+)/[0-9.]+(.+).md$"
		//re := regexp.MustCompile(pattern)
		//params := re.FindStringSubmatch(m.Path)
		//path := ""
		//for i, param := range params {
		//	if i == 0 {
		//		continue
		//	}
		//	path = path + "/" + param

		//}
		var rt returnmatch
		rt.Path = m.Path

		doc, err := model.GetDocByPath(m.Path)
		if err != nil {
			return vm.Page{}
		}

		rt.ID = doc.ID
		rt.Name = doc.Name
		if len(m.Lines) != 0 {
			// markdown -> html -> pure txt
			output := blackfriday.Run([]byte(m.Lines[0].Text), blackfriday.WithNoExtensions())
			str := util.HTML2Txt(string(output[:]))
			if b := strings.Contains(str, "|"); b {
				continue
			}
			if b := strings.Contains(str, "\\t"); b {
				continue
			}

			if str == "" {
				continue
			}
			rt.Text = str
		}
		rm = append(rm, rt)
	}

	list := util.Paging(index, size, rm)
	if list.List == nil {
		list.List = []returnmatch{}
	}

	//err := Mcache.SetPointer(keyword, list, time.Second*20)
	//if err != nil {
	//	return vm.Page{}
	//}
	return list
}
