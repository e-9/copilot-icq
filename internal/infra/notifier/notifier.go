package notifier

// Notification represents something worth alerting the user about.
type Notification struct {
	SessionID string
	Event     string // hook event name
	Title     string // short title for OS notification
	Body      string // detail text
}

// Notifier sends notifications via a specific channel.
type Notifier interface {
	Notify(n Notification) error
}

// Router fans out notifications to multiple backends.
type Router struct {
	backends []Notifier
}

// NewRouter creates a router with the given backends.
func NewRouter(backends ...Notifier) *Router {
	return &Router{backends: backends}
}

// Notify sends a notification to all backends.
func (r *Router) Notify(n Notification) {
	for _, b := range r.backends {
		b.Notify(n)
	}
}

// Add registers a new backend.
func (r *Router) Add(n Notifier) {
	r.backends = append(r.backends, n)
}
