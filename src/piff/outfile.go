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
	"os"

	"github.com/piot/brook-go/src/outstream"
)

type OutFile struct {
	outFile *os.File
}

func NewOutFile(filename string) (*OutFile, error) {
	newFile, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	c := &OutFile{
		outFile: newFile,
	}
	return c, nil
}

func (c *OutFile) WriteChunkTypeIDString(typeID string, payload []byte) error {
	fixedTypeID := [4]byte{
		byte(typeID[0]),
		byte(typeID[1]),
		byte(typeID[2]),
		byte(typeID[3]),
	}
	return c.WriteChunk(fixedTypeID, payload)
}

func (c *OutFile) WriteChunk(typeID [4]byte, payload []byte) error {
	s := outstream.New()
	typeIDOctets := typeID[0:]
	if len(typeIDOctets) != 4 {
		return fmt.Errorf("wrong conversion")
	}
	s.WriteOctets(typeIDOctets)
	octetCount := len(payload)
	s.WriteUint32(uint32(octetCount))
	s.WriteOctets(payload)
	filePayload := s.Octets()
	c.outFile.Write(filePayload)
	c.outFile.Sync()
	return nil
}

func (c *OutFile) Close() {
	c.outFile.Close()
}
