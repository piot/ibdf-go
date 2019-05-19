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

	"github.com/piot/brook-go/src/instream"
	"github.com/piot/piff-go/src/piff"
)

type PacketIndex uint32

type PacketType uint8

const (
	PacketTypeState PacketType = iota
	PacketTypeNormal
)

type HeaderInfo struct {
	packetIndex PacketIndex
	packetType  PacketType
	timestamp   int64
}

type InPacketFile struct {
	inFile            *piff.InSeeker
	schemaPayload     []byte
	infos             []HeaderInfo
	cursorPacketIndex PacketIndex

	startTime int64
	endTime   int64
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
	var infos []HeaderInfo
	foundSomeState := false
	for packetIndex, seekHeader := range c.inFile.AllHeaders() {
		id := seekHeader.Header().TypeIDString()
		var headerInfo HeaderInfo
		switch id {
		case "sta1":
			_, octets, foundErr := c.inFile.FindPartialChunk(packetIndex, 8)
			if foundErr != nil {
				return foundErr
			}
			stream := instream.New(octets)
			timestamp, timestampErr := stream.ReadUint64()
			if timestampErr != nil {
				return timestampErr
			}
			headerInfo = HeaderInfo{packetType: PacketTypeState, packetIndex: PacketIndex(packetIndex), timestamp: int64(timestamp)}
			foundSomeState = true
		case "pkt1":
			headerInfo = HeaderInfo{packetType: PacketTypeNormal, packetIndex: PacketIndex(packetIndex), timestamp: 0}
		case "sch1": // do nothing
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

func (c *InPacketFile) getInfo(packetIndex PacketIndex) HeaderInfo {
	return c.infos[packetIndex]
}

func (c *InPacketFile) CursorAtState() bool {
	info := c.getInfo(c.cursorPacketIndex)
	return info.packetType == PacketTypeState
}

func (c *InPacketFile) CursorAtPacket() bool {
	info := c.getInfo(c.cursorPacketIndex)
	return info.packetType == PacketTypeNormal
}

func (c *InPacketFile) findClosestState(timestamp int64) *HeaderInfo {
	var foundStateInfo *HeaderInfo
	for _, info := range c.infos {
		if info.packetType != PacketTypeState {
			continue
		}
		if info.timestamp > timestamp {
			return foundStateInfo
		}
		foundStateInfo = &info
	}
	return foundStateInfo
}

func (c *InPacketFile) seekToClosestState(timestamp int64) error {
	headerInfo := c.findClosestState(timestamp)
	if headerInfo == nil {
		return fmt.Errorf("couldn't find any states at timestamp %d", timestamp)
	}
	c.cursorPacketIndex = headerInfo.packetIndex
	return nil
}

func (c *InPacketFile) SeekAndGetState(timestamp int64) (uint64, []byte, error) {
	err := c.seekToClosestState(timestamp)
	if err != nil {
		return 0, nil, err
	}
	return c.ReadNextStatePacket()
}

func (c *InPacketFile) Cursor() PacketIndex {
	return c.cursorPacketIndex
}

func NewInPacketFile(filename string) (*InPacketFile, error) {
	newPiffFile, err := piff.NewInSeekerFile(filename)
	if err != nil {
		return nil, err
	}
	c := &InPacketFile{
		inFile:            newPiffFile,
		cursorPacketIndex: 1,
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

const pktHeaderOctetCount = 1 + 8
const pktHeaderStateOctetCount = 8

func deserializeStateHeader(header piff.InHeader, payload []byte) (uint64, error) {
	if header.TypeIDString() != "sta1" {
		return 0, fmt.Errorf("wrong typeid %v", header)
	}
	if len(payload) != pktHeaderStateOctetCount {
		return 0, fmt.Errorf("wrong serialized header size")
	}
	s := instream.New(payload)
	monotonicTimeMs, timeMsErr := s.ReadUint64()
	if timeMsErr != nil {
		return 0, timeMsErr
	}
	return monotonicTimeMs, nil
}

func deserializePacketHeader(header piff.InHeader, payload []byte) (PacketDirection, uint64, error) {
	if header.TypeIDString() != "pkt1" {
		return 0, 0, fmt.Errorf("wrong typeid %v", header)
	}
	if len(payload) != pktHeaderOctetCount {
		return 0, 0, fmt.Errorf("wrong serialized header size")
	}
	s := instream.New(payload)
	cmdValue, cmdErr := s.ReadUint8()
	if cmdErr != nil {
		return 0, 0, cmdErr
	}
	switch cmdValue {
	case CmdIncomingPacket:
	case CmdOutgoingPacket:
	default:
		return CmdIncomingPacket, 0, fmt.Errorf("unknown direction")
	}
	monotonicTimeMs, timeMsErr := s.ReadUint64()
	if timeMsErr != nil {
		return 0, 0, timeMsErr
	}
	return cmdValue, monotonicTimeMs, nil
}

func (c *InPacketFile) advanceCursor() {
	c.cursorPacketIndex++
}

func (c *InPacketFile) IsEOF() bool {
	return int(c.cursorPacketIndex) >= len(c.infos)
}

func (c *InPacketFile) ReadNextPacket() (PacketDirection, uint64, []byte, error) {
	if c.IsEOF() {
		return 0, 0, nil, io.EOF
	}
	for c.CursorAtState() {
		c.advanceCursor()
	}
	if c.IsEOF() {
		return 0, 0, nil, io.EOF
	}
	header, payload, readErr := c.inFile.FindChunk(int(c.cursorPacketIndex))
	if readErr != nil {
		return 0, 0, nil, readErr
	}
	c.advanceCursor()
	pktHeaderOctets := payload[:pktHeaderOctetCount]
	cmd, monotonicTimeMs, serializeErr := deserializePacketHeader(header, pktHeaderOctets)
	if serializeErr != nil {
		return 0, 0, nil, serializeErr
	}
	return cmd, monotonicTimeMs, payload[pktHeaderOctetCount:], nil
}

func (c *InPacketFile) ReadNextStatePacket() (uint64, []byte, error) {
	if c.IsEOF() {
		return 0, nil, io.EOF
	}
	header, payload, readErr := c.inFile.FindChunk(int(c.cursorPacketIndex))
	if readErr != nil {
		return 0, nil, readErr
	}
	c.advanceCursor()
	pktHeaderStateOctets := payload[:pktHeaderStateOctetCount]
	monotonicTimeMs, serializeErr := deserializeStateHeader(header, pktHeaderStateOctets)
	if serializeErr != nil {
		return 0, nil, serializeErr
	}
	return monotonicTimeMs, payload[pktHeaderStateOctetCount:], nil
}

func (c *InPacketFile) Close() {
	c.inFile.Close()
}
