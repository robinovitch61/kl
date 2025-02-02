package linebuffer

import "testing"

var equivalentLineBuffers = [][]LineBufferer{
	{
		New("hello world"),
		NewMulti(
			New("hello"),
			New(" world"),
		),
		NewMulti(
			New("hel"),
			New("lo "),
			New("wo"),
			New("rld"),
		),
		NewMulti(
			New("h"),
			New("e"),
			New("l"),
			New("l"),
			New("o"),
			New(" "),
			New("w"),
			New("o"),
			New("r"),
			New("l"),
			New("d"),
		),
	},
}

func TestMultiLineBuffer_Content(t *testing.T) {
	for _, eq := range equivalentLineBuffers {
		for i, lb := range eq {
			if lb.Content() != eq[0].Content() {
				t.Errorf("expected %q, got %q for line buffer %d", eq[0].Content(), lb.Content(), i)
			}
		}
	}
}

//func TestMultiLin
