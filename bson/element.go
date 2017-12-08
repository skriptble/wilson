package bson

import (
	"encoding/binary"
	"errors"
	"io"

	"github.com/skriptble/wilson/parser/ast"
)

// ErrUninitializedElement is returned whenever any method is invoked on an unintialized Element.
var ErrUninitializedElement = errors.New("wilson/ast/compact: Method call on uninitialized Element")
var ErrTooSmall = errors.New("too small")
var ErrInvalidWriter = errors.New("bson: invalid writer provided")

type ElementTypeError struct {
	Method string
	Type   BSONType
}

func (ete ElementTypeError) Error() string {
	return "Call of " + ete.Method + " on " + ete.Type.String() + " type"
}

type Element struct {
	// NOTE: For SubDocuments and Arrays, the data byte slice may contain just
	// the key, in which case the start will be 0, the value will be the length
	// of the slice, and d will be non-nil.
	ReaderElement
	d *Document
}

func (e *Element) Validate(recursive bool) (uint32, error) {
	var total uint32 = 1
	n, err := e.keySize()
	total += n
	if err != nil {
		return total, err
	}
	n, err = e.validateValue(recursive)
	total += n
	if err != nil {
		return total, err
	}
	return total, nil
}

func (e *Element) validateValue(recursive bool) (uint32, error) {
	var total uint32 = 0
	switch e.data[e.start] {
	case '\x06', '\x0A', '\xFF', '\x7F':
	case '\x01':
		if int(e.value+8) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 8
	case '\x02', '\x0D', '\x0E':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+4+l > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		total += uint32(l)
	case '\x03', '\x04':
		if e.d == nil {
			if int(e.value+4) > len(e.data) {
				return total, errors.New("Too small")
			}
			l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
			total += 4
			if int32(e.value)+4+l > int32(len(e.data)) {
				return total, errors.New("Too small")
			}
			n, err := Reader(e.data[e.value : e.value+4+uint32(l)]).Validate()
			total += n
			if err != nil {
				return total, err
			}
			break
		}

		n, err := e.d.Validate()
		total += uint32(n)
		if err != nil {
			return total, err
		}
	case '\x05':
		if int(e.value+5) > len(e.data) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 5
		if e.data[e.value+4] > '\x05' && e.data[e.value+4] < '\x80' {
			return total, errors.New("Invalid BSON Binary Subtype")
		}
		if int32(e.value)+5+l > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		total += uint32(l)
	case '\x07':
		if int(e.value+12) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 12
	case '\x08':
		if int(e.value+1) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 1
		if e.data[e.value] != '\x00' && e.data[e.value] != '\x01' {
			return total, errors.New("Invalid value for BSON Boolean Type")
		}
	case '\x09':
		if int(e.value+8) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 8
	case '\x0B':
		i := e.value
		for ; int(i) < len(e.data) && e.data[i] != '\x00'; i++ {
			total++
		}
		i++
		if int(i) > len(e.data) {
			return total, errors.New("Too small")
		}
		total++
		for ; int(i) < len(e.data) && e.data[i] != '\x00'; i++ {
			total++
		}
		i++
		if int(i) > len(e.data) {
			return total, errors.New("Too small")
		}
		total++
	case '\x0C':
		if int(e.value+4) > len(e.data) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+4+l+12 > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		total += uint32(l) + 12
	case '\x0F':
		if int(e.value+4) > len(e.data) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+l > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		sLength := int32(binary.LittleEndian.Uint32(e.data[e.value+4 : e.value+8]))
		// If the length of the string is larger than the total length of the
		// field minus the int32 for length, 5 bytes for a minimum document
		// size, and an int32 for the string length
		if sLength > l-13 {
			return total, errors.New("String size is larger than the Code With Scope container")
		}
		total += uint32(sLength)
		if e.d == nil {
			n, err := Reader(e.data[e.value+4+uint32(sLength) : e.value+uint32(l)]).Validate()
			total += n
			if err != nil {
				return total, err
			}
			break
		}

		n, err := e.d.Validate()
		total += uint32(n)
		if err != nil {
			return total, err
		}
	case '\x10':
		if int(e.value+4) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 4
	case '\x11', '\x12':
		if int(e.value+8) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 8
	case '\x13':
		if int(e.value+16) > len(e.data) {
			return total, errors.New("Too small")
		}
		total += 16
	default:
		return total, errors.New("Invalid Element")
	}

	return total, nil
}

