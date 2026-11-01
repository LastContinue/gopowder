package nix

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

type FlakeSettings struct {
	FlakeDir        string          `toml:"flake_dir"`
	FlakeConfigName string          `toml:"flake_config_name"`
	PackageDetails  *PackageDetails `toml:"pkgs"`
	DataGenerator   GetsNixData
}

type PackageDetails struct {
	Ignore    []string
	Overrides map[string]string
}

func (f FlakeSettings) Path() string {
	return f.FlakeDir
}

func (f FlakeSettings) ConfigName() string {
	return f.FlakeConfigName
}

func (f FlakeSettings) GetPackagePaths() ([]string, error) {
	return f.DataGenerator.getPackagePaths(f.FlakeDir, f.FlakeConfigName)
}

func (f FlakeSettings) GetPackage(path string) (*Package, error) {

	pkg, err := f.DataGenerator.getPackageFromPath(path)
	if err != nil {
		return nil, err
	}

	// Don't look for updates for ignored packages. Additionally,
	// some packages, things like darwin-help, darwin-manpages, etc
	// don't have a version and can't be fetched from Nixpkgs (and therefor
	// cannot be updated by the method this app uses)
	// In that case, just skip them. If you're not seeing packages (that you know
	// you have installed) in the output table, this could be why
	if !slices.Contains(f.PackageDetails.Ignore, pkg.Name) && !pkg.isNonVersioned() {

		lookupName := pkg.Name
		overrideName, ok := f.PackageDetails.Overrides[pkg.Name]
		if ok {
			pkg.Name = fmt.Sprintf("%s (%s)", pkg.Name, overrideName)
			lookupName = overrideName
		}

		unstableVersion, err := f.DataGenerator.getUnstableVersion(lookupName)
		if err != nil {
			return nil, err
		}

		if *unstableVersion != *pkg.CurrentVersion {
			pkg.Update = unstableVersion
		}

		return pkg, nil
	}

	return nil, nil
}

func NewFlakeSettings(configPath string,
	decoder Decoder,
	dataGenerator GetsNixData,
) (*FlakeSettings, error) {

	var flakeSettings FlakeSettings

	// Not a big fan of passing by reference and doing changes,
	// but I needed a way to decouple reading a file from creating
	// a FlakeSettings
	err := decoder.DecodeIntoFlakeSettings(configPath, &flakeSettings)
	if err != nil {
		return nil, err
	}

	flakeSettings.FlakeDir = os.ExpandEnv(flakeSettings.FlakeDir)
	flakeSettings.FlakeConfigName = os.ExpandEnv(flakeSettings.FlakeConfigName)

	if flakeSettings.FlakeConfigName == "" {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd := exec.Command("scutil", "--get", "LocalHostName")

		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("hostname failed: %w: %s", err, stderr.String())
		}
		flakeSettings.FlakeConfigName = strings.TrimSpace(stdout.String())
	}

	flakeSettings.DataGenerator = dataGenerator

	return &flakeSettings, nil
}
