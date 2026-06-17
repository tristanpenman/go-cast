package test

import (
	"testing"

	"github.com/tristanpenman/go-cast/internal/sender"
)

func TestParseYouTubeVideoID(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want string
	}{
		{"watch", "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"short host", "https://youtu.be/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"embed", "https://www.youtube.com/embed/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"shorts", "https://youtube.com/shorts/dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"music with extra params", "https://music.youtube.com/watch?v=dQw4w9WgXcQ&list=abc", "dQw4w9WgXcQ"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := sender.ParseYouTubeVideoID(tc.url)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseYouTubeVideoIDErrors(t *testing.T) {
	cases := []struct {
		name string
		url  string
	}{
		{"empty", ""},
		{"no video id", "https://www.youtube.com/"},
		{"unsupported host", "https://example.com/watch?v=dQw4w9WgXcQ"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got, err := sender.ParseYouTubeVideoID(tc.url); err == nil {
				t.Fatalf("expected error, got %q", got)
			}
		})
	}
}
