package voidterm

type Cell struct {
	r    MeasuredRune
	attr CellAttributes
}

func (c *Cell) Attr() CellAttributes {
	return c.attr
}

func (c *Cell) Rune() MeasuredRune {
	return c.r
}
