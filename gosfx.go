package gosfx

import (
	"archive/tar"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz"
)

//Footer is the metadata which is appended to the end of the packed binary
type Footer struct {
	RelArchiveOffset int64
	EntryPoint       []string
}

//Packer handles the embedding of files to the target binary
type Packer struct {
	outFile    *os.File
	outSize    int64
	tw         *tar.Writer
	xzw        io.WriteCloser
	entrypoint string
	extractor  string
	output     string
	addedFiles bool
}

//NewPacker creates a new packer
func NewPacker(entrypoint, extractor, output string) *Packer {
	p := &Packer{
		entrypoint: entrypoint,
		extractor:  extractor,
		output:     output,
	}
	p.copyExtractorBinary()
	return p
}

//Close writes the packer metadata and closes the output file
func (p *Packer) Close() error {
	p.writeFooter()
	return p.outFile.Close()
}

func (p *Packer) writeFooter() error {

	EOF, err := p.outFile.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of output file: %s", err)
	}

	//Write footer
	footer := Footer{
		RelArchiveOffset: EOF - p.outSize,
		EntryPoint:       strings.Split(p.entrypoint, " "),
	}

	gw := gob.NewEncoder(p.outFile)
	err = gw.Encode(footer)
	if err != nil {
		return fmt.Errorf("failed to encode the footer: %s", err)
	}

	newEOF, err := p.outFile.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of output file: %s", err)
	}

	//Write relative offset for footer
	binary.Write(p.outFile, binary.BigEndian, uint16(newEOF-EOF))
	return nil
}

func (p *Packer) copyExtractorBinary() error {
	//Open the extractor binary
	bin, err := os.OpenFile(p.extractor, os.O_RDONLY, 0744)
	if err != nil {
		return fmt.Errorf("failed to open extractor binary: %s", err)
	}
	defer bin.Close()

	//Create the output file
	p.outFile, err = os.OpenFile(p.output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		return fmt.Errorf("failed to open the output file: %s", err)
	}

	//Copy extractor to output file
	if _, err := io.Copy(p.outFile, bin); err != nil {
		return fmt.Errorf("failed to copy data from extractor to output: %s", err)
	}

	//Get the size of the output file
	p.outSize, err = p.outFile.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("failed to seek to the end of output file: %s", err)
	}
	return nil
}

//AddFiles embeds the files into the extractor binary as a tar.xz formatted archive
func (p *Packer) AddFiles(files ...string) (err error) {
	if p.addedFiles {
		return errors.New("only a single execution of AddFiles is valid")
	}
	p.addedFiles = true

	// Create a new xz writer.
	p.xzw, err = xz.NewWriter(p.outFile)
	if err != nil {
		return fmt.Errorf("failed initializing a new xz writer: %s", err)
	}

	// Create a new tar archive.
	p.tw = tar.NewWriter(p.xzw)

	for _, file := range files {
		if err := p.addFile(file, ""); err != nil {
			return err
		}
	}

	if err := p.tw.Close(); err != nil {
		return fmt.Errorf("error when closing tar stream: %s", err)
	}

	if err := p.xzw.Close(); err != nil {
		return fmt.Errorf("error when closing xz stream: %s", err)
	}
	return
}

func (p *Packer) addFile(file, pName string) error {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %s", file, err)
	}

	if pName == "" {
		pName = fileInfo.Name()
	}

	if fileInfo.IsDir() {
		entries, err := ioutil.ReadDir(file)
		if err != nil {
			return fmt.Errorf("failed to list directory entries %s: %s", file, err)
		}
		for _, entry := range entries {
			if err := p.addFile(
				filepath.Join(file, entry.Name()),
				filepath.Join(fileInfo.Name(), entry.Name()),
			); err != nil {
				return err
			}
		}
		return nil
	}

	hdr := &tar.Header{
		Name:       pName,
		Mode:       int64(fileInfo.Mode()),
		Size:       int64(fileInfo.Size()),
		ModTime:    fileInfo.ModTime(),
		AccessTime: fileInfo.ModTime(),
	}

	if err := p.tw.WriteHeader(hdr); err != nil {
		fmt.Printf("%#v", hdr)
		return fmt.Errorf("failed to write tar header: %s", err)
	}

	fileHandle, err := os.OpenFile(file, os.O_RDONLY, fileInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to open %s for archiving: %s", file, err)
	}

	if _, err := io.Copy(p.tw, fileHandle); err != nil {
		return fmt.Errorf("failed to archive %s: %s", file, err)
	}

	if err := fileHandle.Close(); err != nil {
		return fmt.Errorf("error when closing file %s: %s", file, err)
	}
	return nil
}
