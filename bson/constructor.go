package bson

import (
	"github.com/skriptble/wilson/bson/decimal"
	"github.com/skriptble/wilson/bson/elements"
	"github.com/skriptble/wilson/bson/objectid"
)

var C Constructor
var AC ArrayConstructor

type Constructor struct{}
type ArrayConstructor struct{}

func (Constructor) Double(key string, f float64) *Element {
	b := make([]byte, 1+len(key)+1+8)
	elem := newElement(0, 1+uint32(len(key))+1)
	_, err := elements.Double.Element(0, b, key, f)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	return elem
}

func (Constructor) String(key string, val string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(val) + 1)
	b := make([]byte, size)
	elem := newElement(0, 1+uint32(len(key))+1)
	_, err := elements.String.Element(0, b, key, val)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	return elem
}

func (Constructor) SubDocument(key string, d *Document) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := newElement(0, size)
	_, err := elements.Byte.Encode(0, b, '\x03')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	elem.value.d = d
	return elem
}
func (Constructor) SubDocumentFromReader(key string, r Reader) *Element {
	size := uint32(1 + len(key) + 1 + len(r))
	b := make([]byte, size)
	elem := newElement(0, uint32(1+len(key)+1))
	_, err := elements.Byte.Encode(0, b, '\x03')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	// NOTE: We don't validate the Reader here since we don't validate the
	// Document when provided to SubDocument.
	copy(b[1+len(key)+1:], r)
	elem.value.data = b
	return elem
}
func (c Constructor) SubDocumentFromElements(key string, elems ...*Element) *Element {
	return c.SubDocument(key, NewDocument(elems...))
}
func (Constructor) Array(key string, a *Array) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := newElement(0, size)
	_, err := elements.Byte.Encode(0, b, '\x04')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	elem.value.d = a.doc
	return elem
}
func (c Constructor) ArrayFromElements(key string, values ...*Value) *Element {
	return c.Array(key, NewArray(values...))
}

func (c Constructor) Binary(key string, b []byte) *Element {
	return c.BinaryWithSubtype(key, b, 0)
}

func (Constructor) BinaryWithSubtype(key string, b []byte, btype byte) *Element {
	size := uint32(1 + len(key) + 1 + 4 + 1 + len(b))
	if btype == 2 {
		size += 4
	}

	buf := make([]byte, size)
	elem := newElement(0, 1+uint32(len(key))+1)
	_, err := elements.Binary.Element(0, buf, key, b, btype)
	if err != nil {
		panic(err)
	}

	elem.value.data = buf
	return elem
}

func (Constructor) Undefined(key string) *Element {
	size := 1 + uint32(len(key)) + 1
	b := make([]byte, size)
	elem := newElement(0, size)
	_, err := elements.Byte.Encode(0, b, '\x06')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	return elem
}

