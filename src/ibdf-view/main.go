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

package main

import (
	"encoding/base64"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/piot/ibdf-go/src/ibdf"
	"github.com/piot/log-go/src/clog"
)

func options() string {
	//var piffFile string
	//	flag.StringVar(&piffFile, "filename", "", "file to view")
	flag.Parse()
	count := flag.NArg()
	if count < 1 {
		return ""
	}
	ibdfFilename := flag.Arg(0)
	return ibdfFilename
}

func cmdToString(direction ibdf.PacketDirection) string {
	switch direction {
	case ibdf.CmdIncomingPacket:
		return "<< (in) "
	case ibdf.CmdOutgoingPacket:
		return ">> (out) "
	default:
		panic("unknown direction")
	}
}

func openReadSeeker(filename string) (io.ReadSeeker, error) {
	var seekerToUse io.ReadSeeker
	if filename == "" {
		seekerToUse = os.Stdin
	} else {
		newFile, err := os.Open(filename)
		if err != nil {
			return nil, err
		}
		seekerToUse = newFile
	}

	return seekerToUse, nil
}

func octetsToString(payload []byte) string {
	base64String := base64.StdEncoding.EncodeToString(payload)
	return strings.TrimSpace(hex.Dump(payload)) + "\n" + base64String + "\n"
}

func run(filename string, log *clog.Log) error {
	seekerToUse, seekerErr := openReadSeeker(filename)
	if seekerErr != nil {
		return seekerErr
	}
	inStream, err := ibdf.NewInPacketStream(seekerToUse)
	if err != nil {
		_, isStateError := err.(*ibdf.MissingStateError)
		if !isStateError {
			return err
		}
	}

	header, headerErr := inStream.ReadNextFileHeader()
	if headerErr != nil {
		return headerErr
	}
	color.HiMagenta(header.String())

	fmt.Printf("schema:\n")
	if !inStream.IsNextSchema() {
		return fmt.Errorf("Must start with schema")
	}
	schemaString, schemaErr := inStream.ReadNextSchemaTextPacket()
	if schemaErr != nil {
		return schemaErr
	}
	color.HiGreen("%v\n", schemaString)

	for packetIndex := 0; ; packetIndex++ {
		if inStream.IsEOF() {
			break
		}
		if inStream.IsNextPacket() {
			cmd, time, payload, readErr := inStream.ReadNextPacket()
			if readErr == io.EOF {
				break
			}
			cmdString := cmdToString(cmd)
			headerColor := color.New(color.FgMagenta)
			payloadColor := color.New(color.FgHiMagenta)
			if cmd == ibdf.CmdOutgoingPacket {
				headerColor = color.New(color.FgBlue)
				payloadColor = color.New(color.FgHiBlue)
			}

			headerColor.Printf("#%04d %s time:%v (%v octets)\n", packetIndex, cmdString, time, len(payload))
			payloadColor.Println(octetsToString(payload))
		} else if inStream.IsNextState() {
			time, statePayload, readErr := inStream.ReadNextStatePacket()
			if readErr == io.EOF {
				break
			}
			color.Cyan("#%04d * (state) time:%v (%v octets)", packetIndex, time, len(statePayload))
			color.HiCyan(octetsToString(statePayload))
			fmt.Println("")
		}
	}
	return nil
}

func main() {
	log := clog.DefaultLog()
	log.Info("ibdf viewer")
	filename := options()
	err := run(filename, log)
	if err != nil {
		log.Err(err)
		os.Exit(1)
	}
	log.Info("Done!")
}
