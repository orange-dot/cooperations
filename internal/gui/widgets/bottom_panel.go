// internal/gui/widgets/bottom_panel.go
package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// BottomPanel displays decision prompts with action buttons and an optional text input.
// It is only visible when Visible is true (typically when WaitingForInput).
type BottomPanel struct {
	Title   string
	Prompt  string
	Options []string
	Visible bool

	// Callbacks for button actions
	OnApprove func()
	OnReject  func()
	OnEdit    func(comment string)

	// Internal widgets
	approveBtn NeonButton
	rejectBtn  NeonButton
	editBtn    NeonButton
	editor     widget.Editor
}

// NewBottomPanel creates a BottomPanel with default button configuration.
func NewBottomPanel() *BottomPanel {
	bp := &BottomPanel{
		editor: widget.Editor{
			SingleLine: false,
			Submit:     true,
		},
	}

	// Configure buttons with neon colors
	bp.approveBtn = NeonButton{
		Text:  "Approve",
		Color: color.NRGBA{R: 0x00, G: 0xFF, B: 0x88, A: 0xFF}, // Success green
	}
	bp.rejectBtn = NeonButton{
		Text:  "Reject",
		Color: color.NRGBA{R: 0xFF, G: 0x44, B: 0x66, A: 0xFF}, // Error red
	}
	bp.editBtn = NeonButton{
		Text:  "Edit",
		Color: color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF}, // Cyan
	}

	return bp
}

// Layout renders the bottom panel. Returns zero dimensions if not visible.
func (bp *BottomPanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if !bp.Visible {
		return layout.Dimensions{}
	}

	// Wire up button callbacks
	bp.approveBtn.OnClick = bp.OnApprove
	bp.rejectBtn.OnClick = bp.OnReject
	bp.editBtn.OnClick = func() {
		if bp.OnEdit != nil {
			bp.OnEdit(bp.editor.Text())
		}
	}

	// Panel background color (dark with slight transparency)
	panelBg := color.NRGBA{R: 0x0d, G: 0x15, B: 0x20, A: 0xFF}
	borderColor := color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF}

	// Fixed height for the panel
	panelHeight := gtx.Dp(unit.Dp(160))
	size := image.Pt(gtx.Constraints.Max.X, panelHeight)

	// Draw background
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, panelBg)

	// Draw top border
	borderRect := image.Rect(0, 0, size.X, gtx.Dp(unit.Dp(2)))
	st := clip.Rect(borderRect).Push(gtx.Ops)
	paint.Fill(gtx.Ops, borderColor)
	st.Pop()

	// Content with padding
	inset := layout.Inset{
		Top:    unit.Dp(16),
		Bottom: unit.Dp(12),
		Left:   unit.Dp(20),
		Right:  unit.Dp(20),
	}

	inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title and prompt
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if bp.Title == "" {
					return layout.Dimensions{}
				}
				lbl := material.H6(th, bp.Title)
				lbl.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if bp.Prompt == "" {
					return layout.Dimensions{}
				}
				lbl := material.Body2(th, bp.Prompt)
				lbl.Color = color.NRGBA{R: 0x88, G: 0x99, B: 0xAA, A: 0xFF}
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),

			// Buttons row
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Spacing: layout.SpaceStart}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return bp.approveBtn.Layout(gtx, th)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return bp.rejectBtn.Layout(gtx, th)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(12)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return bp.editBtn.Layout(gtx, th)
					}),
					layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
						// Text input area
						return bp.layoutEditor(gtx, th)
					}),
				)
			}),
		)
	})

	return layout.Dimensions{Size: size}
}

func (bp *BottomPanel) layoutEditor(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Editor background
	editorBg := color.NRGBA{R: 0x0a, G: 0x0e, B: 0x17, A: 0xFF}
	editorBorder := color.NRGBA{R: 0x1a, G: 0x3a, B: 0x4a, A: 0xFF}

	return layout.Inset{Left: unit.Dp(16)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Draw editor background
		size := gtx.Constraints.Max
		if size.Y > gtx.Dp(unit.Dp(40)) {
			size.Y = gtx.Dp(unit.Dp(40))
		}

		// Background
		defer clip.RRect{
			Rect: image.Rectangle{Max: size},
			NE:   gtx.Dp(unit.Dp(6)),
			NW:   gtx.Dp(unit.Dp(6)),
			SE:   gtx.Dp(unit.Dp(6)),
			SW:   gtx.Dp(unit.Dp(6)),
		}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, editorBg)

		// Border
		borderWidth := gtx.Dp(unit.Dp(1))
		rr := clip.RRect{
			Rect: image.Rectangle{Max: size},
			NE:   gtx.Dp(unit.Dp(6)),
			NW:   gtx.Dp(unit.Dp(6)),
			SE:   gtx.Dp(unit.Dp(6)),
			SW:   gtx.Dp(unit.Dp(6)),
		}
		st := clip.Stroke{
			Path:  rr.Path(gtx.Ops),
			Width: float32(borderWidth),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, editorBorder)
		st.Pop()

		// Editor content with padding
		return layout.Inset{
			Top:    unit.Dp(8),
			Bottom: unit.Dp(8),
			Left:   unit.Dp(12),
			Right:  unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			ed := material.Editor(th, &bp.editor, "Add comment...")
			ed.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
			ed.HintColor = color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF}
			ed.TextSize = unit.Sp(14)
			ed.Editor.Alignment = text.Start
			return ed.Layout(gtx)
		})
	})
}

// SetText sets the editor text content.
func (bp *BottomPanel) SetText(text string) {
	bp.editor.SetText(text)
}

// Text returns the current editor text.
func (bp *BottomPanel) Text() string {
	return bp.editor.Text()
}
