package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
)

// OSNotifier sends native OS desktop notifications.
type OSNotifier struct{}

// NewOS creates an OS notification backend.
func NewOS() *OSNotifier {
	return &OSNotifier{}
}

// Notify sends a native desktop notification.
func (o *OSNotifier) Notify(n Notification) error {
	title := n.Title
	if title == "" {
		title = "Copilot ICQ"
	}
	body := n.Body
	if body == "" {
		body = fmt.Sprintf("Event: %s", n.Event)
	}

	switch runtime.GOOS {
	case "darwin":
		return exec.Command("osascript", "-e",
			fmt.Sprintf(`display notification %q with title %q`, body, title),
		).Run()

	case "linux":
		return exec.Command("notify-send", title, body).Run()

	case "windows":
		// PowerShell toast notification
		script := fmt.Sprintf(
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] > $null; `+
				`$xml = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent(1); `+
				`$text = $xml.GetElementsByTagName('text'); `+
				`$text.Item(0).AppendChild($xml.CreateTextNode('%s')) > $null; `+
				`$text.Item(1).AppendChild($xml.CreateTextNode('%s')) > $null; `+
				`$toast = [Windows.UI.Notifications.ToastNotification]::new($xml); `+
				`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('Copilot ICQ').Show($toast)`,
			title, body,
		)
		return exec.Command("powershell", "-Command", script).Run()

	default:
		return fmt.Errorf("OS notifications not supported on %s", runtime.GOOS)
	}
}
