package bson

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

// ReaderElement represents a BSON document element. An unintialized ReaderElement will
// panic if any of it's methods are invoked.
//
// This type is composed of an identifier, the length of the key, a value type,
// and the underlying slice that represents the BSON document. The identifier
// is the index into the slice where the ReaderElement starts. The keylen property
// is the length of the key. The keylen property includes the ending null value
// of the c-style string, making the total length of the element
// 1+keylen+vtype.length().
type ReaderElement struct {
	start uint32
	value uint32

	data []byte
}

// TODO(skriptble): Do we need this function? It's only useful if we want to
// allow users to construct their own Elements and not use a constructor.
func NewElement(start, value uint32, data []byte) *ReaderElement {
	return &ReaderElement{start: start, value: value, data: data}
}

// Validates the element and returns its total size.
func (e *ReaderElement) Validate() (uint32, error) {
	var total uint32 = 1
	n, err := e.keySize()
	total += n
	if err != nil {
		return total, err
	}
	n, err = e.validateValue(true)
	total += n
	if err != nil {
		return total, err
	}
	return total, nil
}

// TODO(skriptble): Rename this validateKey to match validateValue.
func (e *ReaderElement) keySize() (uint32, error) {
	pos, end := e.start+1, e.value
	var total uint32 = 0
	for ; pos < end && e.data[pos] != '\x00'; pos++ {
		total++
	}
	if pos == end || e.data[pos] != '\x00' {
		return total, ErrInvalidKey
	}
	total++
	return total, nil
}

// valueSize returns the size of the value in bytes.
func (e *ReaderElement) valueSize() (uint32, error) {
	switch e.data[e.start] {
	case '\x03', '\x04':
		if int(e.value+4) > len(e.data) {
			return 0, errors.New("Too small")
		}
		size := readi32(e.data[e.value : e.value+4])
		if int32(e.value)+size > int32(len(e.data)) {
			return 0, errors.New("Too small")
		}
		// TODO(skriptble): We need a check here to ensure that the size is
		// not negative.
		return uint32(size), nil
	default:
		return e.validateValue(true)
	}
}

func (e *ReaderElement) validateValue(recursive bool) (uint32, error) {
	var total uint32 = 0
	switch e.data[e.start] {
	case '\x06', '\x0A', '\xFF', '\x7F':
	case '\x01':
		if int(e.value+8) >= len(e.data) {
			return total, ErrTooSmall
		}
		total += 8
	case '\x02', '\x0D', '\x0E':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+4+l > int32(len(e.data)) {
			return total, errors.New("Too small")
		}
		total += uint32(l)
	case '\x03', '\x04':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		// TODO(skriptble): This is wrong and could cause a panic.
		l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
		total += 4
		if int32(e.value)+l > int32(len(e.data)) {
			return total, ErrInvalidReadOnlyDocument
		}
		if recursive {
			n, err := Reader(e.data[e.value : e.value+uint32(l)]).Validate()
			total += n - 4
			if err != nil {
				return total, err
			}
			break
		}
		total += uint32(l) - 4
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
		total += uint32(l) - 4
		// TODO(skriptble): This is wrong and could cause a panic.
		sLength := int32(binary.LittleEndian.Uint32(e.data[e.value+4 : e.value+8]))
		// If the length of the string is larger than the total length of the
		// field minus the int32 for length, 5 bytes for a minimum document
		// size, and an int32 for the string length the value is invalid.
		if sLength > l-13 {
			return total, errors.New("String size is larger than the Code With Scope container")
		}
		n, err := Reader(e.data[e.value+4+uint32(sLength) : e.value+uint32(l)]).Validate()
		total += n
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

// Key returns the key for this element.
// It panics if e is uninitialized.
func (e *ReaderElement) Key() string {
	if e == nil || e.start == 0 {
		panic(ErrUninitializedElement)
	}
	return string(e.data[e.start+1 : e.value-1])
}

// Type returns the identifying element byte for this element.
// It panics if e is uninitialized.
func (e *ReaderElement) Type() byte {
	if e == nil || e.start == 0 {
		panic(ErrUninitializedElement)
	}
	return e.data[e.start]
}

// Double returns the float64 value for this element.
// It panics if e's BSON type is not Double ('\x01') or if e is uninitialized.
func (e *ReaderElement) Double() float64 {
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
func (e *ReaderElement) String() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x02' {
		panic(ElementTypeError{"compact.Element.String", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *ReaderElement) Document() Reader {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x03' {
		panic(ElementTypeError{"compact.Element.Document", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return Reader(e.data[e.value : e.value+uint32(l)])
}

func (e *ReaderElement) Array() Reader {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x04' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return Reader(e.data[e.value : e.value+uint32(l)])
}

func (e *ReaderElement) Binary() (subtype byte, data []byte) {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x05' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	st := e.data[e.value+5]
	b := make([]byte, l)
	copy(b, e.data[e.value+5:int32(e.value)+5+l])
	return st, b
}

func (e *ReaderElement) ObjectID() [12]byte {
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

func (e *ReaderElement) Boolean() bool {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x08' {
		panic(ElementTypeError{"compact.Element.Boolean", BSONType(e.data[e.start])})
	}
	return e.data[e.value] == '\x01'
}

func (e *ReaderElement) DateTime() time.Time {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x09' {
		panic(ElementTypeError{"compact.Element.DateTime", BSONType(e.data[e.start])})
	}
	i := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	return time.Unix(int64(i)/1000, int64(i)%1000*1000000)
}

func (e *ReaderElement) Regex() (pattern, options string) {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0B' {
		panic(ElementTypeError{"compact.Element.Regex", BSONType(e.data[e.start])})
	}
	// TODO(skriptble): Use the elements package here.
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

func (e *ReaderElement) DBPointer() [12]byte {
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

func (e *ReaderElement) Javascript() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0D' {
		panic(ElementTypeError{"compact.Element.Javascript", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *ReaderElement) Symbol() string {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0E' {
		panic(ElementTypeError{"compact.Element.Symbol", BSONType(e.data[e.start])})
	}
	l := int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
	return string(e.data[e.value+4 : int32(e.value)+4+l])
}

func (e *ReaderElement) JavascriptWithScope() (code string, rdr Reader) {
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
	r := Reader(e.data[e.value+4+uint32(sLength) : e.value+uint32(l)])
	return str, r
}

func (e *ReaderElement) Int32() int32 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x10' {
		panic(ElementTypeError{"compact.Element.Int32", BSONType(e.data[e.start])})
	}
	return int32(binary.LittleEndian.Uint32(e.data[e.value : e.value+4]))
}

func (e *ReaderElement) Uint64() uint64 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x11' {
		panic(ElementTypeError{"compact.Element.Uint64", BSONType(e.data[e.start])})
	}
	return binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
}

func (e *ReaderElement) Int64() int64 {
	if e == nil || e.start == 0 || e.value == 0 {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x12' {
		panic(ElementTypeError{"compact.Element.Int64", BSONType(e.data[e.start])})
	}
	return int64(binary.LittleEndian.Uint64(e.data[e.value : e.value+8]))
}

func (e *ReaderElement) Decimal128() Decimal128 {
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
