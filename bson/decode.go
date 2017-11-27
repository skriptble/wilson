package bson

import "io"

type Unmarshaler interface {
	UnmarshalBSON([]byte) error
}

type Decoder struct {
}

func NewDecoder(r io.Reader) *Decoder {
	return nil
}

func (d *Decoder) Decode(v interface{}) error {
	return nil
}
