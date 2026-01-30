// Package views provides TUI view components.
package views

import (
	"strings"

	"cooperations/internal/tui/styles"
	"cooperations/internal/tui/widgets"
	"github.com/charmbracelet/lipgloss"
)

// DashboardView is the main dashboard view with three panels.
type DashboardView struct {
	Width  int
	Height int

	// Left panel - workflow and agents
	WorkflowSteps *widgets.WorkflowSteps
	AgentPanel    *widgets.AgentPanel

	// Center panel - streaming content
	StreamingText *widgets.StreamingText
	CodeBlock     *widgets.CodeBlock
	DiffBlock     *widgets.DiffBlock

	// Right panel - activity and files
	ActivityLog *widgets.ActivityLog
	FileTree    *widgets.FileTree
	Metrics     *widgets.MetricsPanel

	// Footer
	ToastStack    *widgets.ToastStack
	ProgressBar   *widgets.ProgressBar
	CostTracker   *widgets.CostTracker

	// State
	ActivePanel  int // 0=left, 1=center, 2=right
	CenterMode   int // 0=streaming, 1=code, 2=diff
	RightMode    int // 0=activity, 1=files, 2=metrics
	ShowProgress bool
}

// NewDashboardView creates a new dashboard view.
func NewDashboardView(width, height int) *DashboardView {
	// Calculate panel widths (25% - 50% - 25%)
	leftWidth := width / 4
	rightWidth := width / 4
	centerWidth := width - leftWidth - rightWidth

	// Calculate heights (subtract header and footer)
	contentHeight := height - 4 // header + footer

	// Create widgets
	workflowSteps := widgets.NewWorkflowSteps(leftWidth - 2)
	agentPanel := widgets.NewAgentPanel(leftWidth-2, contentHeight/2, 1)
	agentPanel.InitAgents()

	streamingText := widgets.NewStreamingText(centerWidth-4, contentHeight-2)
	codeBlock := widgets.NewCodeBlock(centerWidth-4, contentHeight-2)
	diffBlock := widgets.NewDiffBlock(centerWidth-4, contentHeight-2)

	activityLog := widgets.NewActivityLog(rightWidth-4, contentHeight/2)
	fileTree := widgets.NewFileTree(rightWidth-4, contentHeight/2)
	metricsPanel := widgets.NewMetricsPanel(rightWidth-4, 1)

	toastStack := widgets.NewToastStack(5, width-4)
	progressBar := widgets.NewProgressBar(width - 20)
	costTracker := widgets.NewCostTracker(20)

	return &DashboardView{
		Width:         width,
		Height:        height,
		WorkflowSteps: &workflowSteps,
		AgentPanel:    &agentPanel,
		StreamingText: &streamingText,
		CodeBlock:     &codeBlock,
		DiffBlock:     &diffBlock,
		ActivityLog:   &activityLog,
		FileTree:      &fileTree,
		Metrics:       &metricsPanel,
		ToastStack:    &toastStack,
		ProgressBar:   &progressBar,
		CostTracker:   &costTracker,
	}
}

// Resize adjusts the view to new dimensions.
func (d *DashboardView) Resize(width, height int) {
	d.Width = width
	d.Height = height

	// Recalculate panel dimensions
	leftWidth := width / 4
	rightWidth := width / 4
	centerWidth := width - leftWidth - rightWidth
	contentHeight := height - 4

	// Update widget dimensions
	d.WorkflowSteps.Width = leftWidth - 2
	d.AgentPanel.Width = leftWidth - 2
	d.AgentPanel.Height = contentHeight / 2

	d.StreamingText.Width = centerWidth - 4
	d.StreamingText.Height = contentHeight - 2
	d.CodeBlock.Width = centerWidth - 4
	d.CodeBlock.Height = contentHeight - 2
	d.DiffBlock.Width = centerWidth - 4
	d.DiffBlock.Height = contentHeight - 2

	d.ActivityLog.Width = rightWidth - 4
	d.ActivityLog.Height = contentHeight / 2
	d.FileTree.Width = rightWidth - 4
	d.FileTree.Height = contentHeight / 2
	d.Metrics.Width = rightWidth - 4

	d.ProgressBar.Width = width - 20
}

// FocusLeft focuses the left panel.
func (d *DashboardView) FocusLeft() {
	d.ActivePanel = 0
}

// FocusCenter focuses the center panel.
func (d *DashboardView) FocusCenter() {
	d.ActivePanel = 1
}

// FocusRight focuses the right panel.
func (d *DashboardView) FocusRight() {
	d.ActivePanel = 2
}

// CycleCenter cycles through center panel modes.
func (d *DashboardView) CycleCenter() {
	d.CenterMode = (d.CenterMode + 1) % 3
}

// CycleRight cycles through right panel modes.
func (d *DashboardView) CycleRight() {
	d.RightMode = (d.RightMode + 1) % 3
}

// View renders the dashboard view.
func (d *DashboardView) View() string {
	var result strings.Builder

	// Header
	result.WriteString(d.renderHeader())
	result.WriteString("\n")

	// Main content (3 panels)
	leftPanel := d.renderLeftPanel()
	centerPanel := d.renderCenterPanel()
	rightPanel := d.renderRightPanel()

	content := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, centerPanel, rightPanel)
	result.WriteString(content)
	result.WriteString("\n")

	// Footer
	result.WriteString(d.renderFooter())

	// Toasts overlay
	if d.ToastStack.Count() > 0 {
		// Position toasts in top-right
		toasts := d.ToastStack.View()
		result.WriteString("\n")
		result.WriteString(toasts)
	}

	return result.String()
}

