package builder

import (
	"errors"

	"github.com/skriptble/wilson/elements"
)

var ErrTooShort = errors.New("builder: The provided slice's length is too short")
var ErrInvalidWriter = errors.New("builder: Invalid writer provided")

var C Constructor

type Constructor struct{}

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
	Key         string
	funcs       []ElementWriter
	sizers      []ElementSizer
	starts      []uint
	required    uint // number of required bytes. Should start at 4
	initialized bool
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
	// TODO(skriptble): This method should cache it's results and Append should
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
	return db.requiredSize(false)
}

func (db *DocumentBuilder) embeddedSize() uint {
	return db.requiredSize(true)
}

func (db *DocumentBuilder) requiredSize(embedded bool) uint {
	db.calculateStarts()
	if db.required < 5 {
		return 5
	}
	if embedded {
		return db.required + 3 + uint(len(db.Key)) // 1 byte for type, 1 byte for '\x00', 1 byte for doc '\x00'
	}
	return db.required + 1 // We add 1 because we don't include the ending null byte for the document

}

func (db *DocumentBuilder) WriteDocument(writer interface{}) (int64, error) {
	n, err := db.writeDocument(0, writer, false)
	return int64(n), err
}

func (db *DocumentBuilder) writeDocument(start uint, writer interface{}, embedded bool) (int, error) {
	db.Init()
	db.calculateStarts()

	var total, n int
	var err error

	if embedded {
		n, err = elements.Byte.Encode(start, writer, '\x02')
		start += uint(n)
		total += n
		if err != nil {
			return total, err
		}
		n, err = elements.CString.Encode(start, writer, db.Key)
		start += uint(n)
		total += n
		if err != nil {
			return total, err
		}
	}
	if b, ok := writer.([]byte); ok {
		if uint(len(b)) < start+db.required+1 {
			return 0, ErrTooShort
		}
	}

	n, err = db.writeElements(start, writer)
	start += uint(n)
	total += n
	if err != nil {
		return n, err
	}

	n, err = elements.Byte.Encode(start, writer, '\x00')
	total += n
	return total, err
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

func (db *DocumentBuilder) Element() (ElementSizer, ElementWriter) {
	return db.embeddedSize, func(start uint, writer interface{}) (n int, err error) {
		return db.writeDocument(start, writer, true)
	}
}

func (Constructor) SubDocument(key string, elems ...Elementer) *DocumentBuilder {
	return (&DocumentBuilder{Key: key}).Append(elems...)
}

func (Constructor) Double(key string, f float64) ElementFunc {
	return func() (ElementSizer, ElementWriter) {
		// A double will always take 1 + key length + 1 + 8 bytes
		return func() uint {
				return uint(10 + len(key))
			},
			func(start uint, writer interface{}) (int, error) {
				var total int
				n, err := elements.Byte.Encode(start, writer, '\x01')
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
				n, err = elements.Double.Encode(start, writer, f)
				total += n
				if err != nil {
					return total, err
				}
				return total, nil
			}
	}
}
