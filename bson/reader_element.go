package bson

import (
	"encoding/binary"
	"errors"
	"math"
	"time"
)

const validateMaxDepthDefault = 2048

var ErrNilReaderElement = errors.New("ReaderElement is nil")
var ErrInvalidString = errors.New("Invalid string value")
var ErrInvalidBinarySubtype = errors.New("Invalid BSON Binary Subtype")
var ErrInvalidBooleanType = errors.New("Invalid value for BSON Boolean Type")
var ErrStringLargerThanContainer = errors.New("String size is larger than the Code With Scope container")
var ErrInvalidElement = errors.New("Invalid Element")

// ReaderElement represents a read-only BSON element. An unintialized
// ReaderElement will panic if the Key method or any of it's value methods are
// invoked. Validate will return an error when invoked on an unintialized
// ReaderElement.
type ReaderElement struct {
	// start is the offset into the data slice of bytes where this element
	// begins.
	start uint32
	// value is the offset into the data slice of bytes where this element's
	// value begins.
	value uint32

	// data is a potentially shared slice of bytes that contains the actual
	// element. Most of the methods of this type directly index into this slice
	// of bytes.
	data []byte
}

// TODO(skriptble): Do we need this function? It's only useful if we want to
// allow users to construct their own Elements and not use a constructor.
func NewElement(start, value uint32, data []byte) *ReaderElement {
	return &ReaderElement{start: start, value: value, data: data}
}

// Validates the element and returns its total size.
func (e *ReaderElement) Validate() (uint32, error) {
	if e == nil {
		return 0, ErrNilReaderElement
	}
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

// validate is a common validation method for elements.
//
// TODO(skriptble): Fill out this method and ensure all validation routines
// pass through this method.
func (e *ReaderElement) validate(recursive bool, currentDepth, maxDepth uint32) (uint32, error) {
	return 0, nil
}

// TODO(skriptble): Rename this validateKey to match validateValue.
func (e *ReaderElement) keySize() (uint32, error) {
	pos, end := e.start+1, e.value
	var total uint32 = 0
	if end > uint32(len(e.data)) {
		end = uint32(len(e.data))
	}
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
	return e.validateValue(false)
}

// TODO(skriptble): Rename the recursive parameter. It's more like a deep
// validation, e.g. if it's false we only validate what we have to, so we
// validate there are enough bytes in the slice to hold the string, but we
// don't check if the last byte of the string is 0x00. If we do a deep
// validation then we would check if there is a 0x00 as the last byte in the
// string. Similarly, we would actually validate subdocuments and arrays.
func (e *ReaderElement) validateValue(recursive bool) (uint32, error) {
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
		l := readi32(e.data[e.value : e.value+4])
		total += 4
		if int32(e.value)+4+l > int32(len(e.data)) {
			return total, ErrTooSmall
		}
		// We check if the value that is the last element of the string is a
		// null terminator. We take the value offset, add 4 to account for the
		// length, add the length of the string, and subtract one since the size
		// isn't zero indexed.
		if recursive && e.data[e.value+4+uint32(l)-1] != 0x00 {
			return total, ErrInvalidString
		}
		total += uint32(l)
	// TODO(skriptble): Should we validate that an Array's keys are numeric and
	// in increasing order?
	case '\x03', '\x04':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		l := readi32(e.data[e.value : e.value+4])
		total += 4
		if l < 5 {
			return total, ErrInvalidReadOnlyDocument
		}
		if int32(e.value)+l > int32(len(e.data)) {
			return total, ErrTooSmall
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
			return total, ErrTooSmall
		}
		l := readi32(e.data[e.value : e.value+4])
		total += 5
		if e.data[e.value+4] > '\x05' && e.data[e.value+4] < '\x80' {
			return total, ErrInvalidBinarySubtype
		}
		if int32(e.value)+5+l > int32(len(e.data)) {
			return total, ErrTooSmall
		}
		total += uint32(l)
	case '\x07':
		if int(e.value+12) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 12
	case '\x08':
		if int(e.value+1) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 1
		if e.data[e.value] != '\x00' && e.data[e.value] != '\x01' {
			return total, ErrInvalidBooleanType
		}
	case '\x09':
		if int(e.value+8) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 8
	case '\x0B':
		i := e.value
		for ; int(i) < len(e.data) && e.data[i] != '\x00'; i++ {
			total++
		}
		if int(i) == len(e.data) || e.data[i] != '\x00' {
			return total, ErrInvalidString
		}
		i++
		total++
		for ; int(i) < len(e.data) && e.data[i] != '\x00'; i++ {
			total++
		}
		if int(i) == len(e.data) || e.data[i] != '\x00' {
			return total, ErrInvalidString
		}
		total++
	case '\x0C':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		l := readi32(e.data[e.value : e.value+4])
		total += 4
		if int32(e.value)+4+l+12 > int32(len(e.data)) {
			return total, ErrTooSmall
		}
		total += uint32(l) + 12
	case '\x0F':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		l := readi32(e.data[e.value : e.value+4])
		total += 4
		if int32(e.value)+l > int32(len(e.data)) {
			return total, ErrTooSmall
		}
		if recursive {
			sLength := readi32(e.data[e.value+4 : e.value+8])
			total += 4
			// If the length of the string is larger than the total length of the
			// field minus the int32 for length, 5 bytes for a minimum document
			// size, and an int32 for the string length the value is invalid.
			//
			// TODO(skriptble): We should actually validate that the string
			// doesn't consume any of the bytes used by the document.
			if sLength > l-13 {
				return total, ErrStringLargerThanContainer
			}
			// We check if the value that is the last element of the string is a
			// null terminator. We take the value offset, add 4 to account for the
			// length, add the length of the string, and subtract one since the size
			// isn't zero indexed.
			if e.data[e.value+8+uint32(sLength)-1] != 0x00 {
				return total, ErrInvalidString
			}
			total += uint32(sLength)
			n, err := Reader(e.data[e.value+8+uint32(sLength) : e.value+uint32(l)]).Validate()
			total += n
			if err != nil {
				return total, err
			}
			break
		}
		total += uint32(l) - 4
	case '\x10':
		if int(e.value+4) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 4
	case '\x11', '\x12':
		if int(e.value+8) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 8
	case '\x13':
		if int(e.value+16) > len(e.data) {
			return total, ErrTooSmall
		}
		total += 16
	default:
		return total, ErrInvalidElement
	}

	return total, nil
}