// renderHeader renders the header bar.
func (d *DashboardView) renderHeader() string {
	titleStyle := styles.TitleStyle
	title := titleStyle.Render("🤖 COOPERATIONS")

	// Agent badges
	badgeBar := widgets.NewAgentBadgeBar()
	for _, role := range d.AgentPanel.ActiveAgents() {
		badgeBar.SetActive(role, true)
	}

	padding := d.Width - lipgloss.Width(title) - lipgloss.Width(badgeBar.View()) - 4
	if padding < 0 {
		padding = 0
	}

	return title + strings.Repeat(" ", padding) + badgeBar.View()
}

// renderLeftPanel renders the left panel.
func (d *DashboardView) renderLeftPanel() string {
	width := d.Width / 4
	height := d.Height - 4

	var panelStyle lipgloss.Style
	if d.ActivePanel == 0 {
		panelStyle = styles.ActivePanelStyle
	} else {
		panelStyle = styles.PanelStyle
	}
	panelStyle = panelStyle.Width(width - 2).Height(height)

	// Workflow section
	workflowHeader := styles.SubHeaderStyle.Render("Workflow")
	workflowContent := d.WorkflowSteps.View()

	// Agent section
	agentHeader := styles.SubHeaderStyle.Render("Agents")
	agentContent := d.AgentPanel.View()

	content := workflowHeader + "\n" + workflowContent + "\n\n" + agentHeader + "\n" + agentContent

	return panelStyle.Render(content)
}

// renderCenterPanel renders the center panel.
func (d *DashboardView) renderCenterPanel() string {
	width := d.Width - (d.Width/4)*2
	height := d.Height - 4

	var panelStyle lipgloss.Style
	if d.ActivePanel == 1 {
		panelStyle = styles.ActivePanelStyle
	} else {
		panelStyle = styles.PanelStyle
	}
	panelStyle = panelStyle.Width(width - 2).Height(height)

	var header string
	var content string

	switch d.CenterMode {
	case 0:
		header = styles.SubHeaderStyle.Render("Response")
		content = d.StreamingText.View()
	case 1:
		header = styles.SubHeaderStyle.Render("Code")
		content = d.CodeBlock.View()
	case 2:
		header = styles.SubHeaderStyle.Render("Diff")
		content = d.DiffBlock.View()
	}

	// Mode indicator
	modes := []string{"Response", "Code", "Diff"}
	var modeIndicators []string
	for i, mode := range modes {
		if i == d.CenterMode {
			modeIndicators = append(modeIndicators, styles.PrimaryStyle.Render("["+mode+"]"))
		} else {
			modeIndicators = append(modeIndicators, styles.MutedStyle.Render(mode))
		}
	}
	modeBar := strings.Join(modeIndicators, " ")

	return panelStyle.Render(header + " " + modeBar + "\n" + content)
}

// renderRightPanel renders the right panel.
func (d *DashboardView) renderRightPanel() string {
	width := d.Width / 4
	height := d.Height - 4

	var panelStyle lipgloss.Style
	if d.ActivePanel == 2 {
		panelStyle = styles.ActivePanelStyle
	} else {
		panelStyle = styles.PanelStyle
	}
	panelStyle = panelStyle.Width(width - 2).Height(height)

	var header string
	var content string

	switch d.RightMode {
	case 0:
		header = styles.SubHeaderStyle.Render("Activity")
		content = d.ActivityLog.View()
	case 1:
		header = styles.SubHeaderStyle.Render("Files")
		content = d.FileTree.View()
	case 2:
		header = styles.SubHeaderStyle.Render("Metrics")
		content = d.Metrics.View()
	}

	// Mode indicator
	modes := []string{"Activity", "Files", "Metrics"}
	var modeIndicators []string
	for i, mode := range modes {
		if i == d.RightMode {
			modeIndicators = append(modeIndicators, styles.PrimaryStyle.Render("["+mode+"]"))
		} else {
			modeIndicators = append(modeIndicators, styles.MutedStyle.Render(mode))
		}
	}
	modeBar := strings.Join(modeIndicators, " ")

	return panelStyle.Render(header + " " + modeBar + "\n" + content)
}

// renderFooter renders the footer bar.
func (d *DashboardView) renderFooter() string {
	// Progress bar on the left
	var progress string
	if d.ShowProgress {
		progress = d.ProgressBar.View()
	}

	// Cost tracker on the right
	cost := d.CostTracker.View()

	// Help hint in the center
	helpHint := styles.MutedStyle.Render("Press ? for help")

	padding := d.Width - lipgloss.Width(progress) - lipgloss.Width(cost) - lipgloss.Width(helpHint) - 4
	if padding < 0 {
		padding = 0
	}

	leftPad := padding / 2
	rightPad := padding - leftPad

	return progress + strings.Repeat(" ", leftPad) + helpHint + strings.Repeat(" ", rightPad) + cost
}
