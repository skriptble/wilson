package bson

import "time"

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
func (rwd *RWDocument) UpdateIndex() {
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

type RWModifier func(rw *RWElement)

// RWElement is a hybrid structure that can represent any BSON element. For
// types that are not a document nor array, only the Element field will be
// populated and RWDocument will be nil. For a document or array, both
// properties will be non-nil. The Element property will contain the key for
// the document or array, and the RWDocument will hold the value.
//
// Just like the Element type, RWElement expects the user to check the type
// before calling other methods and will therefore panic if the type is not the
// type specified by the method's documentation.
type RWElement struct {
	e *Element
	d *RWDocument
}

func (rwe *RWElement) Modify(rwm RWModifier) {
}

func (rwe *RWElement) Key() string                         { return rwe.e.Key() }
func (rwe *RWElement) Type() byte                          { return rwe.e.Type() }
func (rwe *RWElement) Double() float64                     { return rwe.e.Double() }
func (rwe *RWElement) String() string                      { return rwe.e.String() }
func (rwe *RWElement) Binary() *Binary                     { return rwe.e.Binary() }
func (rwe *RWElement) ObjectID() [12]byte                  { return rwe.e.ObjectID() }
func (rwe *RWElement) Boolean() bool                       { return rwe.e.Boolean() }
func (rwe *RWElement) DateTime() time.Time                 { return rwe.e.DateTime() }
func (rwe *RWElement) Regex() (pattern, options string)    { return rwe.e.Regex() }
func (rwe *RWElement) DBPointer() [12]byte                 { return rwe.e.DBPointer() }
func (rwe *RWElement) Javascript() string                  { return rwe.e.Javascript() }
func (rwe *RWElement) Symbol() string                      { return rwe.e.Symbol() }
func (rwe *RWElement) JavascriptWithScope() *CodeWithScope { return rwe.e.JavascriptWithScope() }
func (rwe *RWElement) Int32() int32                        { return rwe.e.Int32() }

func (rwe *RWElement) Insert(elems ...*RWElement) *RWElement {
	rwe.d.Insert(elems...)
	return rwe
}

func (rwe *RWElement) Lookup(key ...string) (*RWElement, error) {
	return rwe.d.Lookup(key...)
}

func (rwe *RWElement) Delete(key ...string) *RWElement {
	rwe.d.Delete(key...)
	return rwe
}

func (rwe *RWElement) Update(rwm RWModifier, key ...string) *RWElement {
	rwe.d.Update(rwm, key...)
	return rwe
}

func (rwe *RWElement) Err() error {
	return rwe.d.Err()
}

func (rwe *RWElement) ElementAt(index uint) (*RWElement, error) {
	return rwe.d.ElementAt(index)
}

func (rwe *RWElement) Iterator() *Iterator {
	return rwe.d.Iterator()
}

// doc must be one of the following:
//
// - *Document
// - []byte
// - io.Reader
func (rwe *RWElement) Append(doc interface{}) error {
	return rwe.d.Append(doc)
}

// UpdateIndex is used to repair the index when the keys of elements have been
// modified. This method should be called whenever a RWModifier is used to
// directly change the key on an element. If the same RWModifier is used via
// the Update method on RWElement, this method does not need to be called.
func (rwe *RWElement) UpdateIndex() {
	rwe.d.UpdateIndex()
}

func (rwe *RWElement) Validate(recursive bool) error {
	return rwe.d.Validate(recursive)
}
