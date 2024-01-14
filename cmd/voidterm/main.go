package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/1f349/voidterm"
	"github.com/1f349/voidterm/termutil"
	"github.com/creack/pty"
	docker "github.com/fsouza/go-dockerclient"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

func main() {
	fmt.Println("PID:", os.Getpid())
	updateChan := make(chan struct{})

	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		log.Fatal(err)
	}
	execInst, err := client.CreateExec(docker.CreateExecOptions{
		Cmd:          []string{"/bin/sh"},
		Container:    "aa5e8ebe40c4",
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
		err := client.StartExec(execInst.ID, docker.StartExecOptions{
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

	go func() {
		r := io.TeeReader(or, hex.NewEncoder(os.Stdout))
		_, _ = io.Copy(tty1, r)
	}()

	term := termutil.New(termutil.WithWindowManipulator(&voidterm.FakeWindow{}))
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

	go func() {
		for {
			<-updateChan
			fmt.Println("Update buffer")
			a := viewToString(drawContent(term.GetActiveBuffer()))
			outputString.Store(&a)
		}
	}()

	go func() {
		err := http.ListenAndServe(":8080", http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if req.Method == http.MethodPost {
				switch req.FormValue("format") {
				case "cmd":
					iw.Write([]byte(req.FormValue("value") + "\n"))
				case "text":
					iw.Write([]byte(req.FormValue("value")))
				case "hex":
					dec, err := hex.DecodeString(req.FormValue("value"))
					if err != nil {
						http.Error(rw, "Invalid hex", http.StatusBadRequest)
						return
					}
					iw.Write(dec)
				}
				time.Sleep(100 * time.Millisecond)
				http.Redirect(rw, req, "/", http.StatusFound)
				return
			}

			rw.Header().Set("Content-Type", "text/html; charset=utf-8")
			rw.WriteHeader(http.StatusOK)
			view := *outputString.Load()
			rw.Write([]byte("<textarea style='width:100%;height:750;'>" + view + "</textarea>"))
			rw.Write([]byte("<form method=post>CMD: <input type=hidden name=format value=cmd><input type=text name=value><input type=submit value=Submit></form>"))
			rw.Write([]byte("<form method=post>Text: <input type=hidden name=format value=text><input type=text name=value><input type=submit value=Submit></form>"))
			rw.Write([]byte("<form method=post>Hex: <input type=hidden name=format value=hex><input type=text name=value><input type=submit value=Submit></form>"))
			rw.Write([]byte("<form method=post><input type=hidden name=format value=hex>"))
			for _, i := range []byte{'C' - '@', 'G' - '@', 'X' - '@', '\n'} {
				t := fmt.Sprintf("Ctrl+%c", i+'@')
				if i == '\n' {
					t = "Enter"
				}
				rw.Write([]byte("<button type=submit name=value value=" + fmt.Sprintf("%02x", i) + ">" + t + "</button>"))
			}
			rw.Write([]byte("</form>"))
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
