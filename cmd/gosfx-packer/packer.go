package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/netbrain/gosfx"
)

var outputFile *string
var extractorBinary = flag.String("extractor", findExtractorBinary(), "the path to the sfx extractor")
var entrypoint = flag.String("main", "", "command to run after extraction")

func init() {
	var suffix string
	if runtime.GOOS == "windows" {
		suffix = ".exe"
	}
	outputFile = flag.String("output", "gosfx.out"+suffix, "name of the output file")
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.PrintDefaults()
		return
	}

	p := gosfx.NewPacker(*entrypoint, *extractorBinary, *outputFile)
	if err := p.AddFiles(flag.Args()...); err != nil {
		log.Fatal(err)
	}
	p.Close()
}

func findExtractorBinary() string {
	searchLocations := []string{"."}
	for _, path := range []string{
		os.Getenv("GOPATH"),
		os.Getenv("PATH"),
	} {
		for _, location := range filepath.SplitList(path) {
			searchLocations = append(searchLocations, location)
		}
	}

	for _, location := range searchLocations {
		extractorBinary := filepath.Join(location, "gosfx-extractor")
		_, err := os.Stat(extractorBinary)
		if err != nil && err != os.ErrNotExist {
			continue
		}
		return extractorBinary
	}
	return ""
}
