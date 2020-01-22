package lib

import (
	"flag"
	"os"
)

// Enter : Main function of alpmbuild's command line interface
func Enter() {
	file := flag.String("file", "", "The file to build.")

	flag.Parse()

	if *file == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := Build(*file)
	if err != nil {
		println(err.Error())
	}
}
