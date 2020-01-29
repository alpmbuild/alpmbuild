package lib

import (
	"fmt"
	"os"
	"strings"
)

func highlight(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;34m%s\033[0m\033[1;37m", message)
	}
	return message
}

func bold(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;37m%s\033[0m", message)
	} else {
		return message
	}
}

func green(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;32m%s\033[0m", message)
	} else {
		return message
	}
}

func red(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;31m%s\033[0m", message)
	} else {
		return message
	}
}

func yellow(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;93m%s\033[0m", message)
	} else {
		return message
	}
}

func outputStatus(message string) {
	println(green("==> ") + bold(message))
}

func outputError(message string) {
	println(red("ERROR ==> ") + bold(message))
	os.Exit(1)
}

func outputErrorHighlight(message, lineToHighlight, additionalMessage string, startIndex, length int) {
	lineToHighlight = strings.Replace(lineToHighlight, lineToHighlight[startIndex:startIndex+length], highlight(lineToHighlight[startIndex:startIndex+length]), 1)
	println(red("ERROR ==> ") + bold(message))
	fmt.Printf(
		"%s%s\n%s%s%s\n",
		strings.Repeat(" ", len("ERROR ==> ")),
		bold(lineToHighlight),
		strings.Repeat(" ", len("ERROR ==> ")),
		strings.Repeat(" ", startIndex),
		strings.Repeat(red("^"), length),
	)
	if additionalMessage != "" {
		fmt.Printf(
			"\n%s%s\n",
			strings.Repeat(" ", len("ERROR ==> ")),
			bold(additionalMessage),
		)
	}
	os.Exit(1)
}

func outputWarningHighlight(message, lineToHighlight, additionalMessage string, startIndex, length int) {
	lineToHighlight = strings.Replace(lineToHighlight, lineToHighlight[startIndex:startIndex+length], highlight(lineToHighlight[startIndex:startIndex+length]), 1)
	println(yellow("WARNING ==> ") + bold(message))
	fmt.Printf(
		"%s%s\n%s%s%s\n",
		strings.Repeat(" ", len("WARNING ==> ")),
		bold(lineToHighlight),
		strings.Repeat(" ", len("WARNING ==> ")),
		strings.Repeat(" ", startIndex),
		strings.Repeat(yellow("^"), length),
	)
	if additionalMessage != "" {
		fmt.Printf(
			"\n%s%s\n",
			strings.Repeat(" ", len("WARNING ==> ")),
			bold(additionalMessage),
		)
	}
}
