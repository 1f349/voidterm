package voidterm

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLine_Wrap(t *testing.T) {
	t.Run("too wide", func(t *testing.T) {
		l := LineFromRunes([]rune("Hello world!"), CellAttributes{}).Wrap(16)
		assert.Equal(t, WrappedLine{
			LineFromRunes([]rune("Hello world!"), CellAttributes{}),
		}, l)
	})
	t.Run("too thin", func(t *testing.T) {
		l := LineFromRunes([]rune("Hello world!"), CellAttributes{}).Wrap(4)
		assert.Equal(t, WrappedLine{
			LineFromRunes([]rune("Hell"), CellAttributes{}),
			LineFromRunes([]rune("o wo"), CellAttributes{}),
			LineFromRunes([]rune("rld!"), CellAttributes{}),
		}, l)
	})
}
