package viewport

import "github.com/robinovitch61/kl/internal/viewport/linebuffer"

type RenderableComparable interface {
	Render() linebuffer.LineBufferer
	Equals(other interface{}) bool
}

type RenderableString struct {
	LineBuffer linebuffer.LineBufferer
}

func (r RenderableString) Render() linebuffer.LineBufferer {
	return r.LineBuffer
}

func (r RenderableString) Equals(other interface{}) bool {
	otherStr, ok := other.(RenderableString)
	if !ok {
		return false
	}
	if r.LineBuffer == nil || otherStr.LineBuffer == nil {
		return false
	}
	return r.LineBuffer.Content() == otherStr.LineBuffer.Content()
}

// assert RenderableString implements viewport.RenderableComparable
var _ RenderableComparable = RenderableString{}
