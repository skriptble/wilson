package bson

import (
	"io"
	"strconv"

	"github.com/skriptble/wilson/bson/elements"
)

// Array represents an array in BSON. The methods of this type are more
// expensive than those on Document because they require potentially updating
// multiple keys to ensure the array stays valid at all times.
type Array struct {
	doc *Document
}

func NewArray(numberOfElems uint) *Array {
	return &Array{doc: NewDocument(numberOfElems)}
}

func (a *Array) Validate() (uint32, error) {
	var size uint32 = 4 + 1
	for i, elem := range a.doc.elems {
		n, err := elem.value.validate(false)
		if err != nil {
			return 0, err
		}

		// type
		size++
		// key
		size += uint32(len(strconv.Itoa(i))) + 1
		// value
		size += n
	}

	return size, nil
}

func (a *Array) Lookup(index uint) (*Value, error) {
	v, err := a.doc.ElementAt(index)
	if err != nil {
		return nil, err
	}

	return v.value, nil
}

func (a *Array) Append(values ...*Value) *Array {
	a.doc.Append(elemsFromValues(values)...)

	return a
}

func (a *Array) Prepend(values ...*Value) *Array {
	a.doc.Prepend(elemsFromValues(values)...)

	return a
}

func (a *Array) Set(index uint, value *Value) *Array {
	if index >= uint(len(a.doc.elems)) {
		panic(ErrOutOfBounds)
	}

	a.doc.elems = append(append(a.doc.elems[:index], &Element{value}), a.doc.elems[index:]...)

	return a
}

func (a *Array) Combine(doc interface{}) error {
	return nil
}

func (a *Array) Delete(index uint) *Value {
	if index >= uint(len(a.doc.elems)) {
		return nil
	}

	elem := a.doc.elems[index]
	a.doc.elems = append(a.doc.elems[:index], a.doc.elems[index+1:]...)

	return elem.value
}

// WriteTo implements the io.WriterTo interface.
func (a *Array) WriteTo(w io.Writer) (int64, error) {
	b, err := a.MarshalBSON()
	if err != nil {
		return 0, err
	}
	n, err := w.Write(b)
	return int64(n), err
}

// WriteDocument will serialize this document to the provided writer beginning
// at the provided start position.
func (a *Array) WriteArray(start uint, writer []byte) (int64, error) {
	var total int64
	var pos = start

	size, err := a.Validate()
	if err != nil {
		return total, err
	}

	n, err := a.writeByteSlice(pos, size, writer)
	total += n
	pos += uint(n)
	if err != nil {
		return total, err
	}

	return total, nil
}

// writeByteSlice handles serializing this document to a slice of bytes starting
// at the given start position.
func (a *Array) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	var total int64
	var pos = start

	if len(b) < int(start)+int(size) {
		return 0, ErrTooSmall
	}
	n, err := elements.Int32.Encode(start, b, int32(size))
	total += int64(n)
	pos += uint(n)
	if err != nil {
		return total, err
	}

	for i, elem := range a.doc.elems {
		b[pos] = elem.value.data[elem.value.start]
		total += 1
		pos += 1

		key := []byte(strconv.Itoa(i))
		key = append(key, 0)
		copy(b[pos:], key)
		total += int64(len(key))
		pos += uint(len(key))

		n, err := elem.writeElement(false, pos, b)
		total += int64(n)
		pos += uint(n)
		if err != nil {
			return total, err
		}
	}

	n, err = elements.Byte.Encode(pos, b, '\x00')
	total += int64(n)
	pos += uint(n)
	if err != nil {
		return total, err
	}
	return total, nil
}

// MarshalBSON implements the Marshaler interface.
func (a *Array) MarshalBSON() ([]byte, error) {
	size, err := a.Validate()
	if err != nil {
		return nil, err
	}
	b := make([]byte, size)
	_, err = a.writeByteSlice(0, size, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}
