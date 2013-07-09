package protocol

import (
	"encoding/json"
	"io"
)

type MessageType int

const (
	TypeCreate MessageType = iota
	TypeChange
	TypeFailure
)

type Message struct {
	Pixel   string      `json:"pixel"`
	Type    MessageType `json:"type"`
	Payload string      `json:"payload"`
}

type Decoder struct {
	dec *json.Decoder
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		dec: json.NewDecoder(r),
	}
}

func (d *Decoder) Decode() (*Message, error) {
	var m *Message
	err := d.dec.Decode(&m)
	return m, err
}

type Encoder struct {
	enc *json.Encoder
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		enc: json.NewEncoder(w),
	}
}

func (e *Encoder) Encode(m *Message) error {
	return e.enc.Encode(m)
}
