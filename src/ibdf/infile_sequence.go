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
)

type InPacketFileSequence struct {
	inFile            *InPacketFile
	cursorPacketIndex PacketIndex
}

func (c *InPacketFileSequence) CursorAtState() bool {
	if c.IsEOF() {
		return false
	}
	info := c.inFile.getInfo(c.cursorPacketIndex)
	return info.packetType == PacketTypeState
}

func (c *InPacketFileSequence) CursorAtPacket() bool {
	if c.IsEOF() {
		return false
	}
	info := c.inFile.getInfo(c.cursorPacketIndex)
	return info.packetType == PacketTypeNormal
}

func (c *InPacketFileSequence) seekToClosestState(timestamp int64) error {
	headerInfo := c.inFile.FindClosestStateBeforeOrAt(timestamp)
	if headerInfo == nil {
		return fmt.Errorf("couldn't find any states at timestamp %d", timestamp)
	}
	c.cursorPacketIndex = headerInfo.packetIndex
	return nil
}

func (c *InPacketFileSequence) SeekAndGetState(timestamp int64) (uint64, []byte, error) {
	err := c.seekToClosestState(timestamp)
	if err != nil {
		return 0, nil, err
	}
	return c.ReadNextStatePacket()
}

func (c *InPacketFileSequence) Cursor() PacketIndex {
	return c.cursorPacketIndex
}

func NewInPacketFileSequenceFromInFile(inFile *InPacketFile) (*InPacketFileSequence, error) {
	c := &InPacketFileSequence{
		inFile:            inFile,
		cursorPacketIndex: 1,
	}

	return c, nil
}

func (c *InPacketFileSequence) advanceCursor() {
	c.cursorPacketIndex++
}

func (c *InPacketFileSequence) IsEOF() bool {
	return int(c.cursorPacketIndex) >= len(c.inFile.infos)
}

func (c *InPacketFileSequence) ReadNextPacket() (PacketDirection, uint64, []byte, error) {
	if c.IsEOF() {
		return 0, 0, nil, io.EOF
	}
	for c.CursorAtState() {
		c.advanceCursor()
	}
	if c.IsEOF() {
		return 0, 0, nil, io.EOF
	}
	direction, time, payload, readErr := c.inFile.ReadPacket(c.cursorPacketIndex)
	if readErr != nil {
		return 0, 0, nil, readErr
	}
	c.advanceCursor()
	return direction, time, payload, nil
}

func (c *InPacketFileSequence) ReadNextStatePacket() (uint64, []byte, error) {
	if c.IsEOF() {
		return 0, nil, io.EOF
	}
	time, payload, readErr := c.inFile.ReadStatePacket(c.cursorPacketIndex)
	if readErr != nil {
		return 0, nil, readErr
	}
	c.advanceCursor()
	return time, payload, nil
}

func (c *InPacketFileSequence) Close() {
	c.inFile.Close()
}
