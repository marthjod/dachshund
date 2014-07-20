package errorcategory

import (
	"github.com/wsxiaoys/terminal"
	"github.com/wsxiaoys/terminal/color"
)

type ErrorCategory struct {
	kind      string
	color     string
	colorEmph string
}

func NewErrorCategory(kind, color, colorEmph string) *ErrorCategory {
	cat := new(ErrorCategory)
	cat.kind = kind
	cat.color = color
	cat.colorEmph = colorEmph
	return cat
}

func (e *ErrorCategory) MarkError(expr string) {
	color.Print(e.colorEmph + expr)
	terminal.Stdout.Reset()
}

func (e *ErrorCategory) MarkHint(expr string) {
	color.Printf(e.color + expr)
	terminal.Stdout.Reset()
}
