package voidterm

import (
	"github.com/1f349/voidterm/termutil"
	"log"
)

type FakeWindow struct{}

var _ termutil.WindowManipulator = &FakeWindow{}

func (f *FakeWindow) State() termutil.WindowState {
	return termutil.StateNormal
}

func (f *FakeWindow) Minimise() {
	log.Println("Minimise")
}

func (f *FakeWindow) Maximise() {
	log.Println("Maximise")
}

func (f *FakeWindow) Restore() {
	log.Println("Restore")
}

func (f *FakeWindow) SetTitle(title string) {
	log.Println("SetTitle", title)
}

func (f *FakeWindow) Position() (int, int) {
	log.Println("Position")
	return 0, 0
}

func (f *FakeWindow) SizeInPixels() (int, int) {
	log.Println("SizeInPixels")
	return 100, 80
}

func (f *FakeWindow) CellSizeInPixels() (int, int) {
	log.Println("CellSizeInPixels")
	return 14, 11
}

func (f *FakeWindow) SizeInChars() (int, int) {
	log.Println("SizeInChars")
	return 14, 11
}

func (f *FakeWindow) ResizeInPixels(x int, y int) {
	log.Println("ResizeInPixels", x, y)
}

func (f *FakeWindow) ResizeInChars(x int, y int) {
	log.Println("ResizeInChars", x, y)
}

func (f *FakeWindow) ScreenSizeInPixels() (int, int) {
	log.Println("ScreenSizeInPixels")
	return 1920, 1080
}

func (f *FakeWindow) ScreenSizeInChars() (int, int) {
	log.Println("ScreenSizeInChars")
	return 96, 108
}

func (f *FakeWindow) Move(x, y int) {
	log.Println("Move", x, y)
}

func (f *FakeWindow) IsFullscreen() bool {
	log.Println("IsFullscreen")
	return false
}

func (f *FakeWindow) SetFullscreen(enabled bool) {
	log.Println("SetFullscreen", enabled)
}

func (f *FakeWindow) GetTitle() string {
	log.Println("GetTitle")
	return "Title"
}

func (f *FakeWindow) SaveTitleToStack() {
	log.Println("SaveTitleToStack")
}

func (f *FakeWindow) RestoreTitleFromStack() {
	log.Println("RestoreTitleFromStack")
}

func (f *FakeWindow) ReportError(err error) {
	log.Println("ReportError", err)
}
