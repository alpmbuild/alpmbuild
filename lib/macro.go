package lib

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/appadeia/alpmbuild/lib/librpm"
)

/*
   alpmbuild — a tool to build arch packages from RPM specfiles

   Copyright (C) 2020  Carson Black

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

var macros = map[string]string{
	"alpmbuild": version,
}

var expanded = false

func getExtractCommandForName(name string, quiet bool) string {
	if strings.HasSuffix(name, ".zip") {
		if quiet {
			return "%{__unzip} -qq %{_sourcedir}/" + path.Base(name)
		}
		return "%{__unzip} %{_sourcedir}/" + path.Base(name)
	}
	if quiet {
		return "%{__tar} -xf %{_sourcedir}/" + path.Base(name)
	}
	return "%{__tar} -xvvf %{_sourcedir}/" + path.Base(name)
}

func evalInlineMacros(input string, context PackageContext) string {
	mutate := input

	if !expanded {
		for macro, expandTo := range macros {
			librpm.DefineMacro(macro+" "+expandTo, 256)
		}
		home, err := os.UserHomeDir()
		if err != nil {
			outputError("Could not get user's home directory.")
		}
		librpm.LoadFromFile("/usr/lib/rpm/macros")
		librpm.DefineMacro(fmt.Sprintf("buildroot %s", filepath.Join(home, "alpmbuild/package")), 0)
		librpm.DefineMacro(fmt.Sprintf("_sourcedir %s", filepath.Join(home, "alpmbuild/sources")), 0)
	}
	if context.Name != "" {
		librpm.DefineMacro("name "+context.Name, 0)
	}
	if context.Version != "" {
		librpm.DefineMacro("version "+context.Version, 0)
	}
	if context.Name != "" && context.Version != "" {
		librpm.DefineMacro(fmt.Sprintf("buildsubdir %s-%s", context.Name, context.Version), 0)
	}
	if strings.Contains(input, "%setup") {
		set := flag.NewFlagSet("setup", flag.ContinueOnError)

		createDir := set.Bool("c", false, "")
		doNotDeleteDirectory := set.Bool("D", false, "")
		unpackQuietly := set.Bool("q", false, "")
		skipDefaultSource := set.Bool("T", false, "")

		set.Parse(strings.Split(input, " ")[1:])

		script := ""

		if !*doNotDeleteDirectory {
			script += "rm -rf %{buildsubdir}\n"
		}
		if *createDir {
			script += `mkdir -p %{buildsubdir}
cd %{buildsubdir}` + "\n"
		} else if !*skipDefaultSource {
			script += fmt.Sprintf("%s\n", getExtractCommandForName(context.Sources[0], *unpackQuietly))
		}

		if !*createDir {
			script += "cd %{buildsubdir}\n"
		} else if !*skipDefaultSource {
			script += fmt.Sprintf("%s\n", getExtractCommandForName(context.Sources[0], *unpackQuietly))
		}

		for _, source := range context.Sources[1:] {
			script += fmt.Sprintf("%s\n", getExtractCommandForName(source, *unpackQuietly))
		}

		mutate = script
	}

	return librpm.ExpandMacro(mutate)
}
