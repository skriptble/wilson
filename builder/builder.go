package builder

import (
	"errors"

	"github.com/skriptble/wilson/elements"
	"github.com/skriptble/wilson/parser/ast"
)

var ErrTooShort = errors.New("builder: The provided slice's length is too short")
var ErrInvalidWriter = errors.New("builder: Invalid writer provided")

var C Constructor
var AC ArrayConstructor

type Constructor struct{}
type ArrayConstructor struct{}

// Elementer is the interface implemented by types that can serialize
// themselves into a BSON element.
type Elementer interface {
	Element() (ElementSizer, ElementWriter)
}

type ElementFunc func() (ElementSizer, ElementWriter)

func (ef ElementFunc) Element() (ElementSizer, ElementWriter) {
	return ef()
}

// Element is a function type used to insert BSON element values into a BSON
// document using a DocumentBuilder.
// type Element func() (length uint, ew ElementWriter)

// ElementWriter handles writing an element's BSON representation to a writer.
//
// writer can be:
//
// - []byte
// - io.WriterAt
// - io.WriteSeeker
// - io.Writer
//
// If it is not one of these values, the implementations should panic.
type ElementWriter func(start uint, writer interface{}) (n int, err error)

// ElementSizer handles retrieving the size of an element's BSON representation.
type ElementSizer func() (size uint)

// DocumentBuilder allows the creation of a BSON document by appending elements
// and then writing the document. The document can be written multiple times so
// appending then writing and then appending and writing again is a valid usage
// pattern.
type DocumentBuilder struct {
	funcs       []ElementWriter
	sizers      []ElementSizer
	starts      []uint
	required    uint // number of required bytes. Should start at 4
	initialized bool
}

func NewDocumentBuilder() *DocumentBuilder {
	var b DocumentBuilder
	b.Init()

	return &b
}

func (db *DocumentBuilder) Init() {
	if db.initialized {
		return
	}
	sizer, f := db.documentHeader()
	db.funcs = append(db.funcs, f)
	db.sizers = append(db.sizers, sizer)
	db.initialized = true
}

// Append adds the given elements to the BSON document.
func (db *DocumentBuilder) Append(elems ...Elementer) *DocumentBuilder {
	db.Init()
	for _, elem := range elems {
		sizer, f := elem.Element()
		db.funcs = append(db.funcs, f)
		db.sizers = append(db.sizers, sizer)
	}
	return db
}

func (db *DocumentBuilder) documentHeader() (ElementSizer, ElementWriter) {
	return func() uint { return 4 },
		func(start uint, writer interface{}) (n int, err error) {
			return elements.Int32.Encode(start, writer, int32(db.RequiredBytes()))
		}
}

func (db *DocumentBuilder) calculateStarts() {
	// TODO(skriptble): This method should cache its results and Append should
	// invalidate the cache.
	db.required = 0
	db.starts = db.starts[:0]
	for _, sizer := range db.sizers {
		db.starts = append(db.starts, db.required)
		db.required += sizer()
	}
}

// RequireBytes returns the number of bytes required to write the entire BSON
// document.
func (db *DocumentBuilder) RequiredBytes() uint {
	db.calculateStarts()
	if db.required < 5 {
		return 5
	}
	return db.required + 1 // We add 1 because we don't include the ending null byte for the document

}

func (db *DocumentBuilder) WriteDocument(writer interface{}) (int64, error) {
	db.Init()
	db.calculateStarts()

	var total, n int
	var start uint
	var err error

	if b, ok := writer.([]byte); ok {
		if uint(len(b)) < db.required+1 {
			return 0, ErrTooShort
		}
	}

	n, err = db.writeElements(start, writer)
	start += uint(n)
	total += n
	if err != nil {
		return int64(n), err
	}

	n, err = elements.Byte.Encode(start, writer, '\x00')
	total += n
	return int64(total), err
}

func (db *DocumentBuilder) writeElements(start uint, writer interface{}) (total int, err error) {
	for idx := range db.funcs {
		n, err := db.funcs[idx](uint(db.starts[idx])+start, writer)
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func (Constructor) SubDocument(key string, subdoc *DocumentBuilder) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A subdocument will always take (1 + key length + 1) + len(subdoc) bytes
		return func() uint {
				return 2 + uint(len(key)) + subdoc.RequiredBytes()
			},
			func(start uint, writer interface{}) (int, error) {
				subdocBytes := make([]byte, subdoc.RequiredBytes())
				_, err := subdoc.WriteDocument(subdocBytes)
				if err != nil {
					return 0, err
				}

				return elements.Document.Element(start, writer, key, subdocBytes)
			}
	}
}

func (c Constructor) SubDocumentWithElements(key string, elems ...Elementer) ElementFunc {
	var b DocumentBuilder
	b.Init()
	b.Append(elems...)

	return c.SubDocument(key, &b)
}

func (c Constructor) Array(key string, array *ArrayBuilder) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A subdocument will always take (1 + key length + 1) + len(subdoc) bytes
		return func() uint {
				return 2 + uint(len(key)) + array.RequiredBytes()
			},
			func(start uint, writer interface{}) (int, error) {
				arrayBytes := make([]byte, array.RequiredBytes())
				_, err := array.WriteDocument(arrayBytes)
				if err != nil {
					return 0, err
				}

				return elements.Array.Element(start, writer, key, arrayBytes)
			}
	}
}

func (c Constructor) ArrayWithElements(key string, elems ...ArrayElementer) ElementFunc {
	var b ArrayBuilder
	b.Init()
	b.Append(elems...)

	return C.Array(key, &b)
}

