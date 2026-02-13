package sidebar

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/e-9/copilot-icq/internal/domain"
	"github.com/e-9/copilot-icq/internal/ui/theme"
)

// Item wraps a domain.Session for display in the sidebar list.
type Item struct {
	Session domain.Session
}

func (i Item) Title() string       { return i.Session.DisplayName() }
func (i Item) Description() string { return i.Session.ShortID() + " · " + i.Session.CWD }
func (i Item) FilterValue() string { return i.Session.DisplayName() + " " + i.Session.CWD }

// ItemDelegate renders each session item in the list.
type ItemDelegate struct{}

func (d ItemDelegate) Height() int                             { return 2 }
func (d ItemDelegate) Spacing() int                            { return 0 }
func (d ItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d ItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(Item)
	if !ok {
		return
	}

	title := item.Title()
	desc := item.Session.ShortID() + " · " + shortenPath(item.Session.CWD, 20)

	if index == m.Index() {
		title = theme.SelectedItemStyle.Render("◉ " + title)
		desc = theme.DimItemStyle.Render("  " + desc)
	} else {
		title = theme.NormalItemStyle.Render("○ " + title)
		desc = theme.DimItemStyle.Render("  " + desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

// Model wraps the bubbles list component for our sidebar.
type Model struct {
	List   list.Model
	Width  int
	Height int
}

// New creates a new sidebar model.
func New(sessions []domain.Session, width, height int) Model {
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = Item{Session: s}
	}

	l := list.New(items, ItemDelegate{}, width, height)
	l.Title = "🌸 Sessions"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = theme.TitleStyle
	l.SetShowHelp(false)

	return Model{List: l, Width: width, Height: height}
}

// SelectedSession returns the currently selected session, if any.
func (m Model) SelectedSession() *domain.Session {
	item, ok := m.List.SelectedItem().(Item)
	if !ok {
		return nil
	}
	return &item.Session
}

// SetSize updates the sidebar dimensions.
func (m *Model) SetSize(w, h int) {
	m.Width = w
	m.Height = h
	m.List.SetSize(w, h)
}

// Update handles messages for the sidebar list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// View renders the sidebar.
func (m Model) View() string {
	return theme.SidebarStyle.Height(m.Height).Render(m.List.View())
}

func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

// SetItems replaces the session list.
func (m *Model) SetItems(sessions []domain.Session) {
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = Item{Session: s}
	}
	m.List.SetItems(items)
}
