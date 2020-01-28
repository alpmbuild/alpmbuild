package lib

import (
	"strconv"
	"strings"
)

func parseInts(string1, string2 string) (int64, int64) {
	before, err := strconv.ParseInt(string1, 10, 64)
	if err != nil {
		outputError("Error parsing integer: " + err.Error())
	}
	after, err := strconv.ParseInt(string2, 10, 64)
	if err != nil {
		outputError("Error parsing integer: " + err.Error())
	}
	return before, after
}

func evalIf(inputLine string, pkg PackageContext, currentLine int) bool {
	// %if first relation second
	fields := strings.Fields(inputLine)

	if len(fields) < 4 {
		return false
	}

	relationResult := false

	switch fields[2] {
	case "==":
		if evalInlineMacros(fields[1], pkg) == evalInlineMacros(fields[3], pkg) {
			relationResult = true
		}
	case "<=":
		before, after := parseInts(evalInlineMacros(fields[1], pkg), evalInlineMacros(fields[3], pkg))
		if before <= after {
			relationResult = true
		}
	case ">=":
		before, after := parseInts(evalInlineMacros(fields[1], pkg), evalInlineMacros(fields[3], pkg))
		if before >= after {
			relationResult = true
		}
	case "<":
		before, after := parseInts(evalInlineMacros(fields[1], pkg), evalInlineMacros(fields[3], pkg))
		if before < after {
			relationResult = true
		}
	case ">":
		before, after := parseInts(evalInlineMacros(fields[1], pkg), evalInlineMacros(fields[3], pkg))
		if before > after {
			relationResult = true
		}
	default:
		outputErrorHighlight(
			"Invalid comparison operand "+highlight(fields[2])+" on line "+strconv.Itoa(currentLine),
			inputLine,
			"Valid operands are ==, <=, >=, <, and >",
			strings.Index(inputLine, fields[2]), len(fields[2]),
		)
	}

	return relationResult
}
