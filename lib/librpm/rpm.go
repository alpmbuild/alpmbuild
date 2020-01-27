package librpm

// #cgo pkg-config: rpm
// #include <rpm/rpmmacro.h>
//
// static char* expandy(char* in) {
//	char* result = rpmExpand(in, NULL);
//	return result;
// }
// static int definemacro(char* macro, int level) {
//	return rpmDefineMacro(NULL, macro, level);
// }
//
import "C"
import (
	"unsafe"
)

func ExpandMacro(macro string) string {
	cs := C.CString(macro)
	defer C.free(unsafe.Pointer(cs))

	expanded := C.expandy(cs)
	defer C.free(unsafe.Pointer(expanded))

	goStr := C.GoString(expanded)
	return goStr
}

func DefineMacro(macro string, level int) int {
	cs := C.CString(macro)
	defer C.free(unsafe.Pointer(cs))

	result := C.definemacro(cs, C.int(level))

	return int(result)
}