package aglite

type match struct {
	Path  string `json:"path"`
	Lines []line `json:"lines"`
}

type returnmatch struct {
	ID   uint   `json:"id"`
	Path string `json:"path"`
	Name string `json:"name"`
	Text string `json:"text"`
}

type line struct {
	Num  int    `json:"num"`
	Text string `json:"text"`
}

func (m *match) add(num int, text string) {
	m.Lines = append(m.Lines, line{Num: num, Text: text})
}
