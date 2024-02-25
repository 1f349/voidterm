package main

import (
	_ "embed"
	"flag"
	"fmt"
	"github.com/1f349/voidterm"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
	"os"
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
	dockerEndpoint := flag.String("docker", "unix:///var/run/docker.sock", "docker endpoint")
	contId := flag.String("c", "", "container ID")
	contUser := flag.String("u", "", "container user")
	rows := flag.Uint64("rows", 40, "number of rows")
	cols := flag.Uint64("cols", 132, "number of columns")
	flag.Parse()

	fmt.Println("PID:", os.Getpid())
	term, err := voidterm.New(*dockerEndpoint, *contId, *contUser, "/", []string{"/bin/sh"}, uint16(*rows), uint16(*cols))
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := term.Run(); err != nil {
			log.Fatal(err)
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
				_ = c.WriteJSON(TermSend{Text: term.LastFrame().RenderRawString()})
				go func() {
					for {
						a := <-term.FrameChan
						err := c.WriteJSON(TermSend{Text: a.RenderRawString()})
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
					_, _ = term.PipeInput.Write([]byte{a.Code})
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
