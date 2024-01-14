package voidterm

import (
	"os"
	"sync"
)

const (
	MainBuffer     uint8 = 0
	AltBuffer      uint8 = 1
	InternalBuffer uint8 = 2
)

type VoidTerm struct {
	mu             sync.Mutex
	pty            *os.File
	updateChan     chan struct{}
	processChan    chan MeasuredRune
	closeChan      chan struct{}
	buffers        []*Buffer
	activeBuffer   *Buffer
	mouseMode      MouseMode
	mouseExtMode   MouseExtMode
	running        bool
	shell          string
	initialCommand string
}
