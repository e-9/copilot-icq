package app

import (
"fmt"
"os"
"path/filepath"
"strings"
"time"

tea "github.com/charmbracelet/bubbletea"
sdk "github.com/github/copilot-sdk/go"

"github.com/e-9/copilot-icq/internal/copilot"
"github.com/e-9/copilot-icq/internal/domain"
"github.com/e-9/copilot-icq/internal/ui/chat"
"github.com/e-9/copilot-icq/internal/ui/theme"
"gopkg.in/yaml.v3"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
var cmds []tea.Cmd

switch msg := msg.(type) {
case tea.KeyMsg:
// Help overlay: any key dismisses
if m.showHelp {
m.showHelp = false
return m, nil
}

switch msg.String() {
case "ctrl+c":
// If a message is being sent via SDK, abort it instead of quitting
if m.selected != nil && m.pendingSends[m.selected.ID] && m.sdkResumed[m.selected.ID] {
cmds = append(cmds, sdkAbort(m.adapter, m.selected.ID))
return m, tea.Batch(cmds...)
}
return m, tea.Quit
case "q":
if m.focus != FocusInput {
return m, tea.Quit
}
case "esc":
if m.renaming {
m.renaming = false
m.input.Reset()
m.input.ClearRenaming()
m.focus = FocusSidebar
m.input.Blur()
return m, nil
}
if m.focus == FocusInput {
m.focus = FocusChat
m.input.Blur()
return m, nil
}
case "?":
if m.focus != FocusInput {
m.showHelp = !m.showHelp
return m, nil
}
case "e":
if m.focus != FocusInput && m.selected != nil {
return m, m.exportConversation()
}
case "R":
// Shift+R: rename session (sidebar only)
if m.focus == FocusSidebar {
if s := m.sidebar.SelectedSession(); s != nil {
m.renaming = true
m.focus = FocusInput
m.input.SetRenaming(s.DisplayName())
}
return m, nil
}
case "r":
if m.focus != FocusInput {
cmds = append(cmds, sdkListSessions(m.adapter))
}
case "t":
if m.focus == FocusChat {
m.chat.ToggleAllToolCalls()
return m, nil
}
case "tab":
switch m.focus {
case FocusSidebar:
m.focus = FocusChat
m.input.Blur()
case FocusChat:
if m.selected != nil && m.canSend() {
m.focus = FocusInput
m.input.Focus()
} else {
m.focus = FocusSidebar
m.input.Blur()
}
case FocusInput:
m.focus = FocusSidebar
m.input.Blur()
}
return m, nil
case "shift+tab":
switch m.focus {
case FocusSidebar:
if m.selected != nil && m.canSend() {
m.focus = FocusInput
m.input.Focus()
} else {
m.focus = FocusChat
m.input.Blur()
}
case FocusChat:
m.focus = FocusSidebar
m.input.Blur()
case FocusInput:
m.focus = FocusChat
m.input.Blur()
}
return m, nil
case "enter":
if m.focus == FocusSidebar && !m.sidebar.IsFiltering() {
if s := m.sidebar.SelectedSession(); s != nil {
m.selected = s
m.focus = FocusChat
m.unread[s.ID] = 0
m.sidebar.SetActiveID(s.ID)
m.sidebar.SetUnread(m.unread)
m.sidebar.ClearFilterAndSetItems(m.sessions)
m.input.SetSending(m.pendingSends[s.ID])
if !m.pendingSends[s.ID] {
m.input.Reset()
}
if !m.sdkResumed[s.ID] {
cmds = append(cmds, sdkResumeSession(m.adapter, s.ID))
} else {
cmds = append(cmds, sdkLoadHistory(m.adapter, s.ID))
}
m.chat.SetPendingTools(m.pendingToolsForChat())
}
} else if m.focus == FocusInput {
if m.renaming {
// Complete rename
newName := m.input.Value()
if newName != "" && m.sidebar.SelectedSession() != nil {
s := m.sidebar.SelectedSession()
cmds = append(cmds, m.renameSession(s.ID, newName))
}
m.renaming = false
m.input.Reset()
m.input.ClearRenaming()
m.focus = FocusSidebar
m.input.Blur()
} else {
text := m.input.Value()
if text != "" && m.selected != nil && !m.pendingSends[m.selected.ID] && m.sdkResumed[m.selected.ID] {
m.pendingSends[m.selected.ID] = true
m.input.SetSending(true)
m.sidebar.SetPendingSends(m.pendingSends)
m.input.Reset()
cmds = append(cmds, sdkSendMessage(m.adapter, m.selected.ID, text))
}
}
}
}

case tea.MouseMsg:
// Determine which panel was clicked based on X coordinate
if msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion {
if msg.Button == tea.MouseButtonLeft {
if msg.X < m.sidebar.Width+2 {
if m.focus != FocusSidebar {
m.focus = FocusSidebar
m.input.Blur()
}
} else if m.height > 0 && msg.Y >= m.height-5 && m.selected != nil {
if m.focus != FocusInput && m.canSend() {
m.focus = FocusInput
m.input.Focus()
}
} else {
if m.focus != FocusChat {
m.focus = FocusChat
m.input.Blur()
}
}
}
}

case tea.WindowSizeMsg:
m.width = msg.Width
m.height = msg.Height
m.ready = true
borderH := 2
borderW := 2
headerH := 1
statusBarH := 1

panelHeight := m.height - headerH - statusBarH - borderH
if panelHeight < 1 {
panelHeight = 1
}
sidebarInnerW := theme.SidebarWidth
chatInnerW := m.width - sidebarInnerW - borderW*2
if chatInnerW < 1 {
chatInnerW = 1
}

m.sidebar.SetSize(sidebarInnerW, panelHeight)

inputHeight := 1
inputBorderH := inputHeight + borderH
chatInnerH := panelHeight - inputBorderH
m.chat.SetSize(chatInnerW, chatInnerH)
m.input.SetWidth(chatInnerW)

case SessionsLoadedMsg:
if msg.Err != nil {
m.err = msg.Err
return m, nil
}
m.sessions = msg.Sessions
m.sidebar.SetItems(msg.Sessions)

case EventsLoadedMsg:
if msg.Err != nil {
m.err = msg.Err
return m, nil
}
if m.selected != nil && m.selected.ID == msg.SessionID {
m.chat.SetMessages(msg.Messages)
}

case TickMsg:
cmds = append(cmds, sdkListSessions(m.adapter))
cmds = append(cmds, tickEvery(5*time.Second))

case MessageSentMsg:
delete(m.pendingSends, msg.SessionID)
m.sidebar.SetPendingSends(m.pendingSends)
if m.selected != nil && m.selected.ID == msg.SessionID {
m.input.SetSending(false)
}
if msg.Err != nil {
m.err = msg.Err
} else {
m.input.Reset()
}

case SessionRenamedMsg:
if msg.Err == nil {
cmds = append(cmds, sdkListSessions(m.adapter))
}

case ExportCompleteMsg:
// Nothing to do in the model

case SDKConnectedMsg:
if msg.Err != nil {
m.statusFlash = fmt.Sprintf("⚠️  SDK init failed: %v", msg.Err)
cmds = append(cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return ClearFlashMsg{} }))
} else {
m.statusFlash = "🔗 SDK connected"
cmds = append(cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return ClearFlashMsg{} }))
cmds = append(cmds, sdkListSessions(m.adapter))
}

