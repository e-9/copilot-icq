package domain

import "testing"

func TestSessionDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		session Session
		want    string
	}{
		{"with summary", Session{Summary: "My Session", CWD: "/tmp"}, "My Session"},
		{"without summary, with cwd", Session{CWD: "/Users/test/project"}, "project"},
		{"no summary no cwd", Session{ID: "abc12345-def"}, "abc12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.session.DisplayName()
			if got != tt.want {
				t.Errorf("DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSessionShortID(t *testing.T) {
	s := Session{ID: "28424ebb-c4ea-4590-ac5e-d86f5b0b98a8"}
	if got := s.ShortID(); got != "28424ebb" {
		t.Errorf("ShortID() = %q, want %q", got, "28424ebb")
	}

	short := Session{ID: "abc"}
	if got := short.ShortID(); got != "abc" {
		t.Errorf("ShortID() = %q, want %q", got, "abc")
	}
}
