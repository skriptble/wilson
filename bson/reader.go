package bson

import (
	"errors"
)

// Reader is a wrapper around a byte slice. It will interpret the slice as a
// BSON document. Most of the methods on Reader are low cost and are meant for
// simple operations that are run a few times. Because there is no metadata
// stored all methods run in O(n) time. If a more efficient lookup method is
// necessary then the Document type should be used.
type Reader []byte

// Validates the document. This method only validates the first document in
// the slice, to validate other documents, the slice must be resliced
func (r Reader) Validate() (uint32, error) {
	// validate the length, setup a length tracker
	// read each element
	// 		- Validate the key
	// 		- Validate the value
	if len(r) < 5 {
		return 0, errors.New("Too short")
	}
	// TODO(skriptble): We could support multiple documents in the same byte
	// slice without reslicing if we have pos as a parameter and use that to
	// get the length of the document.
	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return 0, errors.New("Incorrect length in document header")
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var elem = new(ReaderElement)
	end := uint32(givenLength)
	for {
		if r[pos] == '\x00' {
			pos++
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
		n, err = elem.Validate()
		pos += n
		if err != nil {
			return pos, err
		}
		pos++
	}
	return pos, nil
}

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

	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return nil, errors.New("Incorrect length in document header")
	}
	if r[givenLength-1] != '\x00' {
		return nil, errors.New("Incorrect document termination")
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var elem = new(ReaderElement)
	end := uint32(givenLength)
	for {
		// TODO(skriptble): Handle the out of bounds error better.
		if pos >= end || r[pos] == '\x00' {
			pos++
			break
		}
		elemStart = pos
		pos++
		n, err := r.keySize(pos, end)
		pos += n
		if err != nil {
			return nil, err
		}
		elemValStart = pos
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = r
		n, err = elem.Validate()
		pos += n
		if err != nil {
			return nil, err
		}
		if key[0] == elem.Key() {
			if len(key) > 1 {
				switch elem.Type() {
				case '\x03':
					return elem.Document().Lookup(key[1:]...)
				case '\x04':
					return elem.Array().Lookup(key[1:]...)
				default:
					// TODO(skriptble): This error message is pretty awful.
					// Please fix.
					return nil, errors.New("Incorrect type for depth")
				}
			}
			return elem, nil
		}
	}
	return nil, nil
}

// ElementAt searches for a retrieves the element at the given index. This
// method will validate all the elements up to and including the element at
// the given index.
func (r Reader) ElementAt(index uint) (*ReaderElement, error) {
	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return nil, errors.New("Incorrect length in document header")
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var current uint
	var elem = new(ReaderElement)
	end := uint32(givenLength)
	for {
		// TODO(skriptble): Handle the out of bounds error better.
		if pos >= end || r[pos] == '\x00' {
			break
		}
		elemStart = pos
		pos++
		n, err := r.keySize(pos, end)
		pos += n
		if err != nil {
			return nil, err
		}
		elemValStart = pos
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = r
		n, err = elem.Validate()
		pos += n
		if err != nil {
			return nil, err
		}
		if current != index {
			current++
			continue
		}
		return elem, nil
	}
	return nil, errors.New("index out of bounds")
}

// Keys returns all of the keys for this document. If recursive is true then
// this method will also return all of the keys for subdocuments and arrays.
//
// The keys will be return in order.
func (r Reader) Keys(recursive bool) (Keys, error) {
	return r.recursiveKeys(recursive)
}

func (r Reader) recursiveKeys(recursive bool, prefix ...string) (Keys, error) {
	givenLength := readi32(r[0:4])
	if len(r) < int(givenLength) {
		return nil, errors.New("Incorrect length in document header")
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var elem = new(ReaderElement)
	end := uint32(givenLength)
	ks := make(Keys, 0)
	for {
		// TODO(skriptble): Handle the out of bounds error better.
		if pos >= end || r[pos] == '\x00' {
			pos++
			break
		}
		elemStart = pos
		pos++
		n, err := r.keySize(pos, end)
		pos += n
		if err != nil {
			return nil, err
		}
		elemValStart = pos
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = r
		n, err = elem.Validate()
		pos += n
		if err != nil {
			return nil, err
		}
		key := elem.Key()
		ks = append(ks, Key{Prefix: prefix, Name: key})
		if recursive {
			prefix = append(prefix, key)
			switch elem.Type() {
			case '\x03':
				recurKeys, err := elem.Document().recursiveKeys(recursive, prefix...)
				if err != nil {
					return nil, err
				}
				ks = append(ks, recurKeys...)
			case '\x04':
				recurKeys, err := elem.Array().recursiveKeys(recursive, prefix...)
				if err != nil {
					return nil, err
				}
				ks = append(ks, recurKeys...)
			}
		}
	}
	return ks, nil
}

// Keys represents the keys of a BSON document.
type Keys []Key

// Key represents an individual key of a BSON document. The Prefix property is
// used to represent the depth of this key.
type Key struct {
	Prefix []string
	Name   string
}

func readi32(b []byte) int32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}
