package marbl

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

type Reader struct {
	r io.Reader
}

type headerFrame struct {
	ID          string
	MessageType MessageType
	Name        string
	Value       string
}

func (hf headerFrame) String() string {
	return fmt.Sprintf("ID=%s; Type=%d; Name=%s; Value=%s", hf.ID, hf.MessageType, hf.Name, hf.Value)
}

type dataFrame struct {
	ID          string
	MessageType MessageType
	Data        []byte
}

func (df dataFrame) String() string {
	dl := len(df.Data)
	if dl > 20 {
		dl = 20
	}

	return fmt.Sprintf("ID=%s; Type=%d; Data=%q", df.ID, df.MessageType, df.Data[:dl])
}

type Frame interface {
	String() string
}

func NewReader(r io.Reader) *Reader {
	return &Reader{
		r: bufio.NewReader(r),
	}
}

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
	case DataFrame:
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
