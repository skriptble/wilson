// package element holds the logic to encode and decode the BSON element types
// from native Go to BSON binary and vice versa.
//
// The encoder helper methods assume that when provided with an io.Writer that
// the writer is properly positioned to write the value. The decoder helper
// methods similarly assume that when provided with an io.Reader that the reader
// is properly positioned to read the value.
//
// These are low level helper methods, so they do not encode or decode BSON
// elements, only the specific types, e.g. these methods do not encode, decode,
// or identify a BSON element, so they won't read the identifier byte and they
// won't parse out the key string. There are encoder and decoder helper methods
// for the CString BSON element type, so this package can be used to parse
// keys.
package elements

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
	"unsafe"

	"github.com/skriptble/wilson/parser/ast"
)

var ErrInvalidWriter = errors.New("element: Invalid writer provided")
var ErrInvalidReader = errors.New("element: Invalid reader provided")
var ErrTooSmall = errors.New("element: The provided slice is too small")

var Double double
var String str
var Document document
var Array array
var Binary bin
var ObjectId objectid
var Boolean boolean
var DateTime datetime
var Regex regex
var DBPointer dbpointer
var Javascript javascript
var Symbol symbol
var CodeWithScope codewithscope
var Int32 i32
var Timestamp timestamp
var Int64 i64
var Decimal128 decimal128
var CString cstring
var Byte bsonbyte

type double struct{}
type str struct{}
type document struct{}
type array struct{}
type bin struct{}
type objectid struct{}
type boolean struct{}
type datetime struct{}
type regex struct{}
type dbpointer struct{}
type javascript struct{}
type symbol struct{}
type codewithscope struct{}
type i32 struct{}
type timestamp struct{}
type i64 struct{}
type decimal128 struct{}
type cstring struct{}
type bsonbyte struct{}

// Encodes a float64 into a BSON double element and serializes the bytes to the
// provided writer.
//
// writer can be:
//
// - []byte
// - io.WriterAt
// - io.WriteSeeker
// - io.Writer
func (double) Encode(start uint, writer interface{}, f float64) (int, error) {
	var written int
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+8) {
			return 0, ErrTooSmall
		}
		bits := math.Float64bits(f)
		binary.LittleEndian.PutUint64(w[start:start+8], bits)
		written = 8
	default:
		return 0, ErrInvalidWriter
	}
	return written, nil
}

// Decode will unserialize the bytes from the provided reader and decode a BSON
// double element into a float64.
//
// read can be:
//
// - []byte
// - io.ReaderAt
// - io.ReadSeeker
// - io.ByteReader
// - io.Reader
func (double) Decode(start uint, reader interface{}) (float64, error) {
	switch r := reader.(type) {
	case []byte:
		if len(r) < int(start+8) {
			return 0, ErrTooSmall
		}
		bits := binary.LittleEndian.Uint64(r[start : start+8])
		return math.Float64frombits(bits), nil
	default:
		return 0, ErrInvalidReader
	}
}

func (double) Element(start uint, writer interface{}, key string, f float64) (int, error) {
	var total int
	n, err := Byte.Encode(start, writer, '\x01')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}
	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}
	n, err = Double.Encode(start, writer, f)
	total += n
	if err != nil {
		return total, err
	}
	return total, nil
}

func (str) Encode(start uint, writer interface{}, s string) (int, error) {
	var total int

	written, err := Int32.Encode(start, writer, int32(len(s))+1)
	total += written
	if err != nil {
		return total, err
	}

	written, err = CString.Encode(start+uint(total), writer, s)
	total += written

	return total, nil
}

func (str) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (str) Element(start uint, writer interface{}, key string, s string) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x02')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = String.Encode(start, writer, s)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (document) Encode(start uint, writer interface{}, doc []byte) (int, error) {
	return encodeByteSlice(start, writer, doc)
}

func (document) Decode(start uint, reader interface{}) ([]byte, error) {
	return nil, nil
}

func (document) Element(start uint, writer interface{}, key string, doc []byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x03')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Document.Encode(start, writer, doc)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (array) Encode(start uint, writer interface{}, arr []byte) (int, error) {
	return Document.Encode(start, writer, arr)
}

func (array) Decode(start uint, reader interface{}) ([]byte, error) {
	return nil, nil
}

func (array) Element(start uint, writer interface{}, key string, arr []byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x04')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Array.Encode(start, writer, arr)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (bin) Encode(start uint, writer interface{}, b []byte, btype byte) (int, error) {
	if btype == 2 {
		return Binary.encodeSubtype2(start, writer, b)
	}

	var total int

	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start)+5+len(b) {
			return 0, ErrTooSmall
		}

		// write length
		n, err := Int32.Encode(start, writer, int32(len(b)))
		start += uint(n)
		total += n
		if err != nil {
			return total, err
		}

		w[start] = btype
		start++
		total += 1

		total += copy(w[start:], b)

	default:
		return 0, ErrInvalidWriter
	}

	return total, nil
}

