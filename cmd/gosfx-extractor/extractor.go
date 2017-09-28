package main

import (
	"archive/tar"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/netbrain/gosfx"
	"github.com/ulikunitz/xz"
)

type execReader struct {
	offset     int64
	executable string
	r          *os.File
	f          *gosfx.Footer
	outputDir  string
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	eReader, err := newExecReader()
	if err != nil {
		log.Fatal(err)
	}
	defer eReader.Close()

	if err := eReader.ExtractFiles(); err != nil {
		log.Fatal(err)
	}

	if err := eReader.ExecuteEntryPoint(); err != nil {
		log.Fatal(err)
	}
}

func (e *execReader) Close() error {
	return e.r.Close()
}

func newExecReader() (e *execReader, err error) {
	e = &execReader{}

	e.executable, err = os.Executable()
	if err != nil {
		err = fmt.Errorf("failed fetching the current executable: %s", err)
		return
	}

	e.r, err = os.OpenFile(e.executable, os.O_RDONLY, 0744)
	if err != nil {
		err = fmt.Errorf("could not open the current executable for reading: %s", err)
	}

	outputDir, err := ioutil.TempDir("", "gosfx")
	if err != nil {
		err = fmt.Errorf("failet to create output directory: %s", err)
	}
	e.outputDir = outputDir

	return e, e.readFooter()
}

func (e *execReader) readFooter() (err error) {
	if e.f != nil {
		return nil
	}

	//Seek to the start of footer
	e.offset, err = e.r.Seek(-2, io.SeekEnd)
	if err != nil {
		err = fmt.Errorf("error when seeking to uint16 footer offset: %s", err)
		return
	}

	//Read the footer offset value
	var footerOffset uint16
	err = binary.Read(e.r, binary.BigEndian, &footerOffset)
	if err != nil {
		err = fmt.Errorf("could not read the footer offset value: %s", err)
		return
	}

	//Seek to the start of footer
	e.offset, err = e.r.Seek(e.offset-int64(footerOffset), io.SeekStart)

	//Read the footer value
	e.f = &gosfx.Footer{}
	gr := gob.NewDecoder(e.r)
	err = gr.Decode(e.f)
	if err != nil {
		err = fmt.Errorf("could not read the footer value: %s", err)
	}
	return
}

func (e *execReader) ExtractFiles() error {
	// Open xz part for reading
	e.r.Seek(e.offset-e.f.RelArchiveOffset, io.SeekStart)

	xzr, err := xz.NewReader(e.r)
	if err != nil {
		log.Fatalf("failed to read xz stream: %s\n", err)
	}

	// Open the tar archive for reading.
	tr := tar.NewReader(xzr)

	//Extract files
	if err != nil {
		return fmt.Errorf("failed to create output directory: %s", err)
	}

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to extract from archive: %s", err)
		}
		fmt.Printf("Extracting '%s'\n", hdr.Name)
		file := filepath.Join(e.outputDir, hdr.Name)

		if err = os.MkdirAll(filepath.Dir(file), 0744); err != nil {
			return fmt.Errorf("failed to create directory: %s", err)
		}
		fileHandle, err := os.Create(file)

		if err != nil {
			return fmt.Errorf("failed to create output file: %s", err)
		}

		if _, err := io.Copy(fileHandle, tr); err != nil {
			return fmt.Errorf("failed to extract to file: %s", err)
		}

		if err := fileHandle.Close(); err != nil {
			return fmt.Errorf("failed to close file handle: %s", err)
		}

		if err := os.Chmod(file, os.FileMode(hdr.Mode)); err != nil {
			return fmt.Errorf("failed to set file permissions: %s", err)
		}

		if err := os.Chtimes(file, hdr.AccessTime, hdr.ModTime); err != nil {
			return fmt.Errorf("failed to set modification time: %s", err)
		}
	}
	return nil
}

func (e *execReader) ExecuteEntryPoint() (err error) {
	if e.f.EntryPoint != nil && len(e.f.EntryPoint) > 0 {
		if err := os.Chdir(e.outputDir); err != nil {
			return err
		}
		cmd := exec.Command(e.f.EntryPoint[0], e.f.EntryPoint[1:]...)
		cmd.Dir = e.outputDir
		cmd.Stdout, cmd.Stdin, cmd.Stderr = os.Stdout, os.Stdin, os.Stderr
		fmt.Printf("Executing entrypoint '%s'\n", e.f.EntryPoint)
		if err := cmd.Run(); err != nil {
			err = fmt.Errorf("%s failed with: %s", e.f.EntryPoint[0], err)
		}
	}
	return
}
