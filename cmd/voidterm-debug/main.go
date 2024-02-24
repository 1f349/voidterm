package main

import (
	"context"
	_ "embed"
	"encoding/hex"
	"fmt"
	"github.com/1f349/voidterm"
	"github.com/1f349/voidterm/termutil"
	"github.com/creack/pty"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

var upgrader = websocket.Upgrader{}

//go:embed index.go.html
var indexPage string

//go:embed keysight.umd.js
var keysightJs string

type TermSend struct {
	Text string
}

type TermAction struct {
	Code byte
}

func main() {
	fmt.Println("PID:", os.Getpid())
	updateChan := make(chan struct{})

	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatal(err)
	}
	execInst, err := client.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{"/bin/sh"},
		Container:    "07d2ab561d0a",
		User:         "root",
		WorkingDir:   "/",
		Context:      context.Background(),
		AttachStdin:  true,
		AttachStdout: true,
		Tty:          true,
	})
	if err != nil {
		log.Fatal(err)
	}

	var rows uint16 = 40
	var cols uint16 = 132

	pty1, tty1, err := pty.Open()
	if err != nil {
		log.Fatal(err)
	}
	pty.Setsize(pty1, &pty.Winsize{Rows: rows, Cols: cols})

	ir, iw := io.Pipe()
	or, ow := io.Pipe()

	go func() {
		err = client.StartExec(execInst.ID, docker.StartExecOptions{
			InputStream:  ir,
			OutputStream: ow,
			ErrorStream:  nil,
			Tty:          true,
			RawTerminal:  true,
			Context:      context.Background(),
		})
		if err != nil {
			log.Fatal(err)
		}
	}()

	err = client.ResizeExecTTY(execInst.ID, int(rows), int(cols))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		r := io.TeeReader(or, hex.NewEncoder(os.Stdout))
		_, _ = io.Copy(tty1, r)
	}()

	term := termutil.New(termutil.WithWindowManipulator(&voidterm.FakeWindow{Rows: rows, Cols: cols}))
	go func() {
		err = term.Run(updateChan, rows, cols, pty1)
		if err != nil {
			log.Fatal(err)
		}
	}()

	var outputString atomic.Pointer[string]
	{
		a := ""
		outputString.Store(&a)
	}
	outputChan := make(chan string, 1)

	go func() {
		for {
			<-updateChan
			fmt.Println("Update buffer")
			a := viewToString(drawContent(term.GetActiveBuffer()))
			outputString.Store(&a)
			outputChan <- a
		}
	}()

	htmlPageTmpl, err := template.New("page").Parse(indexPage)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		err := http.ListenAndServe(":8080", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Header.Get("Upgrade") == "websocket" {
				c, err := upgrader.Upgrade(rw, req, nil)
				if err != nil {
					return
				}
				_ = c.WriteJSON(TermSend{Text: *outputString.Load()})
				go func() {
					for {
						a := <-outputChan
						err := c.WriteJSON(TermSend{Text: a})
						if err != nil {
							return
						}
					}
				}()
				for {
					var a TermAction
					err := c.ReadJSON(&a)
					if err != nil {
						return
					}
					_, _ = iw.Write([]byte{a.Code})
				}
			}

			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			var vv []map[string]string
			for _, i := range []byte{'C' - '@', 'G' - '@', 'X' - '@', '\n'} {
				t := fmt.Sprintf("Ctrl+%c", i+'@')
				if i == '\n' {
					t = "Enter"
				}
				vv = append(vv, map[string]string{"Hex": fmt.Sprintf("%02x", i), "Text": t})
			}
			_ = htmlPageTmpl.Execute(rw, map[string]any{
				"Keysight": template.JS(keysightJs),
				"Buttons":  vv,
			})
		}))
		if err != nil {
			log.Fatal(err)
		}
	}()

	done := make(chan struct{}, 1)
	<-done
}

func viewToString(view [][]rune) string {
	var sb strings.Builder
	for _, row := range view {
		for _, cell := range row {
			sb.WriteRune(cell)
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

func drawContent(buffer *termutil.Buffer) [][]rune {
	view := make([][]rune, buffer.ViewHeight())
	for i := range view {
		view[i] = make([]rune, buffer.ViewWidth())
	}

	// draw base content for each row
	for viewY := int(buffer.ViewHeight() - 1); viewY >= 0; viewY-- {
		drawRow(view, buffer, viewY)
	}
	return view
}

func drawRow(view [][]rune, buffer *termutil.Buffer, viewY int) {
	rowView := view[viewY]

	for i := range rowView {
		rowView[i] = ' '
	}

	// draw text content of each cell in row
	for viewX := uint16(0); viewX < buffer.ViewWidth(); viewX++ {
		cell := buffer.GetCell(viewX, uint16(viewY))

		// we don't need to draw empty cells
		if cell == nil || cell.Rune().Rune == 0 {
			continue
		}

		// draw the text for the cell
		rowView[viewX] = cell.Rune().Rune
	}
}
