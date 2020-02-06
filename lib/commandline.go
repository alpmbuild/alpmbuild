package lib

import (
	"flag"
	"os"
	"os/exec"
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

var checkFiles *bool
var hideCommandOutput *bool
var useColours *bool
var generateSourcePackage *bool
var buildFile *string
var startPWD string
var compressionType *string

type arrayFlag []string

func (i *arrayFlag) String() string {
	return strings.Join(*i, " ")
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Enter : Main function of alpmbuild's command line interface
func Enter() {
	// Alpmbuild-unique flags
	buildFile = flag.String("file", "", "The file to build.")
	checkFiles = flag.Bool("strictFiles", true, "Strictly check %files for the main package")
	hideCommandOutput = flag.Bool("hideCommandOutput", false, "Hide package command output")
	useColours = flag.Bool("useColours", true, "Use colours for output.")
	generateSourcePackage = flag.Bool("generateSourcePackage", true, "Generate a source package")
	compressionType = flag.String("compression", "zstd", "The compression type to use. Default is zstd. Choose from: gz, xz, bz2, or zstd.")
	fakeroot := flag.Bool("fakeroot", false, "Internal flag. Do not set.")

	flag.Parse()

	var macros arrayFlag

	// Flags that mimic behaviour of rpmbuild
	ba := flag.String("ba", "", "Copies rpmbuild -ba's behaviour")
	flag.Var(&macros, "D", "Define a macro with MACRO EXPR")
	flag.Var(&macros, "define", "Define a macro with MACRO EXPR")

	if !*fakeroot {
		cmd := exec.Command("fakeroot", append(os.Args, "-fakeroot")...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
		return
	}

	if _, ok := CompressionTypes[*compressionType]; !ok {
		outputError(*compressionType + " is not a valid compression method.")
	}

	var err error
	startPWD, err = os.Getwd()
	if err != nil {
		outputError("There was an error getting the current working directory:\n\t" + err.Error())
	}

	if *ba != "" {
		*buildFile = *ba
		*generateSourcePackage = true
	}

	for _, macro := range macros {
		librpm.DefineMacro(macro, 256)
	}

	if *buildFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	err = Build(*buildFile)
	if err != nil {
		println(err.Error())
	}
}
