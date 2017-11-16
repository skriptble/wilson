package bson

type RWDocument struct {
	elems []*RWElement
	index []uint32
	err   error
}

func (rwd *RWDocument) Insert(rwe ...*RWElement) *RWDocument {
	return rwd
}

func (rwd *RWDocument) Lookup(key ...string) (*RWElement, error) {
	return nil, nil
}

func (rwd *RWDocument) Delete(key ...string) *RWDocument {
	return rwd
}

func (rwd *RWDocument) Update(rwm RWModifier, key ...string) *RWDocument {
	return rwd
}

func (rwd *RWDocument) Err() error {
	return nil
}

func (rwd *RWDocument) ElementAt(index uint) (*RWElement, error) {
	return nil, nil
}

func (rwd *RWDocument) Iterator() *Iterator {
	return nil
}

// doc must be one of the following:
//
// - *Document
// - []byte
// - io.Reader
func (rwd *RWDocument) Append(doc interface{}) error {
	return nil
}

// UpdateIndex is used to repair the index when the keys of elements have been
// modified. This method should be called whenever a RWModifier is used to
// directly change the key on an element. If the same RWModifier is used via
// the Update method on RWDocument, this method does not need to be called.
func (rwd *RWDocument) UpdateIndex(recursive bool) {
}

func (rwd *RWDocument) Validate(recursive bool) error {
	return nil
}

type Iterator struct {
}

func (itr *Iterator) Next() bool {
	return false
}

func (itr *Iterator) Close() error {
	return nil
}

func (itr *Iterator) Element() *RWElement {
	return nil
}

func (itr *Iterator) Err() error {
	return nil
}