// writeByteSlice handles writing this element to a slice of bytes.
func (e *ReaderElement) writeByteSlice(start uint, size uint32, b []byte) (int64, error) {
	if len(b) < int(size)+int(start) {
		return 0, ErrTooSmall
	}
	var n int
	n = copy(b[start:start+uint(size)], e.data[e.start:e.start+size])
	return int64(n), nil
}

// MarshalBSON implements the Marshaler interface.
func (e *ReaderElement) MarshalBSON() ([]byte, error) {
	size, err := e.Validate()
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

// Key returns the key for this element.
// It panics if e is uninitialized.
func (e *ReaderElement) Key() string {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	return string(e.data[e.start+1 : e.value-1])
}

// Type returns the identifying element byte for this element.
// It panics if e is uninitialized.
func (e *ReaderElement) Type() byte {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	return e.data[e.start]
}

// Double returns the float64 value for this element.
// It panics if e's BSON type is not Double ('\x01') or if e is uninitialized.
func (e *ReaderElement) Double() float64 {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x01' {
		panic(ElementTypeError{"compact.Element.Double", BSONType(e.data[e.start])})
	}
	bits := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	return math.Float64frombits(bits)
}

// StringValue returns the string balue for this element.
// It panics if e's BSON type is not StringValue ('\x02') or if e is uninitialized.
//
// NOTE: This method is called StringValue to avoid it implementing the
// fmt.Stringer interface.
func (e *ReaderElement) StringValue() string {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x02' {
		panic(ElementTypeError{"compact.Element.String", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	return string(e.data[e.value+4 : int32(e.value)+4+l-1])
}

func (e *ReaderElement) Document() Reader {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x03' {
		panic(ElementTypeError{"compact.Element.Document", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	return Reader(e.data[e.value : e.value+uint32(l)])
}

func (e *ReaderElement) Array() Reader {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x04' {
		panic(ElementTypeError{"compact.Element.Array", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	return Reader(e.data[e.value : e.value+uint32(l)])
}

func (e *ReaderElement) Binary() (subtype byte, data []byte) {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x05' {
		panic(ElementTypeError{"compact.Element.Binary", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	st := e.data[e.value+4]
	b := make([]byte, l)
	copy(b, e.data[e.value+5:int32(e.value)+5+l])
	return st, b
}

func (e *ReaderElement) ObjectID() [12]byte {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x07' {
		panic(ElementTypeError{"compact.Element.ObejctID", BSONType(e.data[e.start])})
	}
	var arr [12]byte
	copy(arr[:], e.data[e.value:e.value+12])
	return arr
}

func (e *ReaderElement) Boolean() bool {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x08' {
		panic(ElementTypeError{"compact.Element.Boolean", BSONType(e.data[e.start])})
	}
	return e.data[e.value] == '\x01'
}

func (e *ReaderElement) DateTime() time.Time {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x09' {
		panic(ElementTypeError{"compact.Element.DateTime", BSONType(e.data[e.start])})
	}
	i := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	return time.Unix(int64(i)/1000, int64(i)%1000*1000000)
}

func (e *ReaderElement) Regex() (pattern, options string) {
	if e == nil || e.value == 0 || e.data == nil {
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

func (e *ReaderElement) DBPointer() (string, [12]byte) {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0C' {
		panic(ElementTypeError{"compact.Element.DBPointer", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	var p [12]byte
	copy(p[:], e.data[e.value+4+uint32(l):e.value+4+uint32(l)+12])
	return string(e.data[e.value+4 : int32(e.value)+4+l-1]), p
}

func (e *ReaderElement) Javascript() string {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0D' {
		panic(ElementTypeError{"compact.Element.Javascript", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	return string(e.data[e.value+4 : int32(e.value)+4+l-1])
}

func (e *ReaderElement) Symbol() string {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0E' {
		panic(ElementTypeError{"compact.Element.Symbol", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	return string(e.data[e.value+4 : int32(e.value)+4+l-1])
}

func (e *ReaderElement) JavascriptWithScope() (code string, rdr Reader) {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x0F' {
		panic(ElementTypeError{"compact.Element.JavascriptWithScope", BSONType(e.data[e.start])})
	}
	l := readi32(e.data[e.value : e.value+4])
	sLength := readi32(e.data[e.value+4 : e.value+8])
	// If the length of the string is larger than the total length of the
	// field minus the int32 for length, 5 bytes for a minimum document
	// size, and an int32 for the string length the value is invalid.
	str := string(e.data[e.value+8 : e.value+8+uint32(sLength)-1])
	r := Reader(e.data[e.value+8+uint32(sLength) : e.value+uint32(l)])
	return str, r
}

func (e *ReaderElement) Int32() int32 {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x10' {
		panic(ElementTypeError{"compact.Element.Int32", BSONType(e.data[e.start])})
	}
	return readi32(e.data[e.value : e.value+4])
}

func (e *ReaderElement) Timestamp() uint64 {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x11' {
		panic(ElementTypeError{"compact.Element.Timestamp", BSONType(e.data[e.start])})
	}
	return binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
}

func (e *ReaderElement) Int64() int64 {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x12' {
		panic(ElementTypeError{"compact.Element.Int64", BSONType(e.data[e.start])})
	}
	return int64(binary.LittleEndian.Uint64(e.data[e.value : e.value+8]))
}

func (e *ReaderElement) Decimal128() Decimal128 {
	if e == nil || e.value == 0 || e.data == nil {
		panic(ErrUninitializedElement)
	}
	if e.data[e.start] != '\x13' {
		panic(ElementTypeError{"compact.Element.Decimal128", BSONType(e.data[e.start])})
	}
	l := binary.LittleEndian.Uint64(e.data[e.value : e.value+8])
	h := binary.LittleEndian.Uint64(e.data[e.value+8 : e.value+16])
	return NewDecimal128(h, l)
}
