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

func evalInlineMacros(input string, context PackageContext) string {
	mutate := input

	// This regex will match data inside %{data}
	grabMacro := regexp.MustCompile(`%{(.+?)}`)

	for _, match := range grabMacro.FindAll([]byte(input), -1) {
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
	}

	return mutate
}
