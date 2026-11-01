package nix

import (
	"bytes"
	"errors"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const decoderErrorString = "Decoder error"

type MockDecoder struct{}

func (m MockDecoder) DecodeIntoFlakeSettings(config string, fs *FlakeSettings) error {
	pd := PackageDetails{
		Ignore:    []string{"PkgIgnore"},
		Overrides: map[string]string{"PkgOverride": "PkgOverridden"},
	}

	fs.FlakeDir = "/a/good/path"
	fs.FlakeConfigName = "Good Config Name"
	fs.PackageDetails = &pd

	return nil
}

type MockDecoderNoConfigName struct{}

func (m MockDecoderNoConfigName) DecodeIntoFlakeSettings(config string, fs *FlakeSettings) error {
	fs.FlakeDir = "/a/good/path"
	return nil
}

type MockDecoderThrowsError struct{}

func (m MockDecoderThrowsError) DecodeIntoFlakeSettings(config string, fs *FlakeSettings) error {
	return errors.New(decoderErrorString)
}

type MockEval struct{}

func (m MockEval) getMockPackages() []Package {

	return []Package{
		{
			Name:           "PkgA",
			CurrentVersion: new("1.0"),
			Update:         new("2.0"),
		},
		{
			Name:           "PkgB",
			CurrentVersion: nil,
			Update:         nil,
		},
		{
			Name:           "PkgC",
			CurrentVersion: new(" "),
			Update:         nil,
		},
		{
			Name:           "PkgD",
			CurrentVersion: new("some_weird_setup"),
			Update:         nil,
		},
		{
			Name:           "PkgIgnore",
			CurrentVersion: new("1.0.1"),
			Update:         nil,
		},
		{
			Name:           "PkgOverride",
			CurrentVersion: new("1.0"),
			Update:         new("2.0"),
		},
	}
}

func (m MockEval) getPackagePaths(_path string, _configName string) ([]string, error) {
	return []string{
		"/a/good/path",
		"/a/non-versioned/path",
	}, nil
}

func (m MockEval) getPackageFromPath(path string) (*Package, error) {
	packages := m.getMockPackages()
	matches := map[string]Package{
		"/a/good/path":                packages[0],
		"/a/non-versioned/path":       packages[1],
		"/another/non-versioned/path": packages[2],
		"/the/best/path":              packages[3],
		"/ignore/path":                packages[4],
		"/override/path":              packages[5],
	}

	// If this doesn't find a match, we got a big problem...
	pkg := matches[path]

	return &pkg, nil
}

func (m MockEval) getUnstableVersion(_packageName string) (*string, error) {
	return new("2.0"), nil
}

type MockUnstableVersionErrorsEval struct {
	MockEval
}

func (m MockUnstableVersionErrorsEval) getUnstableVersion(_packageName string) (*string, error) {
	return nil, errors.New("getUnstableVersion error")
}

type MockAllErrorEval struct {
	MockUnstableVersionErrorsEval
}

func (m MockAllErrorEval) getPackagePaths(_path string, _configName string) ([]string, error) {
	return nil, errors.New("getPackagePaths error")
}

func (m MockAllErrorEval) getPackageFromPath(path string) (*Package, error) {
	return nil, errors.New("getPackageFromPath error")
}

func TestNewFlakeSettings(t *testing.T) {

	flakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, MockEval{})
	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	flakeSettingsNoConfigName, err := NewFlakeSettings("config.toml", MockDecoderNoConfigName{}, MockEval{})
	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	_, badDecoderErr := NewFlakeSettings("config.toml", MockDecoderThrowsError{}, MockEval{})

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd := exec.Command("scutil", "--get", "LocalHostName")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("hostname failed: %v", err)
	}

	defaultConfigName := strings.TrimSpace(stdout.String())

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "Sets FlakeDir Name",
			got:  flakeSettings.FlakeDir,
			want: "/a/good/path",
		},
		{
			name: "Sets FlakeConfig Name",
			got:  flakeSettings.FlakeConfigName,
			want: "Good Config Name",
		},
		{
			name: "Sets FlakeConfig to Default When None Set",
			got:  flakeSettingsNoConfigName.FlakeConfigName,
			want: defaultConfigName,
		},
		{
			name: "Bad decoding throws error",
			got:  badDecoderErr.Error(),
			want: decoderErrorString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.got != tt.want {
				t.Errorf("Expected FlakeSettings value to be %v, was actually, %v", tt.want, tt.got)
			}
		})
	}
}

