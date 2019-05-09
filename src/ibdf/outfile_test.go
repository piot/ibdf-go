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
	writeErr := f.DebugOutgoingPacket(octetPayload, 42)
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	f.Close()

	i, openErr := NewInPacketFile(ibdFilename)
	if openErr != nil {
		t.Fatal(openErr)
	}

	cmd, time, payload, readErr := i.ReadPacket()

	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(payload) != len(octetPayload) {
		t.Errorf("wrong octet length expected %v but got %v ('%s')", len(testString), len(payload), hex.Dump(payload))
	}

	if cmd != 0x81 {
		t.Errorf("wrong direction")
	}

	if time != 42 {
		t.Errorf("wrong time")
	}

	if string(payload) != testString {
		t.Errorf("wrong string")
	}
	_, _, _, nextReadErr := i.ReadPacket()
	if nextReadErr != io.EOF {
		t.Errorf("file should have ended")
	}
	i.Close()
}
