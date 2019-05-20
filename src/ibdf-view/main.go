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
		return "<< (int)"
	case ibdf.CmdOutgoingPacket:
		return ">> (out)"
	default:
		panic("unknown direction")
	}
}

func run(filename string, log *clog.Log) error {
	inFile, err := ibdf.NewInPacketFile(filename)
	if err != nil {
		_, isStateError := err.(*ibdf.MissingStateError)
		if !isStateError {
			return err
		}
	}

	fmt.Printf("schema:\n")
	color.HiGreen("%v\n", string(inFile.SchemaPayload()))

	for packetIndex := 0; ; packetIndex++ {
		if inFile.IsEOF() {
			break
		}
		if inFile.CursorAtPacket() {
			cmd, time, payload, readErr := inFile.ReadNextPacket()
			if readErr == io.EOF {
				break
			}
			cmdString := cmdToString(cmd)
			color.HiMagenta("#%d %s time:%v %v", packetIndex, cmdString, time, strings.TrimSpace(hex.Dump(payload)))
		} else if inFile.CursorAtState() {
			time, statePayload, readErr := inFile.ReadNextStatePacket()
			if readErr == io.EOF {
				break
			}
			color.Cyan("#%d * (state) time:%v %v", packetIndex, time, hex.Dump(statePayload))
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
