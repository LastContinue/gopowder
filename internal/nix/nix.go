package nix

import (
	"fmt"
	"sort"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
)

type GetsPackages interface {
	Path() string
	ConfigName() string
	GetPackagePaths() ([]string, error)
	GetPackage(string) (*Package, error)
}

type GetsNixData interface {
	getPackagePaths(string, string) ([]string, error)
	getPackageFromPath(string) (*Package, error)
	getUnstableVersion(string) (*string, error)
}

type Package struct {
	Name           string  `json:"name"`
	CurrentVersion *string `json:"current_version"`
	Update         *string
}

type RunSummary struct {
	Packages    []Package
	UpdateNames []string
}

func (r RunSummary) ToTable() []table.Row {
	var rows []table.Row

	rows = append(rows, table.Row{"Name", "Current Version", "Update"})

	for _, pkg := range r.Packages {
		var update string
		if pkg.Update != nil {
			update = *pkg.Update
		}
		row := table.Row{
			pkg.Name, *pkg.CurrentVersion, green(update),
		}
		rows = append(rows, row)
	}

	return rows
}

func Run(flakeSettings GetsPackages, progressInitFn func(int) MarksProgress) (*RunSummary, error) {

	spinnerMsg := fmt.Sprintf(
		" Querying Darwin Flake\n    • location:'%s'\n    • config:  '%s' \n",
		flakeSettings.Path(),
		flakeSettings.ConfigName(),
	)

	finalMsg := fmt.Sprintf(
		// "yes", spacing is different on this line due to lack of spinner
		"   Querying Darwin Flake ✅\n    • location:'%s'\n    • config:  '%s' \n",
		flakeSettings.Path(),
		flakeSettings.ConfigName(),
	)

	// Good argument could be made for taking this out of here an moving it it into main...
	paths, err := spinWhile(spinnerMsg, finalMsg, func() ([]string, error) {
		return flakeSettings.GetPackagePaths()
	})
	if err != nil {
		return nil, err
	}

	pm := progressInitFn(len(paths))

	// goroutine logic for this was a bit long, so separated it out
	packages, err := goGetPackages(paths, flakeSettings.GetPackage, pm)
	if err != nil {
		return nil, err
	}

	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Name < packages[j].Name
	})

	updateNames := make([]string, 0)
	for _, p := range packages {
		if p.Update != nil {
			updateNames = append(updateNames, p.Name)
		}
	}

	return &RunSummary{
		Packages:    packages,
		UpdateNames: updateNames,
	}, nil
}

/**
 * Get Nix packages using goroutines
 */
func goGetPackages(paths []string, packageFn func(string) (*Package, error), progressMarker MarksProgress) ([]Package, error) {

	packages := make([]Package, 0)
	results := make(chan *Package)
	errors := make(chan error)

	var wg sync.WaitGroup

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			pkg, err := packageFn(p)
			if err != nil {
				errors <- err
				return
			}
			results <- pkg
		}(path)
	}

	go func() {
		wg.Wait()
		if err := progressMarker.close(); err != nil {
			errors <- err
		}
		close(results)
		close(errors)
	}()

	for pkg := range results {
		if pkg != nil {
			packages = append(packages, *pkg)
		}
		// We count progress regardless
		if err := progressMarker.inc(1); err != nil {
			errors <- err
		}
	}

	for err := range errors {
		if err != nil {
			return nil, err
		}
	}

	return packages, nil
}
