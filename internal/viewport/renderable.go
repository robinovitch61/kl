package viewport

import "github.com/robinovitch61/kl/internal/viewport/linebuffer"

type RenderableComparable interface {
	Render() linebuffer.LineBuffer
	Equals(other interface{}) bool
}

type RenderableString struct {
	LineBuffer linebuffer.LineBuffer
}

func (r RenderableString) Render() linebuffer.LineBuffer {
	return r.LineBuffer
}

func (r RenderableString) String() string {
	return r.Render().Content
}

func (r RenderableString) Equals(other interface{}) bool {
	otherStr, ok := other.(RenderableString)
	if !ok {
		return false
	}
	return r.LineBuffer.Content == otherStr.LineBuffer.Content
}

// assert RenderableString implements viewport.RenderableComparable
var _ RenderableComparable = RenderableString{}
