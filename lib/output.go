package lib

import (
	"fmt"
	"os"
)

func highlight(message string) string {
	if IsStdoutTty() && *useColours {
		return fmt.Sprintf("\033[1;34m%s\033[0m\033[1;37m", message)
	}
	return message
}

func outputStatus(message string) {
	if IsStdoutTty() && *useColours {
		fmt.Printf("\033[1;32m==>\033[0m \033[1;37m%s\033[0m\n", message)
	} else {
		fmt.Printf("==> %s\n", message)
	}
}

func outputError(message string) {
	if IsStdoutTty() && *useColours {
		fmt.Printf("\033[1;31mERROR ==>\033[0m \033[1;37m%s\033[0m\n", message)
	} else {
		fmt.Printf("ERROR ==> %s\n", message)
	}
	os.Exit(1)
}
