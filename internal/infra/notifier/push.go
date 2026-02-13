package notifier

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

// PushNotifier sends push notifications via ntfy.sh.
type PushNotifier struct {
	topic string
}

// NewPush creates a push notifier. Returns nil if COPILOT_ICQ_NTFY_TOPIC is unset.
func NewPush() *PushNotifier {
	topic := os.Getenv("COPILOT_ICQ_NTFY_TOPIC")
	if topic == "" {
		return nil
	}
	return &PushNotifier{topic: topic}
}

// Notify sends a push notification via ntfy.sh.
func (p *PushNotifier) Notify(n Notification) error {
	title := n.Title
	if title == "" {
		title = "Copilot ICQ"
	}
	body := n.Body
	if body == "" {
		body = fmt.Sprintf("Event: %s", n.Event)
	}

	url := fmt.Sprintf("https://ntfy.sh/%s", p.topic)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Title", title)
	req.Header.Set("Tags", "robot")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
