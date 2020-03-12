package lib

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"strconv"
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

func ParsePackage(data string) PackageContext {
	if !*fakeroot {
		outputStatus("Parsing package...")
	}

	data = strings.ReplaceAll(data, "\\\n", "")

	lex := PackageContext{}

	lex.Subpackages = make(map[string]PackageContext)
	lex.Reasons = make(map[string]string)

	currentStage := NoStage
	ifStage := NoStage
	currentSubpackage := ""
	currentFilesSubpackage := ""
	currentChangelogSubpackage := ""
	currentScriptletSubpackage := ""

mainParseLoop:
	for currentLine, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		// Blank lines are useless to us.
		if line == "" {
			continue
		}

		// Let's see if any of our macros don't expand...
		{
			// This regex will match data inside %{data}
			expanded := evalInlineMacros(line, lex)
			grabMacroRegex := regexp.MustCompile(`%{(.+?)}`)
			for _, match := range grabMacroRegex.FindAll([]byte(expanded), -1) {
				matchString := string(match)

				macros := librpm.DumpMacroNamesAsString()

				if len(macros) > 0 {
					if !*fakeroot {
						outputWarningHighlight(
							"Macro not expanded on line "+strconv.Itoa(currentLine+1)+": "+highlight(matchString),
							line,
							"Did you mean to use "+highlight("%{"+ClosestString(matchString, macros)+"}")+"?",
							strings.Index(line, matchString), len(matchString),
						)
					}
				} else {
					if !*fakeroot {
						outputWarningHighlight(
							"Macro not expanded on line "+strconv.Itoa(currentLine+1)+": "+highlight(matchString),
							line,
							"",
							strings.Index(line, matchString), len(matchString),
						)
					}
				}
			}
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
				case "ReasonFor":
					if len(fields) < 3 {
						lineWithAdd := line + "          "
						outputErrorHighlight(
							"Not enough arguments to "+highlight("ReasonFor")+" on line "+strconv.Itoa(currentLine+1),
							lineWithAdd,
							"Add a package you want to give a reason for and the reason like this: "+highlight("PackageName: Reason"),
							strings.Index(lineWithAdd, "          "),
							len("          "),
						)
					}
					if len(fields) < 4 {
						lineWithAdd := line + "          "
						outputErrorHighlight(
							"No reason provided for "+highlight(fields[2])+" on line "+strconv.Itoa(currentLine+1),
							lineWithAdd,
							"Add a description why you want users to install this package",
							strings.Index(lineWithAdd, "          "),
							len("          "),
						)
					}
					if strings.Contains(line, ":") {
						split := strings.Split(line, ":")
						if len(split) < 2 {
							outputError("alpmbuild ran into an error state that should not be possible.\nPlease report this to https://github.com/appadeia/alpmbuild and attach a specfile.")
						}
						name := strings.Fields(split[0])[len(strings.Fields(split[0]))-1]
						lex.Reasons[name] = strings.TrimSpace(split[1])
						continue mainParseLoop
					} else {
						outputErrorHighlight(
							"ReasonFor missing "+highlight(":")+" on line "+strconv.Itoa(currentLine+1),
							line,
							"Add a "+highlight(":")+" after the package name",
							strings.Index(line, fields[2]), len(fields[2]),
						)
					}
				default:
					if !*fakeroot {
						outputWarningHighlight(
							"Invalid #!alpmbuild directive "+highlight(fields[1])+"on line "+strconv.Itoa(currentLine+1),
							line,
							"Did you mean to use "+highlight(ClosestString(fields[1], PossibleDirectives))+"?",
							strings.Index(line, fields[1]), len(fields[1]),
						)
					}
					continue mainParseLoop
				}
			}

			if !*fakeroot {
				outputWarningHighlight(
					"#!alpmbuild directive missing type",
					line,
					"",
					0, 0,
				)
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

			// Let's worry about our sources and patches...
			if strings.HasPrefix(strings.ToLower(words[0]), "source") || strings.HasPrefix(strings.ToLower(words[0]), "patch") {
				source := Source{
					URL: evalInlineMacros(words[1], lex),
				}
				if len(words) > 2 {
					slice := words[2:]
					for index, word := range slice {
						if word == "with" {
							if len(slice) > index+2 {
								hashType := slice[index+1]
								hashWord := slice[index+2]
								switch hashType {
								case "sha1":
									source.Sha1 = hashWord
								case "sha224":
									source.Sha224 = hashWord
								case "sha256":
									source.Sha256 = hashWord
								case "sha384":
									source.Sha384 = hashWord
								case "sha512":
									source.Sha512 = hashWord
								case "md5":
									source.Md5 = hashWord
								default:
									outputErrorHighlight(
										"Invalid hash type on line "+strconv.Itoa(currentLine+1),
										line,
										"Did you mean to use "+highlight(ClosestString(hashType, hashTypes))+"?",
										strings.Index(line, hashType), len(hashType),
									)
								}
							} else {
								if len(slice) > index+1 {
									outputErrorHighlight(
										"Incomplete hash directive on line "+strconv.Itoa(currentLine+1),
										line,
										"Add a hash to the end of the line to resolve this error.",
										0, 0,
									)
								} else {
									outputErrorHighlight(
										"Incomplete hash directive on line "+strconv.Itoa(currentLine+1),
										line,
										"Valid hash types are: "+strings.Join(hashTypes, ", "),
										0, 0,
									)
								}
							}
						}
						if word == "renamed" {
							if len(slice) > index+1 {
								source.Rename = evalInlineMacros(slice[index+1], lex)
							} else {
								outputErrorHighlight(
									"Incomplete rename directive on line "+strconv.Itoa(currentLine+1),
									line,
									"Provide a name to resolve this error.",
									strings.Index(line, word), len(word),
								)
							}
						}
					}
				}
				if strings.HasPrefix(strings.ToLower(words[0]), "source") {
					lex.Sources = append(lex.Sources, source)
				} else {
					lex.Sources = append(lex.Patches, source)
				}
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
				// Special case: VRE (Version, Release, Epoch)
				// Epoch:Version-Release
				if strings.ToLower(words[0]) == "epoverrel:" || strings.ToLower(words[0]) == "evr:" || strings.ToLower(words[0]) == "epochversionrelease:" {
					toLex := evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex)
					split := strings.FieldsFunc(toLex, func(r rune) bool {
						return strings.ContainsRune("-:", r)
					})
					if len(split) < 3 {
						outputErrorHighlight(
							"Invalid Epoch-Versions-Release string on line "+strconv.Itoa(currentLine+1),
							line,
							"Epoch-Versions-Release strings are in the following format: "+highlight("Epoch:Version-Release"),
							strings.Index(line, strings.TrimSpace(strings.TrimPrefix(line, words[0]))),
							len(strings.TrimSpace(strings.TrimPrefix(line, words[0]))),
						)
					}
					currentPackage.Epoch = split[0]
					currentPackage.Version = split[1]
					currentPackage.Release = split[2]
					hasSet = true
				}
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
						itemArray := strings.Fields(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex))
						if !*ignoreDeps && !*fakeroot {
							for _, packageField := range packageFields {
								if field.Tag.Get("keyArray") == packageField {
									for _, item := range itemArray {
										if err, _ := lintPackageName(item); err != ValidName {
											outputErrorHighlight(
												fmt.Sprintf(
													"%s is not a valid package identifier on line %s",
													highlight(item),
													strconv.Itoa(currentLine+1),
												),
												line,
												fmt.Sprintf(
													"Package identifiers can include %s, %s, %s, %s, %s, and %s",
													highlight("alphanumeric characters"),
													highlight("+"),
													highlight("_"),
													highlight("."),
													highlight("@"),
													highlight("-"),
												),
												strings.Index(line, item),
												len(item),
											)
										}
										if correction, needed := lintDependency(item); needed {
											outputWarningHighlight(
												fmt.Sprintf(
													"Dependent package %s does not exist in repositories on line %s",
													highlight(item),
													strconv.Itoa(currentLine+1),
												),
												line,
												fmt.Sprintf(
													"Did you mean to use %s?",
													highlight(correction),
												),
												strings.Index(line, item),
												len(item),
											)
										}
									}
								}
							}
							if field.Tag.Get("keyArray") == "groups:" {
								for _, item := range itemArray {
									if correction, needed := lintGroup(item); needed {
										outputWarningHighlight(
											fmt.Sprintf(
												"Group %s does not exist in repositories on line %s",
												highlight(item),
												strconv.Itoa(currentLine+1),
											),
											line,
											fmt.Sprintf(
												"Did you mean to use %s?",
												highlight(correction),
											),
											strings.Index(line, item),
											len(item),
										)
									}
								}
							}
						}

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
				if !*fakeroot {
					outputErrorHighlight(
						highlight(words[0])+" is not a valid key on line "+strconv.Itoa(currentLine+1),
						line,
						"Did you mean to use "+highlight(ClosestString(words[0], PossibleKeys))+"?",
						strings.Index(line, words[0]), len(words[0]),
					)
				}
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
				if splitString := strings.Fields(line); len(splitString) >= 2 {
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
			var stages = map[string]Stage{
				"%prep":    PrepareStage,
				"%build":   BuildStage,
				"%install": InstallStage,
				"%check":   CheckStage,
			}
			for stageString, stageEnum := range stages {
				if strings.HasPrefix(line, stageString) {
					currentStage = stageEnum
					continue mainParseLoop
				}
			}
			var scriptletStages = map[string]Stage{
				"%pre_install":  PreInstallStage,
				"%post_install": PostInstallStage,
				"%pre_upgrade":  PreUpgradeStage,
				"%post_upgrade": PostUpgradeStage,
				"%pre_remove":   PreRemoveStage,
				"%post_remove":  PostRemoveStage,
			}
			for stageString, stageEnum := range scriptletStages {
				if strings.HasPrefix(line, stageString) {
					if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
						currentScriptletSubpackage = subpackageName
					} else {
						if splitString := strings.Fields(line); len(splitString) >= 2 {
							currentScriptletSubpackage = lex.Name + "-" + splitString[1]
						} else {
							currentScriptletSubpackage = ""
						}
					}
					currentStage = stageEnum
					continue mainParseLoop
				}
			}
			if strings.HasPrefix(line, "%files") {
				if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
					currentFilesSubpackage = subpackageName
				} else {
					if splitString := strings.Fields(line); len(splitString) >= 2 {
						currentFilesSubpackage = lex.Name + "-" + splitString[1]
					} else {
						currentFilesSubpackage = ""
					}
				}
				currentStage = FileStage
				continue mainParseLoop
			}
			if strings.HasPrefix(line, "%changelog") {
				if subpackageName, hasSubpackage := grabFlagFromString(line, "-n", []string{}); hasSubpackage {
					currentChangelogSubpackage = subpackageName
				} else {
					if splitString := strings.Fields(line); len(splitString) >= 2 {
						currentChangelogSubpackage = lex.Name + "-" + splitString[1]
					} else {
						currentChangelogSubpackage = ""
					}
				}
				currentStage = ChangelogStage
				continue mainParseLoop
			}

			// If we're in a stage, we want to append some commands to our list
			m := map[Stage]*[]string{
				PrepareStage: &lex.Commands.Prepare,
				BuildStage:   &lex.Commands.Build,
				InstallStage: &lex.Commands.Install,
				CheckStage:   &lex.Commands.Check,
			}
			if str, ok := m[currentStage]; ok {
				*str = append(*str, evalInlineMacros(line, lex))
				continue mainParseLoop
			}
			if currentScriptletSubpackage == "" {
				scriptletStages := map[Stage]*[]string{
					PreInstallStage:  &lex.Scriptlets.PreInstall,
					PostInstallStage: &lex.Scriptlets.PostInstall,
					PreUpgradeStage:  &lex.Scriptlets.PreUpgrade,
					PostUpgradeStage: &lex.Scriptlets.PostUpgrade,
					PreRemoveStage:   &lex.Scriptlets.PreRemove,
					PostRemoveStage:  &lex.Scriptlets.PostRemove,
				}
				if str, ok := scriptletStages[currentStage]; ok {
					*str = append(*str, evalInlineMacros(line, lex))
					continue mainParseLoop
				}
			} else {
				if val, ok := lex.Subpackages[currentScriptletSubpackage]; ok {
					subpkg := val
					scriptletStages := map[Stage]*[]string{
						PreInstallStage:  &subpkg.Scriptlets.PreInstall,
						PostInstallStage: &subpkg.Scriptlets.PostInstall,
						PreUpgradeStage:  &subpkg.Scriptlets.PreUpgrade,
						PostUpgradeStage: &subpkg.Scriptlets.PostUpgrade,
						PreRemoveStage:   &subpkg.Scriptlets.PreRemove,
						PostRemoveStage:  &subpkg.Scriptlets.PostRemove,
					}
					if str, ok := scriptletStages[currentStage]; ok {
						*str = append(*str, evalInlineMacros(line, lex))
						lex.Subpackages[currentScriptletSubpackage] = subpkg
						continue mainParseLoop
					}
				} else {
					outputError("You cannot specify scriptlets for a subpackage that has not been declared")
				}
			}
			if currentStage == ChangelogStage {
				if currentChangelogSubpackage == "" {
					lex.Changelog = append(lex.Changelog, line)
				} else {
					if val, ok := lex.Subpackages[currentChangelogSubpackage]; ok {
						subpkg := val
						subpkg.Changelog = append(subpkg.Changelog, line)
						lex.Subpackages[currentChangelogSubpackage] = subpkg
					}
				}
				continue mainParseLoop
			}
			if currentStage == FileStage {
				var backup string
				if strings.HasPrefix(line, "%config") {
					backup = strings.TrimPrefix(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, "%config")), lex), "/")
				}
				if currentFilesSubpackage == "" {
					if backup != "" {
						lex.Backup = append(lex.Backup, evalInlineMacros(backup, lex))
					} else {
						lex.Files = append(lex.Files, evalInlineMacros(line, lex))
					}
				} else {
					if val, ok := lex.Subpackages[currentFilesSubpackage]; ok {
						subpkg := val
						if backup != "" {
							subpkg.Backup = append(subpkg.Backup, evalInlineMacros(backup, lex))
						} else {
							subpkg.Files = append(subpkg.Files, evalInlineMacros(line, lex))
						}
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

	promptMissingDepsInstall(lex)

	if len(lex.Commands.Prepare) == 0 {
		if !*fakeroot {
			outputStatus("Automatically setting up package...")
		}
		lex.Commands.Prepare = append(lex.Commands.Prepare, evalInlineMacros("%setup -q", lex))
	}

	lex.BuildPackage()

	return lex
}

// Build : Build a specfile, generating an Arch package.
func Build(pathToRecipe string) error {
	if !*fakeroot {
		outputStatus("Reading specfile from " + pathToRecipe + "...")
	}
	data, err := ioutil.ReadFile(pathToRecipe)
	if err != nil {
		return err
	}
	ParsePackage(string(data))
	return nil
}
