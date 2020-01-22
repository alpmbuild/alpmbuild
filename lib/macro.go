package lib

import (
	"reflect"
	"regexp"
	"strings"
)

func evalInlineMacros(input string, context packageContext) string {
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
