package lib

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/appadeia/alpmbuild/lib/libalpm"
)

var pkgNames []string
var pkgNamesInitted bool = false

func lintDependency(name string) (string, bool) {
	if !pkgNamesInitted {
		pkgNames, _ = libalpm.ListPackagesAsString(libalpm.PackageName)
		pkgNamesInitted = true
	}
	for _, pkgName := range pkgNames {
		if pkgName == name {
			return "", false
		}
	}
	return ClosestString(name, pkgNames), true
}

func promptMissingDepsInstall(pkg PackageContext) {
	if *fakeroot || *ignoreDeps {
		return
	}

	var missingDeps []string

	for _, pkg := range append(pkg.Requires, pkg.BuildRequires...) {
		if !libalpm.PackageInstalled(pkg) {
			missingDeps = append(missingDeps, pkg)
		}
	}

	println(red("==>"), highlight(strconv.Itoa(len(missingDeps))), bold("package(s) need installation to build"), highlight(pkg.Name)+bold(":"))
	println(bold("    Actions: Install (i), List (l), Abort (default: a)"))
	println()
	print("  " + bold("-> "))

	reader := bufio.NewReader(os.Stdin)

	abort := func() {
		outputError("Cannot build package without dependencies, aborting...")
		os.Exit(1)
	}

	text, err := reader.ReadString('\n')
	if err != nil {
		abort()
	}
	text = strings.ReplaceAll(text, "\n", "")

	if strings.Contains(text, "i") {
		cmd := exec.Command("pkexec", append([]string{"pacman", "-S"}, missingDeps...)...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		if cmd.ProcessState.ExitCode() != 0 {
			abort()
		}
	} else if strings.Contains(text, "l") {
		outputStatus(strings.Join(missingDeps, " "))
		os.Exit(0)
	} else if strings.Contains(text, "a") {
		abort()
	}
}

//* Built package linting

func (pkg PackageContext) trimPath(in string) string {
	return strings.TrimPrefix(in, pkg.PackageRoot())
}

func (pkg PackageContext) lintAll() {
	outputStatus("Linting package " + highlight(pkg.GetNevra()) + "...")
	pkg.lintForReferencesToBuildDirectory()
	pkg.lintForDotfilesInPackageRoot()
	pkg.lintForNewlinesInFilenames()
}

func (pkg PackageContext) lintForReferencesToBuildDirectory() {
	home, err := os.UserHomeDir()
	if err != nil {
		outputError("Could not get user's home directory.")
	}
	err = filepath.Walk(
		pkg.PackageRoot(),
		func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			content, err := ioutil.ReadFile(path)
			if err != nil {
				outputError("Failed to open file " + highlight(pkg.trimPath(path)) + " in package " + highlight(pkg.GetNevra()) + ": " + err.Error())
			}
			strContent := string(content)
			if strings.Contains(strContent, filepath.Join(home, "alpmbuild")) {
				outputWarning(
					fmt.Sprintf(
						"Package %s contains a reference to the build directory in file %s",
						highlight(pkg.GetNevra()),
						highlight(pkg.trimPath(path)),
					),
				)
			}
			return nil
		},
	)
	if err != nil {
		outputError(err.Error())
	}
}

func (pkg PackageContext) lintForDotfilesInPackageRoot() {
	files, _ := ioutil.ReadDir(pkg.PackageRoot())
	for _, file := range files {
		if strings.HasPrefix(file.Name(), ".") {
			outputWarning(
				fmt.Sprintf(
					"Package %s contains a dotfile %s in the package root",
					highlight(pkg.GetNevra()),
					highlight(pkg.trimPath(path.Join(pkg.PackageRoot(), file.Name()))),
				),
			)
		}
	}
}

func (pkg PackageContext) lintForNewlinesInFilenames() {
	err := filepath.Walk(
		pkg.PackageRoot(),
		func(path string, info os.FileInfo, err error) error {
			if strings.Contains(info.Name(), "\n") {
				outputError(
					fmt.Sprintf(
						"Package %s has paths with a newline: %s",
						highlight(pkg.GetNevra()),
						highlight(pkg.trimPath(path)),
					),
				)
			}
			return nil
		},
	)
	if err != nil {
		outputError(err.Error())
	}
}
