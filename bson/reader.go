package bson

import (
	"errors"
	"strings"
)

var validateDone = errors.New("validation loop complete")

// Reader is a wrapper around a byte slice. It will interpret the slice as a
// BSON document. Most of the methods on Reader are low cost and are meant for
// simple operations that are run a few times. Because there is no metadata
// stored all methods run in O(n) time. If a more efficient lookup method is
// necessary then the Document type should be used.
type Reader []byte

// Validates the document. This method only validates the first document in
// the slice, to validate other documents, the slice must be resliced.
func (r Reader) Validate() (size uint32, err error) {
	return r.readElements(func(elem *ReaderElement) error {
		var err error
		switch elem.Type() {
		case '\x03':
			_, err = elem.Document().Validate()
		case '\x04':
			_, err = elem.Array().Validate()
		}
		return err
	})
}

// keySize will ensure the key is valid and return the length of the key
// including the null terminator.
func (r Reader) keySize(pos, end uint32) (uint32, error) {
	// Read a CString, return the length, including the '\x00'
	var total uint32 = 0
	for ; pos < end && r[pos] != '\x00'; pos++ {
		total++
	}
	if pos == end || r[pos] != '\x00' {
		return total, ErrInvalidKey
	}
	total++
	return total, nil
}

// Lookup search the document, potentially recursively, for the given key. If
// there are multiple keys provided, this method will recurse down, as long as
// the top and intermediate nodes are either documents or arrays. If any key
// except for the last is not a document or an array, an error will be returned.
//
// TODO(skriptble): Implement better error messages.
//
// TODO(skriptble): Determine if this should return an error on empty key and
// key not found.
func (r Reader) Lookup(key ...string) (*ReaderElement, error) {
	if len(key) < 1 {
		return nil, nil
	}

	var elem *ReaderElement
	_, err := r.readElements(func(e *ReaderElement) error {
		if key[0] == e.Key() {
			if len(key) > 1 {
				switch e.Type() {
				case '\x03':
					e, err := e.Document().Lookup(key[1:]...)
					if err != nil {
						return err
					}
					elem = e
					return validateDone
				case '\x04':
					e, err := e.Array().Lookup(key[1:]...)
					if err != nil {
						return err
					}
					elem = e
					return validateDone
				default:
					// TODO(skriptble): This error message is pretty awful.
					// Please fix.
					return errors.New("Incorrect type for depth")
				}
			}
			elem = e
			return validateDone
		}
		return nil
	})
	return elem, err
}

// ElementAt searches for a retrieves the element at the given index. This
// method will validate all the elements up to and including the element at
// the given index.
//
// TODO(skriptble): Should this be zero indexed? Prevents the error case and
// aligns with byte slice style access pattern.
func (r Reader) ElementAt(index uint) (*ReaderElement, error) {
	var current uint
	var elem *ReaderElement
	_, err := r.readElements(func(e *ReaderElement) error {
		if current != index {
			current++
			return nil
		}
		elem = e
		return validateDone
	})
	if err != nil {
		return nil, err
	}
	if elem == nil {
		return nil, errors.New("index out of bounds")
	}
	return elem, nil
}

// Iterator returns a ReaderIterator that can be used to iterate through the
// elements of this Reader.
func (r Reader) Iterator() (*ReaderIterator, error) {
	return NewReadIterator(r)
}

// Keys returns the keys for this document. If recursive is true then this
// method will also return the keys for subdocuments and arrays.
//
// The keys will be return in order.
func (r Reader) Keys(recursive bool) (Keys, error) {
	return r.recursiveKeys(recursive)
}

// recursiveKeys implements the logic for the Keys method. This is a separate
// function to facilitate recursive calls.
func (r Reader) recursiveKeys(recursive bool, prefix ...string) (Keys, error) {
	ks := make(Keys, 0)
	_, err := r.readElements(func(elem *ReaderElement) error {
		key := elem.Key()
		ks = append(ks, Key{Prefix: prefix, Name: key})
		if recursive {
			switch elem.Type() {
			case '\x03':
				recursivePrefix := append(prefix, key)
				recurKeys, err := elem.Document().recursiveKeys(recursive, recursivePrefix...)
				if err != nil {
					return err
				}
				ks = append(ks, recurKeys...)
			case '\x04':
				recursivePrefix := append(prefix, key)
				recurKeys, err := elem.Array().recursiveKeys(recursive, recursivePrefix...)
				if err != nil {
					return err
				}
				ks = append(ks, recurKeys...)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return ks, nil
}

// readElements is an internal method used to traverse the document. It will
// validate the document and the underlying elements. If the provided function
// is non-nil it will be called for each element. If `validateDone` is returned
// from the function, this method will return. This method will return nil when
// the function returns validateDone, in all other cases a non-nil error will
// be returned by this method.
func (r Reader) readElements(f func(e *ReaderElement) error) (uint32, error) {
	if len(r) < 5 {
		return 0, ErrTooSmall
	}
	// TODO(skriptble): We could support multiple documents in the same byte
	// slice without reslicing if we have pos as a parameter and use that to
	// get the length of the document.
	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return 0, ErrInvalidLength
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var elem = new(ReaderElement)
	end := uint32(givenLength)
	for {
		if pos >= end {
			// We've gone off the end of the buffer and we're missing
			// a null terminator.
			return pos, ErrInvalidReadOnlyDocument
		}
		if r[pos] == '\x00' {
			break
		}
		elemStart = pos
		pos++
		n, err := r.keySize(pos, end)
		pos += n
		if err != nil {
			return pos, err
		}
		elemValStart = pos
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = r
		n, err = elem.validateValue(false)
		pos += n
		if err != nil {
			return pos, err
		}
		if f != nil {
			err = f(elem)
			if err != nil {
				if err == validateDone {
					break
				}
				return pos, err
			}
		}
	}

	// The size is always 1 larger than the position, since position is 0
	// indexed.
	return pos + 1, nil

}

// Keys represents the keys of a BSON document.
type Keys []Key

// Key represents an individual key of a BSON document. The Prefix property is
// used to represent the depth of this key.
type Key struct {
	Prefix []string
	Name   string
}

// String implements the fmt.Stringer interface.
func (k Key) String() string {
	str := strings.Join(k.Prefix, ".")
	if str != "" {
		return str + "." + k.Name
	}
	return k.Name
}

// readi32 is a helper function for reading an int32 from slice of bytes.
func readi32(b []byte) int32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}
