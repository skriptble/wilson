package bson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"sort"

	"github.com/skriptble/wilson/elements"
)

var ErrInvalidReadOnlyDocument = errors.New("Invalid read-only document")
var ErrInvalidKey = errors.New("invalid document key")
var ErrInvalidLength = errors.New("document length is invalid")
var ErrEmptyKey = errors.New("empty key provided")

type Document struct {
	// TODO(skriptble): I'm not sure how useful this is, but I'm having trouble
	// justifying adding an error to the return value of the Append and Prepend
	// methods since there is only a single failure case, i.e. a nil element
	// provided. If there is a nil element 4 elements into an 8 element array
	// this seems like a severe application error and there doesn't seem to be
	// a general way to handle that. This seems to support a panic if there is
	// a nil. This bool would allow users to explicitly exempt out of that
	// behavior.
	IgnoreNilInsert bool
	elems           []*Element
	index           []uint32
}

func NewDocument() *Document {
	return new(Document)
}

func ReadDocument(b []byte) (*Document, error) {
	var doc = new(Document)
	err := doc.UnmarshalBSON(b)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func (d *Document) Walk(recursive bool, f func(prefix []string, e *Element)) {
}

// TODO(skriptble): Perhaps we should let the user pass in a []string and put
// as many keys as we can inside?
// TODO(skriptble): Should we even bother getting recursive keys? This seems
// like an edge case but maybe not.
func (d *Document) Keys(recursive bool) [][]string {
	return nil
}

// Appends each element to the document, in order. If a nil element is passed
// as a parameter and IgnoreNilInsert is set to false, this method will panic.
func (d *Document) Append(elems ...*Element) *Document {
	for _, elem := range elems {
		if elem == nil {
			if d.IgnoreNilInsert {
				continue
			}
			// TODO(skriptble): Maybe Append and Prepend should return an error
			// instead of panicking here.
			panic(errors.New("nil Element provided"))
		}
		d.elems = append(d.elems, elem)
		i := sort.Search(len(d.index), func(i int) bool {
			return bytes.Compare(
				d.keyFromIndex(i), elem.data[elem.start+1:elem.value]) >= 0
		})
		if i < len(d.index) {
			d.index = append(d.index, 0)
			copy(d.index[i+1:], d.index[i:])
			d.index[i] = uint32(len(d.elems) - 1)
		} else {
			d.index = append(d.index, uint32(len(d.elems)-1))
		}
	}
	return d
}

// Prepends each element to the document, in order. If a nil element is passed
// as a parameter and IgnoreNilInsert is set to false, this method will panic.
func (d *Document) Prepend(elems ...*Element) *Document {
	// In order to insert the prepended elements in order we need to make space
	// at the front of the elements slice.
	d.elems = append(d.elems, elems...)
	copy(d.elems[len(elems):], d.elems)

	remaining := len(elems)
	for idx, elem := range elems {
		if elem == nil {
			if d.IgnoreNilInsert {
				// Having nil elements in a document would be problematic.
				copy(d.elems[idx:], d.elems[idx+1:])
				d.elems[len(d.elems)-1] = nil
				d.elems = d.elems[:len(d.elems)-1]
				continue
			}
			// Not very efficient, but we're about to blow up so ¯\_(ツ)_/¯
			for j := idx; j < remaining; j++ {
				copy(d.elems[j:], d.elems[j+1:])
				d.elems[len(d.elems)-1] = nil
				d.elems = d.elems[:len(d.elems)-1]
			}
			panic(errors.New("nil Element provided"))
		}
		remaining--
		d.elems[idx] = elem
		i := sort.Search(len(d.index), func(i int) bool {
			return bytes.Compare(
				d.keyFromIndex(i), elem.data[elem.start+1:elem.value]) >= 0
		})
		if i < len(d.index) {
			d.index = append(d.index, 0)
			copy(d.index[i+1:], d.index[i:])
			d.index[i] = uint32(len(d.elems) - 1)
		} else {
			d.index = append(d.index, uint32(len(d.elems)-1))
		}
	}
	return d
}

// TODO: Do we really need an error here? It doesn't seem possible to insert a nil element.
func (d *Document) Lookup(key ...string) (*Element, error) {
	if len(key) == 0 {
		return nil, ErrEmptyKey
	}
	i := sort.Search(len(d.index), func(i int) bool { return bytes.Compare(d.keyFromIndex(i), []byte(key[0])) >= 0 })
	if i < len(d.index) && bytes.Compare(d.keyFromIndex(i), []byte(key[0])) == 0 {
		return d.elems[d.index[i]], nil
	}
	return nil, nil
}

// Delete will delete the key from the Document. The deleted element is
// returned. If the key does not exist, then nil is returned and the delete is
// a no-op.
func (d *Document) Delete(key ...string) *Element {
	if len(key) == 0 {
		return nil
	}
	// Do a binary search through the index, delete the element from
	// the index and delete the element from the elems array.
	var elem *Element
	i := sort.Search(len(d.index), func(i int) bool { return bytes.Compare(d.keyFromIndex(i), []byte(key[0])) >= 0 })
	if i < len(d.index) && bytes.Compare(d.keyFromIndex(i), []byte(key[0])) == 0 {
		keyIndex := d.index[i]
		elem = d.elems[keyIndex]
		if len(key) == 1 {
			d.index = append(d.index[:i], d.index[i+1:]...)
			d.elems = append(d.elems[:keyIndex], d.elems[keyIndex+1:]...)
			return elem
		}
		switch elem.Type() {
		case '\x03':
			elem = elem.Document().Delete(key[1:]...)
		case '\x04':
			elem = elem.Array().Document.Delete(key[1:]...)
		default:
			elem = nil
		}
	}
	return elem
}

func (d *Document) RenameKey(newKey string, key ...string) error {
	if len(key) == 0 {
		return ErrEmptyKey
	}
	var err error
	var elem *Element

	i := sort.Search(len(d.index), func(i int) bool { return bytes.Compare(d.keyFromIndex(i), []byte(key[0])) >= 0 })
	if i < len(d.index) && bytes.Compare(d.keyFromIndex(i), []byte(key[0])) == 0 {
		keyIndex := d.index[i]
		elem = d.elems[keyIndex]
		if len(key) == 1 {
			return elem.updateKey(newKey)
		}
		switch elem.Type() {
		case '\x03':
			err = elem.Document().RenameKey(newKey, key[1:]...)
		case '\x04':
			err = elem.Array().Document.RenameKey(newKey, key[1:]...)
		default:
			err = ErrInvalidKey
		}
	}
	return err
}

func (d *Document) ElementAt(index uint) (*Element, error) {
	if int(index) >= len(d.elems) {
		return nil, errors.New("Out of bounds")
	}
	return d.elems[index], nil
}

func (d *Document) Iterator() *Iterator {
	return nil
}

// doc must be one of the following:
//
// - *Document
// - []byte
// - io.Reader
func (d *Document) Combine(doc interface{}) error {
	return nil
}

// Reset clears a document so it can be reused. This method clears references
// to the underlying pointers to elements so they can be garbage collected.
func (d *Document) Reset() {
	for idx := range d.elems {
		d.elems[idx] = nil
	}
	d.elems = d.elems[:0]
	d.index = d.index[:0]
}

// Validates the document and returns its total size.
func (d *Document) Validate() (uint32, error) {
	// Header and Footer
	var size uint32 = 4 + 1
	for _, elem := range d.elems {
		n, err := elem.Validate(true)
		if err != nil {
			return 0, err
		}
		size += n
	}
	return size, nil
}

// WriteTo implements the io.WriterTo interface.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	return d.WriteDocument(0, w)
}

