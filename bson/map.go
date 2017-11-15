package bson

import (
	"bytes"
	"sort"

	"github.com/skriptble/wilson/elements"
)

type SlicePool interface {
	GetSlice() []byte
	PutSlice([]byte)
}

// Basic design:
// The map wraps a "core" which handles collecting all the actual data. The map has a prefix
// field that's used for embedded documents. While constructing the map, everything is appended
// to the data slice inside the core. The index inside the core is kept sorted at all times.
// Since '\x00' is not a legal value in a string, it is used to separate out prefixes that indicate
// a subdocument. This also means that the entire document is indexed at all times and searching at
// any depth remains O(lg n) time.
//
// The write method is more complicated though. It has to do a stable sort by the prefix, then write
// the document out to the underlying writer after this has been done. This shouldn't be too difficult
// and shouldn't incur too large a performance overhead when actually serializing. This cost could in theory
// be amortized across the insert calls, but that would add more complexity to the insertion code. This
// might be preferable.
type Map struct {
	data  []*mapelem
	index []mapnode
	// embedded []*Map

	pool     SlicePool
	prefix   string
	embedded bool
	core     *core
}

type core struct {
	err   error
	data  [][]byte
	index []mapnode
}

var EC = &ElementConstructor{}

type ElementConstructor struct {
	pool SlicePool
}

type mapelem struct {
	b []byte
	m *Map
}

type mapnode [3]uint32

// type Iterator struct {
// }

func (c *core) insert(prefix string, e *Element) {
	key := e.Key()
	if len(prefix) > 0 {
		prefix += string('\x00')
	}
	pos := c.search(prefix + key)
	c.index = append(c.index, mapnode{})
	copy(c.index[pos+1:], c.index[pos:])
	c.index[pos] = mapnode{0: e.start, 1: e.value, 2: uint32(len(c.data))}
}

func (m *Map) Insert(e *Element) *Map {
	if m.core.err != nil {
		return m
	}
	m.core.insert(m.prefix, e)
	return m
}

// TODO(skriptble): Make this variadic so we can create deeply nested documents
// more easily.
func (m *Map) InsertDocument(key string) *Map {
	return &Map{prefix: m.prefix + string('\x00') + key, embedded: true, core: m.core}
}

func (m *Map) EmbeddedDocument(key string, embed *Map) *Map {
	// pos := m.search(key)
	// m.index = append(m.index)
	return m
}

func (c *core) search(key string) int {
	i := sort.Search(len(c.index), func(j int) bool {
		return bytes.Compare(c.data[c.index[j][0]][0:c.index[j][1]-1], []byte(key)) >= 0
	})
	return i
}

func (m *Map) Retrieve(key ...string) (*Element, error) {
	// if len(key) == 0 {
	// 	return nil, nil
	// }
	// pos := m.search(key[0])
	// if pos >= len(m.index) {
	// 	return nil, errors.New("key not found")
	// }
	// e := &Element{start: index[pos][0], value: index[pos][1], data: m.data[index[pos][2]]}
	// if len(key) == 1 {
	// 	return e, nil
	// }
	// if e.Type() != '\x02' || e.Type != '\x03' {
	// 	return nil, errors.New("key not found")
	// }
	// // TODO(skriptble): If we have an embedded document or array that is not a Map, we have to
	// // search the Element itself.
	//
	// return
	return nil, nil
}

func (m *Map) Update(e *Element) *Map {
	return m
}

func (m *Map) Delete(key ...string) *Map {
	return m
}

func (m *Map) Err() error {
	return nil
}

func (m *Map) ElementAt(index uint) (*Element, error) {
	return nil, nil
}

func (m *Map) Elements() *Iterator {
	return nil
}

// func (i *Iterator) Next() bool {
// 	return false
// }
//
// func (i *Iterator) Close() error {
// 	return nil
// }
//
// func (i *Iterator) Element() *Element {
// 	return nil
// }
//
// func (i *Iterator) Err() error {
// 	return nil
// }

func (ec *ElementConstructor) Double(key string, f float64) *Element {
	var b []byte
	if ec.pool != nil {
		b = ec.pool.GetSlice()
	}
	if len(b) < 10+len(key) {
		b = make([]byte, 10+len(key))
	}
	elements.Double.Element(0, b, key, f)
	return &Element{start: 0, value: uint32(2 + len(key)), data: b}
}
