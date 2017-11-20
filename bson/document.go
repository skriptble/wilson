package bson

import (
	"encoding/binary"
	"errors"
)

// TODO(skriptble): Handle subdocuments better here. The current setup requires
// we allocate an additional slice header when creating a subdocument since there
// is not start element. Increasing the size of this struct by 8 bytes for 2
// uint32's (one for start and one for end/length) would mean we don't have to
// reslice the main document's slice. It also means we could have a longer
// []bytes that contains many BSON documents.
type Document struct {
	// One of these should be nil
	data  []byte
	elems []*Element

	index  []uint32
	start  uint32
	err    error
	parent *Document
}

// enableWrite handles the necessary set up to ensure the document is in write
// mode. This mode requires all elements to be indexed and have entries in the
// elems slice. This is required because once we insert or update an entry the
// bytes are moved from the read-only data to the read-write data. In order to
// reconstruct the document from this point, we need to use the elems slice.
func (d *Document) enableWriting() {
	// The last step of enabling writing should nil the data property. The
	// underlying elements can still hold references to it though.
	if d.elems != nil || d.data == nil {
		return
	}
	// Read through through the data slice and turn it into elements
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
		var elem = new(Element)
		end := d.start + uint32(givenLength)
		for {
			elemStart = pos
			pos++
			n, err := d.validateKey(pos, end)
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
		}
	}
	return 0, nil
}

func (d *Document) Walk(recursive bool, f func(prefix []string, e *Element)) {
}

// TODO(skriptble): Perhaps we should let the user pass in a []string and put
// as many keys as we can inside?
func (d *Document) Keys() []string {
	return nil
}

func (d *Document) validateKey(pos, end uint32) (uint32, error) {
	// Read a CString, return the length, including the '\x00'
	return 0, nil
}

func (d *Document) Index(recursive bool) error {
	// This doesn't need to call the Validate method directly, it can instead
	// build the index and validate as it goes.
	return nil
}

func (d *Document) Append(elems ...*Element) *Document {
	// If we aren't in writing mode, enable it.
	if d.elems == nil {
		d.enableWriting()
	}
	return d
}

func (d *Document) Prepend(elems ...*Element) *Document {
	// If we aren't in writing mode, enable it.
	if d.elems == nil {
		d.enableWriting()
	}
	return d
}

func (d *Document) Lookup(key ...string) (*Element, error) {
	// Do a binary search through the index
	return nil, nil
}

func (d *Document) Delete(key ...string) *Document {
	if d.elems == nil {
		d.enableWriting()
	}
	// Do a binary search through the index, delete the elemnt from
	// the index and delete the element from the elems array.
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
	return d
}

func (d *Document) Err() error {
	return nil
}

func (d *Document) ElementAt(index uint) (*Element, error) {
	// If writing mode is enabled, find the element in the elems slice,
	// if it's not search through the byte slice.
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