func TestFlakeSettings_SliceReturningMethods(t *testing.T) {

	flakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, MockEval{})

	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	tests := []struct {
		name     string
		settings GetsPackages
		want     []string
	}{
		{
			name:     "Gets (Package) Paths",
			settings: flakeSettings,
			want: []string{
				"/a/good/path",
				"/a/non-versioned/path"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			paths, error := tt.settings.GetPackagePaths()
			if error != nil {
				t.Error(error)
			}

			if !slices.Equal(paths, tt.want) {
				t.Errorf("Expected paths to be %v, was actually, %v", tt.want, paths)
			}
		})
	}
}

func TestFlakeSettings_StringReturningMethods(t *testing.T) {

	flakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, MockEval{})

	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	tests := []struct {
		name       string
		method     func() string
		methodName string // There's a way to do this with reflect, but this is more clear for me right now
		want       string
	}{
		{
			name:       "Gets Path",
			method:     flakeSettings.Path,
			methodName: ".Path()",
			want:       "/a/good/path",
		}, {
			name:       "Gets ConfigName",
			method:     flakeSettings.ConfigName,
			methodName: ".ConfigName()",
			want:       "Good Config Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			have := tt.method()

			if have != tt.want {
				t.Errorf("Expected '%s' to be '%v', was actually, '%v'", tt.methodName, tt.want, have)
			}
		})
	}
}

func TestFlakeSettings_GetPackage(t *testing.T) {
	eval := MockEval{}
	flakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, eval)
	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	allErrorEval := MockAllErrorEval{}
	allErrorsFlakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, allErrorEval)
	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	unstableEval := MockUnstableVersionErrorsEval{}
	unstableFlakeSettings, err := NewFlakeSettings("config.toml", MockDecoder{}, unstableEval)
	if err != nil {
		t.Fatalf("NewFlakeSettings instantiation error: %v", err)
	}

	allPackages := eval.getMockPackages()

	overridePkg := &allPackages[5]
	overridePkg.Name = "PkgOverride (PkgOverridden)"

	tests := []struct {
		name     string
		settings *FlakeSettings
		wantErr  bool
		pkgPath  string
		want     *Package
	}{
		{
			name:     "Gets Packages",
			settings: flakeSettings,
			wantErr:  false,
			pkgPath:  "/a/good/path",
			want:     &allPackages[0],
		},
		{
			name:     "Does not return package for a nil currentVersion",
			settings: flakeSettings,
			wantErr:  false,
			pkgPath:  "/a/non-versioned/path",
			want:     nil,
		},
		{
			name:     "Does not return package for an empty string currentVersion",
			settings: flakeSettings,
			wantErr:  false,
			pkgPath:  "/another/non-versioned/path",
			want:     nil,
		},
		{
			name:     "Does not return package in PackageDetails.Ignore",
			settings: flakeSettings,
			wantErr:  false,
			pkgPath:  "/ignore/path",
			want:     nil,
		},
		{
			name:     "Overrides Name in PackageDetails.Override",
			settings: flakeSettings,
			wantErr:  false,
			pkgPath:  "/override/path",
			want:     overridePkg,
		},
		{
			name: "Handles errors from getPackageFromPath",
			//getPackageFromPath is the first GetNixData method to get hit, so "allErrors" will fire from there first
			settings: allErrorsFlakeSettings,
			wantErr:  true,
			pkgPath:  "",
		},
		{
			name: "Handles errors from getUnstableVersion",
			//We need to get deep enough getUnstableVersion to fire, so we need a different FlakeSettings
			settings: unstableFlakeSettings,
			wantErr:  true,
			pkgPath:  "/a/good/path",
			want:     &allPackages[0],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			got, err := tt.settings.GetPackage(tt.pkgPath)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPackage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Expected GetPackage(%s) to be '%v', was actually, '%v'", tt.pkgPath, tt.want, got)
			}
		})
	}
}
