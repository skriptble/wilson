package bson

import "io"

type Marshaler interface {
	MarshalBSON() ([]byte, error)
}

type Encoder struct {
}

func NewEncoder(w io.Writer) *Encoder {
	return nil
}

func (e *Encoder) Encode(v interface{}) error {
	return nil
}
