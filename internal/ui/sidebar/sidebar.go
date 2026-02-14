package sidebar

import (
	"fmt"
	"io"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
type ItemDelegate struct {
	Unread   map[string]int
	LastSeen map[string]time.Time
	ActiveID string
}

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

	// Active indicator: session seen in the last 30 seconds
	isActive := false
	if d.LastSeen != nil {
		if t, ok := d.LastSeen[item.Session.ID]; ok {
			isActive = time.Since(t) < 30*time.Second
		}
	}

	// Unread badge
	badge := ""
	if d.Unread != nil {
		if n, ok := d.Unread[item.Session.ID]; ok && n > 0 {
			badge = theme.UnreadBadgeStyle.Render(fmt.Sprintf(" (%d)", n))
		}
	}

	icon := "○"
	if isActive {
		icon = "◉"
	}

	// Group separator: render a dim line between unread and idle groups
	separator := ""
	if d.ActiveID != "" && index > 0 {
		prevItem, ok := m.Items()[index-1].(Item)
		if ok {
			prevUnread := d.Unread[prevItem.Session.ID] > 0
			prevIsActive := prevItem.Session.ID == d.ActiveID
			curUnread := d.Unread[item.Session.ID] > 0
			curIsActive := item.Session.ID == d.ActiveID

			// Show separator when transitioning between groups
			if (prevIsActive && !curIsActive) || (prevUnread && !curUnread && !curIsActive) {
				separator = lipgloss.NewStyle().Foreground(theme.Subtle).Render("  ─────────────────────") + "\n"
			}
		}
	}

	if index == m.Index() {
		title = theme.SelectedItemStyle.Render(icon+" "+title) + badge
		desc = theme.DimItemStyle.Render("  " + desc)
	} else {
		title = theme.NormalItemStyle.Render(icon+" "+title) + badge
		desc = theme.DimItemStyle.Render("  " + desc)
	}

	fmt.Fprintf(w, "%s%s\n%s", separator, title, desc)
}

// Model wraps the bubbles list component for our sidebar.
type Model struct {
	List     list.Model
	Width    int
	Height   int
	delegate *ItemDelegate
	activeID string // currently viewed session ID
}

// New creates a new sidebar model.
func New(sessions []domain.Session, width, height int) Model {
	items := make([]list.Item, len(sessions))
	for i, s := range sessions {
		items[i] = Item{Session: s}
	}

	d := &ItemDelegate{
		Unread:   make(map[string]int),
		LastSeen: make(map[string]time.Time),
	}

	l := list.New(items, d, width, height)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowTitle(false)

	return Model{List: l, Width: width, Height: height, delegate: d}
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

// SetUnread updates the unread message counts.
func (m *Model) SetUnread(unread map[string]int) {
	m.delegate.Unread = unread
}

// SetLastSeen updates the last-seen timestamps for activity indicators.
func (m *Model) SetLastSeen(lastSeen map[string]time.Time) {
	m.delegate.LastSeen = lastSeen
}

// IsFiltering returns true if the list is in active filter mode.
func (m Model) IsFiltering() bool {
	return m.List.FilterState() == list.Filtering
}

// ClearFilterAndSetItems resets the filter then replaces items with sorting.
// Use this when the user has completed a filter+select action.
func (m *Model) ClearFilterAndSetItems(sessions []domain.Session) {
	m.List.ResetFilter()
	m.setItemsInternal(sessions)
}

// Update handles messages for the sidebar list.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

// View renders the sidebar.
func (m Model) View() string {
	// The list.View() may render more lines than allocated height.
	// We must truncate to prevent pushing other UI elements off-screen.
	content := m.List.View()
	return lipgloss.NewStyle().
		Width(m.Width).
		Height(m.Height).
		MaxHeight(m.Height).
		Render(content)
}

func shortenPath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "…" + path[len(path)-maxLen+1:]
}

// SetActiveID sets the currently viewed session so it sorts to the top.
func (m *Model) SetActiveID(id string) {
	m.activeID = id
	m.delegate.ActiveID = id
}

// SetItems replaces the session list with smart sorting.
// Skips update if user is actively filtering to avoid disrupting their input.
func (m *Model) SetItems(sessions []domain.Session) {
	// Don't replace items while user is actively filtering — it would clear their input
	if m.List.FilterState() == list.Filtering || m.List.FilterState() == list.FilterApplied {
		return
	}
	m.setItemsInternal(sessions)
}

func (m *Model) setItemsInternal(sessions []domain.Session) {
	// Remember which session the cursor is on
	var cursorID string
	if sel := m.SelectedSession(); sel != nil {
		cursorID = sel.ID
	}

	sorted := m.sortSessions(sessions)

	items := make([]list.Item, len(sorted))
	for i, s := range sorted {
		items[i] = Item{Session: s}
	}
	m.List.SetItems(items)

	// Restore cursor to the same session
	if cursorID != "" {
		for i, s := range sorted {
			if s.ID == cursorID {
				m.List.Select(i)
				break
			}
		}
	}
}

// sortSessions orders sessions by priority:
// 1. Active session (currently viewed) at top
// 2. Sessions with unread messages, sorted by most recent activity
// 3. Idle sessions, sorted by most recently updated
func (m *Model) sortSessions(sessions []domain.Session) []domain.Session {
	result := make([]domain.Session, len(sessions))
	copy(result, sessions)

	sort.SliceStable(result, func(i, j int) bool {
		si, sj := result[i], result[j]

		// Active session always first
		if si.ID == m.activeID && sj.ID != m.activeID {
			return true
		}
		if sj.ID == m.activeID && si.ID != m.activeID {
			return false
		}

		// Unread sessions before idle
		ui := m.delegate.Unread[si.ID]
		uj := m.delegate.Unread[sj.ID]
		hasUnreadI := ui > 0
		hasUnreadJ := uj > 0

		if hasUnreadI && !hasUnreadJ {
			return true
		}
		if hasUnreadJ && !hasUnreadI {
			return false
		}

		// Within same group, sort by most recent activity
		ti := m.lastActivity(si)
		tj := m.lastActivity(sj)
		return ti.After(tj)
	})

	return result
}

// lastActivity returns the most recent timestamp for a session.
func (m *Model) lastActivity(s domain.Session) time.Time {
	if t, ok := m.delegate.LastSeen[s.ID]; ok {
		if t.After(s.UpdatedAt) {
			return t
		}
	}
	return s.UpdatedAt
}
