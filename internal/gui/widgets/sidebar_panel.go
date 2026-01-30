// internal/gui/widgets/sidebar_panel.go
package widgets

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

// WorkflowStep represents a single step in the workflow for display.
type WorkflowStep struct {
	ID       string
	Label    string
	Status   string  // "pending", "inprogress", "complete", "waiting"
	Progress float32 // 0.0 to 1.0
	Subtext  string
}

// HandoffEntry represents a role handoff for display.
type HandoffEntry struct {
	FromRole  string
	ToRole    string
	Timestamp time.Time
}

// SidebarPanel displays workflow steps and handoff history.
type SidebarPanel struct {
	Steps          []WorkflowStep
	HandoffHistory []HandoffEntry
	CurrentStep    int

	stepsList   widget.List
	handoffList widget.List

	// Progress widgets for each step (reused)
	stepProgress []NeonProgress
}

// NewSidebarPanel creates a new sidebar panel.
func NewSidebarPanel() *SidebarPanel {
	return &SidebarPanel{
		stepsList: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
		handoffList: widget.List{
			List: layout.List{Axis: layout.Vertical},
		},
	}
}

// Layout renders the sidebar panel.
func (sp *SidebarPanel) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Panel background
	panelBg := color.NRGBA{R: 0x0d, G: 0x15, B: 0x20, A: 0xFF}
	borderColor := color.NRGBA{R: 0x1a, G: 0x3a, B: 0x4a, A: 0xFF}

	// Draw background
	size := gtx.Constraints.Max
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, panelBg)

	// Draw right border
	borderWidth := gtx.Dp(unit.Dp(1))
	borderRect := image.Rect(size.X-borderWidth, 0, size.X, size.Y)
	st := clip.Rect(borderRect).Push(gtx.Ops)
	paint.Fill(gtx.Ops, borderColor)
	st.Pop()

	// Ensure we have enough progress widgets
	sp.ensureProgressWidgets(len(sp.Steps))

	// Content
	inset := layout.UniformInset(unit.Dp(12))
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Section header: Workflow
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return sp.sectionHeader(gtx, th, "WORKFLOW")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

			// Workflow steps list (takes 60% of space)
			layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
				return sp.layoutSteps(gtx, th)
			}),

			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),

			// Section header: Handoffs
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return sp.sectionHeader(gtx, th, "HANDOFFS")
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),

			// Handoff history list (takes remaining space)
			layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
				return sp.layoutHandoffs(gtx, th)
			}),
		)
	})
}

func (sp *SidebarPanel) sectionHeader(gtx layout.Context, th *material.Theme, title string) layout.Dimensions {
	lbl := material.Caption(th, title)
	lbl.Color = color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF} // Cyan
	return lbl.Layout(gtx)
}

func (sp *SidebarPanel) layoutSteps(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(sp.Steps) == 0 {
		lbl := material.Body2(th, "No workflow steps")
		lbl.Color = color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF}
		return lbl.Layout(gtx)
	}

	return material.List(th, &sp.stepsList).Layout(gtx, len(sp.Steps), func(gtx layout.Context, i int) layout.Dimensions {
		step := sp.Steps[i]
		isCurrent := i == sp.CurrentStep

		return sp.layoutStep(gtx, th, step, isCurrent, i)
	})
}

func (sp *SidebarPanel) layoutStep(gtx layout.Context, th *material.Theme, step WorkflowStep, isCurrent bool, idx int) layout.Dimensions {
	// Step container with padding
	return layout.Inset{Bottom: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Step label with status indicator
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal, Alignment: layout.Middle}.Layout(gtx,
					// Status indicator dot
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return sp.statusDot(gtx, step.Status, isCurrent)
					}),
					layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
					// Label
					layout.Flexed(1.0, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Body2(th, step.Label)
						if isCurrent {
							lbl.Color = color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}
						} else {
							lbl.Color = color.NRGBA{R: 0x88, G: 0x99, B: 0xAA, A: 0xFF}
						}
						return lbl.Layout(gtx)
					}),
				)
			}),

			// Progress bar for in-progress steps
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if step.Status != "inprogress" || step.Progress <= 0 {
					return layout.Dimensions{}
				}
				return layout.Inset{Left: unit.Dp(16), Top: unit.Dp(4)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					sp.stepProgress[idx].Progress = step.Progress
					sp.stepProgress[idx].Glow = true
					return sp.stepProgress[idx].Layout(gtx, unit.Dp(4))
				})
			}),

			// Subtext
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if step.Subtext == "" {
					return layout.Dimensions{}
				}
				return layout.Inset{Left: unit.Dp(16), Top: unit.Dp(2)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Caption(th, step.Subtext)
					lbl.Color = color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF}
					return lbl.Layout(gtx)
				})
			}),
		)
	})
}

