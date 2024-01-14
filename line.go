package voidterm

type Line struct {
	wrapped bool
	cells   []Cell
}

func LineFromRunes(runes []rune, style CellAttributes) Line {
	l := make(Line, len(runes))
	for i, r := range runes {
		l[i] = Cell{
			r: r,
			s: style,
		}
	}
	return l
}

func (l Line) Wrap(width uint64) WrappedLine {
	if uint64(len(l)) <= width {
		return WrappedLine{l}
	}
	a := l
	w := make(WrappedLine, 0, 1+uint64(len(l)-1)/width)
	for uint64(len(a)) > width {
		w = append(w, a[:width])
		a = a[width:]
	}
	return w
}

type WrappedLine []Line
