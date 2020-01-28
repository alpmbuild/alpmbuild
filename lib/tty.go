package lib

// #include <unistd.h>
// #include <stdio.h>
// #include <stdbool.h> 
// static bool isStdoutTty() {
//	return isatty(fileno(stdout));
// }
import "C"

func IsStdoutTty() bool {
	return bool(C.isStdoutTty())
}