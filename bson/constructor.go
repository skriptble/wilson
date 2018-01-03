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

func (ArrayConstructor) Double(f float64) *Element            { return nil }
func (ArrayConstructor) String(val string) *Element           { return nil }
func (ArrayConstructor) Document(elems ...*Element) *Element  { return nil }
func (ArrayConstructor) Array(elemens ...*Element) *Element   { return nil }
func (ArrayConstructor) Binary(b []byte, btype uint) *Element { return nil }
func (ArrayConstructor) ObjectID(obj [12]byte) *Element       { return nil }
func (ArrayConstructor) Boolean(b bool) *Element              { return nil }
func (ArrayConstructor) DateTime(dt int64) *Element           { return nil }
func (ArrayConstructor) Null() *Element {
	return nil
	// size := uint32(1 + len(key) + 1)
	// b := make([]byte, size)
	// elem := new(Element)
	// elem.start = 0
	// elem.value = uint32(1 + len(key) + 1)
	// _, err := elements.Byte.Encode(0, b, '\x0A')
	// if err != nil {
	// 	panic(err)
	// }
	// _, err = elements.CString.Encode(1, b, key)
	// if err != nil {
	// 	panic(err)
	// }
	// elem.data = b
	// return elem
}
func (ArrayConstructor) Regex(pattern, options string) *Element            { return nil }
func (ArrayConstructor) DBPointer(dbpointer [12]byte) *Element             { return nil }
func (ArrayConstructor) Javascript(js string) *Element                     { return nil }
func (ArrayConstructor) Symbol(symbol string) *Element                     { return nil }
func (ArrayConstructor) CodeWithScope(js string, scope *Document) *Element { return nil }
func (ArrayConstructor) Int32(i int32) *Element                            { return nil }
func (ArrayConstructor) Uint64(u uint64) *Element                          { return nil }
func (ArrayConstructor) Int64(i int64) *Element                            { return nil }
func (ArrayConstructor) Decimal128(d ast.Decimal128) *Element              { return nil }
