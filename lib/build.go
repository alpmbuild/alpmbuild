package lib

import (
	"io/ioutil"
	"reflect"
	"strconv"
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

func ParsePackage(data string) PackageContext {
	outputStatus("Parsing package...")

	lex := PackageContext{}

	lex.Subpackages = make(map[string]PackageContext)

	currentStage := NoStage
	ifStage := NoStage
	currentSubpackage := ""
	currentFilesSubpackage := ""

mainParseLoop:
	for currentLine, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		// Blank lines are useless to us.
		if line == "" {
			continue
		}

		// If we're currently in an if statement that's false, we want to ignore
		// everything else and go straight to checking for ifs
		if ifStage == IfFalseStage {
			goto Conditionals
		}

		// Let's look at #!alpmbuild directives
		if strings.HasPrefix(line, "#!alpmbuild") {
			fields := strings.Fields(line)

			if len(fields) >= 2 {
				switch fields[1] {
				case "NoFileCheck":
					*checkFiles = false
					continue mainParseLoop
				}
			}
		}

		// Let's parse the key-value lines
		if strings.Contains(line, ": ") {
			// Split our line by whitespace
			words := strings.Fields(line)

			// Because we need at least two values for a Key: Value PAIR, make
			// sure we have at least two values
			if len(words) < 2 {
				continue mainParseLoop
			}

			// Let's worry about our sources first...
			if strings.HasPrefix(strings.ToLower(words[0]), "source") {
				lex.Sources = append(lex.Sources, evalInlineMacros(words[1], lex))
				continue mainParseLoop
			}

			// And then our patches...
			if strings.HasPrefix(strings.ToLower(words[0]), "patch") {
				lex.Patches = append(lex.Patches, evalInlineMacros(words[1], lex))
				continue mainParseLoop
			}

			// Now we produce macros based off our package context.

			fields := reflect.TypeOf(lex)

			// If we're currently operating on a subpackage, then we want to write
			// changes to that subpackage.
			if currentSubpackage != "" {
				fields = reflect.TypeOf(lex.Subpackages[currentSubpackage])
			}

			currentPackage := lex
			if currentSubpackage != "" {
				currentPackage = lex.Subpackages[currentSubpackage]
			}

			num := fields.NumField()

			hasSet := false

			// Loop through the fields of the package context in order to see if any of the annotated key values match the line we're on
			for i := 0; i < num; i++ {
				field := fields.Field(i)
				// These are the single-value keys, such as Name, Version, Release, and Summary
				if strings.ToLower(words[0]) == field.Tag.Get("key") {
					// We assert that packageContext only has string fields here.
					// If it doesn't, our code will break.
					key := reflect.ValueOf(&currentPackage).Elem().FieldByName(field.Name)
					if key.IsValid() {
						key.SetString(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex))
						hasSet = true
					}
				}
				// These are the multi-value keys, such as Requires and BuildRequires
				if strings.ToLower(words[0]) == field.Tag.Get("keyArray") {
					// We assume that packageContext only has string array fields here.
					// If it doesn't, our code will break.
					key := reflect.ValueOf(&currentPackage).Elem().FieldByName(field.Name)
					if key.IsValid() {
						itemArray := strings.Split(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex), " ")

						key.Set(reflect.AppendSlice(key, reflect.ValueOf(itemArray)))
						hasSet = true
					}
				}
			}

			if currentSubpackage != "" {
				lex.Subpackages[currentSubpackage] = currentPackage
			} else {
				lex = currentPackage
			}

			if !hasSet {
				outputErrorHighlight(
					highlight(words[0])+" is not a valid key on line "+strconv.Itoa(currentLine+1),
					line,
					"Did you mean to use "+highlight(ClosestString(words[0], PossibleKeys))+"?",
					strings.Index(line, words[0]), len(words[0]),
				)
			}
			continue mainParseLoop
		}

		// How about some conditional stuff?
	Conditionals:
		if strings.Contains(line, "%if") ||
			strings.Contains(line, "%elif") ||
			strings.Contains(line, "%else") ||
			strings.Contains(line, "%endif") {

			fields := strings.Fields(line)
			switch fields[0] {
			case "%endif":
				ifStage = NoStage
				continue mainParseLoop
			case "%if":
				if evalIf(line, lex, currentLine+1) {
					ifStage = IfTrueStage
				} else {
					ifStage = IfFalseStage
				}
				continue mainParseLoop
			case "%else":
				if ifStage == IfTrueStage {
					ifStage = IfFalseStage
				} else {
					ifStage = IfTrueStage
				}
				continue mainParseLoop
			case "%elseif":
				if ifStage == IfTrueStage {
					ifStage = IfFalseStage
				} else {
					if evalIf(line, lex, currentLine+1) {
						ifStage = IfTrueStage
					} else {
						ifStage = IfFalseStage
					}
				}
				continue mainParseLoop
			}
		}

		// If we're here because it's an if false stage,
		// let's skip to the next iteration
		if ifStage == IfFalseStage {
			continue mainParseLoop
		}

		// We need to be able to handle subpackages
		if strings.Contains(line, "%package") {
			if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
				currentSubpackage = subpackageName
			} else {
				if splitString := strings.Split(line, " "); len(splitString) >= 2 {
					currentSubpackage = lex.Name + "-" + splitString[1]
				} else {
					outputError("%package needs to have a name")
				}
			}
			if currentSubpackage != "" {
				lex.Subpackages[currentSubpackage] = PackageContext{
					Name:          currentSubpackage,
					IsSubpackage:  true,
					parentPackage: &lex,
				}
			}
			continue mainParseLoop
		}

		// Time for the sections!
		{
			// Let's make sure we don't accidentally include directives in
			// commands
			if isStringInSlice(line, otherDirectives) {
				currentStage = NoStage
				continue
			}

			// Now we check to see if we're switching to a new stage
			if strings.HasPrefix(line, "%prep") {
				currentStage = PrepareStage
				continue mainParseLoop
			}
			if strings.HasPrefix(line, "%build") {
				currentStage = BuildStage
				continue mainParseLoop
			}
			if strings.HasPrefix(line, "%install") {
				currentStage = InstallStage
				continue mainParseLoop
			}
			if strings.HasPrefix(line, "%files") {
				if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
					currentFilesSubpackage = subpackageName
				} else {
					if splitString := strings.Split(line, " "); len(splitString) >= 2 {
						currentFilesSubpackage = lex.Name + "-" + splitString[1]
					} else {
						currentFilesSubpackage = ""
					}
				}
				currentStage = FileStage
				continue mainParseLoop
			}

			// If we're in a stage, we want to append some commands to our list
			switch currentStage {
			case PrepareStage:
				lex.Commands.Prepare = append(lex.Commands.Prepare, evalInlineMacros(line, lex))
				continue mainParseLoop
			case BuildStage:
				lex.Commands.Build = append(lex.Commands.Build, evalInlineMacros(line, lex))
				continue mainParseLoop
			case InstallStage:
				lex.Commands.Install = append(lex.Commands.Install, evalInlineMacros(line, lex))
				continue mainParseLoop
			case FileStage:
				if currentFilesSubpackage == "" {
					lex.Files = append(lex.Files, evalInlineMacros(line, lex))
				} else {
					if val, ok := lex.Subpackages[currentFilesSubpackage]; ok {
						subpkg := val
						subpkg.Files = append(subpkg.Files, evalInlineMacros(line, lex))
						lex.Subpackages[currentFilesSubpackage] = subpkg
					} else {
						outputError("You cannot specify files for a subpackage if the subpackage has not been declared")
					}
				}
				continue mainParseLoop
			}
		}

		// Comment time!
		if strings.HasPrefix(line, "#") {
			continue mainParseLoop
		}

		// If we got to here without continuing, something's wrong.
		outputError("Could not parse line " + strconv.Itoa(currentLine+1) + ":\n          " + line)
	}

	lex.BuildPackage()

	return lex
}

// Build : Build a specfile, generating an Arch package.
func Build(pathToRecipe string) error {
	outputStatus("Reading specfile from " + pathToRecipe + "...")
	data, err := ioutil.ReadFile(pathToRecipe)
	if err != nil {
		return err
	}
	ParsePackage(string(data))
	return nil
}