case SDKSessionResumedMsg:
if msg.Err != nil {
m.statusFlash = fmt.Sprintf("⚠️  Resume failed: %v", msg.Err)
cmds = append(cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return ClearFlashMsg{} }))
} else {
m.sdkResumed[msg.SessionID] = true
cmds = append(cmds, sdkLoadHistory(m.adapter, msg.SessionID))
}

case SDKEventMsg:
evt := msg.Event
switch evt.Type {
case copilot.EventSession:
if evt.SessionEvent != nil {
m.handleSDKSessionEvent(evt.SessionID, *evt.SessionEvent, &cmds)
}
case copilot.EventLifecycle:
cmds = append(cmds, sdkListSessions(m.adapter))
case copilot.EventPermission:
if evt.Permission != nil {
m.pendingTools[evt.SessionID] = append(m.pendingTools[evt.SessionID], PendingTool{
ToolName: evt.Permission.ToolName,
ToolArgs: evt.Permission.Action,
})
if m.selected != nil && m.selected.ID == evt.SessionID {
m.chat.SetPendingTools(m.pendingToolsForChat())
}
// Auto-allow for now
evt.Permission.Response <- copilot.PermissionResponse{Allow: true}
}
case copilot.EventUserInput:
if evt.UserInput != nil {
if m.selected != nil && m.selected.ID == evt.SessionID {
msgs := m.chat.Messages()
msgs = append(msgs, domain.Message{
Role:    domain.RoleSystem,
Content: fmt.Sprintf("🤖 Agent asks: %s", evt.UserInput.Question),
})
m.chat.SetMessages(msgs)
}
// Auto-respond for now
evt.UserInput.Response <- copilot.UserInputResponse{Answer: "", WasFreeform: true}
}
}
// Keep listening
cmds = append(cmds, listenSDKEvents(m.adapter))

