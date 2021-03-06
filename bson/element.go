package bson

import (
	"errors"
	"io"

	"github.com/skriptble/wilson/bson/elements"
)

const validateMaxDepthDefault = 2048

// ErrUninitializedElement is returned whenever any method is invoked on an uninitialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")

// ErrTooSmall indicates that a slice provided to write into is not large enough to fit the data.
var ErrTooSmall = errors.New("too small")

// ErrInvalidWriter indicates that a type that can't be written into was passed to a writer method.
var ErrInvalidWriter = errors.New("bson: invalid writer provided")

// ErrInvalidString indicates that a BSON string value had an incorrect length.
var ErrInvalidString = errors.New("invalid string value")

// ErrInvalidBinarySubtype indicates that a BSON binary value had an undefined subtype.
var ErrInvalidBinarySubtype = errors.New("invalid BSON binary Subtype")

// ErrInvalidBooleanType indicates that a BSON boolean value had an incorrect byte.
var ErrInvalidBooleanType = errors.New("invalid value for BSON Boolean Type")

// ErrStringLargerThanContainer indicates that the code portion of a BSON JavaScript code with scope
// value is larger than the specified length of the entire value.
var ErrStringLargerThanContainer = errors.New("string size is larger than the JavaScript code with scope container")

// ErrInvalidElement indicates that a bson.Element had invalid underlying BSON.
var ErrInvalidElement = errors.New("invalid Element")

// ElementTypeError specifies that a method to obtain a BSON value an incorrect type was called on a bson.Value.
type ElementTypeError struct {
	Method string
	Type   Type
}

// Error implements the error interface.
func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

// Element represents a BSON element, i.e. key-value pair of a BSON document.
type Element struct {
	value *Value
}

func newElement(start uint32, offset uint32) *Element {
	return &Element{&Value{start: start, offset: offset}}
}

// Clone creates a shallow copy of the element/
func (e *Element) Clone() *Element {
	return &Element{
		value: &Value{
			start:  e.value.start,
			offset: e.value.offset,
			data:   e.value.data,
			d:      e.value.d,
		},
	}
}

// Value returns the value associated with the BSON element.
func (e *Element) Value() *Value {
	return e.value
}

// Validate validates the element and returns its total size.
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
	var total uint32
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
	return e.writeElement(true, start, writer)
}

func (e *Element) writeElement(key bool, start uint, writer interface{}) (int64, error) {
	// TODO(skriptble): Figure out if we want to use uint or uint32 and
	// standardize across all packages.
	var total int64
	size, err := e.Validate()
	if err != nil {
		return 0, err
	}
	switch w := writer.(type) {
	case []byte:
		n, err := e.writeByteSlice(key, start, size, w)
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
func (e *Element) writeByteSlice(key bool, start uint, size uint32, b []byte) (int64, error) {
	var startToWrite uint
	needed := start + uint(size)

	if key {
		startToWrite = uint(e.value.start)
	} else {
		startToWrite = uint(e.value.offset)

		// Fewer bytes are needed if the key isn't being written.
		needed -= uint(e.value.offset) - uint(e.value.start) - 1
	}

	if uint(len(b)) < needed {
		return 0, ErrTooSmall
	}

	var n int
	switch e.value.data[e.value.start] {
	case '\x03':
		if e.value.d == nil {
			n = copy(b[start:], e.value.data[startToWrite:e.value.start+size])
			break
		}

		header := e.value.offset - e.value.start
		size -= header
		if key {
			n += copy(b[start:], e.value.data[startToWrite:e.value.offset])
			start += uint(n)
		}

		nn, err := e.value.d.writeByteSlice(start, size, b)
		n += int(nn)
		if err != nil {
			return int64(n), err
		}
	case '\x04':
		if e.value.d == nil {
			n = copy(b[start:], e.value.data[startToWrite:e.value.start+size])
			break
		}

		header := e.value.offset - e.value.start
		size -= header
		if key {
			n += copy(b[start:], e.value.data[startToWrite:e.value.offset])
			start += uint(n)
		}

		arr := &Array{doc: e.value.d}

		nn, err := arr.writeByteSlice(start, size, b)
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
		n = copy(b[start:], e.value.data[startToWrite:e.value.start+size])
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
	_, err = e.writeByteSlice(true, 0, size, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func elemsFromValues(values []*Value) []*Element {
	elems := make([]*Element, len(values))

	for i, v := range values {
		if v == nil {
			elems[i] = nil
		} else {
			elems[i] = &Element{v}
		}
	}

	return elems
}
