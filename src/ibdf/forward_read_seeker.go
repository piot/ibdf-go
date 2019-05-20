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

import "io"

type ForwardReadSeeker struct {
	reader   io.Reader
	position int64
}

func NewForwardReadSeeker(reader io.Reader) *ForwardReadSeeker {
	return &ForwardReadSeeker{reader: reader}
}

func (f *ForwardReadSeeker) Seek(offset int64, whence int) (int64, error) {
	requestedPosition := offset
	switch whence {
	case io.SeekStart: // means relative to the start of the file
	// do nothing
	case io.SeekCurrent: // means relative to the current offset
		requestedPosition = f.position + offset
	case io.SeekEnd: //means relative to the end
		panic("not supported")
	}

	if requestedPosition < f.position {
		panic("not supported")
	}

	return 0, nil
}

func (f *ForwardReadSeeker) Read(p []byte) (n int, err error) {
	octetsRead, readErr := f.reader.Read(p)
	if readErr != nil {
		return 0, readErr
	}
	f.position += int64(octetsRead)
	return octetsRead, readErr
}
