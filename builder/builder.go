package builder

import (
	"encoding/binary"
	"errors"
	"io"
)

var ErrTooShort = errors.New("builder: The provided slice's length is too short")
var ErrInvalidWriter = errors.New("builder: Invalid writer provided")

// Element is a function type used to insert BSON element values into a BSON
// document using a DocumentBuilder.
type Element func() (length uint, ep ElementWriter)

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

// DocumentBuilder allows the creation of a BSON document by appending elements
// and then writing the document. The document can be written multiple times so
// appending then writing and then appending and writing again is a valid usage
// pattern.
type DocumentBuilder struct {
	funcs       []ElementWriter
	starts      []uint
	required    uint // number of required bytes. Should start at 4
	initialized bool
}

func NewDocumentBuilder() *DocumentBuilder {
	return new(DocumentBuilder)
}

func (db *DocumentBuilder) Init() {
	if db.initialized {
		return
	}
	l, f := db.documentHeader()
	db.funcs = append(db.funcs, f)
	db.starts = append(db.starts, db.required)
	db.required += l

}

// Append adds the given elements to the BSON document.
func (db *DocumentBuilder) Append(elems ...Element) {
	db.Init()
	for _, elem := range elems {
		length, f := elem()
		db.funcs = append(db.funcs, f)
		db.starts = append(db.starts, db.required)
		db.required += length
	}
}

func (db *DocumentBuilder) documentHeader() (length uint, ep ElementWriter) {
	return 4, func(start uint, writer interface{}) (n int, err error) {
		switch w := writer.(type) {
		case []byte:
			if len(w) < int(start+4) {
				return 0, ErrTooShort
			}
			binary.LittleEndian.PutUint32(w[start:start+4], uint32(db.required+1))
			return 4, nil
		case io.WriterAt:
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(db.required+1))
			return w.WriteAt(b[:], int64(start))
		case io.WriteSeeker:
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(db.required+1))
			_, err = w.Seek(int64(start), io.SeekStart)
			if err != nil {
				return 0, err
			}
			return w.Write(b[:])
		case io.Writer:
			var b [4]byte
			binary.LittleEndian.PutUint32(b[:], uint32(db.required+1))
			return w.Write(b[:])
		default:
			return 0, ErrInvalidWriter
		}
	}
}

// RequireBytes returns the number of bytes required to write the entire BSON
// document.
func (db *DocumentBuilder) RequiredBytes() uint {
	if db.required < 5 {
		return 5
	}
	return db.required + 1 // We add 1 because we don't include the ending null byte for the document
}

func (db *DocumentBuilder) WriteDocument(writer interface{}) (total int64, err error) {
	db.Init()
	if b, ok := writer.([]byte); ok {
		if uint(len(b)) < db.required+1 {
			return 0, ErrTooShort
		}
	}
	n, err := db.writeElements(writer)
	if err != nil {
		return n, err
	}

	switch w := writer.(type) {
	case []byte:
		w[len(w)-1] = '\x00'
	case io.WriterAt:
		_, err = w.WriteAt([]byte{'\x00'}, int64(db.required-1))
	case io.WriteSeeker:
		_, err = w.Seek(int64(db.required-1), io.SeekStart)
		if err != nil {
			return n, err
		}
		_, err = w.Write([]byte{'\x00'})
	case io.Writer:
		_, err = w.Write([]byte{'\x00'})
	default:
		return 0, ErrInvalidWriter
	}

	return n + 1, nil
}

func (db *DocumentBuilder) writeElements(writer interface{}) (total int64, err error) {
	for idx := range db.funcs {
		n, err := db.funcs[idx](uint(db.starts[idx]), writer)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
