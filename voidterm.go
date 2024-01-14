package voidterm

import (
	"github.com/1f349/voidterm/termutil"
)

type VoidTerm struct {
	term *termutil.Terminal
}

func New(dockerContainer, shell string) *VoidTerm {
	term := termutil.New(termutil.WithShell(shell))
	return &VoidTerm{term: term}
}

func (v *VoidTerm) Run() {
	v.term.Run()
}
