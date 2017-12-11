package bson

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
)

func TestDocument(t *testing.T) {
	t.Run("NewDocument", func(t *testing.T) {
		t.Run("TooShort", func(t *testing.T) {
			want := ErrTooSmall
			_, got := ReadDocument([]byte{'\x00', '\x00'})
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("InvalidLength", func(t *testing.T) {
			want := ErrInvalidLength
			b := make([]byte, 5)
			binary.LittleEndian.PutUint32(b[0:4], 200)
			_, got := ReadDocument(b)
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("keyLength-error", func(t *testing.T) {
			want := ErrInvalidKey
			b := make([]byte, 8)
			binary.LittleEndian.PutUint32(b[0:4], 8)
			b[4], b[5], b[6], b[7] = '\x02', 'f', 'o', 'o'
			_, got := ReadDocument(b)
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("Missing-Null-Terminator", func(t *testing.T) {
			want := ErrInvalidReadOnlyDocument
			b := make([]byte, 9)
			binary.LittleEndian.PutUint32(b[0:4], 9)
			b[4], b[5], b[6], b[7], b[8] = '\x0A', 'f', 'o', 'o', '\x00'
			_, got := ReadDocument(b)
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		t.Run("validateValue-error", func(t *testing.T) {
			want := ErrTooSmall
			b := make([]byte, 11)
			binary.LittleEndian.PutUint32(b[0:4], 11)
			b[4], b[5], b[6], b[7], b[8], b[9], b[10] = '\x01', 'f', 'o', 'o', '\x00', '\x01', '\x02'
			_, got := ReadDocument(b)
			if got != want {
				t.Errorf("Did not get expected error. got %v; want %v", got, want)
			}
		})
		testCases := []struct {
			name string
		}{}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {})
		}
	})
	t.Run("Walk", func(t *testing.T) {})
	t.Run("Keys", func(t *testing.T) {})
	t.Run("Append", func(t *testing.T) {
		testCases := []struct {
			name  string
			elems [][]*Element
			want  []byte
		}{
			{"one-one", tpag.oneOne(), tpag.oneOneAppendBytes()},
			{"two-one", tpag.twoOne(), tpag.twoOneAppendBytes()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDocument()
				for _, elems := range tc.elems {
					d.Append(elems...)
				}
				got, err := d.MarshalBSON()
				if err != nil {
					t.Errorf("Received an unexpected error while marhsaling BSON: %s", err)
				}
				if !bytes.Equal(got, tc.want) {
					t.Errorf("Output from Append is not correct. got %#v; want %v", got, tc.want)
				}
			})
		}
	})
	t.Run("Prepend", func(t *testing.T) {
		testCases := []struct {
			name  string
			elems [][]*Element
			want  []byte
		}{
			{"one-one", tpag.oneOne(), tpag.oneOnePrependBytes()},
			{"two-one", tpag.twoOne(), tpag.twoOnePrependBytes()},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				d := NewDocument()
				for _, elems := range tc.elems {
					d.Prepend(elems...)
				}
				got, err := d.MarshalBSON()
				if err != nil {
					t.Errorf("Received an unexpected error while marhsaling BSON: %s", err)
				}
				if !bytes.Equal(got, tc.want) {
					t.Errorf("Output from Prepend is not correct. got %#v; want %v", got, tc.want)
				}
			})
		}
	})
	t.Run("Lookup", func(t *testing.T) {
		testCases := []struct {
			name string
			doc  *Document
			key  []string
			want *Element
		}{}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {})
		}
	})
	t.Run("Delete", func(t *testing.T) {
		testCases := []struct {
			name string
			doc  *Document
			key  []string
			want []byte
		}{}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {})
		}
	})
	t.Run("Update", func(t *testing.T) {})
	t.Run("Err", func(t *testing.T) {})
	t.Run("ElementAt", func(t *testing.T) {})
	t.Run("Iterator", func(t *testing.T) {})
	t.Run("Combine", func(t *testing.T) {})
}

var tpag testPrependAppendGenerator

type testPrependAppendGenerator struct{}