func (Constructor) ObjectID(key string, oid objectid.ObjectID) *Element {
	size := uint32(1 + len(key) + 1 + 12)
	elem := newElement(0, 1+uint32(len(key))+1)
	elem.value.data = make([]byte, size)

	_, err := elements.ObjectId.Element(0, elem.value.data, key, oid)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Boolean(key string, b bool) *Element {
	size := uint32(1 + len(key) + 1 + 1)
	elem := newElement(0, 1+uint32(len(key))+1)
	elem.value.data = make([]byte, size)

	_, err := elements.Boolean.Element(0, elem.value.data, key, b)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) DateTime(key string, dt int64) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := newElement(0, 1+uint32(len(key))+1)
	elem.value.data = make([]byte, size)

	_, err := elements.DateTime.Element(0, elem.value.data, key, dt)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Null(key string) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := newElement(0, uint32(1+len(key)+1))
	_, err := elements.Byte.Encode(0, b, '\x0A')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.value.data = b
	return elem
}
func (Constructor) Regex(key string, pattern, options string) *Element {
	size := uint32(1 + len(key) + 1 + len(pattern) + 1 + len(options) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Regex.Element(0, elem.value.data, key, pattern, options)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) DBPointer(key string, ns string, oid objectid.ObjectID) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(ns) + 1 + 12)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.DBPointer.Element(0, elem.value.data, key, ns, oid)
	if err != nil {
		panic(err)
	}

	return elem
}
func (Constructor) Javascript(key string, code string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(code) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Javascript.Element(0, elem.value.data, key, code)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Symbol(key string, symbol string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(symbol) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Symbol.Element(0, elem.value.data, key, symbol)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) CodeWithScope(key string, code string, scope *Document) *Element {
	size := uint32(1 + len(key) + 1 + 4 + 4 + len(code) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)
	elem.value.d = scope

	_, err := elements.Byte.Encode(0, elem.value.data, '\x0F')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.value.data, key)
	if err != nil {
		panic(err)
	}

	_, err = elements.Int32.Encode(1+uint(len(key))+1, elem.value.data, int32(size))
	if err != nil {
		panic(err)
	}

	_, err = elements.String.Encode(1+uint(len(key))+1+4, elem.value.data, code)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Int32(key string, i int32) *Element {
	size := uint32(1 + len(key) + 1 + 4)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Int32.Element(0, elem.value.data, key, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Timestamp(key string, t uint32, i uint32) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Timestamp.Element(0, elem.value.data, key, t, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Int64(key string, i int64) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := newElement(0, 1+uint32(len(key))+1)
	elem.value.data = make([]byte, size)

	_, err := elements.Int64.Element(0, elem.value.data, key, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Decimal128(key string, d decimal.Decimal128) *Element {
	size := uint32(1 + len(key) + 1 + 16)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Decimal128.Element(0, elem.value.data, key, d)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) MinKey(key string) *Element {
	size := uint32(1 + len(key) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Byte.Encode(0, elem.value.data, '\xFF')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.value.data, key)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) MaxKey(key string) *Element {
	size := uint32(1 + len(key) + 1)
	elem := newElement(0, uint32(1+len(key)+1))
	elem.value.data = make([]byte, size)

	_, err := elements.Byte.Encode(0, elem.value.data, '\x7F')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.value.data, key)
	if err != nil {
		panic(err)
	}

	return elem
}

func (ArrayConstructor) Double(f float64) *Value {
	return C.Double("", f).value
}

func (ArrayConstructor) String(val string) *Value {
	return C.String("", val).value
}

func (ArrayConstructor) Document(d *Document) *Value {
	return C.SubDocument("", d).value
}

func (ArrayConstructor) DocumentFromReader(r Reader) *Value {
	return C.SubDocumentFromReader("", r).value
}

func (ArrayConstructor) DocumentFromElements(elems ...*Element) *Value {
	return C.SubDocumentFromElements("", elems...).value
}

func (ArrayConstructor) Array(a *Array) *Value {
	return C.Array("", a).value
}

func (ArrayConstructor) ArrayFromValues(values ...*Value) *Value {
	return C.ArrayFromElements("", values...).value
}

func (ac ArrayConstructor) Binary(b []byte) *Value {
	return ac.BinaryWithSubtype(b, 0)
}

func (ArrayConstructor) BinaryWithSubtype(b []byte, btype byte) *Value {
	return C.BinaryWithSubtype("", b, btype).value
}

func (ArrayConstructor) Undefined() *Value {
	return C.Undefined("").value
}

func (ArrayConstructor) ObjectID(oid objectid.ObjectID) *Value {
	return C.ObjectID("", oid).value
}

func (ArrayConstructor) Boolean(b bool) *Value {
	return C.Boolean("", b).value
}

func (ArrayConstructor) DateTime(dt int64) *Value {
	return C.DateTime("", dt).value
}

func (ArrayConstructor) Null() *Value {
	return C.Null("").value
}

func (ArrayConstructor) Regex(pattern, options string) *Value {
	return C.Regex("", pattern, options).value
}

func (ArrayConstructor) DBPointer(ns string, oid objectid.ObjectID) *Value {
	return C.DBPointer("", ns, oid).value
}

func (ArrayConstructor) Javascript(code string) *Value {
	return C.Javascript("", code).value
}

func (ArrayConstructor) Symbol(symbol string) *Value {
	return C.Symbol("", symbol).value
}

func (ArrayConstructor) CodeWithScope(code string, scope *Document) *Value {
	return C.CodeWithScope("", code, scope).value
}

func (ArrayConstructor) Int32(i int32) *Value {
	return C.Int32("", i).value
}

func (ArrayConstructor) Timestamp(t uint32, i uint32) *Value {
	return C.Timestamp("", t, i).value
}

func (ArrayConstructor) Int64(i int64) *Value {
	return C.Int64("", i).value
}

func (ArrayConstructor) Decimal128(d decimal.Decimal128) *Value {
	return C.Decimal128("", d).value
}

func (ArrayConstructor) MinKey() *Value {
	return C.MinKey("").value
}

func (ArrayConstructor) MaxKey() *Value {
	return C.MaxKey("").value
}
