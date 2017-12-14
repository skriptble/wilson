package bson

import (
	"encoding/binary"
	"errors"
	"io"
)

// ErrUninitializedElement is returned whenever any method is invoked on an unintialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")
var ErrTooSmall = errors.New("too small")
var ErrInvalidWriter = errors.New("bson: invalid writer provided")

type ElementTypeError struct {
	Method string
	Type   BSONType
}

func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

type Element struct {
	// NOTE: For SubDocuments and Arrays, the data byte slice may contain just
	// the key, in which case the start will be 0, the value will be the length
	// of the slice, and d will be non-nil.
	ReaderElement
	d *Document
}

// Validates the element and if recursive is true, validates all subdocuments
// and arrays.
func (e *Element) Validate(recursive bool) (uint32, error) {
	var total uint32 = 1
	n, err := e.keySize()
	total += n
	if err != nil {
		return total, err
	}
	n, err = e.validateValue(recursive)
	total += n
	if err != nil {
		return total, err
	}
	return total, nil
}

// validate is a common validation method for elements.
//
// TODO(skriptble): Fill out this method and ensure all validation routines
// pass through this method.
func (e *Element) validate(recursive bool, currentDepth, maxDepth uint32) (uint32, error) {
	return 0, nil
}

// validateValue will validate that the element is properly formed. This method
// wraps the ReaderElement's validateValue command except for the case of an
// element that has a document.
func (e *Element) validateValue(recursive bool) (uint32, error) {
	if e.d == nil {
		return e.ReaderElement.validateValue(recursive)
	}

	var total uint32 = 0

	switch e.data[e.start] {
	case '\x03', '\x04':
		n, err := e.d.Validate()
		total += uint32(n)
		if err != nil {
			return total, err
		}
	case '\x0F':
		if int(e.value+4) > len(e.data) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+l > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		sLength := int32(binary.LittleEndian.Uint32(e.data[e.value+4 : e.value+8]))
		// If the length of the string is larger than the total length of the
		// field minus the int32 for length, 5 bytes for a minimum document
		// size, and an int32 for the string length
		if sLength > l-13 {
			return total, errors.New("String size is larger than the Code With Scope container")
		}
		total += uint32(sLength)

		n, err := e.d.Validate()
		total += uint32(n)
		if err != nil {
			return total, err
		}
	default:
		// These are the only types that have a document attached to them.
		// It's an error if we get here.
		return 0, errors.New("Invalid element")
	}

	return total, nil
}

// Document returns the subdocument for this element.
func (e *Element) Document() *Document {
	if e == nil || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x03' {
		panic(ElementTypeError{"compact.Element.Document", BSONType(e.data[e.start])})
	}
	if e.d == nil {
		var err error
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		e.d, err = ReadDocument(e.data[e.value : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return e.d
}

// Array returns the array for this element.
func (e *Element) Array() *Array {
	if e == nil || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x04' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	if e.d == nil {
		var err error
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		e.d, err = ReadDocument(e.data[e.value : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return &Array{e.d}
}

// JavascriptWithScope returns the javascript code and the scope document for
// this element
func (e *Element) JavascriptWithScope() (code string, d *Document) {
	if e == nil || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0F' {
		panic(ElementTypeError{"compact.Element.JavascriptWithScope", BSONType(e.data[e.start])})
	}
	// TODO(skriptble): This is wrong and could cause a panic.
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	// TODO(skriptble): This is wrong and could cause a panic.
	sLength := int32(binary.LittleEndian.Uint32(e.data[e.value+4 : e.value+8]))
	// If the length of the string is larger than the total length of the
	// field minus the int32 for length, 5 bytes for a minimum document
	// size, and an int32 for the string length the value is invalid.
	str := string(e.data[e.value+4 : e.value+4+uint32(sLength)])
	if e.d == nil {
		var err error
		e.d, err = ReadDocument(e.data[e.value+4+uint32(sLength) : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return str, e.d
}

// WriteTo implements the io.WriterTo interface.
func (e *Element) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

// WriteElement serializes this element to the provided writer starting at the
// provided start position.
func (e *Element) WriteElement(start uint, writer interface{}) (int64, error) {
	// TODO(skriptble): Figure out if we want to use uint or uint32 and
	// standardize across all packages.
	var total int64
	size, err := e.Validate(true)
	if err != nil {
		return 0, err
	}
	switch w := writer.(type) {
	case []byte:
		n, err := e.writeByteSlice(start, size, w)
		if err != nil {
			return 0, ErrTooSmall
		}
		total += int64(n)
	default:
		return 0, ErrInvalidWriter
	}
	return total, nil
}

// MarshalBSON implements the Marshaler interface.
func (e *Element) MarshalBSON() ([]byte, error) {
	size, err := e.Validate(true)
	if err != nil {
		return nil, err
	}
	b := make([]byte, size)
	_, err = e.writeByteSlice(0, size, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// writeByteSlice handles writing this element to a slice of bytes.
func (e *Element) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	if len(b) < int(size)+int(start) {
		return 0, ErrTooSmall
	}
	var n int
	switch e.data[e.start] {
	// TODO(skriptble): For types that contain a document, we need to check if
	// the d property is nil. If it is, we can do a regular copy. If it is
	// non-nil we need to marshal that Document to bytes then copy it into the
	// byte slice.
	case '\x03', '\x04':
		if e.d == nil {
			n = copy(b[start:start+uint(size)], e.data[e.start:e.start+size])
			break
		}
		header := e.value - e.start
		n += copy(b[start:start+uint(header)], e.data[e.start:e.value])
		start += uint(n)
		size -= header
		nn, err := e.d.writeByteSlice(start, size, b)
		n += int(nn)
		if err != nil {
			return int64(n), err
		}
	case '\x0F':
	default:
		n = copy(b[start:start+uint(size)], e.data[e.start:e.start+size])
	}
	return int64(n), nil
}
