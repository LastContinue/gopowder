package nix

import (
	"fmt"
	"sort"
	"strings"
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

func (p Package) isNonVersioned() bool {
	return p.CurrentVersion == nil || strings.TrimSpace(*p.CurrentVersion) == ""
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

	packages := make([]Package, 0, len(paths))
	results := make(chan *Package)
	// There's an off chance we fill in more (unexpected) errors than we can drain
	// so buffer up to number of paths + 1
	errs := make(chan error, len(paths)+1)

	var wg sync.WaitGroup

	for _, path := range paths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			pkg, err := packageFn(p)
			if err != nil {
				errs <- err
				return
			}
			results <- pkg
			if err := progressMarker.inc(1); err != nil {
				errs <- err
			}
		}(path)
	}

	go func() {
		wg.Wait()
		close(results)
		close(errs)
	}()

	for pkg := range results {
		if pkg != nil {
			packages = append(packages, *pkg)
		}
	}

	for err := range errs {
		if err != nil {
			return nil, err
		}
	}

	if err := progressMarker.close(); err != nil {
		return nil, err
	}

	return packages, nil
}
