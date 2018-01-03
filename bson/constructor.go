package bson

import (
	"github.com/skriptble/wilson/bson/elements"
	"github.com/skriptble/wilson/bson/parser/ast"
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
	return c.SubDocument(key, NewDocument(uint(len(elems))).Append(elems...))
}
func (Constructor) Array(key string, d *Document) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := new(Element)
	elem.start = 0
	elem.value = size
	_, err := elements.Byte.Encode(0, b, '\x04')
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
func (c Constructor) ArrayFromElements(key string, elems ...*Element) *Element {
	return c.Array(key, NewDocument(uint(len(elems))).Append(elems...))
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
	elem := new(Element)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	_, err := elements.Binary.Element(0, buf, key, b, btype)
	if err != nil {
		panic(err)
	}

	elem.data = buf
	return elem
}

func (Constructor) Undefined(key string) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := new(Element)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	_, err := elements.Byte.Encode(0, b, '\x06')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.data = b
	return elem
}

func (Constructor) ObjectID(key string, oid [12]byte) *Element {
	size := uint32(1 + len(key) + 1 + 12)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.ObjectId.Element(0, elem.data, key, oid)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Boolean(key string, b bool) *Element {
	size := uint32(1 + len(key) + 1 + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Boolean.Element(0, elem.data, key, b)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) DateTime(key string, dt int64) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.DateTime.Element(0, elem.data, key, dt)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Null(key string) *Element {
	size := uint32(1 + len(key) + 1)
	b := make([]byte, size)
	elem := new(Element)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	_, err := elements.Byte.Encode(0, b, '\x0A')
	if err != nil {
		panic(err)
	}
	_, err = elements.CString.Encode(1, b, key)
	if err != nil {
		panic(err)
	}
	elem.data = b
	return elem
}
func (Constructor) Regex(key string, pattern, options string) *Element {
	size := uint32(1 + len(key) + 1 + len(pattern) + 1 + len(options) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Regex.Element(0, elem.data, key, pattern, options)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) DBPointer(key string, ns string, oid [12]byte) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(ns) + 1 + 12)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.DBPointer.Element(0, elem.data, key, ns, oid)
	if err != nil {
		panic(err)
	}

	return elem
}
func (Constructor) Javascript(key string, code string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(code) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Javascript.Element(0, elem.data, key, code)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Symbol(key string, symbol string) *Element {
	size := uint32(1 + len(key) + 1 + 4 + len(symbol) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Symbol.Element(0, elem.data, key, symbol)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) CodeWithScope(key string, code string, scope *Document) *Element {
	size := uint32(1 + len(key) + 1 + 4 + 4 + len(code) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)
	elem.d = scope

	_, err := elements.Byte.Encode(0, elem.data, '\x0F')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.data, key)
	if err != nil {
		panic(err)
	}

	_, err = elements.Int32.Encode(1+uint(len(key))+1, elem.data, int32(size))
	if err != nil {
		panic(err)
	}

	_, err = elements.String.Encode(1+uint(len(key))+1+4, elem.data, code)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Int32(key string, i int32) *Element {
	size := uint32(1 + len(key) + 1 + 4)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Int32.Element(0, elem.data, key, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Timestamp(key string, t uint32, i uint32) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Timestamp.Element(0, elem.data, key, t, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Int64(key string, i int64) *Element {
	size := uint32(1 + len(key) + 1 + 8)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Int64.Element(0, elem.data, key, i)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) Decimal128(key string, d ast.Decimal128) *Element {
	size := uint32(1 + len(key) + 1 + 16)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Decimal128.Element(0, elem.data, key, d)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) MinKey(key string) *Element {
	size := uint32(1 + len(key) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Byte.Encode(0, elem.data, '\xFF')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.data, key)
	if err != nil {
		panic(err)
	}

	return elem
}

func (Constructor) MaxKey(key string) *Element {
	size := uint32(1 + len(key) + 1)
	elem := new(Element)
	elem.data = make([]byte, size)
	elem.start = 0
	elem.value = uint32(1 + len(key) + 1)

	_, err := elements.Byte.Encode(0, elem.data, '\x7F')
	if err != nil {
		panic(err)
	}

	_, err = elements.CString.Encode(1, elem.data, key)
	if err != nil {
		panic(err)
	}

	return elem
}

func (ArrayConstructor) Double(f float64) *Element {
	return C.Double("0", f)
}

func (ArrayConstructor) String(val string) *Element {
	return C.String("0", val)
}

func (ArrayConstructor) Document(d *Document) *Element {
	return C.SubDocument("0", d)
}

func (ArrayConstructor) DocumentFromElements(elems ...*Element) *Element {
	return C.SubDocumentFromElements("0", elems...)
}

func (ArrayConstructor) Array(d *Document) *Element {
	return C.Array("0", d)
}

func (ArrayConstructor) ArrayFromElements(elems ...*Element) *Element {
	return C.ArrayFromElements("0", elems...)
}

func (ac ArrayConstructor) Binary(b []byte) *Element {
	return ac.BinaryWithSubtype(b, 0)
}

func (ArrayConstructor) BinaryWithSubtype(b []byte, btype byte) *Element {
	return C.BinaryWithSubtype("0", b, btype)
}

func (ArrayConstructor) Undefined() *Element {
	return C.Undefined("0")
}

func (ArrayConstructor) ObjectID(obj [12]byte) *Element {
	return C.ObjectID("0", obj)
}

func (ArrayConstructor) Boolean(b bool) *Element {
	return C.Boolean("0", b)
}

func (ArrayConstructor) DateTime(dt int64) *Element {
	return C.DateTime("0", dt)
}

func (ArrayConstructor) Null() *Element {
	return C.Null("0")
}

func (ArrayConstructor) Regex(pattern, options string) *Element {
	return C.Regex("0", pattern, options)
}

func (ArrayConstructor) DBPointer(ns string, oid [12]byte) *Element {
	return C.DBPointer("0", ns, oid)
}

func (ArrayConstructor) Javascript(code string) *Element {
	return C.Javascript("0", code)
}

func (ArrayConstructor) Symbol(symbol string) *Element {
	return C.Symbol("0", symbol)
}

func (ArrayConstructor) CodeWithScope(code string, scope *Document) *Element {
	return C.CodeWithScope("0", code, scope)
}

func (ArrayConstructor) Int32(i int32) *Element {
	return C.Int32("0", i)
}

func (ArrayConstructor) Timestamp(t uint32, i uint32) *Element {
	return C.Timestamp("0", t, i)
}

func (ArrayConstructor) Int64(i int64) *Element {
	return C.Int64("0", i)
}

func (ArrayConstructor) Decimal128(d ast.Decimal128) *Element {
	return C.Decimal128("0", d)
}

func (ArrayConstructor) MinKey() *Element {
	return C.MinKey("0")
}

func (ArrayConstructor) MaxKey() *Element {
	return C.MaxKey("0")
}
