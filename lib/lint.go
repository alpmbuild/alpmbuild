package lib

import (
	"bufio"
	"os"
	"os/exec"
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
