package tui

type rectangle struct{ x, y, width, height int }
type layout struct {
	width, height, bodyHeight, listWidth, detailsWidth int
	modal                                              rectangle
}

func measure(width, height int) layout {
	w, h := max(0, width), max(0, height)
	l := layout{width: w, height: h, bodyHeight: max(0, h-2)}
	l.listWidth = w/5*2 + w%5*2/5
	l.detailsWidth = w - l.listWidth
	mw, mh := min(72, max(0, w-4)), min(20, max(0, h-4))
	l.modal = rectangle{(w - mw) / 2, 1 + (l.bodyHeight-mh)/2, mw, mh}
	return l
}

func (l layout) usable() bool { return l.width >= 80 && l.height >= 24 }
