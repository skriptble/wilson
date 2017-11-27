package bson

import (
	"bytes"
	"encoding/binary"
	"errors"
	"sort"
)

var ErrInvalidReadOnlyDocument = errors.New("Invalid read-only document")
var ErrInvalidKey = errors.New("invalid document key")

// TODO(skriptble): Handle subdocuments better here. The current setup requires
// we allocate an additional slice header when creating a subdocument since there
// is not start element. Increasing the size of this struct by 8 bytes for 2
// uint32's (one for start and one for end/length) would mean we don't have to
// reslice the main document's slice. It also means we could have a longer
// []bytes that contains many BSON documents.
type Document struct {
	// One of these should be nil
	data  []byte
	elems []*ReaderElement

	index  []node
	start  uint32
	err    error
	parent *Document
}

func NewDocument(ro []byte) *Document {
	return &Document{data: ro}
}

// enableWrite handles the necessary set up to ensure the document is in write
// mode. This mode requires all elements to be indexed and have entries in the
// elems slice. This is required because once we insert or update an entry the
// bytes are moved from the read-only data to the read-write data. In order to
// reconstruct the document from this point, we need to use the elems slice.
func (d *Document) enableWriting() error {
	if d.parent != nil {
		err := d.parent.enableWriting()
		if err != nil {
			return err
		}
	}
	// The last step of enabling writing should nil the data property. The
	// underlying elements can still hold references to it though.
	if d.data == nil {
		return nil
	}
	// Read through through the data slice and turn it into elements
	// Do a first pass and count the number of elements
	// This method also validates the structure of the document so this is the
	// last place we need to check for errors.
	elemCount, err := d.countElements()
	if err != nil {
		return err
	}
	// Allocate the elems slice with a capacity of the number of elements
	elems := make([]*ReaderElement, 0, elemCount)
	// Do a second pass and populate the element pointers
	// NOTE: We don't check the length here because countElements has already
	// validated the length is correct.
	givenLength := int32(binary.LittleEndian.Uint32(d.data[d.start : d.start+4]))

	var pos uint32 = d.start + 4
	var elemStart, elemValStart uint32
	var elem = new(ReaderElement)
	end := d.start + uint32(givenLength)
	for {
		elemStart = pos
		if d.data[pos] == '\x00' {
			break
		}
		pos++
		// NOTE: We throw away this error because we cannot reach this point
		// without having a valid document
		n, _ := d.keySize(pos, end)
		pos += n
		elemValStart = pos
		elem = new(ReaderElement)
		elem.start = elemStart
		elem.value = elemValStart
		elem.data = d.data
		// NOTE: We throw away this error because we cannot reach this point
		// without having a valid document
		n, _ = elem.valueSize()
		pos += n
		elems = append(elems, elem)
		pos++
	}
	d.elems = elems
	d.data = nil
	return nil
}

func (d *Document) countElements() (uint, error) {
	var elemCount uint
	if d.data != nil {
		if len(d.data) < 5 {
			return 0, ErrInvalidReadOnlyDocument
		}
		givenLength := int32(binary.LittleEndian.Uint32(d.data[d.start : d.start+4]))
		if len(d.data) < int(givenLength) {
			// TODO(skriptble): More descriptive error.
			return 0, ErrInvalidReadOnlyDocument
		}
		var pos uint32 = d.start + 4
		var elemStart, elemValStart uint32
		var elem = new(ReaderElement)
		end := d.start + uint32(givenLength)
		for {
			elemStart = pos
			if d.data[pos] == '\x00' {
				break
			}
			pos++
			n, err := d.keySize(pos, end)
			pos += n
			if err != nil {
				return 0, err
			}
			elemValStart = pos
			elem.start = elemStart
			elem.value = elemValStart
			elem.data = d.data
			n, err = elem.valueSize()
			pos += n
			if err != nil {
				return 0, err
			}
			elemCount++
			pos++
		}
	}
	return elemCount, nil
}

func (d *Document) Validate(recursive bool) (uint32, error) {
	// validate the length, setup a length tracker
	// read each element
	// 		- Validate the key
	// 		- Validate the value
	if d.elems == nil {
		if len(d.data) < 5 {
			return 0, errors.New("Too short")
		}
		givenLength := int32(binary.LittleEndian.Uint32(d.data[d.start : d.start+4]))
		if len(d.data) < int(givenLength) {
			return 0, errors.New("Incorrect length in document header")
		}
		var pos uint32 = d.start
		var elemStart, elemValStart uint32
		var elem = new(ReaderElement)
		end := d.start + uint32(givenLength)
		for {
			if d.data[pos] == '\x00' {
				break
			}
			elemStart = pos
			pos++
			n, err := d.keySize(pos, end)
			pos += n
			if err != nil {
				return pos, err
			}
			elemValStart = pos
			elem.start = elemStart
			elem.value = elemValStart
			elem.data = d.data
			n, err = elem.Validate(recursive)
			pos += n
			if err != nil {
				return pos, err
			}
			pos++
		}
	}
	return 0, nil
}

