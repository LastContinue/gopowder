package nix

import (
	"errors"
	"os"
	"testing"
)

func TestExpandTilde(t *testing.T) {

	home, err := os.UserHomeDir()
	if err != nil {
		t.Error("TestExpandTilde couldn't find os.UserHomeDir")
	}

	tests := []struct {
		name string
		have string
		want string
	}{
		{
			name: "Path with tilde matches HOME + path",
			have: "~/foo/bar",
			want: home + "/foo/bar",
		},
		{
			name: "Path with no tilde matches path",
			have: "/foo/bar",
			want: "/foo/bar",
		},
		{
			name: "empty/blank path returns empty/blank path",
			have: "/foo/bar",
			want: "/foo/bar",
		},
		{
			name: "just '~' matches $HOME",
			have: "~",
			want: home,
		},
	}

	for _, tt := range tests {

		getHome := func() (string, error) {
			return os.UserHomeDir()
		}

		expandedPath, err := expandTilde(tt.have, getHome)
		if err != nil {
			t.Error(err)
		}

		if expandedPath != tt.want {
			t.Errorf("Expected expandedPath to be '%s', was actually, '%s'", tt.want, expandedPath)
		}
	}
}

func TestExpandTildeNoHomeDir(t *testing.T) {

	getHome := func() (string, error) {
		return "", errors.New("No Home Error")
	}

	homeless, err := expandTilde("/a/good/path", getHome)
	if err == nil && homeless != "" {
		t.Error("TestExpandTilde still found os.UserHomeDir when $HOME was unset")
	}
}

func TestGreenText(t *testing.T) {
	regularText := "Regular Text"
	greenText := green(regularText)

	// Not going too hard on this, just want to make sure that "foo" != green("foo")
	if greenText == regularText {
		t.Errorf("Expected green(%s) to NOT be '%s', was actually, '%s'", regularText, regularText, greenText)
	}
}