func (e *Element) Document() *Document {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x03' {
		panic(ElementTypeError{"compact.Element.Document", BSONType(e.data[e.start])})
	}
	if e.d == nil {
		var err error
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		e.d, err = ReadDocument(e.data[e.value : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return e.d
}

func (e *Element) Array() *Array {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x04' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	if e.d == nil {
		var err error
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		e.d, err = ReadDocument(e.data[e.value : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return &Array{e.d}
}

func (e *Element) JavascriptWithScope() (code string, d *Document) {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0F' {
		panic(ElementTypeError{"compact.Element.JavascriptWithScope", BSONType(e.data[e.start])})
	}
	// TODO(skriptble): This is wrong and could cause a panic.
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	// TODO(skriptble): This is wrong and could cause a panic.
	sLength := int32(binary.LittleEndian.Uint32(e.data[e.value+4 : e.value+8]))
	// If the length of the string is larger than the total length of the
	// field minus the int32 for length, 5 bytes for a minimum document
	// size, and an int32 for the string length the value is invalid.
	str := string(e.data[e.value+4 : e.value+4+uint32(sLength)])
	if e.d == nil {
		var err error
		e.d, err = ReadDocument(e.data[e.value+4+uint32(sLength) : e.value+uint32(l)])
		if err != nil {
			panic(err)
		}
	}
	return str, e.d
}

func (e *Element) ConvertToDouble(f float64)                         {}
func (e *Element) ConvertToString(val string)                        {}
func (e *Element) ConvertToDocument(doc *Document)                   {}
func (e *Element) ConvertToArray(arr *Array)                         {}
func (e *Element) ConvertToBinary(b []byte, btype uint)              {}
func (e *Element) ConvertToObjectID(obj [12]byte)                    {}
func (e *Element) ConvertToBoolean(b bool)                           {}
func (e *Element) ConvertToDateTime(dt int64)                        {}
func (e *Element) ConvertToRegex(pattern, options string)            {}
func (e *Element) ConvertToDBPointer(dbpointer [12]byte)             {}
func (e *Element) ConvertToJavascript(js string)                     {}
func (e *Element) ConvertToSymbol(symbol string)                     {}
func (e *Element) ConvertToCodeWithScope(js string, scope *Document) {}
func (e *Element) ConvertToInt32(i int32)                            {}
func (e *Element) ConvertToUint64(u uint64)                          {}
func (e *Element) ConvertToInt64(i int64)                            {}
func (e *Element) ConvertToDecimal128(d ast.Decimal128)              {}

// NOTE: This is private since we don't want anything outside of this package
// to change the key of a element that is inside of a document.
func (e *Element) updateKey(key string) error                     { return nil }
func (e *Element) UpdateDouble(f float64)                         {}
func (e *Element) UpdateString(val string)                        {}
func (e *Element) UpdateDocument(doc *Document)                   {}
func (e *Element) UpdateArray(arr *Array)                         {}
func (e *Element) UpdateBinary(b []byte, btype uint)              {}
func (e *Element) UpdateObjectID(obj [12]byte)                    {}
func (e *Element) UpdateBoolean(b bool)                           {}
func (e *Element) UpdateDateTime(dt int64)                        {}
func (e *Element) UpdateRegex(pattern, options string)            {}
func (e *Element) UpdateDBPointer(dbpointer [12]byte)             {}
func (e *Element) UpdateJavascript(js string)                     {}
func (e *Element) UpdateSymbol(symbol string)                     {}
func (e *Element) UpdateCodeWithScope(js string, scope *Document) {}
func (e *Element) UpdateInt32(i int32)                            {}
func (e *Element) UpdateUint64(u uint64)                          {}
func (e *Element) UpdateInt64(i int64)                            {}
func (e *Element) UpdateDecimal128(d ast.Decimal128)              {}

func (e *Element) WriteTo(w io.Writer) (int64, error) {
	return 0, nil
}

func (e *Element) WriteElement(start uint, writer interface{}) (int64, error) {
	// TODO(skriptble): Figure out if we want to use uint or uint32 and
	// standardize across all packages.
	var total int64
	size, err := e.Validate(true)
	if err != nil {
		return 0, err
	}
	switch w := writer.(type) {
	case []byte:
		n, err := e.writeByteSlice(start, size, w)
		if err != nil {
			return 0, ErrTooSmall
		}
		total += int64(n)
	default:
		return 0, ErrInvalidWriter
	}
	return total, nil
}

func (e *Element) MarshalBSON() ([]byte, error) {
	size, err := e.Validate(true)
	if err != nil {
		return nil, err
	}
	b := make([]byte, size)
	_, err = e.writeByteSlice(0, size, b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (e *Element) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	if len(b) < int(size)+int(start) {
		return 0, ErrTooSmall
	}
	var n int
	switch e.data[e.start] {
	// TODO(skriptble): For types that contain a document, we need to check if
	// the d property is nil. If it is, we can do a regular copy. If it is
	// non-nil we need to marshal that Document to bytes then copy it into the
	// byte slice.
	case '\x03':
		if e.d == nil {
			n = copy(b[start:start+uint(size)], e.data[e.start:e.start+size])
			break
		}
		header := e.value - e.start
		n += copy(b[start:start+uint(header)], e.data[e.start:e.value])
		start += uint(n)
		size -= header
		nn, err := e.d.writeByteSlice(start, size, b)
		n += int(nn)
		if err != nil {
			return int64(n), err
		}
	case '\x04':
	case '\x0F':
	default:
		n = copy(b[start:start+uint(size)], e.data[e.start:e.start+size])
	}
	return int64(n), nil
}
