package bson

import (
	"github.com/skriptble/wilson/elements"
	"github.com/skriptble/wilson/parser/ast"
)

var C Constructor
var AC ArrayConstructor

type Constructor struct{}
type ArrayConstructor struct{}

func (Constructor) Double(key string, f float64) *Element {
	b := make([]byte, 1+len(key)+1+8)
	elem := new(Element)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	_, err := elements.Double.Element(0, b, key, f)
	if err != nil {
		panic(err)
	}
	elem.data = b
	return elem
}

func (Constructor) String(key string, val string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(val) + 1)
	b := make([]byte, size)
	elem := new(Element)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	_, err := elements.String.Element(0, b, key, val)
	if err != nil {
		panic(err)
	}
	elem.data = b
	return elem
}

func (Constructor) SubDocument(key string, d *Document) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := new(Element)
	elem.start = 0
	elem.value = size
	_, err := elements.Byte.Encode(0, b, '\x03')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.data = b
	elem.d = d
	return elem
}
func (c Constructor) SubDocumentFromElements(key string, elems ...*Element) *Element {
	return c.SubDocument(key, NewDocument().Append(elems...))
}
func (Constructor) Array(key string, elems ...*Element) *Element                  { return nil }
func (Constructor) Binary(key string, b []byte, btype uint) *Element              { return nil }
func (Constructor) ObjectID(key string, obj [12]byte) *Element                    { return nil }
func (Constructor) Boolean(key string, b bool) *Element                           { return nil }
func (Constructor) DateTime(key string, dt int64) *Element                        { return nil }
func (Constructor) Regex(key string, pattern, options string) *Element            { return nil }
func (Constructor) DBPointer(key string, dbpointer [12]byte) *Element             { return nil }
func (Constructor) Javascript(key string, js string) *Element                     { return nil }
func (Constructor) Symbol(key string, symbol string) *Element                     { return nil }
func (Constructor) CodeWithScope(key string, js string, scope *Document) *Element { return nil }
func (Constructor) Int32(key string, i int32) *Element                            { return nil }
func (Constructor) Uint64(key string, u uint64) *Element                          { return nil }
func (Constructor) Int64(key string, i int64) *Element                            { return nil }
func (Constructor) Decimal128(key string, d ast.Decimal128) *Element              { return nil }

func (ArrayConstructor) Double(f float64) *Element                         { return nil }
func (ArrayConstructor) String(val string) *Element                        { return nil }
func (ArrayConstructor) Document(elems ...*Element) *Element               { return nil }
func (ArrayConstructor) Array(elemens ...*Element) *Element                { return nil }
func (ArrayConstructor) Binary(b []byte, btype uint) *Element              { return nil }
func (ArrayConstructor) ObjectID(obj [12]byte) *Element                    { return nil }
func (ArrayConstructor) Boolean(b bool) *Element                           { return nil }
func (ArrayConstructor) DateTime(dt int64) *Element                        { return nil }
func (ArrayConstructor) Regex(pattern, options string) *Element            { return nil }
func (ArrayConstructor) DBPointer(dbpointer [12]byte) *Element             { return nil }
func (ArrayConstructor) Javascript(js string) *Element                     { return nil }
func (ArrayConstructor) Symbol(symbol string) *Element                     { return nil }
func (ArrayConstructor) CodeWithScope(js string, scope *Document) *Element { return nil }
func (ArrayConstructor) Int32(i int32) *Element                            { return nil }
func (ArrayConstructor) Uint64(u uint64) *Element                          { return nil }
func (ArrayConstructor) Int64(i int64) *Element                            { return nil }
func (ArrayConstructor) Decimal128(d ast.Decimal128) *Element              { return nil }
