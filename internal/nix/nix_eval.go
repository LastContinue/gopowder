package nix

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

// This exists to facilitate a GetsNixData.
// Every method here is tied to the underlying Nix system
// so testing might be "interesting" - I suppose you could do _another_ interface
// for the nixExpr but...
type NixEval struct{}

func (n NixEval) getPackagePaths(path string, configName string) ([]string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	nixExpr := fmt.Sprintf("%s/#darwinConfigurations.\"%s\".config.environment.systemPackages", path, configName)
	cmd := exec.Command("nix", "eval", "--json", nixExpr)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nix eval failed: %w: %s", err, stderr.String())
	}

	var paths []string
	err := json.Unmarshal(stdout.Bytes(), &paths)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w: %s", err, stdout.String())
	}

	return paths, nil
}

func (n NixEval) getPackageFromPath(path string) (*Package, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	nixExpr := fmt.Sprintf(`
        let
            sp = %s;
            base = builtins.baseNameOf sp;

            rest = (let m = builtins.match "^[^-]+-(.*)$" base;
                    in if m == null then base else builtins.elemAt m 0);

            parsed = builtins.parseDrvName rest;

            name =
            if parsed ? name
            then parsed.name
            else rest;

            current_version =
            if parsed ? version && parsed.version != null
            then parsed.version
            else "unknown";
        in
        {
            name = name;
            current_version = current_version;
        }
        `, path)

	cmd := exec.Command("nix", "eval", "--json", "--expr", nixExpr)

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nix eval failed: %w: %s", err, stderr.String())
	}

	var pkg Package
	if err := json.Unmarshal(stdout.Bytes(), &pkg); err != nil {
		return nil, fmt.Errorf("json unmarshal failed: %w: %s", err, stdout.String())
	}

	return &pkg, nil
}

func (n NixEval) getUnstableVersion(packageName string) (*string, error) {

	nixExpr := fmt.Sprintf("github:NixOS/nixpkgs/nixpkgs-unstable#%s.version", packageName)
	cmd := exec.Command("nix", "eval", "--raw", nixExpr)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nix eval failed: %w: %s", err, stderr.String())
	}

	return new(stdout.String()), nil
}
