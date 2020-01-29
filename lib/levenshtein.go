package lib

import "unicode/utf8"

func max(int1, int2 int) int {
	if int1 > int2 {
		return int1
	}
	return int2
}

func min(int1, int2 int) int {
	if int1 > int2 {
		return int2
	}
	return int1
}

// Implementation taken from https://gist.github.com/andrei-m/982927#gistcomment-1931258 and
// https://github.com/agnivade/levenshtein/blob/master/levenshtein.go.
// Both implementations are licensed under the MIT license.
func LevenshteinDistance(string1, string2 string) int {
	// If one of our strings is zero-length,
	// the distance is equivalent to the length
	// of the non-zero string
	if len(string1) == 0 || len(string2) == 0 {
		return max(
			utf8.RuneCountInString(string1),
			utf8.RuneCountInString(string2),
		)
	}

	// If our strings are identical, the distance between them is zero.
	if string1 == string2 {
		return 0
	}

	// Let's convert our strings to arrays of runes
	runes1, runes2 := []rune(string1), []rune(string2)

	// We can optimise this by having runes1 being smaller than runes2.
	if len(runes1) > len(runes2) {
		runes1, runes2 = runes2, runes1
	}

	lengthOfRunes1 := len(runes1)
	lengthOfRunes2 := len(runes2)

	// Let's make our row and fill it up now...
	row := make([]int, lengthOfRunes1+1)
	for i := 1; i < len(row); i++ {
		row[i] = i
	}

	for i := 1; i <= lengthOfRunes2; i++ {
		previous := i
		var current int

		for j := 1; j <= lengthOfRunes1; j++ {
			if runes2[i-1] == runes1[j-1] {
				current = row[j-1]
			} else {
				current = min(
					min(
						row[j-1]+1,
						previous+1,
					),
					row[j]+1,
				)
			}
			row[j-1] = previous
			previous = current
		}

		row[lengthOfRunes1] = previous
	}
	return row[lengthOfRunes1]
}

func ClosestString(stringToLookForMatches string, stringsToLookIn []string) string {
	length, returnString := int(^uint(0)>>1), ""

	for _, stringToCheck := range stringsToLookIn {
		matchNumber := LevenshteinDistance(stringToLookForMatches, stringToCheck)
		if matchNumber <= length {
			length, returnString = matchNumber, stringToCheck
		}
	}

	return returnString
}