func (d *Document) WriteDocument(start uint, writer interface{}) (int64, error) {
	var total int64
	var pos uint = start
	size, err := d.Validate()
	if err != nil {
		return total, err
	}
	switch w := writer.(type) {
	case []byte:
		n, err := d.writeByteSlice(pos, size, w)
		total += n
		pos += uint(n)
		if err != nil {
			return total, err
		}
	default:
		return 0, ErrInvalidWriter
	}
	return total, nil
}

func (d *Document) MarshalBSON() ([]byte, error) {
	size, err := d.Validate()
	if err != nil {
		return nil, err
	}
	b := make([]byte, size)
	_, err = d.writeByteSlice(0, size, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (d *Document) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	var total int64
	var pos uint = start
	if len(b) < int(start)+int(size) {
		return 0, ErrTooSmall
	}
	n, err := elements.Int32.Encode(start, b, int32(size))
	total += int64(n)
	pos += uint(n)
	if err != nil {
		return total, err
	}
	for _, elem := range d.elems {
		n, err := elem.WriteElement(pos, b)
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

func (d *Document) UnmarshalBSON(b []byte) error {
	// Read byte array
	//   - Create an Element for each element found
	//   - Update the index with the key of the element
	//   TODO: Maybe do 2 pass and alloc the elems and index once?
	// 		   We should benchmark 2 pass vs multiple allocs for growing the slice
	if len(b) < 5 {
		return ErrTooSmall
	}
	givenLength := int32(binary.LittleEndian.Uint32(b[0:4]))
	if len(b) < int(givenLength) {
		return ErrInvalidLength
	}
	var pos uint32 = 4
	var elemStart, elemValStart uint32
	var elem *Element
	for {
		if int(pos) >= len(b) {
			return ErrInvalidReadOnlyDocument
		}
		if b[pos] == '\x00' {
			break
		}
		elemStart = pos
		pos++
		n, err := keyLength(pos, b)
		pos += n
		if err != nil {
			return err
		}
		elem = new(Element)
		elemValStart = pos
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = b
		n, err = elem.validateValue(true)
		pos += n
		if err != nil {
			return err
		}
		d.elems = append(d.elems, elem)
		i := sort.Search(len(d.index), func(i int) bool {
			return bytes.Compare(
				d.keyFromIndex(i), elem.data[elem.start+1:elem.value]) >= 0
		})
		if i < len(d.index) {
			d.index = append(d.index, 0)
			copy(d.index[i+1:], d.index[i:])
			d.index[i] = uint32(len(d.elems) - 1)
		} else {
			d.index = append(d.index, uint32(len(d.elems)-1))
		}
		pos++
	}
	return nil
}

// ReadFrom will read one BSON document from the given io.Reader.
func (d *Document) ReadFrom(r io.Reader) (int64, error) {
	return 0, nil
}

// keyLength attempts to read a c style string starting at pos from the byte
// slice b. This method returns the number of bytes read and an error.
func keyLength(pos uint32, b []byte) (uint32, error) {
	// Read a CString, return the length, including the '\x00'
	var total uint32 = 0
	for ; pos < uint32(len(b)) && b[pos] != '\x00'; pos++ {
		total++
	}
	if pos == uint32(len(b)) || b[pos] != '\x00' {
		return total, ErrInvalidKey
	}
	total++
	return total, nil
}

// keyFromIndex returns the key for the element. The idx parameter is the
// position in the index property, not the elems property. This method is
// mainly used when calling sort.Search.
func (d *Document) keyFromIndex(idx int) []byte {
	haystack := d.elems[d.index[idx]]
	return haystack.data[haystack.start+1 : haystack.value]
}
