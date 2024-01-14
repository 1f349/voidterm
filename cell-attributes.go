package voidterm

import (
	"image/color"
)

type CellAttributes struct {
	fg            color.Color
	bg            color.Color
	bold          bool
	italic        bool
	dim           bool
	underline     bool
	strikethrough bool
	blink         bool
	inverse       bool
	hidden        bool
}
