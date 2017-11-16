package bson

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

// ErrUninitializedElement is returned whenever any method is invoked on an unintialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")

type ElementTypeError struct {
	Method string
	Type   BSONType
}

func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

// Element represents a BSON document element. An unintialized Element will
// panic if any of it's methods are invoked.
//
// This type is composed of an identifier, the length of the key, a value type,
// and the underlying slice that represents the BSON document. The identifier
// is the index into the slice where the Element starts. The keylen property
// is the length of the key. The keylen property includes the ending null value
// of the c-style string, making the total length of the element
// 1+keylen+vtype.length().
type Element struct {
	start uint32
	value uint32

	data []byte
}

func NewElement(start, value uint32, data []byte) *Element {
	return &Element{start: start, value: value, data: data}
}

func (e *Element) Recycle(start, value uint32, data []byte) {
	e.start, e.value, e.data = start, value, data
}

func (e *Element) Validate() error {
	return nil
}

// Key returns the key for this element.
// It panics if e is uninitialized.
func (e *Element) Key() string {
	if e == nil || e.start == 0 {
		panic(ErrUninitializedElement)
	}
	return string(e.data[e.start+1 : e.value-1])
}

// Type returns the identifying element byte for this element.
// It panics if e is uninitialized.
func (e *Element) Type() byte {
	if e == nil || e.start == 0 {
		panic(ErrUninitializedElement)
	}
	return e.data[e.start]
}

// Double returns the float64 value for this element.
// It panics if e's BSON type is not Double ('\x01') or if e is uninitialized.
func (e *Element) Double() float64 {
	if e == nil || e.start == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x01' {
		panic(ElementTypeError{"compact.Element.Double", BSONType(e.data[e.start])})
	}
	bits := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	return math.Float64frombits(bits)
}

// String returns the string balue for this element.
// It panics if e's BSON type is not String ('\x02') or if e is uninitialized.
func (e *Element) String() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x02' {
		panic(ElementTypeError{"compact.Element.String", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *Element) Document() *Document {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x03' {
		panic(ElementTypeError{"compact.Element.Document", BSONType(e.data[e.start])})
	}
	d := &Document{
		start: uint64(e.value),
		data:  e.data,
	}
	return d
}

func (e *Element) Array() *Document {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x04' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	d := &Document{
		start: uint64(e.value),
		data:  e.data,
	}
	return d
}

func (e *Element) Binary() *Binary {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x05' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	return &Binary{}
}

func (e *Element) ObjectID() [12]byte {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x07' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	var arr [12]byte
	copy(arr[:], e.data[e.value:e.value+12])
	return arr
}

func (e *Element) Boolean() bool {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x08' {
		panic(ElementTypeError{"compact.Element.Boolean", BSONType(e.data[e.start])})
	}
	return e.data[e.value] == '\x01'
}

func (e *Element) DateTime() time.Time {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x09' {
		panic(ElementTypeError{"compact.Element.DateTime", BSONType(e.data[e.start])})
	}
	i := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	return time.Unix(int64(i)/1000, int64(i)%1000*1000000)
}

func (e *Element) Regex() (pattern, options string) {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0B' {
		panic(ElementTypeError{"compact.Element.Regex", BSONType(e.data[e.start])})
	}
	var pstart, pend, ostart, oend uint32
	i := e.value
	pstart = i
	for ; e.data[i] != '\x00'; i++ {
	}
	pend = i
	i++
	ostart = i
	for ; e.data[i] != '\x00'; i++ {
	}
	oend = i

	return string(e.data[pstart:pend]), string(e.data[ostart:oend])
}

func (e *Element) DBPointer() [12]byte {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0C' {
		panic(ElementTypeError{"compact.Element.DBPointer", BSONType(e.data[e.start])})
	}
	var p [12]byte
	copy(p[:], e.data[e.value:e.value+12])
	return p
}

func (e *Element) Javascript() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0D' {
		panic(ElementTypeError{"compact.Element.Javascript", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *Element) Symbol() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0E' {
		panic(ElementTypeError{"compact.Element.Symbol", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *Element) JavascriptWithScope() *CodeWithScope {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0F' {
		panic(ElementTypeError{"compact.Element.JavascriptWithScope", BSONType(e.data[e.start])})
	}
	return &CodeWithScope{
		start: e.value,
		data:  e.data,
	}
}

func (e *Element) Int32() int32 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x10' {
		panic(ElementTypeError{"compact.Element.Int32", BSONType(e.data[e.start])})
	}
	return int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
}

func (e *Element) Uint64() uint64 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x11' {
		panic(ElementTypeError{"compact.Element.Uint64", BSONType(e.data[e.start])})
	}
	return binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
}

func (e *Element) Int64() int64 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x12' {
		panic(ElementTypeError{"compact.Element.Int64", BSONType(e.data[e.start])})
	}
	return int64(binary.LittleEndian.Uint64(e.data[e.value : e.value+8]))
}

func (e *Element) Decimal128() Decimal128 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x12' {
		panic(ElementTypeError{"compact.Element.Decimal", BSONType(e.data[e.start])})
	}
	l := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	h := binary.LittleEndian.Uint64(e.data[e.value+8 : e.value+16])
	return NewDecimal128(h, l)
}
