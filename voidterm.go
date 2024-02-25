package voidterm

import (
	"context"
	"encoding/hex"
	"github.com/1f349/voidterm/termutil"
	"github.com/creack/pty"
	docker "github.com/fsouza/go-dockerclient"
	"io"
	"os"
	"sync/atomic"
)

type VoidTerm struct {
	rows         uint16
	cols         uint16
	updateChan   chan struct{}
	dockerClient *docker.Client
	execInst     *docker.Exec
	pty1         *os.File
	tty1         *os.File
	execClose    docker.CloseWaiter
	term         *termutil.Terminal
	lastFrame    atomic.Pointer[ViewFrame]
	FrameChan    chan *ViewFrame
	PipeInput    io.Writer
}

func New(dockerEndpoint, container, user, workingDir string, cmd []string, rows, cols uint16) (v *VoidTerm, err error) {
	v = &VoidTerm{
		rows:       rows,
		cols:       cols,
		updateChan: make(chan struct{}),
		FrameChan:  make(chan *ViewFrame, 1),
	}
	v.lastFrame.Store(&ViewFrame{})

	v.dockerClient, err = docker.NewClient(dockerEndpoint)
	if err != nil {
		return nil, err
	}
	v.execInst, err = v.dockerClient.CreateExec(docker.CreateExecOptions{
		Cmd:          cmd,
		Container:    container,
		User:         user,
		WorkingDir:   workingDir,
		Context:      context.Background(),
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          true,
	})
	if err != nil {
		return nil, err
	}

	v.pty1, v.tty1, err = pty.Open()
	if err != nil {
		return nil, err
	}
	if err := pty.Setsize(v.pty1, &pty.Winsize{Rows: rows, Cols: cols}); err != nil {
		return nil, err
	}

	ir, iw := io.Pipe()
	or, ow := io.Pipe()
	v.PipeInput = iw

	v.execClose, err = v.dockerClient.StartExecNonBlocking(v.execInst.ID, docker.StartExecOptions{
		InputStream:  ir,
		OutputStream: ow,
		Tty:          true,
		RawTerminal:  true,
		Context:      context.Background(),
	})
	if err != nil {
		return nil, err
	}
	err = v.dockerClient.ResizeExecTTY(v.execInst.ID, int(rows), int(cols))
	if err != nil {
		return nil, err
	}

	go func() {
		r := io.TeeReader(or, hex.NewEncoder(os.Stdout))
		_, _ = io.Copy(v.tty1, r)
	}()

	v.term = termutil.New(termutil.WithWindowManipulator(&FakeWindow{Rows: rows, Cols: cols}))

	go func() {
		for {
			<-v.updateChan
			frame := ViewFrameFromBuffer(v.term.GetActiveBuffer())
			v.lastFrame.Store(frame)
			v.FrameChan <- frame
		}
	}()

	return
}

func (v *VoidTerm) Run() error {
	return v.term.Run(v.updateChan, v.rows, v.cols, v.pty1)
}

func (v *VoidTerm) LastFrame() *ViewFrame {
	return v.lastFrame.Load()
}
