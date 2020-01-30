package lib

import (
	"strconv"
	"strings"
)

func errorFunc(raw, macro, inputLine string, currentLine int) {
	if raw != macro {
		outputErrorHighlight(
			highlight(raw)+" does not expand to a valid integer on line "+strconv.Itoa(currentLine),
			inputLine,
			highlight(raw)+bold(" -> ")+highlight(macro),
			strings.Index(inputLine, raw), len(raw),
		)
	} else {
		outputErrorHighlight(
			highlight(raw)+" is not a valid integer on line "+strconv.Itoa(currentLine),
			inputLine,
			"",
			strings.Index(inputLine, raw), len(raw),
		)
	}
}

func parseInts(string1, string2 string, pkg PackageContext, inputLine string, currentLine int) (int64, int64) {
	parsedString1 := evalInlineMacros(string1, pkg)
	parsedString2 := evalInlineMacros(string2, pkg)

	before, err := strconv.ParseInt(parsedString1, 10, 64)
	if err != nil {
		errorFunc(string1, parsedString1, inputLine, currentLine)
	}
	after, err := strconv.ParseInt(parsedString2, 10, 64)
	if err != nil {
		errorFunc(string2, parsedString2, inputLine, currentLine)
	}
	return before, after
}

func parseInt(string1 string, pkg PackageContext, inputLine string, currentLine int) int64 {
	parsedString := evalInlineMacros(string1, pkg)

	parsedInt, err := strconv.ParseInt(parsedString, 10, 64)
	if err != nil {
		errorFunc(string1, parsedString, inputLine, currentLine)
	}
	return parsedInt
}

func evalIf(inputLine string, pkg PackageContext, currentLine int) bool {
	// %if first relation second
	fields := strings.Fields(inputLine)

	if len(fields) < 4 {
		// %if number
		if len(fields) == 2 {
			inty := parseInt(fields[1], pkg, inputLine, currentLine)
			if inty > 0 {
				return true
			}
		}
		return false
	}

	relationResult := false

	switch fields[2] {
	case "==":
		if fields[1] == fields[3] {
			relationResult = true
		}
	case "<=":
		before, after := parseInts(fields[1], fields[3], pkg, inputLine, currentLine)
		if before <= after {
			relationResult = true
		}
	case ">=":
		before, after := parseInts(fields[1], fields[3], pkg, inputLine, currentLine)
		if before >= after {
			relationResult = true
		}
	case "<":
		before, after := parseInts(fields[1], fields[3], pkg, inputLine, currentLine)
		if before < after {
			relationResult = true
		}
	case ">":
		before, after := parseInts(fields[1], fields[3], pkg, inputLine, currentLine)
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
