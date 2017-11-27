package bson

import (
	"errors"
)

// ErrUninitializedElement is returned whenever any method is invoked on an unintialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")
var ErrTooSmall = errors.New("too small")

type ElementTypeError struct {
	Method string
	Type   BSONType
}

func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

type Element struct {
	ReaderElement
}
