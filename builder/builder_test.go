package builder

import (
	"bytes"
	"runtime"
	"testing"

	"github.com/skriptble/wilson/parser/ast"
)

func TestDocumentBuilder(t *testing.T) {
	t.Run("Basic-Construction", func(t *testing.T) {
		b := make([]byte, 41)
		d := new(DocumentBuilder).Append(C.Double("foo", 3.14159), C.SubDocumentWithElements("bar", C.Double("baz", 3.14159)))
		_, _ = d.WriteDocument(b)
	})

	t.Run("Static-Functions", func(t *testing.T) {
		t.Run("Double", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				f       float64
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 3.14159, 13,
					[]byte{
						// type
						0x1,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0x6e, 0x86, 0x1b, 0xf0, 0xf9,
						0x21, 0x9, 0x40},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Double(tc.key, tc.f)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("String", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				s       string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "bar", 13,
					[]byte{
						// type
						0x2,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x72, 0x0,
					},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).String(tc.key, tc.s)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("SubDocument", func(t *testing.T) {
			var b DocumentBuilder
			b.Init()
			b.Append(C.String("bar", "baz"))

			testCases := []struct {
				name    string
				key     string
				subdoc  *DocumentBuilder
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo",
					&b,
					23,
					[]byte{
						// type
						0x3,
						// key
						0x66, 0x6f, 0x6f, 0x0,

						// length
						0x12, 0x0, 0x0, 0x0,
						// type
						0x2,
						// key
						0x62, 0x61, 0x72, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x7a, 0x0,

						// null terminator
						0x0,
					},
					0, 23, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).SubDocument(tc.key, tc.subdoc).Element()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("SubDocumentWithElements", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				subdoc  []Elementer
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo",
					[]Elementer{C.String("bar", "baz")},
					23,
					[]byte{
						// type
						0x3,
						// key
						0x66, 0x6f, 0x6f, 0x0,

						// length
						0x12, 0x0, 0x0, 0x0,
						// type
						0x2,
						// key
						0x62, 0x61, 0x72, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x7a, 0x0,

						// null terminator
						0x0,
					},
					0, 23, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).SubDocumentWithElements(tc.key, tc.subdoc...).Element()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Array", func(t *testing.T) {
			var b ArrayBuilder
			b.Init()
			b.Append(AC.String("baz"))

			testCases := []struct {
				name    string
				key     string
				array   *ArrayBuilder
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo",
					&b,
					21,
					[]byte{
						// type
						0x4,
						// key
						0x66, 0x6f, 0x6f, 0x0,

						// length
						0x10, 0x0, 0x0, 0x0,
						// type
						0x2,
						// key
						0x30, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x7a, 0x0,

						// null terminator
						0x0,
					},
					0, 21, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Array(tc.key, tc.array)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Array", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				array   []ArrayElementer
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo",
					[]ArrayElementer{AC.String("baz")},
					21,
					[]byte{
						// type
						0x4,
						// key
						0x66, 0x6f, 0x6f, 0x0,

						// length
						0x10, 0x0, 0x0, 0x0,
						// type
						0x2,
						// key
						0x30, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x7a, 0x0,

						// null terminator
						0x0,
					},
					0, 21, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).ArrayWithElements(tc.key, tc.array...)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Binary", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				b       []byte
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", []byte{8, 6, 7, 5, 3, 0, 9}, 17,
					[]byte{
						// type
						0x5,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - binary length
						0x7, 0x0, 0x0, 0x0,
						// value - binary subtype
						0x0,
						// value - binary data
						0x8, 0x6, 0x7, 0x5, 0x3, 0x0, 0x9,
					},
					0, 17, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Binary(tc.key, tc.b)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("BinaryWithSubtype", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				b       []byte
				btype   byte
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", []byte{8, 6, 7, 5, 3, 0, 9}, 0x2, 21,
					[]byte{
						// type
						0x5,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - binary length
						0xb, 0x0, 0x0, 0x0,
						// value - binary subtype
						0x2,
						//
						0x07, 0x00, 0x00, 0x00,
						// value - binary data
						0x8, 0x6, 0x7, 0x5, 0x3, 0x0, 0x9,
					},
					0, 21, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).BinaryWithSubtype(tc.key, tc.b, tc.btype)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Undefined", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 5,
					[]byte{
						// type
						0x6,
						// key
						0x66, 0x6f, 0x6f, 0x0,
					},
					0, 5, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Undefined(tc.key)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("ObjectId", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				oid     [12]byte
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", [12]byte{0x5a, 0x15, 0xd0, 0xa4, 0xd5, 0xda, 0xa5, 0xf1, 0x0a, 0x5e, 0x10, 0x89}, 17,
					[]byte{
						// type
						0x7,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0x5a, 0x15, 0xd0, 0xa4, 0xd5, 0xda, 0xa5, 0xf1, 0x0a, 0x5e, 0x10, 0x89,
					},
					0, 17, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).ObjectId(tc.key, tc.oid)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Boolean", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				b       bool
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", false, 6,
					[]byte{
						// type
						0x8,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0x0,
					},
					0, 6, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Boolean(tc.key, tc.b)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("DateTime", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				t       int64
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 17, 13,
					[]byte{
						// type
						0x9,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0x11, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
					},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).DateTime(tc.key, tc.t)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Null", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 5,
					[]byte{
						// type
						0xa,
						// key
						0x66, 0x6f, 0x6f, 0x0,
					},
					0, 5, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Null(tc.key)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Regex", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				pattern string
				options string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "bar", "i", 11,
					[]byte{
						// type
						0xb,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - pattern
						0x62, 0x61, 0x72, 0x0,
						// value - options
						0x69, 0x0,
					},
					0, 11, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Regex(tc.key, tc.pattern, tc.options)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("DBPointer", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				ns      string
				oid     [12]byte
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "bar", [12]byte{0x5a, 0x15, 0xd0, 0xa4, 0xd5, 0xda, 0xa5, 0xf1, 0x0a, 0x5e, 0x10, 0x89}, 25,
					[]byte{
						// type
						0xc,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - namespace length
						0x4, 0x0, 0x0, 0x0,
						// value - namespace
						0x62, 0x61, 0x72, 0x0,
						// value - oid
						0x5a, 0x15, 0xd0, 0xa4, 0xd5, 0xda, 0xa5, 0xf1, 0x0a, 0x5e, 0x10, 0x89,
					},
					0, 25, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).DBPointer(tc.key, tc.ns, tc.oid)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("JavaScriptCode", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				code    string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "var bar = 3;", 22,
					[]byte{
						// type
						0xd,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - code length
						0xd, 0x0, 0x0, 0x0,
						// value - code
						0x76, 0x61, 0x72, 0x20, 0x62, 0x61, 0x72, 0x20, 0x3d, 0x20, 0x33, 0x3b, 0x0,
					},
					0, 22, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).JavaScriptCode(tc.key, tc.code)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Symbol", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				s       string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "bar", 13,
					[]byte{
						// type
						0xe,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - string length
						0x4, 0x0, 0x0, 0x0,
						// value - string
						0x62, 0x61, 0x72, 0x0,
					},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Symbol(tc.key, tc.s)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("CodeWithScope", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				code    string
				scope   []byte
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", "var bar = x;",
					[]byte{
						// length
						0x8, 0x0, 0x0, 0x0,
						// type
						0xa,
						// key
						0x78, 0x0,
						// terminal
						0x0,
					},
					34,
					[]byte{
						// type
						0xf,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value - code length
						0x1d, 0x0, 0x0, 0x0,
						// vale -length
						0xd, 0x0, 0x0, 0x0,
						// value - code
						0x76, 0x61, 0x72, 0x20, 0x62, 0x61, 0x72, 0x20, 0x3d, 0x20, 0x78, 0x3b, 0x0,
						// value - scope length
						0x8, 0x0, 0x0, 0x0,
						// value - scope element type
						0xa,
						// value - scope element key
						0x78, 0x0,
						// value - scope terminal
						0x0,
					},
					0, 34, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).CodeWithScope(tc.key, tc.code, tc.scope)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Int32", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				i       int32
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", -27, 9,
					[]byte{
						// type
						0x10,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0xe5, 0xff, 0xff, 0xff,
					},
					0, 9, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Int32(tc.key, tc.i)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Timestamp", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				t       uint32
				i       uint32
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 8, 17, 13,
					[]byte{
						// type
						0x11,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0x11, 0x0, 0x0, 0x0, 0x8, 0x0, 0x0, 0x0,
					},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Timestamp(tc.key, tc.t, tc.i)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Int64", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				i       int64
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", -27, 13,
					[]byte{
						// type
						0x12,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0xe5, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
					},
					0, 13, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Int64(tc.key, tc.i)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("Decimal128", func(t *testing.T) {
			d, _ := ast.ParseDecimal128("-7.50")

			testCases := []struct {
				name    string
				key     string
				d       ast.Decimal128
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", d, 21,
					[]byte{
						// type
						0x13,
						// key
						0x66, 0x6f, 0x6f, 0x0,
						// value
						0xee, 0x02, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
						0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x3c, 0xb0,
					},
					0, 21, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).Decimal(tc.key, tc.d)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("MinKey", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 5,
					[]byte{
						// type
						0xff,
						// key
						0x66, 0x6f, 0x6f, 0x0,
					},
					0, 5, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).MinKey(tc.key)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})

		t.Run("MaxKey", func(t *testing.T) {
			testCases := []struct {
				name    string
				key     string
				size    uint
				repr    []byte
				start   uint
				written int
				err     error
			}{
				{"success", "foo", 5,
					[]byte{
						// type
						0x7f,
						// key
						0x66, 0x6f, 0x6f, 0x0,
					},
					0, 5, nil},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					sizer, f := (Constructor{}).MaxKey(tc.key)()
					if sizer() != tc.size {
						t.Errorf("Element sizes do not match. got %d; want %d", sizer(), tc.size)
					}
					t.Run("[]byte", func(t *testing.T) {
						b := make([]byte, sizer())
						written, err := f(tc.start, b)
						if written != tc.written {
							t.Errorf("Number of bytes written incorrect. got %d; want %d", written, tc.written)
						}
						if err != tc.err {
							t.Errorf("Returned error not expected error. got %s; want %s", err, tc.err)
						}
						if !bytes.Equal(b, tc.repr) {
							t.Errorf("Written bytes do not match. got %#v; want %#v", b, tc.repr)
						}
					})
					t.Run("io.WriterAt", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.WriteSeeker", func(t *testing.T) {
						t.Skip("not implemented")
					})
					t.Run("io.Writer", func(t *testing.T) {
						t.Skip("not implemented")
					})
				})
			}
		})
	})
}

func ExampleDocumentBuilder_ClientDoc() {
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

	f := func(appName string) *DocumentBuilder {
		docbuilder := new(DocumentBuilder)
		docbuilder.Init()
		docbuilder.Append(
			C.SubDocumentWithElements("driver",
				C.String("name", "mongo-go-driver"),
				C.String("version", internalVersion),
			),
			C.SubDocumentWithElements("os",
				C.String("type", runtime.GOOS),
				C.String("architecture", runtime.GOARCH),
			),
			C.String("platform", runtime.Version()),
		)
		if appName != "" {
			docbuilder.Append(C.SubDocumentWithElements("application", C.String("name", appName)))
		}

		return docbuilder
	}
	f("hello-world")
}
