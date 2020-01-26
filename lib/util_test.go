package lib

import "testing"

func TestFlagGrab(t *testing.T) {
	value, hasFlag := grabFlagFromString("%package -n cyanogen -q pingas", "n", []string{"q"})

	t.Logf("\nPackage Name:\n\t%s\nHas Flag:\n\t%t", value, hasFlag)
	if value != "cyanogen" || !hasFlag {
		t.Fail()
	}
}
