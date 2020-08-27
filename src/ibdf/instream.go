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
	"io"

	"github.com/piot/brook-go/src/instream"
	"github.com/piot/piff-go/src/piff"
)

type InStream struct {
	stream *piff.InStream
}

func NewInPacketStream(reader io.Reader) (*InStream, error) {
	seeker := NewForwardReadSeeker(reader)
	stream, readErr := piff.NewInStreamReadSeeker(seeker)
	if readErr != nil {
		return nil, readErr
	}

	return &InStream{stream: stream}, nil
}

func (i *InStream) IsNextSchema() bool {
	piffHeader := i.stream.PendingChunkHeader()
	return piffHeader.TypeIDString() == "sch1"
}

func (i *InStream) IsNextState() bool {
	piffHeader := i.stream.PendingChunkHeader()
	return piffHeader.TypeIDString() == "sta1"
}

func (i *InStream) IsNextPacket() bool {
	piffHeader := i.stream.PendingChunkHeader()
	return piffHeader.TypeIDString() == "pkt1"
}

func (i *InStream) IsNextFileHeader() bool {
	piffHeader := i.stream.PendingChunkHeader()
	return piffHeader.TypeIDString() == "pac1"
}

func (i *InStream) ReadNextPacket() (piff.ChunkIndex, PacketDirection, uint64, []byte, error) {
	return deserializePacketFromStream(i.stream)
}

func (i *InStream) ReadNextStatePacket() (piff.ChunkIndex, uint64, []byte, error) {
	return deserializeStatePacketFromStream(i.stream)
}

func (i *InStream) ReadNextSchemaTextPacket() (string, error) {
	return deserializeSchemaTextFromStream(i.stream)
}

func (i *InStream) ReadNextFileHeader() (Header, error) {
	_, payload, err := i.stream.ReadChunk()
	if err != nil {
		return Header{}, err
	}
	stream := instream.New(payload)
	return readHeader(stream)
}

func (i *InStream) IsEOF() bool {
	return i.stream.IsEOF()
}