func (testPrependAppendGenerator) oneOne() [][]*Element {
	return [][]*Element{
		[]*Element{C.Double("foobar", 3.14159)},
	}
}

func (testPrependAppendGenerator) oneOneAppendBytes() []byte {
	return []byte{
		// size
		0x15, 0x0, 0x0, 0x0,
		// type
		0x1,
		// key
		0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72, 0x0,
		// value
		0x6e, 0x86, 0x1b, 0xf0, 0xf9, 0x21, 0x9, 0x40,
		// null terminator
		0x0,
	}
}

func (testPrependAppendGenerator) oneOnePrependBytes() []byte {
	return []byte{
		// size
		0x15, 0x0, 0x0, 0x0,
		// type
		0x1,
		// key
		0x66, 0x6f, 0x6f, 0x62, 0x61, 0x72, 0x0,
		// value
		0x6e, 0x86, 0x1b, 0xf0, 0xf9, 0x21, 0x9, 0x40,
		// null terminator
		0x0,
	}
}

func (testPrependAppendGenerator) twoOne() [][]*Element {
	return [][]*Element{
		[]*Element{C.Double("foo", 1.234)},
		[]*Element{C.Double("foo", 5.678)},
	}
}

func (testPrependAppendGenerator) twoOneAppendBytes() []byte {
	return []byte{
		// size
		0x1f, 0x0, 0x0, 0x0,
		//type - key - value
		0x1, 0x66, 0x6f, 0x6f, 0x0, 0x58, 0x39, 0xb4, 0xc8, 0x76, 0xbe, 0xf3, 0x3f,
		// type - key - value
		0x1, 0x66, 0x6f, 0x6f, 0x0, 0x83, 0xc0, 0xca, 0xa1, 0x45, 0xb6, 0x16, 0x40,
		// null terminator
		0x0,
	}
}

func (testPrependAppendGenerator) twoOnePrependBytes() []byte {
	return []byte{
		// size
		0x1f, 0x0, 0x0, 0x0,
		// type - key - value
		0x1, 0x66, 0x6f, 0x6f, 0x0, 0x83, 0xc0, 0xca, 0xa1, 0x45, 0xb6, 0x16, 0x40,
		//type - key - value
		0x1, 0x66, 0x6f, 0x6f, 0x0, 0x58, 0x39, 0xb4, 0xc8, 0x76, 0xbe, 0xf3, 0x3f,
		// null terminator
		0x0,
	}
}

func ExampleDocument_ClientDoc() {
	internalVersion := "1234567"

	/*
		func createClientDoc(appName string) bson.M {
			clientDoc := bson.M{
				"driver": bson.M{
					"name":    "mongo-go-driver",
					"version": internal.Version,
				},
				"os": bson.M{
					"type":         runtime.GOOS,
					"architecture": runtime.GOARCH,
				},
				"platform": runtime.Version(),
			}
			if appName != "" {
				clientDoc["application"] = bson.M{"name": appName}
			}

			return clientDoc
		}
	*/

	f := func(appName string) *Document {
		doc := NewDocument()
		doc.Append(
			C.SubDocumentFromElements("driver",
				C.String("name", "mongo-go-driver"),
				C.String("version", internalVersion),
			),
			C.SubDocumentFromElements("os",
				C.String("type", runtime.GOOS),
				C.String("architecture", runtime.GOARCH),
			),
			C.String("platform", runtime.Version()),
		)
		if appName != "" {
			doc.Append(C.SubDocumentFromElements("application", C.String("name", appName)))
		}

		return doc
	}
	f("hello-world").MarshalBSON()
}

func BenchmarkDocument(b *testing.B) {
	b.ReportAllocs()
	internalVersion := "1234567"
	for i := 0; i < b.N; i++ {
		doc := NewDocument()
		doc.Append(
			C.SubDocumentFromElements("driver",
				C.String("name", "mongo-go-driver"),
				C.String("version", internalVersion),
			),
			C.SubDocumentFromElements("os",
				C.String("type", runtime.GOOS),
				C.String("architecture", runtime.GOARCH),
			),
			C.String("platform", runtime.Version()),
		)
		doc.MarshalBSON()
	}
}
