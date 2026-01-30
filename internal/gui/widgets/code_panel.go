// internal/gui/widgets/code_panel.go
package widgets

import (
	"fmt"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

type CodePanel struct {
	Code     string
	Language string

	list widget.List

	// cached derived state to avoid re-lexing on every frame
	lastCode string
	lastLang string
	lines    []codeLine
	errLine  string

	// cached font face
	face font.Font
}

type codeRun struct {
	Text string
	Col  color.NRGBA
}

type codeLine struct {
	Runs []codeRun
}

// NewCodePanel returns a CodePanel with sensible defaults.
func NewCodePanel() *CodePanel {
	cp := &CodePanel{
		list: widget.List{
			List: layout.List{
				Axis: layout.Vertical,
			},
		},
		face: font.Font{Typeface: "monospace"},
	}
	return cp
}

func (cp *CodePanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	cp.ensureParsed()

	bg := color.NRGBA{R: 0x10, G: 0x12, B: 0x17, A: 0xFF}
	defer clip.Rect{Max: gtx.Constraints.Max}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, bg)

	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		if cp.errLine != "" {
			lbl := material.Body2(th, cp.errLine)
			lbl.Color = color.NRGBA{R: 0xFF, G: 0x66, B: 0x66, A: 0xFF}
			lbl.Font = cp.face
			lbl.Alignment = text.Start
			return lbl.Layout(gtx)
		}

		return material.List(th, &cp.list).Layout(gtx, len(cp.lines), func(gtx layout.Context, i int) layout.Dimensions {
			ln := cp.lines[i]
			return layout.Flex{Axis: layout.Horizontal}.Layout(gtx, cp.runFlexChildren(th, ln)...)
		})
	})
}

func (cp *CodePanel) runFlexChildren(th *material.Theme, ln codeLine) []layout.FlexChild {
	if len(ln.Runs) == 0 {
		return []layout.FlexChild{
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(th, " ")
				lbl.Font = cp.face
				lbl.Color = color.NRGBA{R: 0xC9, G: 0xD1, B: 0xD9, A: 0xFF}
				lbl.Alignment = text.Start
				lbl.MaxLines = 1
				// lbl.WrapPolicy removed - not available in this Gio version
				return lbl.Layout(gtx)
			}),
		}
	}

	children := make([]layout.FlexChild, 0, len(ln.Runs))
	for _, r := range ln.Runs {
		run := r
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Body2(th, run.Text)
			lbl.Font = cp.face
			lbl.Color = run.Col
			lbl.Alignment = text.Start
			lbl.MaxLines = 1
			// lbl.WrapPolicy removed - not available in this Gio version
			return lbl.Layout(gtx)
		}))
	}
	return children
}

func (cp *CodePanel) ensureParsed() {
	if cp.lastCode == cp.Code && cp.lastLang == cp.Language && cp.lines != nil {
		return
	}
	cp.lastCode = cp.Code
	cp.lastLang = cp.Language
	cp.errLine = ""
	cp.lines = nil

	lines, err := highlightToLines(cp.Code, cp.Language)
	if err != nil {
		cp.errLine = fmt.Sprintf("highlight error: %v", err)
		return
	}
	cp.lines = lines
}

func highlightToLines(code, lang string) ([]codeLine, error) {
	var lx chroma.Lexer
	lang = strings.TrimSpace(lang)
	if lang != "" {
		lx = lexers.Get(lang)
	}
	if lx == nil {
		lx = lexers.Analyse(code)
	}
	if lx == nil {
		lx = lexers.Fallback
	}
	lx = chroma.Coalesce(lx)

	it, err := lx.Tokenise(nil, code)
	if err != nil {
		return nil, err
	}

	st := styles.Get("github-dark")
	if st == nil {
		st = styles.Fallback
	}

	defaultCol := color.NRGBA{R: 0xC9, G: 0xD1, B: 0xD9, A: 0xFF}

	var lines []codeLine
	cur := codeLine{}

	flushLine := func() {
		lines = append(lines, cur)
		cur = codeLine{}
	}

	addRun := func(text string, col color.NRGBA) {
		if text == "" {
			return
		}
		n := len(cur.Runs)
		if n > 0 && cur.Runs[n-1].Col == col {
			cur.Runs[n-1].Text += text
			return
		}
		cur.Runs = append(cur.Runs, codeRun{Text: text, Col: col})
	}

	for tok := it(); tok != chroma.EOF; tok = it() {
		entry := st.Get(tok.Type)
		col := defaultCol
		if entry.Colour.IsSet() {
			col = chromaColourToNRGBA(entry.Colour)
		}

		val := tok.Value
		for {
			idx := strings.IndexByte(val, '\n')
			if idx < 0 {
				addRun(val, col)
				break
			}
			addRun(val[:idx], col)
			flushLine()
			val = val[idx+1:]
			if val == "" {
				break
			}
		}
	}

	flushLine()

	if code == "" && len(lines) == 1 && len(lines[0].Runs) == 0 {
		return []codeLine{{Runs: nil}}, nil
	}
	return lines, nil
}

func chromaColourToNRGBA(c chroma.Colour) color.NRGBA {
	// chroma.Colour is uint32 with RGBA packed
	// If brightness is 0, the color might be unset, use full alpha
	return color.NRGBA{R: c.Red(), G: c.Green(), B: c.Blue(), A: 0xFF}
}