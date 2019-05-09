package piff

import (
	"io"
	"testing"
)

func TestReadWrite(t *testing.T) {
	const testString = "this is a string"
	const ibdFilename = "test.ibdf"
	const typeID = "cafe"
	f, outErr := NewOutFile(ibdFilename)
	if outErr != nil {
		t.Fatal(outErr)
	}
	writeErr := f.WriteChunkTypeIDString(typeID, []byte(testString))
	if writeErr != nil {
		t.Fatal(writeErr)
	}
	f.Close()

	i, _ := NewInFile(ibdFilename)

	header, payload, readErr := i.ReadChunk()

	if readErr != nil {
		t.Fatal(readErr)
	}
	if header.octetLength != len(testString) {
		t.Errorf("wrong octet length")
	}

	if header.typeID[1] != typeID[1] {
		t.Errorf("wrong typeid")
	}

	if string(payload) != testString {
		t.Errorf("wrong string")
	}
	_, _, nextReadErr := i.ReadChunk()
	if nextReadErr != io.EOF {
		t.Errorf("file should have ended")
	}
	i.Close()
}