func (d *Document) Walk(recursive bool, f func(prefix []string, e *ReaderElement)) {
}

// TODO(skriptble): Perhaps we should let the user pass in a []string and put
// as many keys as we can inside?
func (d *Document) Keys(recursive bool) [][]string {
	return nil
}

// keySize attempts to read a c style string starting at pos. The end
// parameter is the end of the document. This method returns the number of
// bytes read and an error.
func (d *Document) keySize(pos, end uint32) (uint32, error) {
	// Read a CString, return the length, including the '\x00'
	var total uint32 = 0
	for ; pos < end && d.data[pos] != '\x00'; pos++ {
		total++
	}
	if pos == end || d.data[pos] != '\x00' {
		return total, ErrInvalidKey
	}
	return 0, nil
}

func (d *Document) Index(recursive bool) error {
	// This doesn't need to call the Validate method directly, it can instead
	// build the index and validate as it goes.
	if d.data != nil {
		if len(d.index) > 0 {
			d.index = make([]node, 0)
		}
		if len(d.data) < 5 {
			return errors.New("Invalid read-only document")
		}
		givenLength := int32(binary.LittleEndian.Uint32(d.data[d.start : d.start+4]))
		if len(d.data) < int(givenLength) {
			return errors.New("Invalid read-only document")
		}
		var pos uint32 = d.start
		var elemStart, elemValStart uint32
		var elem ReaderElement
		end := d.start + uint32(givenLength)
		for {
			elemStart = pos
			if d.data[pos] == '\x00' {
				break
			}
			pos++
			n, err := d.keySize(pos, end)
			if err != nil {
				return err
			}
			i := sort.Search(len(d.index), func(i int) bool { return bytes.Compare(d.data[d.index[i][0]:d.index[i][1]], d.data[pos:pos+n]) >= 0 })
			if i < len(d.index) {
				d.index = append(d.index, node{})
				copy(d.index[i+1:], d.index[i:])
				d.index[i] = node{pos, pos + n}
			} else {
				d.index = append(d.index, node{pos, pos + n})
			}
			pos += n
			elemValStart = pos
			elem.start = elemStart
			elem.value = elemValStart
			elem.data = d.data
			n, err = elem.valueSize()
			pos += n
			if err != nil {
				return err
			}
			pos++
		}
	}
	return nil
}

func (d *Document) Append(elems ...*ReaderElement) *Document {
	// If we aren't in writing mode, enable it.
	if d.elems == nil {
		d.enableWriting()
	}
	d.elems = append(d.elems, elems...)
	return d
}

func (d *Document) Prepend(elems ...*ReaderElement) *Document {
	// If we aren't in writing mode, enable it.
	if d.elems == nil {
		d.enableWriting()
	}
	d.elems = append(elems, d.elems...)
	return d
}

func (d *Document) Lookup(key ...string) (*ReaderElement, error) {
	if len(key) == 0 {
		return nil, nil
	}
	// Do a binary search through the index
	switch {
	case d.data != nil && d.index != nil:
		i := sort.Search(len(d.index), func(i int) bool { return bytes.Compare(d.data[d.index[i][0]:d.index[i][1]], []byte(key[0])) >= 0 })
		if i < len(d.index) && bytes.Compare(d.data[d.index[i][0]:d.index[i][1]], []byte(key[0])) == 0 {
			return &ReaderElement{start: d.index[i][0], value: d.index[i][1], data: d.data}, nil
		}
		return nil, nil
	case d.data != nil && d.index == nil:
		return nil, nil
	case d.elems != nil && d.index != nil:
		return nil, nil
	case d.elems != nil && d.index == nil:
		return nil, nil
	default:
		// d.elems == nil && d.data == nil
		return nil, nil
	}
}

func (d *Document) Delete(key ...string) *Document {
	if d.elems == nil {
		d.enableWriting()
	}
	// Do a binary search through the index, delete the elemnt from
	// the index and delete the element from the elems array.
	if d.index != nil {
		return d
	}
	return d
}

func (d *Document) Update(m Modifier, key ...string) *Document {
	if d.elems == nil {
		d.enableWriting()
	}
	// When modifying an Element, we need to check if the start is 0. If it is
	// then the Element can be modified in place. If it isn't, that means it's
	// part of a read-only Document's data byte slice and a new slice needs to
	// be allocated.
	if d.index != nil {
		return d
	}
	return d
}

func (d *Document) Err() error {
	return d.err
}

func (d *Document) ElementAt(index uint) (*ReaderElement, error) {
	// If writing mode is enabled, find the element in the elems slice,
	// if it's not search through the byte slice.
	if d.data != nil {
		return nil, nil
	}
	return nil, nil
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
