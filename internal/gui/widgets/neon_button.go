// internal/gui/widgets/neon_button.go
package widgets

import (
	"image"
	"image/color"
	"strings"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// NeonButton is a clickable button with a neon border and subtle glow background.
// It tracks hover state to brighten the border.
type NeonButton struct {
	Text    string
	Color   color.NRGBA
	OnClick func()

	Clickable widget.Clickable
}

func (b *NeonButton) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if th == nil {
		panic("widgets.NeonButton: nil theme")
	}

	for b.Clickable.Clicked(gtx) {
		if b.OnClick != nil {
			b.OnClick()
		}
	}

	// Use Clickable's built-in hover detection
	hovered := b.Clickable.Hovered()

	const (
		cornerRadiusDp = unit.Dp(12)
		borderWidthDp  = unit.Dp(2)
		hPadDp         = unit.Dp(16)
		vPadDp         = unit.Dp(10)
		minHeightDp    = unit.Dp(40)
	)

	r := gtx.Dp(cornerRadiusDp)
	bw := gtx.Dp(borderWidthDp)
	hPad := gtx.Dp(hPadDp)
	vPad := gtx.Dp(vPadDp)

	// Prepare label.
	lbl := material.Label(th, unit.Sp(16), strings.TrimSpace(b.Text))
	lbl.Alignment = text.Middle
	lbl.Font.Weight = font.Medium
	lbl.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}

	// Measure label with a recording.
	rec := op.Record(gtx.Ops)
	labelDims := lbl.Layout(gtx)
	labelCall := rec.Stop()

	minW := labelDims.Size.X + 2*hPad
	minH := labelDims.Size.Y + 2*vPad
	if mh := gtx.Dp(minHeightDp); minH < mh {
		minH = mh
	}

	// Avoid mutating gtx.Constraints in-place.
	gtxBtn := gtx
	gtxBtn.Constraints.Min = image.Pt(minW, minH)
	size := gtxBtn.Constraints.Constrain(image.Pt(minW, minH))

	return b.Clickable.Layout(gtxBtn, func(gtx layout.Context) layout.Dimensions {
		// Hover effect: brighten border a bit.
		borderCol := b.Color
		if hovered {
			borderCol = brightenNRGBA(borderCol, 1.25)
		}

		// Background is a dimmed version of the neon color.
		bgCol := brightenNRGBA(b.Color, 0.30)
		bgCol.A = 0x60

		// Draw background.
		{
			stack := clip.UniformRRect(image.Rectangle{Max: size}, r).Push(gtx.Ops)
			paint.Fill(gtx.Ops, bgCol)
			stack.Pop()
		}

		// Draw border (stroke a rounded rect).
		{
			rr := clip.RRect{
				Rect: image.Rectangle{Max: size},
				NE:   r, NW: r, SE: r, SW: r,
			}
			st := clip.Stroke{
				Path:  rr.Path(gtx.Ops),
				Width: float32(bw),
			}
			stack := st.Op().Push(gtx.Ops)
			paint.Fill(gtx.Ops, borderCol)
			stack.Pop()
		}

		// Draw label centered.
		{
			gtx2 := gtx
			gtx2.Constraints.Min = image.Point{}
			gtx2.Constraints.Max = size

			layout.Center.Layout(gtx2, func(gtx layout.Context) layout.Dimensions {
				labelCall.Add(gtx.Ops)
				return labelDims
			})
		}

		return layout.Dimensions{Size: size}
	})
}

func brightenNRGBA(c color.NRGBA, factor float32) color.NRGBA {
	clamp := func(v float32) uint8 {
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		return uint8(v + 0.5)
	}
	return color.NRGBA{
		R: clamp(float32(c.R) * factor),
		G: clamp(float32(c.G) * factor),
		B: clamp(float32(c.B) * factor),
		A: c.A,
	}
}