func (sp *SidebarPanel) statusDot(gtx layout.Context, status string, isCurrent bool) layout.Dimensions {
	dotSize := gtx.Dp(unit.Dp(8))
	size := image.Pt(dotSize, dotSize)

	var dotColor color.NRGBA
	switch status {
	case "complete":
		dotColor = color.NRGBA{R: 0x00, G: 0xFF, B: 0x88, A: 0xFF} // Green
	case "inprogress":
		dotColor = color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF} // Cyan
	case "waiting":
		dotColor = color.NRGBA{R: 0xFF, G: 0xAA, B: 0x00, A: 0xFF} // Orange
	default: // pending
		dotColor = color.NRGBA{R: 0x44, G: 0x55, B: 0x66, A: 0xFF} // Gray
	}

	if isCurrent && status != "complete" {
		dotColor = color.NRGBA{R: 0x00, G: 0xFF, B: 0xFF, A: 0xFF} // Cyan for current
	}

	// Draw circle
	radius := dotSize / 2
	center := image.Pt(radius, radius)
	defer clip.Ellipse{
		Min: image.Pt(center.X-radius, center.Y-radius),
		Max: image.Pt(center.X+radius, center.Y+radius),
	}.Push(gtx.Ops).Pop()
	paint.Fill(gtx.Ops, dotColor)

	return layout.Dimensions{Size: size}
}

func (sp *SidebarPanel) layoutHandoffs(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(sp.HandoffHistory) == 0 {
		lbl := material.Body2(th, "No handoffs yet")
		lbl.Color = color.NRGBA{R: 0x66, G: 0x77, B: 0x88, A: 0xFF}
		return lbl.Layout(gtx)
	}

	// Show most recent handoffs (reversed order)
	return material.List(th, &sp.handoffList).Layout(gtx, len(sp.HandoffHistory), func(gtx layout.Context, i int) layout.Dimensions {
		// Reverse index to show newest first
		idx := len(sp.HandoffHistory) - 1 - i
		handoff := sp.HandoffHistory[idx]
		return sp.layoutHandoff(gtx, th, handoff)
	})
}

func (sp *SidebarPanel) layoutHandoff(gtx layout.Context, th *material.Theme, h HandoffEntry) layout.Dimensions {
	return layout.Inset{Bottom: unit.Dp(6)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// From -> To
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				text := fmt.Sprintf("%s → %s", h.FromRole, h.ToRole)
				lbl := material.Body2(th, text)
				lbl.Color = color.NRGBA{R: 0xCC, G: 0xDD, B: 0xEE, A: 0xFF}
				return lbl.Layout(gtx)
			}),
			// Timestamp
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				timeStr := h.Timestamp.Format("15:04:05")
				lbl := material.Caption(th, timeStr)
				lbl.Color = color.NRGBA{R: 0x55, G: 0x66, B: 0x77, A: 0xFF}
				return lbl.Layout(gtx)
			}),
		)
	})
}

func (sp *SidebarPanel) ensureProgressWidgets(n int) {
	if len(sp.stepProgress) >= n {
		return
	}
	for i := len(sp.stepProgress); i < n; i++ {
		sp.stepProgress = append(sp.stepProgress, NeonProgress{})
	}
}

// SetSteps updates the workflow steps display.
func (sp *SidebarPanel) SetSteps(steps []WorkflowStep) {
	sp.Steps = steps
}

// SetHandoffs updates the handoff history display.
func (sp *SidebarPanel) SetHandoffs(history []HandoffEntry) {
	sp.HandoffHistory = history
}

// SetCurrentStep sets the currently active step index.
func (sp *SidebarPanel) SetCurrentStep(idx int) {
	sp.CurrentStep = idx
}
