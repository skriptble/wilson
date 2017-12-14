package bson

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
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
	t.Run("Keys", testDocumentKeys)
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
			d    *Document
			key  []string
			want *Element
			err  error
		}{
			{"first", (&Document{}).Append(C.Null("x")), []string{"x"},
				&Element{ReaderElement: ReaderElement{start: 0, value: 3}}, nil},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.d.Lookup(tc.key...)
				if err != tc.err {
					t.Errorf("Returned error does not match. got %v; want %v", err, tc.err)
				}
				if !elementEqual(got, tc.want) {
					t.Errorf("Returned element does not match expected element. got %v; want %v", got, tc.want)
				}
			})
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

func testDocumentKeys(t *testing.T) {
	testCases := []struct {
		name      string
		d         *Document
		want      Keys
		err       error
		recursive bool
	}{
		{"one", (&Document{}).Append(C.String("foo", "")), Keys{{Name: "foo"}}, nil, false},
		{"two", (&Document{}).Append(C.Null("x"), C.Null("y")), Keys{{Name: "x"}, {Name: "y"}}, nil, false},
		{"one-flat", (&Document{}).Append(C.SubDocumentFromElements("foo", C.Null("a"), C.Null("b"))),
			Keys{{Name: "foo"}}, nil, false,
		},
		{"one-recursive", (&Document{}).Append(C.SubDocumentFromElements("foo", C.Null("a"), C.Null("b"))),
			Keys{{Name: "foo"}, {Prefix: []string{"foo"}, Name: "a"}, {Prefix: []string{"foo"}, Name: "b"}}, nil, true,
		},
		// {"one-array-recursive", (&Document{}).Append(c.ArrayFromElements("foo", AC.Null(())),
		// 	Keys{{Name: "foo"}, {Prefix: []string{"foo"}, Name: "1"}, {Prefix: []string{"foo"}, Name: "2"}}, nil, true,
		// },
		// {"invalid-subdocument",
		// 	Reader{
		// 		'\x15', '\x00', '\x00', '\x00',
		// 		'\x03',
		// 		'f', 'o', 'o', '\x00',
		// 		'\x0B', '\x00', '\x00', '\x00', '\x01', '1', '\x00',
		// 		'\x0A', '2', '\x00', '\x00', '\x00',
		// 	},
		// 	nil, ErrTooSmall, true,
		// },
		// {"invalid-array",
		// 	Reader{
		// 		'\x15', '\x00', '\x00', '\x00',
		// 		'\x04',
		// 		'f', 'o', 'o', '\x00',
		// 		'\x0B', '\x00', '\x00', '\x00', '\x01', '1', '\x00',
		// 		'\x0A', '2', '\x00', '\x00', '\x00',
		// 	},
		// 	nil, ErrTooSmall, true,
		// },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.d.Keys(tc.recursive)
			if err != tc.err {
				t.Errorf("Returned error does not match. got %v; want %v", err, tc.err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Returned keys do not match expected keys. got %v; want %v", got, tc.want)
			}
		})
	}
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

func ExampleDocument() {
	internalVersion := "1234567"

	f := func(appName string) *Document {
		doc := NewDocument()
		doc.Append(
			C.SubDocumentFromElements("driver",
				C.String("name", "mongo-go-driver"),
				C.String("version", internalVersion),
			),
			C.SubDocumentFromElements("os",
				C.String("type", "darwin"),
				C.String("architecture", "amd64"),
			),
			C.String("platform", "go1.9.2"),
		)
		if appName != "" {
			doc.Append(C.SubDocumentFromElements("application", C.String("name", appName)))
		}

		return doc
	}
	buf, err := f("hello-world").MarshalBSON()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(buf)

	// Output: [177 0 0 0 3 100 114 105 118 101 114 0 52 0 0 0 2 110 97 109 101 0 16 0 0 0 109 111 110 103 111 45 103 111 45 100 114 105 118 101 114 0 2 118 101 114 115 105 111 110 0 8 0 0 0 49 50 51 52 53 54 55 0 0 3 111 115 0 46 0 0 0 2 116 121 112 101 0 7 0 0 0 100 97 114 119 105 110 0 2 97 114 99 104 105 116 101 99 116 117 114 101 0 6 0 0 0 97 109 100 54 52 0 0 2 112 108 97 116 102 111 114 109 0 8 0 0 0 103 111 49 46 57 46 50 0 3 97 112 112 108 105 99 97 116 105 111 110 0 27 0 0 0 2 110 97 109 101 0 12 0 0 0 104 101 108 108 111 45 119 111 114 108 100 0 0 0]
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
				C.String("type", "darwin"),
				C.String("architecture", "amd64"),
			),
			C.String("platform", "go1.9.2"),
		)
		doc.MarshalBSON()
	}
}

func elementEqual(e1, e2 *Element) bool {
	if e1.start != e2.start {
		return false
	}
	if e1.value != e2.value {
		return false
	}
	return true
}
