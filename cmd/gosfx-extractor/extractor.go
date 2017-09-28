package main

import (
	"archive/tar"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"

	"encoding/gob"

	"github.com/netbrain/gosfx"
	"github.com/ulikunitz/xz"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	var err error
	var footerOffset uint16
	var footer gosfx.Footer

	//Open the current binary
	exec, err := os.Executable()
	if err != nil {
		log.Fatal("failed fetching the current executable: %s", err)
	}
	execFile, err := os.OpenFile(exec, os.O_RDONLY, 0744)
	if err != nil {
		log.Fatalf("could not open the current executable for reading: %s\n", err)
	}

	var in io.ReadSeeker
	in = &gosfx.Dumper{IO: execFile}

	//Seek to the start of footer
	offset, err := in.Seek(-2, io.SeekEnd)
	if err != nil {
		log.Fatalf("error when seeking to uint16 footer offset: %s\n", err)
	}

	//Read the footer offset value
	err = binary.Read(in, binary.BigEndian, &footerOffset)
	if err != nil {
		log.Fatalf("could not read the footer offset value: %s\n", err)
	}

	//Seek to the start of footer
	offset, err = in.Seek(offset-int64(footerOffset), io.SeekStart)

	//Read the footer value
	//xzr, err := xz.NewReader(in)
	//if err != nil {
	//	log.Fatalf("failed to read xz stream: %s\n", err)
	//}

	gr := gob.NewDecoder(in)
	err = gr.Decode(&footer)
	if err != nil {
		log.Fatalf("could not read the footer value: %s\n", err)
	}

	// Open xz part for reading
	in.Seek(offset-footer.RelArchiveOffset, io.SeekStart)

	xzr, err := xz.NewReader(in)
	if err != nil {
		log.Fatalf("failed to read xz stream: %s\n", err)
	}

	// Open the tar archive for reading.
	tr := tar.NewReader(xzr)

	// Iterate through the files in the archive.
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)
		if _, err := io.Copy(os.Stdout, tr); err != nil {
			log.Fatalln(err)
		}
		fmt.Println()
	}
}
