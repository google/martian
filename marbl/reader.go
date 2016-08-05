// Copyright 2015 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package marbl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// Reader wraps a buffered Reader that read from the reader and emit Frames.
type Reader struct {
	r io.Reader
}

// NewReader returns a Reader initialized with a buffered reader.
func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}

// ReadFrame reads from the reader and returns either a HeaderFrame or a BodyFrame
// and an error.
func (r *Reader) ReadFrame() (Frame, error) {
	fh := make([]byte, 10)

	if _, err := io.ReadFull(r.r, fh); err != nil {
		return nil, err
	}

	switch FrameType(fh[0]) {
	case HeaderFrame:
		hf := headerFrame{
			ID:          string(fh[2:]),
			MessageType: MessageType(fh[1]),
		}

		lens := make([]byte, 8)
		if _, err := io.ReadFull(r.r, lens); err != nil {
			return nil, err
		}

		nl := binary.BigEndian.Uint32(lens[:4])
		vl := binary.BigEndian.Uint32(lens[4:])

		nv := make([]byte, int(nl+vl))
		if _, err := io.ReadFull(r.r, nv); err != nil {
			return nil, err
		}

		hf.Name = string(nv[:nl])
		hf.Value = string(nv[nl:])

		return hf, nil
	case BodyFrame:
		df := dataFrame{
			ID:          string(fh[2:]),
			MessageType: MessageType(fh[1]),
		}

		dlen := make([]byte, 4)
		if _, err := io.ReadFull(r.r, dlen); err != nil {
			return nil, err
		}

		dl := binary.BigEndian.Uint32(dlen[:4])

		data := make([]byte, int(dl))
		if _, err := io.ReadFull(r.r, data); err != nil {
			return nil, err
		}

		df.Data = data

		return df, nil
	default:
		return nil, fmt.Errorf("logstream: unknown frame")
	}
}

// Frame describes the interface for a frame (either BodyFrame or HeaderFrame).
type Frame interface {
	String() string
}

type headerFrame struct {
	ID          string
	MessageType MessageType
	Name        string
	Value       string
}

// String returns the contents of a headerFrame in a format appropriate for debugging and runtime logging.
func (hf headerFrame) String() string {
	return fmt.Sprintf("ID=%s; Type=%d; Name=%s; Value=%s", hf.ID, hf.MessageType, hf.Name, hf.Value)
}

type dataFrame struct {
	ID          string
	MessageType MessageType
	Data        []byte
}

// String returns the contents of a dataFrame in a format appropriate for debugging and runtime logging. In the
// case that the Data is greater than 20 characters, the output is truncated at the 20th character.
func (df dataFrame) String() string {
	dl := len(df.Data)
	if dl > 20 {
		dl = 20
	}

	return fmt.Sprintf("ID=%s; Type=%d; Data=%q", df.ID, df.MessageType, df.Data[:dl])
}
