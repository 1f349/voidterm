package voidterm

import (
	"github.com/1f349/voidterm/termutil"
	"strings"
)

type ViewFrame struct {
	// Cells stores individual cells in [row][col] format
	Cells [][]*termutil.Cell
}

func (v *ViewFrame) RenderRawString() string {
	var sb strings.Builder
	for _, row := range v.Cells {
		for _, cell := range row {
			if cell == nil {
				sb.WriteRune(' ')
			} else {
				sb.WriteRune(cell.Rune().Rune)
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func ViewFrameFromBuffer(buffer *termutil.Buffer) *ViewFrame {
	frame := &ViewFrame{}
	frame.Cells = make([][]*termutil.Cell, buffer.ViewHeight())
	for i := range frame.Cells {
		frame.Cells[i] = make([]*termutil.Cell, buffer.ViewWidth())
	}

	// draw base content for each row
	for viewY := uint16(0); viewY < buffer.ViewHeight(); viewY++ {
		for viewX := uint16(0); viewX < buffer.ViewWidth(); viewX++ {
			cell := buffer.GetCell(viewX, viewY)

			// we don't need to draw empty cells
			if cell == nil || cell.Rune().Rune == 0 {
				continue
			}

			a := *cell
			frame.Cells[viewY][viewX] = &a
		}
	}

	return frame
}
