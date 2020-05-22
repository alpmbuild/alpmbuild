package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/appadeia/alpmbuild/lib/libalpm"
)

type BuildInfo struct {
	System struct {
		Arch      string
		OS        string
		GoVersion string
		Packages  []libalpm.Package
	}
	SpecFile      []byte
	Package       PackageContext
	ParentPackage *PackageContext `json:",omitempty"`
}

func (pkg PackageContext) GenerateBuildInfo() {
	parent := pkg
	different := false
	outputStatus("Generating build info for " + highlight(pkg.GetNevra()) + "...")
	for parent.parentPackage != nil {
		parent = *pkg.parentPackage
		different = true
	}
	pkgdir := ""
	home, err := os.UserHomeDir()
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate build info:\n%s", err.Error()))
	}
	if !pkg.IsSubpackage {
		pkgdir = filepath.Join(home, "alpmbuild/package")
	} else {
		pkgdir = filepath.Join(home, "alpmbuild/subpackages", pkg.GetNevra())
	}
	os.Chdir(pkgdir)
	pkgs, err := libalpm.ListInstalled()
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate build info:\n%s", err.Error()))
	}
	buildInfo := BuildInfo{}
	buildInfo.System = struct {
		Arch      string
		OS        string
		GoVersion string
		Packages  []libalpm.Package
	}{
		Arch:      runtime.GOARCH,
		OS:        runtime.GOOS,
		GoVersion: runtime.Version(),
		Packages:  pkgs,
	}
	if different {
		buildInfo.ParentPackage = &parent
	}
	buildInfo.Package = pkg
	buildInfo.SpecFile = rawdata
	data, err := json.MarshalIndent(buildInfo, "", "\t")
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate build info:\n%s", err.Error()))
	}
	err = ioutil.WriteFile(filepath.Join(pkgdir, ".ALPMBUILD_BUILDINFO"), []byte(data), 0644)
	if err != nil {
		outputError(fmt.Sprintf("Failed to generate build info:\n%s", err.Error()))
	}
}
