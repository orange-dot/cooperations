// internal/gui/widgets/neon_progress.go
package widgets

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

// NeonProgress is an animated progress bar with an optional neon glow.
// Progress is clamped to [0..1]. When Glow is true, the leading edge pulses.
type NeonProgress struct {
	Progress float32
	Glow     bool

	last  time.Time
	phase float32
}

// SetProgress clamps and sets the progress value.
func (p *NeonProgress) SetProgress(v float32) {
	p.Progress = clamp01(v)
}

// Layout draws the progress bar. It expands to the maximum width, and uses the
// provided height (dp) for thickness.
func (p *NeonProgress) Layout(gtx layout.Context, height unit.Dp) layout.Dimensions {
	// Animate phase when glow is enabled.
	now := time.Now()
	if p.last.IsZero() {
		p.last = now
	}
	if p.Glow {
		dt := float32(now.Sub(p.last).Seconds())
		p.phase += dt * 2.5
		// Keep phase bounded to avoid precision loss over long runtimes.
		if p.phase > 2*math.Pi {
			p.phase = float32(math.Mod(float64(p.phase), 2*math.Pi))
		}
	} else {
		p.phase = 0
	}
	p.last = now

	// Ask for next frame while glowing to keep animation running.
	if p.Glow {
		gtx.Execute(op.InvalidateCmd{})
	}

	h := gtx.Dp(height)
	if h <= 0 {
		h = 1
	}
	w := gtx.Constraints.Max.X
	if w <= 0 {
		w = 1
	}

	prog := clamp01(p.Progress)
	fillW := int(math.Round(float64(float32(w) * prog)))
	if fillW < 0 {
		fillW = 0
	}
	if fillW > w {
		fillW = w
	}

	// Colors tuned for a "neon" cyan look.
	bg := color.NRGBA{R: 12, G: 14, B: 18, A: 255}
	bgEdge := color.NRGBA{R: 22, G: 28, B: 36, A: 255}

	// Fill gradient.
	fillL := color.NRGBA{R: 0, G: 180, B: 255, A: 255}
	fillM := color.NRGBA{R: 0, G: 255, B: 220, A: 255}
	fillR := color.NRGBA{R: 160, G: 255, B: 255, A: 255}

	// Glow colors (more transparent).
	glowInner := color.NRGBA{R: 90, G: 255, B: 255, A: 140}
	glowOuter := color.NRGBA{R: 0, G: 200, B: 255, A: 60}

	radiusF := float32(h) * 0.5
	radius := int(radiusF)

	// Draw background track.
	{
		rr := clip.RRect{
			Rect: image.Rect(0, 0, w, h),
			NE:   radius, NW: radius, SE: radius, SW: radius,
		}
		st := rr.Push(gtx.Ops)

		paint.LinearGradientOp{
			Stop1:  f32.Pt(0, 0),
			Stop2:  f32.Pt(float32(w), 0),
			Color1: bgEdge,
			Color2: bg,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)

		st.Pop()
	}

	// Draw filled portion (clipped to track and progress width).
	if fillW > 0 {
		rr := clip.RRect{
			Rect: image.Rect(0, 0, w, h),
			NE:   radius, NW: radius, SE: radius, SW: radius,
		}
		trackSt := rr.Push(gtx.Ops)

		fillSt := clip.Rect(image.Rect(0, 0, fillW, h)).Push(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Pt(0, 0),
			Stop2:  f32.Pt(float32(w), 0),
			Color1: fillL,
			Color2: fillR,
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		fillSt.Pop()

		// Add a subtle inner highlight band near the top for depth.
		hiH := int(math.Max(1, float64(h/3)))
		hiSt := clip.Rect(image.Rect(0, 0, fillW, hiH)).Push(gtx.Ops)
		paint.LinearGradientOp{
			Stop1:  f32.Pt(0, 0),
			Stop2:  f32.Pt(float32(w), 0),
			Color1: withAlpha(fillM, 110),
			Color2: withAlpha(fillR, 40),
		}.Add(gtx.Ops)
		paint.PaintOp{}.Add(gtx.Ops)
		hiSt.Pop()

		trackSt.Pop()
	}

	// Glow around the leading edge (a couple of soft rectangles to simulate bloom).
	if p.Glow && fillW > 0 {
		// Pulse factor in [0..1].
		pulse := 0.5 + 0.5*float32(math.Sin(float64(p.phase)))
		innerA := uint8(90 + 90*pulse)
		outerA := uint8(30 + 50*pulse)

		edgeX := float32(fillW)
		if edgeX < radiusF {
			edgeX = radiusF
		}
		if edgeX > float32(w)-radiusF {
			edgeX = float32(w) - radiusF
		}

		outerW := float32(h) * 2.6
		innerW := float32(h) * 1.5

		// Outer glow.
		{
			x0 := int(math.Round(float64(edgeX - outerW)))
			x1 := int(math.Round(float64(edgeX + outerW)))
			if x0 < 0 {
				x0 = 0
			}
			if x1 > w {
				x1 = w
			}
			if x1 > x0 {
				st := clip.Rect(image.Rect(x0, 0, x1, h)).Push(gtx.Ops)
				paint.LinearGradientOp{
					Stop1:  f32.Pt(float32(x0), 0),
					Stop2:  f32.Pt(float32(x1), 0),
					Color1: withAlpha(glowOuter, 0),
					Color2: withAlpha(glowOuter, outerA),
				}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				st.Pop()
			}
		}

		// Inner glow.
		{
			x0 := int(math.Round(float64(edgeX - innerW)))
			x1 := int(math.Round(float64(edgeX + innerW)))
			if x0 < 0 {
				x0 = 0
			}
			if x1 > w {
				x1 = w
			}
			if x1 > x0 {
				st := clip.Rect(image.Rect(x0, 0, x1, h)).Push(gtx.Ops)
				paint.LinearGradientOp{
					Stop1:  f32.Pt(float32(x0), 0),
					Stop2:  f32.Pt(float32(x1), 0),
					Color1: withAlpha(glowInner, 0),
					Color2: withAlpha(glowInner, innerA),
				}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				st.Pop()
			}
		}

		// A crisp leading edge highlight.
		{
			lineW := int(math.Max(1, float64(h/8)))
			x0 := fillW - lineW
			x1 := fillW + lineW
			if x0 < 0 {
				x0 = 0
			}
			if x1 > w {
				x1 = w
			}
			if x1 > x0 {
				st := clip.Rect(image.Rect(x0, 0, x1, h)).Push(gtx.Ops)
				white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
				paint.LinearGradientOp{
					Stop1:  f32.Pt(float32(x0), 0),
					Stop2:  f32.Pt(float32(x1), 0),
					Color1: withAlpha(white, 0),
					Color2: withAlpha(white, 160),
				}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				st.Pop()
			}
		}
	}

	return layout.Dimensions{Size: image.Pt(w, h)}
}

func clamp01(v float32) float32 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func withAlpha(c color.NRGBA, a uint8) color.NRGBA {
	c.A = a
	return c
}