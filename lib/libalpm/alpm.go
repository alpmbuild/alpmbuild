package libalpm

import (
	"os/exec"
	"strings"
)

type Package struct {
	Repository string
	Name       string
	Version    string
}

type PackageField int

const (
	PackageRepository PackageField = iota
	PackageName
	PackageVersion
)

func ListPackages() ([]Package, error) {
	cmd := exec.Command("pacman", "-Sl")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []Package{}, err
	}
	str := string(output)
	var pkgs []Package
	for _, line := range strings.Split(strings.TrimSpace(str), "\n") {
		split := strings.Split(line, " ")
		pkgs = append(pkgs, Package{
			Repository: split[0],
			Name:       split[1],
			Version:    split[2],
		})
	}
	return pkgs, nil
}

func ListPackagesAsString(field PackageField) ([]string, error) {
	packages, err := ListPackages()
	var fields []string
	switch field {
	case PackageRepository:
		for _, pkg := range packages {
			fields = append(fields, pkg.Repository)
		}
	case PackageName:
		for _, pkg := range packages {
			fields = append(fields, pkg.Name)
		}
	case PackageVersion:
		for _, pkg := range packages {
			fields = append(fields, pkg.Version)
		}
	}
	return fields, err
}

func PackageInstalled(packageName string) bool {
	cmd := exec.Command("pacman", "-Qi", packageName)

	if err := cmd.Start(); err != nil {
		return false
	}

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			return exiterr.ExitCode() == 0
		}
	}

	return false
}
