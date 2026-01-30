// internal/gui/widgets/main_panel.go
package widgets

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// MainPanel displays the activity log and code content.
type MainPanel struct {
	ActivityLog []string
	CodeContent string
	CodeLang    string

	activityList widget.List
	codePanel    *CodePanel
}

// NewMainPanel creates a new main panel with initialized widgets.
func NewMainPanel() *MainPanel {
	return &MainPanel{
		activityList: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		codePanel: NewCodePanel(),
	}
}

// Layout renders the main panel with activity log and code display.
func (mp *MainPanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Panel background
	panelBg := color.NRGBA{R: 0x0a, G: 0x0e, B: 0x17, A: 0xFF}

	// Draw background
	size := gtx.Constraints.Max
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, panelBg)

	// Content with padding
	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		// Decide layout based on whether we have code to show
		hasCode := mp.CodeContent != ""

		if hasCode {
			// Split view: activity log on top, code below
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				// Activity log section (40%)
				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					return mp.layoutActivitySection(gtx, th)
				}),

				// Separator
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return mp.layoutSeparator(gtx)
				}),

				// Code section (60%)
				layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
					return mp.layoutCodeSection(gtx, th)
				}),
			)
		}

		// No code - just show activity log full height
		return mp.layoutActivitySection(gtx, th)
	})
}

func (mp *MainPanel) layoutActivitySection(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Section header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Caption(th, "ACTIVITY")
			lbl.Color = color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF}
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

		// Activity log list
		layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
			if len(mp.ActivityLog) == 0 {
				lbl := material.Body2(th, "Waiting for activity...")
				lbl.Color = color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF}
				return lbl.Layout(gtx)
			}

			return material.List(th, &mp.activityList).Layout(gtx, len(mp.ActivityLog), func(gtx layout.Context, i int) layout.Dimensions {
				// Show newest entries at the bottom
				entry := mp.ActivityLog[i]
				return mp.layoutActivityEntry(gtx, th, entry, i)
			})
		}),
	)
}

func (mp *MainPanel) layoutActivityEntry(gtx layout.Context, th *material.Theme, entry string, idx int) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Start}.Layout(gtx,
			// Entry number indicator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				// Small dot
				dotSize := gtx.Dp(unit.Dp(4))
				defer clip.Ellipse{
					Min: image.Pt(0, gtx.Dp(unit.Dp(6))),
					Max: image.Pt(dotSize, gtx.Dp(unit.Dp(6))+dotSize),
				}.Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0x88})
				return layout.Dimensions{Size: image.Pt(dotSize, gtx.Dp(unit.Dp(16)))}
			}),
			layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),

			// Entry text
			layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Body2(th, entry)
				lbl.Color = color.NRGBA{R: 0xCC, G: 0xDD, B: 0xEE, A: 0xFF}
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (mp *MainPanel) layoutSeparator(gtx layout.Context) layout.Dimensions {
	return layout.Inset{Top: unit.Dp(8), Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		height := gtx.Dp(unit.Dp(1))
		width := gtx.Constraints.Max.X

		defer clip.Rect{Max: image.Pt(width, height)}.Push(gtx.Ops).Pop()
		paint.Fill(gtx.Ops, color.NRGBA{R: 0x1a, G: 0x3a, B: 0x4a, A: 0xFF})

		return layout.Dimensions{Size: image.Pt(width, height)}
	})
}

func (mp *MainPanel) layoutCodeSection(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Section header
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			title := "CODE"
			if mp.CodeLang != "" {
				title = "CODE (" + mp.CodeLang + ")"
			}
			lbl := material.Caption(th, title)
			lbl.Color = color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF}
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

		// Code panel
		layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
			mp.codePanel.Code = mp.CodeContent
			mp.codePanel.Language = mp.CodeLang
			return mp.codePanel.Layout(gtx, th)
		}),
	)
}

// SetActivityLog updates the activity log entries.
func (mp *MainPanel) SetActivityLog(log []string) {
	mp.ActivityLog = log
	// Auto-scroll to bottom
	if len(log) > 0 {
		mp.activityList.Position.First = len(log) - 1
		mp.activityList.Position.Offset = 0
	}
}

// SetCode updates the code display content.
func (mp *MainPanel) SetCode(content, lang string) {
	mp.CodeContent = content
	mp.CodeLang = lang
}

// AppendActivity adds a new activity entry and scrolls to show it.
func (mp *MainPanel) AppendActivity(entry string) {
	mp.ActivityLog = append(mp.ActivityLog, entry)
	// Auto-scroll to bottom
	mp.activityList.Position.First = len(mp.ActivityLog) - 1
	mp.activityList.Position.Offset = 0
}
