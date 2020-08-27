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
	"github.com/piot/piff-go/src/piff"
)

const pktHeaderOctetCount = 1 + 8
const pktHeaderStateOctetCount = 8

func deserializeStateHeader(header piff.InHeader, payload []byte) (uint64, error) {
	if header.TypeIDString() != "sta1" {
		return 0, fmt.Errorf("wrong typeid %v", header)
	}
	if len(payload) < pktHeaderStateOctetCount {
		return 0, fmt.Errorf("wrong serialized header size")
	}
	s := instream.New(payload)
	monotonicTimeMs, timeMsErr := s.ReadUint64()
	if timeMsErr != nil {
		return 0, timeMsErr
	}
	return monotonicTimeMs, nil
}

func deserializeStateHeaderFromStream(stream *piff.InStream) (uint64, error) {
	header, payload, readErr := stream.ReadPartChunk(pktHeaderStateOctetCount)
	if readErr != nil {
		return 0, readErr
	}
	return deserializeStateHeader(header, payload)
}

func deserializePacketHeader(header piff.InHeader, payload []byte) (PacketDirection, uint64, error) {
	if header.TypeIDString() != "pkt1" {
		return 0, 0, fmt.Errorf("wrong typeid %v", header)
	}
	if len(payload) < pktHeaderOctetCount {
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

func deserializePacketFromStream(stream *piff.InStream) (piff.ChunkIndex, PacketDirection, uint64, []byte, error) {
	header, payload, readErr := stream.ReadChunk()
	if readErr != nil {
		return 0, CmdIncomingPacket, 0, nil, readErr
	}
	return deserializePacketFromPiffPayload(header, payload)
}

func deserializePacketFromPiffPayload(header piff.InHeader, payload []byte) (piff.ChunkIndex, PacketDirection, uint64, []byte, error) {
	pktHeaderOctets := payload[:pktHeaderOctetCount]
	cmd, monotonicTimeMs, serializeErr := deserializePacketHeader(header, pktHeaderOctets)
	if serializeErr != nil {
		return 0, 0, 0, nil, serializeErr
	}
	return header.ChunkIndex(), cmd, monotonicTimeMs, payload[pktHeaderOctetCount:], nil
}

func deserializeStatePacketFromStream(stream *piff.InStream) (piff.ChunkIndex, uint64, []byte, error) {
	header, payload, readErr := stream.ReadChunk()
	if readErr != nil {
		return 0, 0, nil, readErr
	}
	return deserializeStatePacketFromPiffPayload(header, payload)
}

func deserializeStatePacketFromPiffPayload(header piff.InHeader, payload []byte) (piff.ChunkIndex, uint64, []byte, error) {
	stateHeaderOctets := payload[:pktHeaderStateOctetCount]
	monotonicTimeMs, serializeErr := deserializeStateHeader(header, stateHeaderOctets)
	if serializeErr != nil {
		return 0, 0, nil, serializeErr
	}
	return header.ChunkIndex(), monotonicTimeMs, payload[pktHeaderStateOctetCount:], nil
}

func deserializeSchemaTextFromStream(stream *piff.InStream) (string, error) {
	header, payload, readErr := stream.ReadChunk()
	if readErr != nil {
		return "", readErr
	}
	return deserializeSchemaTextFromPiffPayload(header, payload)
}

func deserializeSchemaTextFromPiffPayload(header piff.InHeader, payload []byte) (string, error) {
	return string(payload), nil
}
