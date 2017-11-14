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
var ObjectID objectid
var Boolean boolean
var DateTime datetime
var Regex regex
var DBPointer dbpointer
var Javascript javascript
var Symbol symbol
var CodeWithScope codewithscope
var Int32 i32
var Uint64 u64
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
type u64 struct{}
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

func (str) Encode(start uint, writer interface{}, s string) error {
	return nil
}

func (str) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (document) Encode(start uint, writer interface{}, doc []byte) error {
	return nil
}

func (document) Decode(start uint, reader interface{}) ([]byte, error) {
	return nil, nil
}

func (array) Encode(start uint, writer interface{}, arr []byte) error {
	return nil
}

func (array) Decode(start uint, reader interface{}) ([]byte, error) {
	return nil, nil
}

func (bin) Encode(start uint, writer interface{}, b []byte, btype uint) error {
	return nil
}

func (bin) Decode(start uint, reader interface{}) (b []byte, btype uint, err error) {
	return nil, 0, nil
}

func (objectid) Encode(start uint, writer interface{}, obj [12]byte) error {
	return nil
}

func (objectid) Decode(start uint, reader interface{}) ([12]byte, error) {
	var obj [12]byte
	return obj, nil
}

func (boolean) Encode(start uint, writer interface{}, b bool) error {
	return nil
}

func (boolean) Decode(start uint, reader interface{}) (bool, error) {
	return false, nil
}

func (datetime) Encode(start uint, writer interface{}, dt int64) error {
	return nil
}

func (datetime) Decode(start uint, reader interface{}) (int64, error) {
	return 0, nil
}

func (regex) Encode(start uint, writer interface{}, pattern, options string) error {
	return nil
}

func (regex) Decode(start uint, reader interface{}) (string, string, error) {
	return "", "", nil
}

func (dbpointer) Encode(start uint, writer interface{}, dbpointer [12]byte) error {
	return nil
}

func (dbpointer) Decode(start uint, reader interface{}) ([12]byte, error) {
	var dbpointer [12]byte
	return dbpointer, nil
}

func (javascript) Encode(start uint, writer interface{}, js string) error {
	return nil
}

func (javascript) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (symbol) Encode(start uint, writer interface{}, symbol string) error {
	return nil
}

func (symbol) Decode(start uint, reader interface{}) (string, error) {
	return "", nil
}

func (codewithscope) Encode(start uint, writer interface{}, js string, doc []byte) error {
	return nil
}

func (codewithscope) Decode(start uint, reader interface{}) (string, []byte, error) {
	return "", nil, nil
}

func (i32) Encode(start uint, writer interface{}, i int32) (int, error) {
	switch w := writer.(type) {
	case []byte:
		if len(w) < int(start+4) {
			return 0, ErrTooSmall
		}
		binary.LittleEndian.PutUint32(w[start:start+4], uint32(i))
		return 4, nil
	case io.WriterAt:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(i))
		return w.WriteAt(b[:], int64(start))
	case io.WriteSeeker:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(i))
		_, err := w.Seek(int64(start), io.SeekStart)
		if err != nil {
			return 0, err
		}
		return w.Write(b[:])
	case io.Writer:
		var b [4]byte
		binary.LittleEndian.PutUint32(b[:], uint32(i))
		return w.Write(b[:])
	default:
		return 0, ErrInvalidWriter
	}
}

func (i32) Deocde(start uint, reader interface{}) (int32, error) {
	return 0, nil
}

func (u64) Encode(start uint, writer interface{}, u uint64) error {
	return nil
}

func (u64) Decode(start uint, reader interface{}) (uint64, error) {
	return 0, nil
}

func (i64) Encode(start uint, writer interface{}, i int64) error {
	return nil
}

func (i64) Decode(start uint, reader interface{}) (int64, error) {
	return 0, nil
}

func (decimal128) Encode(start uint, writer interface{}, d ast.Decimal128) error {
	return nil
}

func (decimal128) Decode(start uint, reader interface{}) (ast.Decimal128, error) {
	return ast.Decimal128{}, nil
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