case SDKDisconnectedMsg:
m.statusFlash = "⚠️  SDK disconnected"
cmds = append(cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return ClearFlashMsg{} }))

case ClearFlashMsg:
m.statusFlash = ""
}

// Route input to focused panel
switch m.focus {
case FocusSidebar:
var cmd tea.Cmd
m.sidebar, cmd = m.sidebar.Update(msg)
cmds = append(cmds, cmd)
case FocusChat:
var cmd tea.Cmd
m.chat, cmd = m.chat.Update(msg)
cmds = append(cmds, cmd)
case FocusInput:
var cmd tea.Cmd
m.input, cmd = m.input.Update(msg)
cmds = append(cmds, cmd)
}

return m, tea.Batch(cmds...)
}

// renameSession updates the summary field in workspace.yaml.
func (m Model) renameSession(sessionID, newName string) tea.Cmd {
return func() tea.Msg {
wsPath := filepath.Join(m.sessionBasePath, sessionID, "workspace.yaml")
data, err := os.ReadFile(wsPath)
if err != nil {
return SessionRenamedMsg{SessionID: sessionID, Err: err}
}

var ws map[string]interface{}
if err := yaml.Unmarshal(data, &ws); err != nil {
return SessionRenamedMsg{SessionID: sessionID, Err: err}
}

ws["summary"] = newName
out, err := yaml.Marshal(ws)
if err != nil {
return SessionRenamedMsg{SessionID: sessionID, Err: err}
}

if err := os.WriteFile(wsPath, out, 0600); err != nil {
return SessionRenamedMsg{SessionID: sessionID, Err: err}
}

return SessionRenamedMsg{SessionID: sessionID}
}
}

// exportConversation writes the current conversation to a markdown file.
func (m Model) exportConversation() tea.Cmd {
return func() tea.Msg {
if m.selected == nil {
return ExportCompleteMsg{Err: fmt.Errorf("no session selected")}
}

var sb strings.Builder
sb.WriteString(fmt.Sprintf("# Copilot Session: %s\n\n", m.selected.DisplayName()))
sb.WriteString(fmt.Sprintf("- **Session ID**: `%s`\n", m.selected.ID))
sb.WriteString(fmt.Sprintf("- **CWD**: `%s`\n", m.selected.CWD))
sb.WriteString(fmt.Sprintf("- **Created**: %s\n", m.selected.CreatedAt.Format(time.RFC3339)))
sb.WriteString(fmt.Sprintf("- **Updated**: %s\n\n", m.selected.UpdatedAt.Format(time.RFC3339)))
sb.WriteString("---\n\n")

msgs := m.chat.Messages()
for _, msg := range msgs {
ts := msg.Timestamp.Format("15:04:05")
switch msg.Role {
case domain.RoleUser:
sb.WriteString(fmt.Sprintf("### 🧑 You (%s)\n\n%s\n\n", ts, msg.Content))
case domain.RoleAssistant:
sb.WriteString(fmt.Sprintf("### 🤖 Copilot (%s)\n\n", ts))
if msg.Content != "" {
sb.WriteString(msg.Content + "\n\n")
}
for _, tc := range msg.ToolCalls {
status := "⏳"
if tc.Status == domain.ToolCallComplete {
status = "✓"
} else if tc.Status == domain.ToolCallFailed {
status = "✗"
}
sb.WriteString(fmt.Sprintf("- 🔧 **%s** %s", tc.Name, status))
if tc.Summary != "" {
summary := tc.Summary
if len(summary) > 100 {
summary = summary[:97] + "..."
}
sb.WriteString(fmt.Sprintf(": `%s`", summary))
}
sb.WriteString("\n")
}
sb.WriteString("\n")
case domain.RoleSystem:
sb.WriteString(fmt.Sprintf("*ℹ %s* (%s)\n\n", msg.Content, ts))
}
}

exportDir := "."
if m.cfg != nil && m.cfg.ExportDir != "" {
exportDir = m.cfg.ExportDir
}

filename := fmt.Sprintf("copilot-session-%s.md", m.selected.ShortID())
outPath := filepath.Join(exportDir, filename)
if err := os.WriteFile(outPath, []byte(sb.String()), 0644); err != nil {
return ExportCompleteMsg{Err: err}
}

return ExportCompleteMsg{Path: outPath}
}
}

