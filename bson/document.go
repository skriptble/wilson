package bson

import (
	"bytes"
	"encoding/binary"
	"io"
)

// TODO(skriptble): Handle subdocuments better here. The current setup requires
// we allocate an additional slice header when creating a subdocument since there
// is not start element. Increasing the size of this struct by 8 bytes for 2
// uint32's (one for start and one for end/length) would mean we don't have to
// reslice the main document's slice. It also means we could have a longer
// []bytes that contains many BSON documents.
type Document struct {
	data    []byte
	n       []node
	current *Element

	roData io.ReaderAt
	rwData io.ReadWriteSeeker
	index  []uint32
	elems  []*Element
	start  uint64
	err    error
}

func (d *Document) Length() int32 {
	return int32(binary.LittleEndian.Uint32(d.data[d.start : d.start+4]))
}

func (d *Document) ElementList() []*Element {
	return nil
}

// NOTE: this should reuse the same *Element to avoid allocations.
// This comment should should mention that.
func (d *Document) Element(index uint) *Element {
	return nil
}

func (d *Document) Recycle(e *Element) {
}

func (d *Document) Len() int { return len(d.n) }

func (d *Document) Less(i, j int) bool {
	return bytes.Compare(
		d.data[d.n[i][0]+1:d.n[i][1]-1],
		d.data[d.n[j][0]+1:d.n[j][1]-1],
	) < 0
}

func (d *Document) Swap(i, j int) { d.n[i], d.n[j] = d.n[j], d.n[i] }

func (d *Document) Validate(shallow bool) error {
	return nil
}

func (d *Document) Index(shallow bool) error {
	return nil
}

func (d *Document) Parse() error {
	return nil
}
