package voidterm

import (
	"io"
	"sync/atomic"
)

type Buffer struct {
	lines  []Line
	cursor Position

	inputStream io.Reader

	width, height atomic.Uint64

	done         chan struct{}
	updateBuffer chan struct{}
}

func NewBuffer(width, height uint64, input io.Reader) *Buffer {
	v := &Buffer{
		lines:        make([]Line, 0),
		cursor:       Position{0, 0},
		inputStream:  input,
		done:         make(chan struct{}),
		updateBuffer: make(chan struct{}),
	}
	v.width.Store(width)
	v.height.Store(height)
	return v
}

func (v *Buffer) renderBuffer() [][]Cell {
	width := v.width.Load()
	height := v.height.Load()
	buf := make([][]Cell, height)
	y := v.height.Load() - 1
	lineIdx := len(v.lines) - 1
	for _, i := range v.lines {
		w := i.Wrap(width)
		for _, j := range w {
			buf[y] = j
			y++
		}
	}
	return nil
}

var specialChars = map[rune]func(t *Buffer){
	0x07: handleOutputBell,
	0x08: handleOutputBackspace,
	'\n': handleOutputLineFeed,
	'\v': handleOutputLineFeed,
	'\f': handleOutputLineFeed,
	'\r': handleOutputCarriageReturn,
	'\t': handleOutputTab,
	0x0e: handleShiftOut, // handle switch to G1 character set
	0x0f: handleShiftIn,  // handle switch to G0 character set
}