// pendingToolsForChat returns pending tools for the currently selected session.
func (m Model) pendingToolsForChat() []chat.PendingTool {
if m.selected == nil {
return nil
}
tools := m.pendingTools[m.selected.ID]
if len(tools) == 0 {
return nil
}
result := make([]chat.PendingTool, len(tools))
for i, t := range tools {
result[i] = chat.PendingTool{
ToolName:   t.ToolName,
ToolArgs:   t.ToolArgs,
Denied:     t.Denied,
DenyReason: t.DenyReason,
}
}
return result
}

// canSend returns true if the user can send messages to the selected session.
func (m Model) canSend() bool {
return m.selected != nil && m.sdkResumed[m.selected.ID]
}

// handleSDKSessionEvent processes a single SDK session event.
func (m *Model) handleSDKSessionEvent(sessionID string, event sdk.SessionEvent, cmds *[]tea.Cmd) {
m.lastSeen[sessionID] = time.Now()
m.sidebar.SetLastSeen(m.lastSeen)

switch event.Type {
case sdk.AssistantMessageDelta:
if m.selected != nil && m.selected.ID == sessionID {
if event.Data.DeltaContent != nil {
msgs := m.chat.Messages()
if len(msgs) > 0 && msgs[len(msgs)-1].Role == domain.RoleAssistant {
msgs[len(msgs)-1].Content += *event.Data.DeltaContent
} else {
msgs = append(msgs, domain.Message{
Role:    domain.RoleAssistant,
Content: *event.Data.DeltaContent,
})
}
m.chat.SetMessages(msgs)
}
} else {
m.unread[sessionID]++
m.sidebar.SetUnread(m.unread)
}

case sdk.AssistantMessage:
if m.selected != nil && m.selected.ID == sessionID {
if event.Data.Content != nil {
msgs := m.chat.Messages()
if len(msgs) > 0 && msgs[len(msgs)-1].Role == domain.RoleAssistant {
msgs[len(msgs)-1].Content = *event.Data.Content
} else {
msgs = append(msgs, domain.Message{
Role:    domain.RoleAssistant,
Content: *event.Data.Content,
})
}
m.chat.SetMessages(msgs)
}
}

case sdk.ToolExecutionStart:
toolName := ""
if event.Data.ToolName != nil {
toolName = *event.Data.ToolName
}
m.pendingTools[sessionID] = append(m.pendingTools[sessionID], PendingTool{
ToolName: toolName,
})
if m.selected != nil && m.selected.ID == sessionID {
m.chat.SetPendingTools(m.pendingToolsForChat())
}

case sdk.ToolExecutionComplete:
toolName := ""
if event.Data.ToolName != nil {
toolName = *event.Data.ToolName
}
if tools, ok := m.pendingTools[sessionID]; ok {
var remaining []PendingTool
removed := false
for _, t := range tools {
if !removed && t.ToolName == toolName {
removed = true
continue
}
remaining = append(remaining, t)
}
m.pendingTools[sessionID] = remaining
}
if m.selected != nil && m.selected.ID == sessionID {
m.chat.SetPendingTools(m.pendingToolsForChat())
}

case sdk.SessionIdle:
delete(m.pendingSends, sessionID)
m.sidebar.SetPendingSends(m.pendingSends)
if m.selected != nil && m.selected.ID == sessionID {
m.input.SetSending(false)
}

case sdk.SessionError:
errMsg := "unknown error"
if event.Data.Message != nil {
errMsg = *event.Data.Message
}
m.statusFlash = fmt.Sprintf("⚠️  %s", errMsg)
*cmds = append(*cmds, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg { return ClearFlashMsg{} }))
}

m.sidebar.SetItems(m.sessions)
}