func (bin) encodeSubtype2(start uint, writer interface{}, b []byte) (int, error) {
	var total int

	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start)+9+len(b) {
			return 0, ErrTooSmall
		}

		// write length
		n, err := Int32.Encode(start, writer, int32(len(b))+4)
		start += uint(n)
		total += n
		if err != nil {
			return total, err
		}

		w[start] = 2
		start++
		total += 1

		n, err = Int32.Encode(start, writer, int32(len(b)))
		start += uint(n)
		total += n
		if err != nil {
			return total, err
		}

		total += copy(w[start:], b)
	}

	return total, nil
}

func (bin) Decode(start uint, reader interface{}) (b []byte, btype byte, err error) {
	return nil, 0, nil
}

func (bin) Element(start uint, writer interface{}, key string, b []byte, btype byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x05')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Binary.Encode(start, writer, b, btype)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (objectid) Encode(start uint, writer interface{}, oid [12]byte) (int, error) {
	return encodeByteSlice(start, writer, oid[:])
}

func (objectid) Decode(start uint, reader interface{}) ([12]byte, error) {
	var obj [12]byte
	return obj, nil
}

func (objectid) Element(start uint, writer interface{}, key string, oid [12]byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x07')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = ObjectId.Encode(start, writer, oid)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (boolean) Encode(start uint, writer interface{}, b bool) (int, error) {
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start)+1 {
			return 0, ErrTooSmall
		}

		if b {
			w[start] = 1
		} else {
			w[start] = 0
		}

	default:
		return 0, ErrInvalidWriter
	}

	return 1, nil
}

func (boolean) Decode(start uint, reader interface{}) (bool, error) {
	return false, nil
}

func (boolean) Element(start uint, writer interface{}, key string, b bool) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x08')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Boolean.Encode(start, writer, b)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (datetime) Encode(start uint, writer interface{}, dt int64) (int, error) {
	return Int64.Encode(start, writer, dt)
}

func (datetime) Decode(start uint, reader interface{}) (int64, error) {
	return 0, nil
}

func (datetime) Element(start uint, writer interface{}, key string, dt int64) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x09')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = DateTime.Encode(start, writer, dt)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (regex) Encode(start uint, writer interface{}, pattern, options string) (int, error) {
	var total int

	written, err := CString.Encode(start, writer, pattern)
	total += written
	if err != nil {
		return total, err
	}

	written, err = CString.Encode(start+uint(total), writer, options)
	total += written

	return total, err
}

func (regex) Decode(start uint, reader interface{}) (string, string, error) {
	return "", "", nil
}

func (regex) Element(start uint, writer interface{}, key string, pattern, options string) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x0B')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, pattern)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, options)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (dbpointer) Encode(start uint, writer interface{}, ns string, oid [12]byte) (int, error) {
	var total int

	written, err := String.Encode(start, writer, ns)
	total += written
	if err != nil {
		return total, err
	}

	written, err = ObjectId.Encode(start+uint(written), writer, oid)
	total += written

	return total, err
}

func (dbpointer) Decode(start uint, reader interface{}) (string, [12]byte, error) {
	var ns string
	var oid [12]byte
	return ns, oid, nil
}

func (dbpointer) Element(start uint, writer interface{}, key string, ns string, oid [12]byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x0C')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = DBPointer.Encode(start, writer, ns, oid)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil

}

func (javascript) Encode(start uint, writer interface{}, code string) (int, error) {
	return String.Encode(start, writer, code)
}

func (javascript) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (javascript) Element(start uint, writer interface{}, key string, code string) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x0D')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Javascript.Encode(start, writer, code)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (symbol) Encode(start uint, writer interface{}, symbol string) (int, error) {
	return String.Encode(start, writer, symbol)
}

func (symbol) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (symbol) Element(start uint, writer interface{}, key string, symbol string) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x0E')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Symbol.Encode(start, writer, symbol)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (codewithscope) Encode(start uint, writer interface{}, code string, doc []byte) (int, error) {
	var total int

	// Length of CodeWithScope is 4 + 4 + len(code) + 1 + len(doc)
	n, err := Int32.Encode(start, writer, 9+int32(len(code))+int32(len(doc)))
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = String.Encode(start, writer, code)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = encodeByteSlice(start, writer, doc)
	total += n

	return total, err
}

func (codewithscope) Decode(start uint, reader interface{}) (string, []byte, error) {
	return "", nil, nil
}

