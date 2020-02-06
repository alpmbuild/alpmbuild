package lib

import "github.com/appadeia/alpmbuild/lib/librpm"

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

func evalInlineMacros(input string, context PackageContext) string {
	if !expanded {
		for macro, expandTo := range macros {
			librpm.DefineMacro(macro+" "+expandTo, 256)
		}
		librpm.LoadFromFile("/usr/lib/rpm/macros")
	}
	if context.Name != "" {
		librpm.DefineMacro("name "+context.Name, 0)
	}
	if context.Version != "" {
		librpm.DefineMacro("version "+context.Version, 0)
	}
	mutate := input
	return librpm.ExpandMacro(mutate)
}
