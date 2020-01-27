package lib

import (
	"reflect"
	"regexp"
	"strings"
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
	"_sysconfdir":     "/etc",
	"_prefix":         "/usr",
	"_exec_prefix":    "%{_prefix}",
	"_includedir":     "%{_prefix}/include",
	"_bindir":         "%{_exec_prefix}/bin",
	"_libdir":         "%{_exec_prefix}/%{_lib}",
	"_libexecdir":     "%{_exec_prefix}/libexec",
	"_sbindir":        "%{_exec_prefix}/sbin",
	"_datadir":        "%{_datarootdir}",
	"_infodir":        "%{_datarootdir}/info",
	"_mandir":         "%{_datarootdir}/man",
	"_docdir":         "%{_datadir}/doc",
	"_rundir":         "/run",
	"_localstatedir":  "/var",
	"_sharedstatedir": "/var/lib",
	"_lib":            "lib",
}

func evalInlineMacros(input string, context PackageContext) string {
	mutate := input

	// This regex will match data inside %{data}
	grabMacro := regexp.MustCompile(`%{(.+?)}`)

	timesLooped := 0

macroLoop:
	for _, match := range grabMacro.FindAll([]byte(mutate), -1) {
		// Let's turn our match into a string...
		matchString := string(match)

		// ... and remove the %{} from %{data} to get data
		matchContent := strings.TrimPrefix(strings.TrimSuffix(matchString, "}"), "%{")

		// Loop through the fields of the package context in order to see if any of the annotated macros are in this line
		fields := reflect.TypeOf(context)
		num := fields.NumField()

		for i := 0; i < num; i++ {
			field := fields.Field(i)
			if strings.ToLower(matchContent) == field.Tag.Get("macro") {
				// We assert that packageContext only has string fields here.
				// If it doesn't, our code will break.
				key := reflect.ValueOf(&context).Elem().FieldByName(field.Name)
				if key.IsValid() {
					mutate = strings.ReplaceAll(mutate, matchString, key.String())
				}
			}
		}

		// Now let's see if our hardcoded macros have anything in store for us
		for macro, value := range macros {
			if strings.ToLower(matchContent) == macro {
				mutate = strings.ReplaceAll(mutate, matchString, value)
			}
		}
	}

	// We want to keep going over macros until there's no more remaining
	// But not for forever, as we don't want to trip over invalid macros
	if len(grabMacro.FindAll([]byte(mutate), -1)) != 0 && timesLooped < 256 {
		timesLooped++
		goto macroLoop
	}

	return mutate
}
