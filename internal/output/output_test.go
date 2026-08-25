package output

import "testing"

func TestFormatUpdatedString(t *testing.T) {
	tests := map[string][]string{
		"0 Updated Packages":              make([]string, 0),
		"1 Updated Package - PkgA":        {"PkgA"},
		"2 Updated Packages - PkgA, PkgB": {"PkgA", "PkgB"},
	}

	for want, have := range tests {
		output := FormatUpdatedString(have)

		if want != output {
			t.Errorf("Expected %s, got %s", want, output)
		}
	}
}
