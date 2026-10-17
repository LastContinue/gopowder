package nix

import (
	"reflect"
	"slices"
	"testing"

	"github.com/jedib0t/go-pretty/v6/table"
)

type mockSetting struct{}

func (m mockSetting) getMockPackages() []Package {

	return []Package{
		{
			Name:           "PkgA",
			CurrentVersion: new("1.0.1"),
			Update:         new("1.0.2"),
		},
		{
			Name:           "PkgB",
			CurrentVersion: new("1.0.1"),
			Update:         nil,
		},
		{
			Name:           "PkgC",
			CurrentVersion: new("some_weird_setup"),
			Update:         new("even_weirder_setup"),
		},
	}
}

func (m mockSetting) Path() string {
	return "mock/flake/path"
}

func (m mockSetting) ConfigName() string {
	return "mockConfigName"
}

func (m mockSetting) GetPackagePaths() ([]string, error) {
	return []string{"/a/good/path", "/a/better/path", "/the/best/path"}, nil
}

func (m mockSetting) GetPackage(path string) (*Package, error) {
	packages := m.getMockPackages()
	matches := map[string]Package{
		"/a/good/path":   packages[0],
		"/a/better/path": packages[1],
		"/the/best/path": packages[2],
	}

	// If this doesn't find a match, we got a big problem...
	pkg := matches[path]

	return &pkg, nil
}

type mockProgress struct{}

func (m mockProgress) inc(int) error {
	return nil
}

func (m mockProgress) close() error {
	return nil
}

func initMockProgress(_ int) MarksProgress {
	return mockProgress{}
}

func TestRun_Packages(t *testing.T) {
	ms := mockSetting{}

	tests := []struct {
		name     string
		settings GetsPackages
		progress func(int) MarksProgress
		want     []Package
	}{
		{
			name:     "Gets Packages",
			settings: ms,
			progress: initMockProgress,
			want:     ms.getMockPackages(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rs, error := Run(tt.settings, tt.progress)
			if error != nil {
				t.Error(error)
			}

			if !reflect.DeepEqual(rs.Packages, tt.want) {
				t.Errorf("Expected rs.Packages to be %v, was actually, %v", tt.want, rs.Packages)
			}
		})
	}
}

// 'Run' does control what goes into RunSummary.UpdateNames so we can test it here
func TestRun_Updates(t *testing.T) {

	tests := []struct {
		name     string
		settings GetsPackages
		progress func(int) MarksProgress
		want     []string
	}{
		{
			name:     "Packages with updates are returned in RunSummary.UpdateNames",
			settings: mockSetting{},
			progress: initMockProgress,
			want:     []string{"PkgA", "PkgC"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rs, error := Run(tt.settings, initMockProgress)
			if error != nil {
				t.Error(error)
			}

			if !slices.Equal(rs.UpdateNames, tt.want) {
				t.Errorf("Expected rs.UpdateNames to be %v, was actually, %v", tt.want, rs.UpdateNames)
			}
		})
	}
}

func TestRunSummary_ToTable(t *testing.T) {

	packages := []Package{
		{
			Name:           "PkgA",
			CurrentVersion: new("1.0.0"),
			Update:         new("1.0.1"),
		},
		{

			Name:           "PkgB",
			CurrentVersion: new("2.0"),
			Update:         nil,
		},
		{

			Name:           "PkgC",
			CurrentVersion: new("something else"),
			Update:         nil,
		},
	}

	tests := []struct {
		name         string
		summary      RunSummary
		updatedNames []string
		want         []table.Row
	}{
		{
			name: "ToTable returns correct rows",
			summary: RunSummary{
				Packages: packages,
			},
			want: []table.Row{
				{"Name", "Current Version", "Update"},
				{"PkgA", "1.0.0", green("1.0.1")},
				{"PkgB", "2.0", green("")}, //yes, an empty green string is different than an empty string
				{"PkgC", "something else", green("")},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableRows := tt.summary.ToTable()

			if !reflect.DeepEqual(tableRows, tt.want) {
				t.Errorf("Expected tableRows to be %v, was actually, %v", tt.want, tableRows)
			}
		})
	}
}

func TestPackage_isNonVersioned(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion *string
		want           bool
	}{
		{
			name: "nil version",
			want: true,
		},
		{
			name:           "empty version",
			currentVersion: new(""),
			want:           true,
		},
		{
			name:           "whitespace only version",
			currentVersion: new(" \t\n "),
			want:           true,
		},
		{
			name:           "version with surrounding whitespace",
			currentVersion: new(" 1.2.3 "),
			want:           false,
		},
		{
			name:           "version",
			currentVersion: new("1.2.3"),
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Package{CurrentVersion: tt.currentVersion}

			if got := p.isNonVersioned(); got != tt.want {
				t.Errorf("isNonVersioned() = %t, want %t", got, tt.want)
			}
		})
	}
}
