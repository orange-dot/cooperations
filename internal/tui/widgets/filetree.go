// Package widgets provides TUI components.
package widgets

import (
	"path/filepath"
	"sort"
	"strings"

	"cooperations/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// FileStatus represents the status of a file.
type FileStatus int

const (
	FileStatusNone FileStatus = iota
	FileStatusModified
	FileStatusAdded
	FileStatusDeleted
	FileStatusRenamed
)

// FileNode represents a node in the file tree.
type FileNode struct {
	Name     string
	Path     string
	IsDir    bool
	Expanded bool
	Status   FileStatus
	Children []*FileNode
	Depth    int
}

// FileTree displays a hierarchical file tree.
type FileTree struct {
	Root      *FileNode
	Width     int
	Height    int
	ScrollPos int
	Selected  int
	ShowIcons bool
	flat      []*FileNode // Flattened visible nodes
}

// NewFileTree creates a new file tree widget.
func NewFileTree(width, height int) FileTree {
	return FileTree{
		Root: &FileNode{
			Name:     ".",
			IsDir:    true,
			Expanded: true,
		},
		Width:     width,
		Height:    height,
		ShowIcons: true,
	}
}

// Clear resets the tree to an empty state.
func (t *FileTree) Clear() {
	t.Root = &FileNode{
		Name:     ".",
		IsDir:    true,
		Expanded: true,
	}
	t.flat = nil
	t.ScrollPos = 0
	t.Selected = 0
}

// AddPath adds a file or directory to the tree.
func (t *FileTree) AddPath(path string, status FileStatus, isDir bool) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	current := t.Root

	for i, part := range parts {
		isLast := i == len(parts)-1
		found := false

		for _, child := range current.Children {
			if child.Name == part {
				current = child
				found = true
				break
			}
		}

		if !found {
			node := &FileNode{
				Name:     part,
				Path:     strings.Join(parts[:i+1], "/"),
				IsDir:    !isLast || isDir,
				Expanded: true,
				Depth:    i + 1,
			}
			if isLast {
				node.Status = status
			}
			current.Children = append(current.Children, node)

			// Sort children: directories first, then alphabetically
			sort.Slice(current.Children, func(i, j int) bool {
				a, b := current.Children[i], current.Children[j]
				if a.IsDir != b.IsDir {
					return a.IsDir
				}
				return a.Name < b.Name
			})

			current = node
		} else if isLast {
			current.IsDir = current.IsDir || isDir
			if status != FileStatusNone {
				current.Status = status
			}
		}
	}

	t.flatten()
}

// AddFile adds a file to the tree.
func (t *FileTree) AddFile(path string, status FileStatus) {
	t.AddPath(path, status, false)
}

// RemoveFile removes a file from the tree.
func (t *FileTree) RemoveFile(path string) {
	parts := strings.Split(filepath.ToSlash(path), "/")
	t.removeNode(t.Root, parts, 0)
	t.flatten()
}

// removeNode recursively removes a node.
func (t *FileTree) removeNode(node *FileNode, parts []string, index int) bool {
	if index >= len(parts) {
		return false
	}

	for i, child := range node.Children {
		if child.Name == parts[index] {
			if index == len(parts)-1 {
				// Remove this node
				node.Children = append(node.Children[:i], node.Children[i+1:]...)
				return true
			}
			// Recurse
			removed := t.removeNode(child, parts, index+1)
			// Remove empty directories
			if removed && len(child.Children) == 0 && child.IsDir {
				node.Children = append(node.Children[:i], node.Children[i+1:]...)
			}
			return removed
		}
	}
	return false
}

// flatten creates a flat list of visible nodes.
func (t *FileTree) flatten() {
	t.flat = nil
	t.flattenNode(t.Root)
}

// flattenNode recursively flattens the tree.
func (t *FileTree) flattenNode(node *FileNode) {
	if node != t.Root {
		t.flat = append(t.flat, node)
	}

	if node.IsDir && node.Expanded {
		for _, child := range node.Children {
			t.flattenNode(child)
		}
	}
}

// Toggle expands or collapses the selected node.
func (t *FileTree) Toggle() {
	if t.Selected >= 0 && t.Selected < len(t.flat) {
		node := t.flat[t.Selected]
		if node.IsDir {
			node.Expanded = !node.Expanded
			t.flatten()
		}
	}
}

// MoveUp moves selection up.
func (t *FileTree) MoveUp() {
	if t.Selected > 0 {
		t.Selected--
		if t.Selected < t.ScrollPos {
			t.ScrollPos = t.Selected
		}
	}
}

