/*

MIT License

Copyright (c) 2019 Peter Bjorklund

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

*/

package piff

import (
	"fmt"
	"io"
	"os"

	"github.com/piot/brook-go/src/instream"
)

type InHeader struct {
	typeID      [4]byte
	octetLength int
}

func (i InHeader) TypeIDString() string {
	return string(i.typeID[0:])
}

func (i InHeader) OctetCount() int {
	return i.octetLength
}

func (i InHeader) String() string {
	return fmt.Sprintf("header %v %v", i.TypeIDString(), i.OctetCount())
}

type InFile struct {
	inFile *os.File
	header InHeader
	isEOF  bool
}

func NewInFile(filename string) (*InFile, error) {
	newFile, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	c := &InFile{
		inFile: newFile,
	}
	headerErr := c.readHeader()
	return c, headerErr
}

func (c *InFile) readHeader() error {
	header := make([]byte, 8)
	octetCount, err := c.inFile.Read(header)

	if err == io.EOF {
		c.isEOF = true
		return nil
	}
	if err != nil {
		return err
	}
	if octetCount != 8 {
		return fmt.Errorf("piff: couldn't read whole header")
	}
	s := instream.New(header)
	typeID, readErr := s.ReadOctets(4)
	if readErr != nil {
		return readErr
	}
	c.header.typeID = [4]byte{
		typeID[0],
		typeID[1],
		typeID[2],
		typeID[3],
	}
	chunkOctetCount, countErr := s.ReadUint32()
	if countErr != nil {
		return countErr
	}
	c.header.octetLength = int(chunkOctetCount)
	return nil
}

func (c *InFile) ReadChunk() (InHeader, []byte, error) {
	if c.isEOF {
		return InHeader{}, nil, io.EOF
	}

	payload := make([]byte, c.header.octetLength)
	octetsRead, err := c.inFile.Read(payload)
	if err != nil {
		return InHeader{}, nil, err
	}
	savedHeader := c.header

	if octetsRead != savedHeader.octetLength {
		return InHeader{}, nil, fmt.Errorf("couldnt read the whole payload")
	}
	headerErr := c.readHeader()
	return savedHeader, payload, headerErr
}

func (c *InFile) Close() {
	c.inFile.Close()
}
