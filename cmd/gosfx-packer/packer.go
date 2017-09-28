package main

import (
	"io"
	"log"
	"os"

	"archive/tar"
	"encoding/binary"
	"flag"

	"path/filepath"

	"strings"

	"fmt"

	"encoding/gob"

	"github.com/netbrain/gosfx"
	"github.com/ulikunitz/xz"
)

const (
	EXTRACTOR_BINARY_NAME = "gosfx-extractor"
)

var extractorBinary = flag.String("extractor", findExtractorBinary(), "the path to the sfx extractor")
var outputFile = flag.String("output", "gosfx.out"+OUTPUT_FILE_SUFFIX, "name of the output file")
var entrypoint = flag.String("main", "", "command to run after extraction")

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	flag.Parse()

	if flag.NArg() == 0 {
		flag.PrintDefaults()
		return
	}

	//Open the extractor binary
	bin, err := os.OpenFile(*extractorBinary, os.O_RDONLY, 0744)
	if err != nil {
		log.Fatalf("failed to open extractor binary: %s\n", err)
	}
	defer bin.Close()

	//Create the output file
	outFile, err := os.OpenFile(*outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		log.Fatalf("failed to open the output file: %s\n", err)
	}
	defer outFile.Close()

	//Copy extractor to output file
	if _, err := io.Copy(outFile, bin); err != nil {
		log.Fatalf("failed to copy data from extractor to output: %s\n", err)
	}

	//Get the size of the output file
	absoluteOffset, err := outFile.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatalf("failed to seek to the end of output file: %s\n", err)
	}

	// Create a new xz writer.
	xzw, err := xz.NewWriter(outFile)
	if err != nil {
		log.Fatalf("failed initializing a new xz writer: %s", err)
	}

	// Create a new tar archive.
	tw := tar.NewWriter(xzw)

	// Add some files to the archive.
	for _, file := range flag.Args() {
		fileInfo, err := os.Stat(file)
		if err != nil {
			log.Fatalf("failed to stat %s: %s\n", file, err)
		}

		if fileInfo.IsDir() {
			log.Fatal("Dirs not supported")
		}

		hdr := &tar.Header{
			Name:    fileInfo.Name(),
			Mode:    int64(fileInfo.Mode()),
			Size:    int64(fileInfo.Size()),
			ModTime: fileInfo.ModTime(),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatalf("failed to write tar header: %s\n", err)
		}

		fileHandle, err := os.OpenFile(file, os.O_RDONLY, fileInfo.Mode())
		if err != nil {
			log.Fatalf("failed to open %s for archiving: %s\n", file, err)
		}

		if _, err := io.Copy(tw, fileHandle); err != nil {
			log.Fatalf("failed to archive %s: %s\n", file, err)
		}

		if err := fileHandle.Close(); err != nil {
			log.Fatalf("error when closing file %s: %s\n", file, err)
		}

	}
	// Make sure to check the error on Close.
	if err := tw.Close(); err != nil {
		log.Fatalf("error when closing tar stream: %s\n", err)
	}

	// Make sure to check the error on Close.
	if err := xzw.Close(); err != nil {
		log.Fatalf("error when closing xz stream: %s\n", err)
	}
	eOF, err := outFile.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatalf("failed to seek to the end of output file: %s\n", err)
	}

	//Write footer
	footer := gosfx.Footer{
		RelArchiveOffset: eOF - absoluteOffset,
		EntryPoint:       strings.Split(*entrypoint, " "),
	}

	//Write for real with correct archive offset
	var out io.Writer
	out = &gosfx.Dumper{outFile}
	//out = outFile
	//xzw, err = xz.NewWriter(out)
	//if err != nil {
	//	log.Fatal(err)
	//}
	gw := gob.NewEncoder(out)
	err = gw.Encode(footer)
	if err != nil {
		log.Fatal(err)
	}

	newEof, err := outFile.Seek(0, io.SeekEnd)
	if err != nil {
		log.Fatalf("failed to seek to the end of output file: %s\n", err)
	}

	//Write relative offset for footer
	binary.Write(out, binary.BigEndian, uint16(newEof-eOF))
	fmt.Println(newEof - eOF)
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
		extractorBinary := filepath.Join(location, EXTRACTOR_BINARY_NAME)
		_, err := os.Stat(extractorBinary)
		if err != nil && err != os.ErrNotExist {
			continue
		}
		return extractorBinary
	}
	return ""
}