func (Constructor) Double(key string, f float64) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A double will always take (1 + key length + 1) + 8 bytes
		return func() uint {
				return uint(10 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Double.Element(start, writer, key, f)
			}
	}
}

func (Constructor) String(key string, value string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A string's length is (1 + key length + 1) + (4 + value length + 1)
		return func() uint {
				return uint(7 + len(key) + len(value))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.String.Element(start, writer, key, value)
			}
	}
}

func (c Constructor) Binary(key string, b []byte) ElementFunc {
	return c.BinaryWithSubtype(key, b, 0)
}

func (Constructor) BinaryWithSubtype(key string, b []byte, btype byte) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A binary of subtype 2 has length (1 + key length + 1) + (4 + 1 + 4 + b length)
		// All other binary subtypes have length  (1 + key length + 1) + (4 + 1 + b length)
		return func() uint {
				//
				if btype == 2 {
					return uint(11 + len(key) + len(b))
				}

				return uint(7 + len(key) + len(b))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Binary.Element(start, writer, key, b, btype)
			}
	}
}

func (Constructor) Undefined(key string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// Undefined's length is 1 + key length + 1
		return func() uint {
				return uint(2 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				var total int

				n, err := elements.Byte.Encode(start, writer, '\x06')
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				n, err = elements.CString.Encode(start, writer, key)
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				return total, nil
			}
	}
}

func (Constructor) ObjectId(key string, oid [12]byte) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An ObjectId's length is (1 + key length + 1) + 12
		return func() uint {
				return uint(14 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.ObjectId.Element(start, writer, key, oid)
			}
	}
}

func (Constructor) Boolean(key string, b bool) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An ObjectId's length is (1 + key length + 1) + 1
		return func() uint {
				return uint(3 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Boolean.Element(start, writer, key, b)
			}
	}
}

func (Constructor) DateTime(key string, dt int64) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// Datetime's length is (1 + key length + 1) + 8
		return func() uint {
				return uint(10 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.DateTime.Element(start, writer, key, dt)
			}
	}
}

func (Constructor) Null(key string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// Null's length is 1 + key length + 1
		return func() uint {
				return uint(2 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				var total int

				n, err := elements.Byte.Encode(start, writer, '\x0A')
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				n, err = elements.CString.Encode(start, writer, key)
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				return total, nil
			}
	}
}

func (Constructor) Regex(key string, pattern, options string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// Null's length is (1 + key length + 1) + (pattern length + 1) + (options length + 1)
		return func() uint {
				return uint(4 + len(key) + len(pattern) + len(options))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Regex.Element(start, writer, key, pattern, options)
			}
	}
}

func (Constructor) DBPointer(key string, ns string, oid [12]byte) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An dbpointer's length is (1 + key length + 1) + (4 + ns length + 1) + 12
		return func() uint {
				return uint(19 + len(key) + len(ns))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.DBPointer.Element(start, writer, key, ns, oid)
			}
	}
}

func (Constructor) JavaScriptCode(key string, code string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// JavaScript code's length is (1 + key length + 1) + (4 + code length + 1)
		return func() uint {
				return uint(7 + len(key) + len(code))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Javascript.Element(start, writer, key, code)
			}
	}
}

func (Constructor) Symbol(key string, symbol string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A symbol's length is (1 + key length + 1) + (4 + symbol length + 1)
		return func() uint {
				return uint(7 + len(key) + len(symbol))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Symbol.Element(start, writer, key, symbol)
			}
	}
}

func (Constructor) CodeWithScope(key string, code string, scope []byte) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// JavaScript code with scope's length is (1 + key length + 1) + 4 +  (4 + len key + 1) + len(scope)
		return func() uint {
				return uint(11 + len(key) + len(code) + len(scope))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.CodeWithScope.Element(start, writer, key, code, scope)
			}
	}
}

func (Constructor) Int32(key string, i int32) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An int32's length is (1 + key length + 1) + 4 bytes
		return func() uint {
				return uint(6 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Int32.Element(start, writer, key, i)
			}
	}
}

func (Constructor) Timestamp(key string, t uint32, i uint32) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An decimal's length is (1 + key length + 1) + 8 bytes
		return func() uint {
				return uint(10 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Timestamp.Element(start, writer, key, t, i)
			}
	}
}

func (Constructor) Int64(key string, i int64) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An int64's length is (1 + key length + 1) + 8 bytes
		return func() uint {
				return uint(10 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Int64.Element(start, writer, key, i)
			}
	}
}

func (Constructor) Decimal(key string, d ast.Decimal128) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An decimal's length is (1 + key length + 1) + 16 bytes
		return func() uint {
				return uint(18 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				return elements.Decimal128.Element(start, writer, key, d)
			}
	}
}

func (Constructor) MinKey(key string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An min key's length is (1 + key length + 1)
		return func() uint {
				return uint(2 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				var total int

				n, err := elements.Byte.Encode(start, writer, '\xFF')
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				n, err = elements.CString.Encode(start, writer, key)
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				return total, nil
			}
	}
}

func (Constructor) MaxKey(key string) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// An max key's length is (1 + key length + 1)
		return func() uint {
				return uint(2 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				var total int

				n, err := elements.Byte.Encode(start, writer, '\x7F')
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				n, err = elements.CString.Encode(start, writer, key)
				start += uint(n)
				total += n
				if err != nil {
					return total, err
				}

				return total, nil
			}
	}
}
