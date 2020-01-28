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

func evalIf(inputLine string) bool {
	// %if first relation second
	fields := strings.Fields(inputLine)

	if len(fields) < 4 {
		return false
	}

	relationResult := false

	switch fields[2] {
	case "==":
		if fields[1] == fields[3] {
			relationResult = true
		}
	case "<=":
		before, after := parseInts(fields[1], fields[3])
		if before <= after {
			relationResult = true
		}
	case ">=":
		before, after := parseInts(fields[1], fields[3])
		if before >= after {
			relationResult = true
		}
	case "<":
		before, after := parseInts(fields[1], fields[3])
		if before < after {
			relationResult = true
		}
	case ">":
		before, after := parseInts(fields[1], fields[3])
		if before > after {
			relationResult = true
		}
	default:
		outputError("Invalid comparison operand: " + fields[2])
	}

	return relationResult
}
