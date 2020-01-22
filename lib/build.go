package lib

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"
)

func lex(data string) packageContext {
	lex := packageContext{}

	// Let's parse the Key: Value things first
	for _, line := range strings.Split(strings.TrimSuffix(data, "\n"), "\n") {
		// We expect a ': ' to be present in any Key: Value line
		if strings.Contains(line, ": ") {
			// Split our line by whitespace
			words := strings.Fields(line)

			// Because we need at least two values for a Key: Value PAIR, make
			// sure we have at least two values
			if len(words) < 2 {
				continue
			}

			// Let's worry about our sources first...
			if strings.HasPrefix(strings.ToLower(words[0]), "source") {
				lex.Sources = append(lex.Sources, evalInlineMacros(words[1], lex))
			}

			// Now we produce macros based off our package context.

			fields := reflect.TypeOf(lex)
			num := fields.NumField()

			// Loop through the fields of the package context in order to see if any of the annotated key values match the line we're on
			for i := 0; i < num; i++ {
				field := fields.Field(i)
				if strings.ToLower(words[0]) == field.Tag.Get("key") {
					// We assert that packageContext only has string fields here.
					// If it doesn't, our code will break.
					key := reflect.ValueOf(&lex).Elem().FieldByName(field.Name)
					if key.IsValid() {
						key.SetString(evalInlineMacros(strings.TrimSpace(strings.TrimPrefix(line, words[0])), lex))
					}
				}
			}
		}
	}

	fmt.Printf("%+v\n", lex)

	return lex
}

// Build : Build a specfile, generating an Arch package.
func Build(pathToRecipe string) error {
	data, err := ioutil.ReadFile(pathToRecipe)
	if err != nil {
		return err
	}
	lex(string(data))
	return nil
}
