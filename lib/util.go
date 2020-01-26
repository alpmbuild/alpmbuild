package lib

import (
	"encoding/json"
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

func containsInsensitive(larger, substring string) bool {
	return strings.Contains(strings.ToLower(larger), strings.ToLower(substring))
}

func grabFlagFromString(parent, grabFlag string, dontGrabFlags []string) (string, bool) {
	splitString := strings.Split(parent, " ")
	returnValue := ""

	for index, splitWord := range splitString {
		if strings.HasPrefix(splitWord, "-"+grabFlag) {
			flagValue := ""
		parentFlagLoop:
			for _, ii := range splitString[index+1:] {
				for _, dontGrab := range dontGrabFlags {
					if !strings.HasPrefix(ii, "-"+dontGrab) {
						flagValue = flagValue + " " + ii
					} else {
						break parentFlagLoop
					}
				}
			}
			if strings.TrimSpace(flagValue) != "" {
				returnValue = strings.TrimSpace(flagValue)
			}
		}
	}

	if returnValue != "" {
		return returnValue, true
	} else {
		return "", false
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
