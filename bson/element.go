package bson

import (
	"errors"
	"io"

	"github.com/skriptble/wilson/bson/elements"
)

const validateMaxDepthDefault = 2048

// ErrUninitializedElement is returned whenever any method is invoked on an uninitialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")
var ErrTooSmall = errors.New("too small")
var ErrInvalidWriter = errors.New("bson: invalid writer provided")
var ErrInvalidString = errors.New("Invalid string value")
var ErrInvalidBinarySubtype = errors.New("Invalid BSON Binary Subtype")
var ErrInvalidBooleanType = errors.New("Invalid value for BSON Boolean Type")
var ErrStringLargerThanContainer = errors.New("String size is larger than the Code With Scope container")
var ErrInvalidElement = errors.New("Invalid Element")

type ElementTypeError struct {
	Method string
	Type   BSONType
}

func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

type Element struct {
	value *Value
}

func newElement(start uint32, offset uint32) *Element {
	return &Element{&Value{start: start, offset: offset}}
}

// Validates the element and returns its total size.
func (e *Element) Validate() (uint32, error) {
	if e == nil {
		return 0, ErrNilElement
	}
	if e.value == nil {
		return 0, ErrUninitializedElement
	}

	var total uint32 = 1
	n, err := e.validateKey()
	total += n
	if err != nil {
		return total, err
	}
	n, err = e.value.validate(false)
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

func (e *Element) validateKey() (uint32, error) {
	if e.value.data == nil {
		return 0, ErrUninitializedElement
	}

	pos, end := e.value.start+1, e.value.offset
	var total uint32 = 0
	if end > uint32(len(e.value.data)) {
		end = uint32(len(e.value.data))
	}
	for ; pos < end && e.value.data[pos] != '\x00'; pos++ {
		total++
	}
	if pos == end || e.value.data[pos] != '\x00' {
		return total, ErrInvalidKey
	}
	total++
	return total, nil
}

// Key returns the key for this element.
// It panics if e is uninitialized.
func (e *Element) Key() string {
	if e == nil || e.value == nil || e.value.offset == 0 || e.value.data == nil {
		panic(ErrUninitializedElement)
	}
	return string(e.value.data[e.value.start+1 : e.value.offset-1])
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
	size, err := e.Validate()
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

// writeByteSlice handles writing this element to a slice of bytes.
func (e *Element) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	if len(b) < int(size)+int(start) {
		return 0, ErrTooSmall
	}
	var n int
	switch e.value.data[e.value.start] {
	case '\x03', '\x04':
		if e.value.d == nil {
			n = copy(b[start:start+uint(size)], e.value.data[e.value.start:e.value.start+size])
			break
		}

		header := e.value.offset - e.value.start
		n += copy(b[start:start+uint(header)], e.value.data[e.value.start:e.value.offset])
		start += uint(n)
		size -= header
		nn, err := e.value.d.writeByteSlice(start, size, b)
		n += int(nn)
		if err != nil {
			return int64(n), err
		}
	case '\x0F':
		// Get length of code
		codeStart := e.value.offset + 4
		codeLength := readi32(e.value.data[codeStart : codeStart+4])

		if e.value.d != nil {
			lengthWithoutScope := 4 + 4 + codeLength

			scopeLength, err := e.value.d.Validate()
			if err != nil {
				return 0, err
			}

			codeWithScopeLength := lengthWithoutScope + int32(scopeLength)
			_, err = elements.Int32.Encode(uint(e.value.offset), e.value.data, codeWithScopeLength)
			if err != nil {
				return int64(n), err
			}

			typeAndKeyLength := e.value.offset - e.value.start
			n += copy(
				b[start:start+uint(typeAndKeyLength)+uint(lengthWithoutScope)],
				e.value.data[e.value.start:e.value.start+typeAndKeyLength+uint32(lengthWithoutScope)])
			start += uint(n)

			nn, err := e.value.d.writeByteSlice(start, scopeLength, b)
			n += int(nn)
			if err != nil {
				return int64(n), err
			}

			break
		}

		// Get length of scope
		scopeStart := codeStart + 4 + uint32(codeLength)
		scopeLength := readi32(e.value.data[scopeStart : scopeStart+4])

		// Calculate end of entire CodeWithScope value
		codeWithScopeEnd := int32(scopeStart) + scopeLength

		// Set the length of the value
		codeWithScopeLength := codeWithScopeEnd - int32(e.value.offset)
		_, err := elements.Int32.Encode(uint(e.value.offset), e.value.data, codeWithScopeLength)
		if err != nil {
			return 0, err
		}

		fallthrough
	default:
		n = copy(b[start:start+uint(size)], e.value.data[e.value.start:e.value.start+size])
	}
	return int64(n), nil
}

// MarshalBSON implements the Marshaler interface.
func (e *Element) MarshalBSON() ([]byte, error) {
	size, err := e.Validate()
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
