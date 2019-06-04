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
	"encoding/hex"
	"io"
	"testing"
)

func TestReadWritePacket(t *testing.T) {
	const testString = "this is a string"
	const ibdFilename = "test.ibdf"
	const typeID = "cafe"
	octetPayload := []byte(testString)
	f, outErr := NewOutPacketFile(ibdFilename, nil)
	if outErr != nil {
		t.Fatal(outErr)
	}
	firstKnownTime := int64(41)

	writeStateErr := f.DebugState([]byte("this is a state"), firstKnownTime)
	if writeStateErr != nil {
		t.Fatal(writeStateErr)
	}
	writeErr := f.DebugOutgoingPacket(octetPayload, 42)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	f.Close()

	pf, openErr := NewInPacketFile(ibdFilename)
	if openErr != nil {
		t.Fatal(openErr)
	}
	i, iErr := NewInPacketFileSequenceFromInFile(pf)
	if iErr != nil {
		t.Fatal(iErr)
	}
	if !i.CursorAtState() {
		t.Errorf("strange, we should be at a state packet now")
	}

	time, statePayload, stateReadErr := i.ReadNextStatePacket()
	if stateReadErr != nil {
		t.Error(stateReadErr)
	}
	if time != 41 {
		t.Errorf("wrong time")
	}
	statePayloadString := string(statePayload)
	if statePayloadString != "this is a state" {
		t.Errorf("wrong state payload")
	}

	if !i.CursorAtPacket() {
		t.Errorf("strange, we should be at a normal packet now")
	}
	cmd, time, payload, readErr := i.ReadNextPacket()

	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(payload) != len(octetPayload) {
		t.Errorf("wrong octet length expected %v but got %v ('%s')", len(testString), len(payload), hex.Dump(payload))
	}

	if cmd != CmdOutgoingPacket {
		t.Errorf("wrong direction")
	}

	if time != 42 {
		t.Errorf("wrong time")
	}

	if string(payload) != testString {
		t.Errorf("wrong string")
	}
	_, _, _, nextReadErr := i.ReadNextPacket()
	if nextReadErr != io.EOF {
		t.Errorf("file should have ended")
	}

	stateTime, state, seekErr := i.SeekAndGetState(firstKnownTime + 10)
	if seekErr != nil {
		t.Errorf("couldnt seek %v", seekErr)
	}
	if stateTime != 41 {
		t.Errorf("unexpected state time %v", stateTime)
	}
	if string(state) != "this is a state" {
		t.Errorf("problem with state")
	}

	afterStateCmd, afterStateTime, payload, readErr := i.ReadNextPacket()
	if cmd != 0x81 {
		t.Errorf("wrong direction")
	}
	if afterStateTime != 42 {
		t.Errorf("wrong after time")
	}
	if afterStateCmd != CmdOutgoingPacket {
		t.Errorf("wrong after packet")
	}

	i.Close()
}
