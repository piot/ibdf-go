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
	"io"
	"os"

	"github.com/piot/brook-go/src/instream"
	"github.com/piot/piff-go/src/piff"
)

type PacketIndex uint32

type PacketType uint8

const (
	PacketTypeState PacketType = iota
	PacketTypeNormal
	PacketTypeOther
)

type HeaderInfo struct {
	packetIndex PacketIndex
	packetType  PacketType
	timestamp   int64
	direction   PacketDirection
	octetCount  int
}

func (h HeaderInfo) PacketIndex() PacketIndex {
	return h.packetIndex
}

func (h HeaderInfo) PacketType() PacketType {
	return h.packetType
}

func (h HeaderInfo) Timestamp() int64 {
	return h.timestamp
}

func (h HeaderInfo) PacketDirection() PacketDirection {
	return h.direction
}

func (h HeaderInfo) OctetCount() int {
	return h.octetCount
}

func (h HeaderInfo) String() string {
	return fmt.Sprintf("index:%v type:%v time:%v octetCount:%v", h.packetIndex, h.packetType, h.timestamp, h.octetCount)
}

type InPacketFile struct {
	inFile        *piff.InSeeker
	schemaPayload []byte
	infos         []*HeaderInfo

	startTime int64
	endTime   int64
}

func (c *InPacketFile) AllHeaders() []*HeaderInfo {
	return c.infos
}

func (c *InPacketFile) SchemaPayload() []byte {
	return c.schemaPayload
}

func (c *InPacketFile) readSchema() ([]byte, error) {
	header, payload, readErr := c.inFile.FindChunk(0)
	if readErr != nil {
		return nil, fmt.Errorf("read schema %v", readErr)
	}
	if header.TypeIDString() != "sch1" {
		return nil, fmt.Errorf("wrong schema typeid %v", header)
	}

	return payload, nil
}

func (c *InPacketFile) scanAllChunks() error {
	var infos []*HeaderInfo
	foundSomeState := false
	for packetIndex, seekHeader := range c.inFile.AllHeaders() {
		id := seekHeader.Header().TypeIDString()
		var headerInfo *HeaderInfo
		switch id {
		case "sta1":
			header, octets, foundErr := c.inFile.FindPartialChunk(packetIndex, 8)
			if foundErr != nil {
				return foundErr
			}
			stream := instream.New(octets)
			timestamp, timestampErr := stream.ReadUint64()
			if timestampErr != nil {
				return timestampErr
			}
			headerInfo = &HeaderInfo{packetType: PacketTypeState, packetIndex: PacketIndex(packetIndex), timestamp: int64(timestamp), octetCount: header.OctetCount()}
			foundSomeState = true
		case "pkt1":
			header, payload, foundErr := c.inFile.FindPartialChunk(packetIndex, pktHeaderOctetCount)
			if foundErr != nil {
				return foundErr
			}
			direction, time, deserializeErr := deserializePacketHeader(header, payload)
			if deserializeErr != nil {
				return deserializeErr
			}
			headerInfo = &HeaderInfo{packetType: PacketTypeNormal, packetIndex: PacketIndex(packetIndex), timestamp: int64(time), direction: direction, octetCount: header.OctetCount()}
		case "sch1": // do nothing
			headerInfo = &HeaderInfo{packetType: PacketTypeOther, packetIndex: PacketIndex(packetIndex), direction: CmdIncomingPacket}
		default:
			return fmt.Errorf("unknown type id %s", id)
		}
		infos = append(infos, headerInfo)
	}

	c.infos = infos
	if !foundSomeState {
		return &MissingStateError{}
	}

	return nil
}

func (c *InPacketFile) getInfo(packetIndex PacketIndex) *HeaderInfo {
	return c.infos[packetIndex]
}

func (c *InPacketFile) IsState(packetIndex PacketIndex) bool {
	if c.IsEOF(packetIndex) {
		return false
	}
	info := c.getInfo(packetIndex)
	return info.packetType == PacketTypeState
}

func (c *InPacketFile) IsPacket(packetIndex PacketIndex) bool {
	if c.IsEOF(packetIndex) {
		return false
	}
	info := c.getInfo(packetIndex)
	return info.packetType == PacketTypeNormal
}

func (c *InPacketFile) FindClosestStateBeforeOrAt(timestamp int64) *HeaderInfo {
	var foundStateInfo *HeaderInfo
	for _, info := range c.infos {
		if info.packetType != PacketTypeState {
			continue
		}
		if info.timestamp > timestamp {
			return foundStateInfo
		}
		foundStateInfo = info
	}
	return foundStateInfo
}

func NewInPacketFile(filename string) (*InPacketFile, error) {
	seeker, openErr := os.Open(filename)
	if openErr != nil {
		return nil, openErr
	}
	return NewInPacketFileFromSeeker(seeker)
}

func NewInPacketFileFromSeeker(readSeeker io.ReadSeeker) (*InPacketFile, error) {
	newPiffFile, err := piff.NewInSeeker(readSeeker)
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

	scanChunksErr := c.scanAllChunks()
	if scanChunksErr != nil {
		return c, scanChunksErr
	}
	return c, err
}

func (c *InPacketFile) IsEOF(packetIndex PacketIndex) bool {
	return int(packetIndex) >= len(c.infos)
}

func (c *InPacketFile) ReadPacket(packetIndex PacketIndex) (PacketDirection, uint64, []byte, error) {
	if c.IsEOF(packetIndex) {
		return 0, 0, nil, io.EOF
	}
	if c.IsState(packetIndex) {
		return 0, 0, nil, fmt.Errorf("read packet (%v): wrong packet type (encountered a state)", packetIndex)
	}
	header, payload, readErr := c.inFile.FindChunk(int(packetIndex))
	if readErr != nil {
		return 0, 0, nil, readErr
	}
	return deserializePacketFromPiffPayload(header, payload)
}

func (c *InPacketFile) ReadStatePacket(packetIndex PacketIndex) (uint64, []byte, error) {
	if c.IsEOF(packetIndex) {
		return 0, nil, io.EOF
	}
	if !c.IsState(packetIndex) {
		return 0, nil, fmt.Errorf("read state packet (%v): wrong packet type", packetIndex)
	}
	header, payload, readErr := c.inFile.FindChunk(int(packetIndex))
	if readErr != nil {
		return 0, nil, readErr
	}
	return deserializeStatePacketFromPiffPayload(header, payload)
}

func (c *InPacketFile) Close() {
	c.inFile.Close()
}
