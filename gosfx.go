package gosfx

import (
	"encoding/hex"
	"fmt"
	"io"
)

type Footer struct {
	RelArchiveOffset int64
	EntryPoint       []string
}

type Dumper struct {
	IO io.ReadWriteSeeker
}

func (d *Dumper) Seek(offset int64, whence int) (int64, error) {
	return d.IO.Seek(offset, whence)
}

func (d *Dumper) Read(p []byte) (n int, err error) {
	n, err = d.IO.Read(p)
	fmt.Println(hex.Dump(p))
	return
}

func (d *Dumper) Write(p []byte) (n int, err error) {
	fmt.Println(hex.Dump(p))
	n, err = d.IO.Write(p)
	return
}
