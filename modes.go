package voidterm

type Modes struct {
	ShowCursor            bool
	ApplicationCursorKeys bool
	BlinkingCursor        bool
	ReplaceMode           bool
	OriginMode            bool
	LineFeedMode          bool
	ScreenMode            bool
	AutoWrap              bool
	SixelScrolling        bool
	BracketedPasteMode    bool
}

type MouseMode uint
type MouseExtMode uint

const (
	MouseModeNone MouseMode = iota
	MouseModeX10
	MouseModeVT200
	MouseModeVT200Highlight
	MouseModeButtonEvent
	MouseModeAnyEvent
)

const (
	MouseExtNone MouseExtMode = iota
	MouseExtUTF
	MouseExtSGR
	MouseExtURXVT
)
