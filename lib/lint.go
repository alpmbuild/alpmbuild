package lib

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/appadeia/alpmbuild/lib/libalpm"
)

var pkgNames []string
var pkgNamesInitted bool = false
var groupNames []string
var groupNamesInitted bool = false

type BadNameReason int

const (
	ValidName BadNameReason = iota
	EmptyName
	StartsWithHyphen
	StartsWithDot
	NonASCIICharacter
	NonAlphanumericCharacters
)

func lintPackageName(name string) (BadNameReason, int) {
	if name == "" {
		return EmptyName, -1
	}
	if strings.HasPrefix(name, "-") {
		return StartsWithHyphen, 0
	}
	if strings.HasPrefix(name, ".") {
		return StartsWithDot, 0
	}
	for i := 0; i < len(name); i++ {
		if name[i] > unicode.MaxASCII {
			return NonASCIICharacter, strings.Index(name, string(name[i]))
		}
	}
	isNonAlphaNum := regexp.MustCompile(`[^a-zA-Z0-9 +_.@-]{1,255}`)
	if matches := isNonAlphaNum.FindAllString(name, -1); len(matches) > 0 {
		return NonAlphanumericCharacters, strings.Index(name, matches[0])
	}
	return ValidName, -1
}

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

func lintGroup(name string) (string, bool) {
	if !groupNamesInitted {
		groupNames = libalpm.ListGroups()
		groupNamesInitted = true
	}
	for _, groupName := range groupNames {
		if groupName == name {
			return "", false
		}
	}
	return ClosestString(name, groupNames), true
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
		return
	} else if strings.Contains(text, "l") {
		outputStatus(strings.Join(missingDeps, " "))
		os.Exit(0)
	} else if strings.Contains(text, "a") {
		abort()
	}
	abort()
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
