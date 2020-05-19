package lib

import "github.com/appadeia/alpmbuild/lib/librpm"

var archMacros = map[string]string{
	// Rust
	"alp_cargo_build": "cargo build --release --locked",
	"alp_cargo_test":  "cargo test --release --locked",
	// Go
	"alpm_go_build": `go build -trimpath -buildmode=pie -mod=readonly -modcacherw`,
}

func init() {
	for name, value := range archMacros {
		librpm.DefineMacro(name+" "+value, 256)
	}
}
