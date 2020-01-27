package librpm

import "testing"

func TestExpandMacro(t *testing.T) {
	DefineMacro("_libdir flyingpingas", 0)
	if ExpandMacro("%{_libdir}") != "flyingpingas" {
		t.Fail()
	}
	if ExpandMacro("%{_wonky}") != "%{_wonky}" {
		t.Fail()
	}
}
