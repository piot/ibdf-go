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

	writeStateErr := f.DebugState([]byte("this is a state"), 41)
	if writeStateErr != nil {
		t.Fatal(writeStateErr)
	}
	writeErr := f.DebugOutgoingPacket(octetPayload, 42)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	f.Close()

	i, openErr := NewInPacketFile(ibdFilename)
	if openErr != nil {
		t.Fatal(openErr)
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

	stateTime, state, seekErr := i.SeekAndGetState(0)
	if seekErr != nil {
		t.Errorf("couldnt seek")
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
