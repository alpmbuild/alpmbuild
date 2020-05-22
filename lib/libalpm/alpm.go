package libalpm

import (
	"os/exec"
	"strings"
)

type Package struct {
	Repository string `json:",omitempty"`
	Name       string
	Version    string
}

type PackageField int

const (
	PackageRepository PackageField = iota
	PackageName
	PackageVersion
)

func ListInstalled() (pkgs []Package, err error) {
	cmd := exec.Command("pacman", "-Q")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	str := string(output)
	for _, line := range strings.Split(strings.TrimSpace(str), "\n") {
		split := strings.Fields(line)
		pkgs = append(pkgs, Package{
			Name:    split[0],
			Version: split[1],
		})
	}
	return
}

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

func ListGroups() []string {
	cmd := exec.Command("pacman", "-Sg")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return []string{}
	}
	outputStr := strings.TrimSpace(string(output))
	var groups []string
	for _, line := range strings.Split(outputStr, "\n") {
		groups = append(groups, line)
	}
	return groups
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
