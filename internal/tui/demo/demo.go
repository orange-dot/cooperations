// Package demo provides a demo mode for the TUI.
package demo

import (
	"time"

	"cooperations/internal/tui/stream"
)

// Run runs a simulated workflow for demo purposes.
func Run(s *stream.WorkflowStream) {
	// Phase 1: Initialization
	s.SendProgress(stream.ProgressUpdate{
		Percent: 0,
		Stage:   "Initialize",
		Message: "Starting workflow...",
	})
	time.Sleep(500 * time.Millisecond)

	// Phase 2: Architect
	s.SendHandoff(stream.HandoffEvent{
		From:   "",
		To:     "architect",
		Reason: "Analyzing task requirements",
	})
	time.Sleep(300 * time.Millisecond)

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "architect",
		Level:     "info",
		Message:   "Reading task description and requirements",
	})
	time.Sleep(200 * time.Millisecond)

	// Stream architect thinking
	architectResponse := []string{
		"# Design Analysis\n\n",
		"## Requirements\n",
		"- ",
		"Implement",
		" a ",
		"modern",
		" TUI",
		" interface\n",
		"- ",
		"Support",
		" real-time",
		" streaming\n",
		"- ",
		"Use",
		" Bubble Tea",
		" framework\n\n",
		"## ",
		"Architecture",
		" Decision\n",
		"The ",
		"system",
		" will",
		" use",
		" an",
		" event-driven",
		" architecture",
		" with",
		" channels",
		" for",
		" communication.\n",
	}

	for _, token := range architectResponse {
		s.SendToken(stream.TokenChunk{
			AgentRole: "architect",
			Token:     token,
		})
		time.Sleep(50 * time.Millisecond)
	}

	s.SendToken(stream.TokenChunk{
		AgentRole: "architect",
		Token:     "",
		IsFinal:   true,
	})

	s.SendProgress(stream.ProgressUpdate{
		Percent: 25,
		Stage:   "Design",
		Message: "Architecture complete",
	})

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "architect",
		Level:     "info",
		Message:   "Design phase completed successfully",
	})

	s.SendMetrics(stream.MetricsSnapshot{
		TotalTokens:      450,
		PromptTokens:     200,
		CompletionTokens: 250,
		EstimatedCostUSD: 0.0067,
		AgentCycles:      1,
		CurrentAgent:     "architect",
	})

	time.Sleep(500 * time.Millisecond)

	// Phase 3: Implementer
	s.SendHandoff(stream.HandoffEvent{
		From:   "architect",
		To:     "implementer",
		Reason: "Design approved, starting implementation",
	})
	time.Sleep(300 * time.Millisecond)

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "implementer",
		Level:     "info",
		Message:   "Generating code based on design specifications",
	})

	// Stream implementer response
	codeContent := `package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	message string
	ready   bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Hello, %s!\n\nPress q to quit.", m.message)
}

func main() {
	p := tea.NewProgram(model{message: "World"})
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
`

	// Stream code character by character with realistic timing
	for i, char := range codeContent {
		s.SendToken(stream.TokenChunk{
			AgentRole: "implementer",
			Token:     string(char),
		})
		// Vary speed: faster for whitespace, slower for important characters
		if char == '\n' || char == '\t' || char == ' ' {
			time.Sleep(10 * time.Millisecond)
		} else {
			time.Sleep(20 * time.Millisecond)
		}

		// Update progress periodically
		if i%100 == 0 {
			progress := 25 + float64(i)/float64(len(codeContent))*40
			s.SendProgress(stream.ProgressUpdate{
				Percent: progress,
				Stage:   "Implement",
				Message: "Generating code...",
			})
		}
	}

	s.SendToken(stream.TokenChunk{
		AgentRole: "implementer",
		Token:     "",
		IsFinal:   true,
	})

	// Send code update
	s.SendCode(stream.CodeUpdate{
		Path:     "main.go",
		Language: "go",
		Content:  codeContent,
	})

	s.SendProgress(stream.ProgressUpdate{
		Percent: 65,
		Stage:   "Implement",
		Message: "Code generation complete",
	})

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "implementer",
		Level:     "info",
		Message:   "Implementation completed, created main.go",
	})

	s.SendMetrics(stream.MetricsSnapshot{
		TotalTokens:      1250,
		PromptTokens:     400,
		CompletionTokens: 850,
		EstimatedCostUSD: 0.0187,
		AgentCycles:      2,
		CurrentAgent:     "implementer",
	})

	time.Sleep(500 * time.Millisecond)

	// Phase 4: Reviewer
	s.SendHandoff(stream.HandoffEvent{
		From:   "implementer",
		To:     "reviewer",
		Reason: "Implementation complete, starting review",
	})
	time.Sleep(300 * time.Millisecond)

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "reviewer",
		Level:     "info",
		Message:   "Reviewing code quality and best practices",
	})

	// Stream reviewer response
	reviewResponse := []string{
		"# Code Review\n\n",
		"## ",
		"Quality",
		" Assessment\n\n",
		"**Overall",
		" Score:**",
		" 8.5/10\n\n",
		"### ",
		"Strengths\n",
		"- ✅ ",
		"Clean",
		" Bubble Tea",
		" implementation\n",
		"- ✅ ",
		"Proper",
		" Init/Update/View",
		" pattern\n",
		"- ✅ ",
		"Graceful",
		" exit",
		" handling\n\n",
		"### ",
		"Suggestions\n",
		"- 💡 ",
		"Consider",
		" adding",
		" error",
		" handling\n",
		"- 💡 ",
		"Add",
		" window",
		" resize",
		" support\n\n",
		"**Verdict:**",
		" APPROVED ✓\n",
	}

	for _, token := range reviewResponse {
		s.SendToken(stream.TokenChunk{
			AgentRole: "reviewer",
			Token:     token,
		})
		time.Sleep(40 * time.Millisecond)
	}

	s.SendToken(stream.TokenChunk{
		AgentRole: "reviewer",
		Token:     "",
		IsFinal:   true,
	})

	s.SendProgress(stream.ProgressUpdate{
		Percent: 90,
		Stage:   "Review",
		Message: "Code review complete",
	})

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "reviewer",
		Level:     "info",
		Message:   "Code approved with minor suggestions",
	})

	s.SendMetrics(stream.MetricsSnapshot{
		TotalTokens:      1650,
		PromptTokens:     600,
		CompletionTokens: 1050,
		EstimatedCostUSD: 0.0247,
		AgentCycles:      3,
		CurrentAgent:     "reviewer",
	})

	time.Sleep(500 * time.Millisecond)

	// Phase 5: Completion
	s.SendProgress(stream.ProgressUpdate{
		Percent: 100,
		Stage:   "Complete",
		Message: "Workflow completed successfully",
	})

	s.SendLog(stream.AgentLogEntry{
		AgentRole: "",
		Level:     "info",
		Message:   "All agents completed their tasks",
	})

	s.SendToast(stream.ToastNotification{
		Level:   "success",
		Title:   "Success",
		Message: "Demo workflow completed successfully!",
	})

	time.Sleep(300 * time.Millisecond)
	s.SignalDone()
}

// RunFast runs a faster version of the demo for testing.
func RunFast(s *stream.WorkflowStream) {
	s.SendProgress(stream.ProgressUpdate{
		Percent: 0,
		Stage:   "Start",
		Message: "Quick demo...",
	})

	agents := []string{"architect", "implementer", "reviewer"}
	for i, agent := range agents {
		s.SendHandoff(stream.HandoffEvent{
			To:     agent,
			Reason: "Processing",
		})
		time.Sleep(100 * time.Millisecond)

		s.SendProgress(stream.ProgressUpdate{
			Percent: float64((i + 1) * 33),
			Stage:   agent,
			Message: "Working...",
		})
		time.Sleep(200 * time.Millisecond)
	}

	s.SendProgress(stream.ProgressUpdate{
		Percent: 100,
		Stage:   "Done",
		Message: "Complete",
	})

	s.SendToast(stream.ToastNotification{
		Level:   "success",
		Message: "Quick demo done!",
	})

	s.SignalDone()
}