// MoveDown moves selection down.
func (t *FileTree) MoveDown() {
	if t.Selected < len(t.flat)-1 {
		t.Selected++
		if t.Selected >= t.ScrollPos+t.Height {
			t.ScrollPos = t.Selected - t.Height + 1
		}
	}
}

// ScrollToTop jumps to the top of the tree.
func (t *FileTree) ScrollToTop() {
	t.flatten()
	t.ScrollPos = 0
	if len(t.flat) > 0 {
		t.Selected = 0
	}
}

// ScrollToBottom jumps to the bottom of the tree.
func (t *FileTree) ScrollToBottom() {
	t.flatten()
	if len(t.flat) == 0 {
		t.ScrollPos = 0
		t.Selected = 0
		return
	}
	t.Selected = len(t.flat) - 1
	if len(t.flat) > t.Height {
		t.ScrollPos = len(t.flat) - t.Height
	} else {
		t.ScrollPos = 0
	}
}

// GetSelected returns the currently selected path.
func (t *FileTree) GetSelected() string {
	if t.Selected >= 0 && t.Selected < len(t.flat) {
		return t.flat[t.Selected].Path
	}
	return ""
}

// fileIcon returns the icon for a file type.
func fileIcon(name string, isDir bool) string {
	if isDir {
		return "📁"
	}

	ext := strings.ToLower(filepath.Ext(name))
	switch ext {
	case ".go":
		return "🔷"
	case ".py":
		return "🐍"
	case ".js", ".ts":
		return "📜"
	case ".json":
		return "📋"
	case ".yaml", ".yml":
		return "⚙️"
	case ".md":
		return "📝"
	case ".html":
		return "🌐"
	case ".css":
		return "🎨"
	case ".sql":
		return "🗃️"
	case ".sh", ".bash":
		return "💻"
	default:
		return "📄"
	}
}

// statusIndicator returns the status indicator for a file.
func statusIndicator(status FileStatus) string {
	switch status {
	case FileStatusModified:
		return styles.StatusWaiting.Render("M")
	case FileStatusAdded:
		return styles.StatusComplete.Render("A")
	case FileStatusDeleted:
		return styles.StatusError.Render("D")
	case FileStatusRenamed:
		return styles.SecondaryStyle.Render("R")
	default:
		return " "
	}
}

// View renders the file tree.
func (t FileTree) View() string {
	if t.Width <= 0 || t.Height <= 0 {
		return ""
	}
	if len(t.flat) == 0 {
		return styles.MutedStyle.Render("No files")
	}

	var lines []string

	// Calculate visible range
	end := t.ScrollPos + t.Height
	if end > len(t.flat) {
		end = len(t.flat)
	}

	for i := t.ScrollPos; i < end; i++ {
		node := t.flat[i]

		// Build indent
		indent := strings.Repeat("  ", node.Depth-1)

		// Build tree characters
		var prefix string
		if node.IsDir {
			if node.Expanded {
				prefix = "▼ "
			} else {
				prefix = "▶ "
			}
		} else {
			prefix = "  "
		}

		// Build line
		var line string
		line += indent
		line += prefix

		if t.ShowIcons {
			line += fileIcon(node.Name, node.IsDir) + " "
		}

		// Apply style based on selection and type
		nameStyle := lipgloss.NewStyle().Foreground(styles.Current.Foreground)
		if node.IsDir {
			nameStyle = nameStyle.Foreground(styles.Current.Primary)
		}
		if i == t.Selected {
			nameStyle = nameStyle.Reverse(true)
		}

		line += nameStyle.Render(node.Name)

		// Add status indicator
		if node.Status != FileStatusNone {
			line += " " + statusIndicator(node.Status)
		}

		// Truncate if too wide
		if lipgloss.Width(line) > t.Width {
			if t.Width <= 3 {
				line = line[:maxInt(t.Width, 0)]
			} else {
				line = line[:t.Width-3] + "..."
			}
		}

		lines = append(lines, line)
	}

	// Add scroll indicator
	if len(t.flat) > t.Height && t.Width > 12 {
		indicator := styles.MutedStyle.Render(
			strings.Repeat(" ", t.Width-10) +
				"[" + strings.Repeat("▓", (t.ScrollPos*10)/len(t.flat)) +
				strings.Repeat("░", 10-(t.ScrollPos*10)/len(t.flat)) + "]",
		)
		lines = append(lines, indicator)
	}

	return strings.Join(lines, "\n")
}
