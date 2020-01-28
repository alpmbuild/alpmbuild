package lib

import (
	"flag"
	"os"
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

// Enter : Main function of alpmbuild's command line interface
func Enter() {
	file := flag.String("file", "", "The file to build.")
	checkFiles = flag.Bool("strictFiles", true, "Strictly check %files for the main package")
	hideCommandOutput = flag.Bool("hideCommandOutput", false, "Hide package command output")
	useColours = flag.Bool("useColours", true, "Use colours for output.")

	flag.Parse()

	if *file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := Build(*file)
	if err != nil {
		println(err.Error())
	}
}
