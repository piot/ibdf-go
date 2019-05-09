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

package ibdf

import (
	"fmt"

	"github.com/piot/brook-go/src/instream"
	"github.com/piot/ibdf-go/src/piff"
)

type InPacketFile struct {
	inFile        *piff.InFile
	schemaPayload []byte
}

func (c *InPacketFile) SchemaPayload() []byte {
	return c.schemaPayload
}

func (c *InPacketFile) readSchema() ([]byte, error) {
	header, payload, readErr := c.inFile.ReadChunk()
	if readErr != nil {
		return nil, readErr
	}
	if header.TypeIDString() != "sch1" {
		return nil, fmt.Errorf("wrong schema typeid %v", header)
	}

	return payload, nil
}

func NewInPacketFile(filename string) (*InPacketFile, error) {
	newPiffFile, err := piff.NewInFile(filename)
	if err != nil {
		return nil, err
	}
	c := &InPacketFile{
		inFile: newPiffFile,
	}
	schemaOctets, err := c.readSchema()
	if err != nil {
		return nil, err
	}
	c.schemaPayload = schemaOctets
	return c, err
}

func (c *InPacketFile) ReadPacket() (uint8, uint64, []byte, error) {
	header, payload, readErr := c.inFile.ReadChunk()
	if readErr != nil {
		return 0, 0, nil, readErr
	}
	if header.TypeIDString() != "pkt1" {
		return 0, 0, nil, fmt.Errorf("wrong typeid %v", header)
	}
	s := instream.New(payload)
	cmd, cmdErr := s.ReadUint8()
	if cmdErr != nil {
		return 0, 0, nil, cmdErr
	}
	monotonicTimeMs, timeMsErr := s.ReadUint64()
	if timeMsErr != nil {
		return 0, 0, nil, timeMsErr
	}

	return cmd, monotonicTimeMs, payload[1+8:], nil
}

func (c *InPacketFile) Close() {
	c.inFile.Close()
}
