package viewport

type RenderableComparable interface {
	Render() string
	Equals(other interface{}) bool
}

type RenderableString struct {
	Content string
}

func (r RenderableString) Render() string {
	return r.Content
}

func (r RenderableString) Equals(other interface{}) bool {
	otherStr, ok := other.(RenderableString)
	if !ok {
		return false
	}
	return r.Content == otherStr.Content
}

// assert RenderableString implements viewport.RenderableComparable
var _ RenderableComparable = RenderableString{}
