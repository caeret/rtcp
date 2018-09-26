package rtcp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

const prefix = "\x0F\x0F"

type Header struct {
	CMD        [16]byte
	BodyLength uint32
}

func (h Header) CMDStr() string {
	return string(bytes.TrimRight(h.CMD[:], "\x00"))
}

func Read(r io.Reader) (header Header, data []byte, err error) {
	b := make([]byte, 2)
	_, err = io.ReadAtLeast(r, b, 2)
	if err != nil {
		return
	}

	if string(b) != prefix {
		err = errors.New("invalid packet prefix")
		return
	}

	b = make([]byte, 20)
	n, err := r.Read(b)
	if err != nil {
		return
	}

	err = binary.Read(bytes.NewReader(b[:n]), binary.BigEndian, &header)
	if err != nil {
		return
	}
	data = make([]byte, header.BodyLength)
	_, err = io.ReadFull(r, data)
	return
}

func Write(w io.Writer, CMD []byte, data []byte) (err error) {
	_, err = w.Write([]byte(prefix))
	if err != nil {
		return
	}
	var header Header
	copy(header.CMD[:], CMD)
	header.BodyLength = uint32(len(data))
	err = binary.Write(w, binary.BigEndian, &header)
	if err != nil {
		return
	}
	if header.BodyLength > 0 {
		_, err = w.Write(data)
	}
	return
}
