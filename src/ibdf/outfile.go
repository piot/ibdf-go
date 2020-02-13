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
	"os"

	"github.com/piot/brook-go/src/outstream"
	"github.com/piot/piff-go/src/piff"
)

type PacketDirection = uint8

const (
	CmdOutgoingPacket PacketDirection = 0x81
	CmdIncomingPacket PacketDirection = 0x01
)

type OutPacketFile struct {
	outFile *piff.OutStream
}

func NewOutPacketFile(filename string, header Header, schemaPayload []byte) (*OutPacketFile, error) {
	newPiffFile, err := piff.NewOutStream(filename)
	if err != nil {
		return nil, err
	}
	return internalCreate(newPiffFile, header, schemaPayload)
}

func NewOutPacketFileUsingFile(file *os.File, header Header, schemaPayload []byte) (*OutPacketFile, error) {
	newPiffFile, err := piff.NewOutStreamFile(file)
	if err != nil {
		return nil, err
	}

	return internalCreate(newPiffFile, header, schemaPayload)
}

func writeString(out *outstream.OutStream, s string) error {
	lengthErr := out.WriteUint8(uint8(len(s)))
	if lengthErr != nil {
		return lengthErr
	}
	applicationErr := out.WriteOctets([]byte(s))
	if applicationErr != nil {
		return applicationErr
	}

	return nil
}

func writeNameAndVersion(stream *outstream.OutStream, nameAndVersion NameAndVersion) error {
	nameErr := writeString(stream, nameAndVersion.Name)
	if nameErr != nil {
		return nameErr
	}
	versionErr := writeString(stream, nameAndVersion.Version)
	if versionErr != nil {
		return versionErr
	}

	return nil
}

func writeHeader(headerStream *outstream.OutStream, header Header) error {
	companyErr := writeString(headerStream, header.CompanyName)
	if companyErr != nil {
		return companyErr
	}
	appErr := writeNameAndVersion(headerStream, header.Application)
	if appErr != nil {
		return appErr
	}
	schemaErr := writeNameAndVersion(headerStream, header.Schema)
	if schemaErr != nil {
		return schemaErr
	}

	networkErr := writeNameAndVersion(headerStream, header.NetworkEngine)
	if networkErr != nil {
		return networkErr
	}

	protocolErr := writeNameAndVersion(headerStream, header.Protocol)
	return protocolErr
}

func internalCreate(newPiffFile *piff.OutStream, header Header, schemaPayload []byte) (*OutPacketFile, error) {
	c := &OutPacketFile{
		outFile: newPiffFile,
	}

	headerStream := outstream.New()
	headerErr := writeHeader(headerStream, header)
	if headerErr != nil {
		return nil, headerErr
	}
	c.outFile.WriteChunkTypeIDString("pac1", headerStream.Octets())

	writeErr := c.outFile.WriteChunkTypeIDString("sch1", schemaPayload)
	if writeErr != nil {
		return nil, writeErr
	}
	return c, nil
}

func (c *OutPacketFile) writePacket(cmd PacketDirection, monotonicTimeMs int64, b []byte) error {
	s := outstream.New()
	s.WriteUint8(cmd)
	s.WriteUint64(uint64(monotonicTimeMs))
	s.WriteOctets(b)
	return c.outFile.WriteChunkTypeIDString("pkt1", s.Octets())
}

func (c *OutPacketFile) DebugIncomingPacket(b []byte, monotonicTimeMs int64) error {
	return c.writePacket(CmdIncomingPacket, monotonicTimeMs, b)
}

func (c *OutPacketFile) DebugOutgoingPacket(b []byte, monotonicTimeMs int64) error {
	return c.writePacket(CmdOutgoingPacket, monotonicTimeMs, b)
}

func (c *OutPacketFile) DebugState(stateOctets []byte, monotonicTimeMs int64) error {
	s := outstream.New()
	s.WriteUint64(uint64(monotonicTimeMs))
	s.WriteOctets(stateOctets)
	return c.outFile.WriteChunkTypeIDString("sta1", s.Octets())
}

func (c *OutPacketFile) Close() {
	c.outFile.Close()
}