func (codewithscope) Element(start uint, writer interface{}, key string, code string, scope []byte) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x0F')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CodeWithScope.Encode(start, writer, code, scope)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (i32) Encode(start uint, writer interface{}, i int32) (int, error) {
	u := signed32ToUnsigned(i)

	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start)+4 {
			return 0, ErrTooSmall
		}
		binary.LittleEndian.PutUint32(w[start:start+4], u)
		return 4, nil
	case io.WriterAt:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		return w.WriteAt(b[:], int64(start))
	case io.WriteSeeker:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		_, err := w.Seek(int64(start), io.SeekStart)
		if err != nil {
			return 0, err
		}
		return w.Write(b[:])
	case io.Writer:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		return w.Write(b[:])
	default:
		return 0, ErrInvalidWriter
	}
}

func (i32) Decode(start uint, reader interface{}) (int32, error) {
	return 0, nil
}

func (i32) Element(start uint, writer interface{}, key string, i int32) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x10')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Int32.Encode(start, writer, i)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (timestamp) Encode(start uint, writer interface{}, t uint32, i uint32) (int, error) {
	var total int

	n, err := encodeUint32(start, writer, i)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = encodeUint32(start, writer, t)
	start += uint(n)
	total += n

	return total, err
}

func (timestamp) Decode(start uint, reader interface{}) (uint32, uint32, error) {
	return 0, 0, nil
}

func (timestamp) Element(start uint, writer interface{}, key string, t uint32, i uint32) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x11')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Timestamp.Encode(start, writer, t, i)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (i64) Encode(start uint, writer interface{}, i int64) (int, error) {
	u := signed64ToUnsigned(i)

	return encodeUint64(start, writer, u)
}

func (i64) Decode(start uint, reader interface{}) (int64, error) {
	return 0, nil
}

func (i64) Element(start uint, writer interface{}, key string, i int64) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x12')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Int64.Encode(start, writer, i)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (decimal128) Encode(start uint, writer interface{}, d ast.Decimal128) (int, error) {
	var total int
	high, low := d.GetBytes()

	written, err := encodeUint64(start, writer, low)
	total += written
	if err != nil {
		return total, err
	}

	written, err = encodeUint64(start+uint(total), writer, high)
	total += written

	return total, err
}

func (decimal128) Decode(start uint, reader interface{}) (ast.Decimal128, error) {
	return ast.Decimal128{}, nil
}

func (decimal128) Element(start uint, writer interface{}, key string, d ast.Decimal128) (int, error) {
	var total int

	n, err := Byte.Encode(start, writer, '\x13')
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = CString.Encode(start, writer, key)
	start += uint(n)
	total += n
	if err != nil {
		return total, err
	}

	n, err = Decimal128.Encode(start, writer, d)
	total += n
	if err != nil {
		return total, err
	}

	return total, nil
}

func (cstring) Encode(start uint, writer interface{}, str string) (int, error) {
	var written int
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+1)+len(str) {
			return 0, ErrTooSmall
		}
		end := int(start) + len(str)
		written += copy(w[start:end], str)
		w[end] = '\x00'
		written += 1
	default:
		return 0, ErrInvalidWriter
	}
	return written, nil
}

func (cstring) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (bsonbyte) Encode(start uint, writer interface{}, t byte) (int, error) {
	var written int
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+1) {
			return 0, ErrTooSmall
		}
		w[start] = t
		written = 1
	default:
		return 0, ErrInvalidWriter
	}
	return written, nil
}

func encodeByteSlice(start uint, writer interface{}, b []byte) (int, error) {
	var total int

	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start)+len(b) {
			return 0, ErrTooSmall
		}

		total += copy(w[start:], b)

	default:
		return 0, ErrInvalidWriter
	}

	return total, nil
}

func encodeUint32(start uint, writer interface{}, u uint32) (int, error) {
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+4) {
			return 0, ErrTooSmall
		}
		binary.LittleEndian.PutUint32(w[start:], u)
		return 4, nil
	case io.WriterAt:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		return w.WriteAt(b[:], int64(start))
	case io.WriteSeeker:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		_, err := w.Seek(int64(start), io.SeekStart)
		if err != nil {
			return 0, err
		}
		return w.Write(b[:])
	case io.Writer:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], u)
		return w.Write(b[:])
	default:
		return 0, ErrInvalidWriter
	}
}

func encodeUint64(start uint, writer interface{}, u uint64) (int, error) {
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+8) {
			return 0, ErrTooSmall
		}
		binary.LittleEndian.PutUint64(w[start:], u)
		return 8, nil
	case io.WriterAt:
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], u)
		return w.WriteAt(b[:], int64(start))
	case io.WriteSeeker:
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], u)
		_, err := w.Seek(int64(start), io.SeekStart)
		if err != nil {
			return 0, err
		}
		return w.Write(b[:])
	case io.Writer:
		var b [8]byte
		binary.LittleEndian.PutUint64(b[:], u)
		return w.Write(b[:])
	default:
		return 0, ErrInvalidWriter
	}
}

func signed32ToUnsigned(i int32) uint32 {
	return *(*uint32)(unsafe.Pointer(&i))
}

func signed64ToUnsigned(i int64) uint64 {
	return *(*uint64)(unsafe.Pointer(&i))
}
